package main

import (
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"net/http/httputil"
	"strings"
	"time"

	"github.com/samuel/go-zookeeper/zk"
)

var targets []string

// Function to connect with Zookeeper
func connect() *zk.Conn {
	conn, _, err := zk.Connect([]string{"zookeeper:2181"}, time.Second)
	if err != nil {
		panic(err)
	}
	return conn
}

// Function to keep monitoring the gserver nodes
func gserver_monitoring(conn *zk.Conn, child_path string, running_gserver chan []string) {

	for {
		children, _, _, _ := conn.ChildrenW(child_path)
		running_gserver <- children
		time.Sleep(time.Millisecond * 500)
	}
}

// NewMultipleHostReverseProxy creates a reverse proxy that will randomly select a host from targets if the url request contains library
// Targets are explicitly defined in main.
func NewMultipleHostReverseProxy() *httputil.ReverseProxy {
	director := func(req *http.Request) {
		if strings.Contains(req.URL.Path, "library") {
			fmt.Println("This should redirect to gserve servers after catching library in round robin.")
			target := targets[rand.Int()%len(targets)]
			req.URL.Scheme = "http"
			req.URL.Host = target
		} else {
			fmt.Println("This should redirect to nginx for all other cases.")
			req.URL.Scheme = "http"
			req.URL.Host = "nginx"
		}
	}
	return &httputil.ReverseProxy{Director: director}
}

func main() {

	//Create connection to zookeeper and holding function.
	conn := connect()
	defer conn.Close()

	for conn.State() != zk.StateHasSession {
		time.Sleep(5)
	}

	//create ZNode
	grproxy_flag := int32(0)
	grproxy_ac := zk.WorldACL(zk.PermAll)
	grproxy_seq, err := conn.Create("/grproxy", []byte("grproxy"), grproxy_flag, grproxy_ac)
	if err != nil {
		fmt.Printf("Error while creating ZNODE: %v\n", err)
	}
	fmt.Println("Znode created: ", grproxy_seq)

	running_gservers := make(chan []string)
	// Return the running instances of gservers
	go gserver_monitoring(conn, "/grproxy", running_gservers)

	go func() {
		for {
			children := <-running_gservers
			var temp []string
			for _, child := range children {
				gserver_urls, _, rerr := conn.Get("/grproxy/" + child)
				temp = append(temp, string(gserver_urls))
				if rerr != nil {
					fmt.Printf("Gserver node error in proxy routine: %+v\n", rerr)
				}
			}
			targets = temp
		}
	}()

	proxy := NewMultipleHostReverseProxy()
	log.Fatal(http.ListenAndServe(":8080", proxy))
}
