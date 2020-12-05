package main

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/samuel/go-zookeeper/zk"
)

var my_name string = "Undecided."

// Function to connect with Zookeeper
func connect() *zk.Conn {
	conn, _, err := zk.Connect([]string{"zookeeper:2181"}, time.Second)
	if err != nil {
		panic(err)
	}
	return conn
}

func library(w http.ResponseWriter, req *http.Request) {

	fmt.Fprintf(w, "Proudly served by %s\n", my_name)
}

func main() {
	// Get environment gsname. Defined in docker-compose in our case.
	my_name = os.Getenv("gsname")

	// Connect to Zookeeper and hold any function till session is built.
	conn := connect()
	defer conn.Close()

	for conn.State() != zk.StateHasSession {
		time.Sleep(5)
	}
	if conn.State() == zk.StateHasSession {
		fmt.Printf(" %s has successfully connected to Zookeeper\n", my_name)
	}

	http.HandleFunc("/library", library)

	http.ListenAndServe(":80", nil)
}
