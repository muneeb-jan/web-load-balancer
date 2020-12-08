package main

import (
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"time"

	"github.com/samuel/go-zookeeper/zk"
)

// Function to connect with Zookeeper
func connect() *zk.Conn {
	conn, _, err := zk.Connect([]string{"zookeeper:2181"}, time.Second)
	if err != nil {
		panic(err)
	}
	return conn
}

// NewMultipleHostReverseProxy creates a reverse proxy that will randomly select a host from the passed param `targets` if the url request contains /library
// Targets are explicitly defined in main.
func NewMultipleHostReverseProxy(targets []*url.URL) *httputil.ReverseProxy {
	director := func(req *http.Request) {
		if strings.Contains(req.URL.Path, "library") {
			fmt.Println("This should redirect to gserve servers after catching library in round robin.")
			target := targets[rand.Int()%len(targets)]
			req.URL.Scheme = target.Scheme
			req.URL.Host = target.Host
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

	grproxy_flag := int32(0)
	grproxy_ac := zk.WorldACL(zk.PermAll)
	grproxy_seq, err := conn.Create("/grproxy", []byte("grproxy"), grproxy_flag, grproxy_ac)
	if err != nil {
		fmt.Printf("Error while creating ZNODE: %v\n", err)
	}
	fmt.Println("Znode created: ", grproxy_seq)

	proxy := NewMultipleHostReverseProxy([]*url.URL{
		{
			Scheme: "http",
			Host:   "gserve1",
		},
		{
			Scheme: "http",
			Host:   "gserve2",
		},
	})
	log.Fatal(http.ListenAndServe(":8080", proxy))
}
