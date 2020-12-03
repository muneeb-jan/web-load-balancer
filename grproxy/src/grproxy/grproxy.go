package main

import (
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"net/http/httputil"
	"net/url"
)

// NewMultipleHostReverseProxy creates a reverse proxy that will randomly select a host from the passed param `targets` if the url request contains /library
// Targets are explicitly defined in main.
func NewMultipleHostReverseProxy(targets []*url.URL) *httputil.ReverseProxy {
	director := func(req *http.Request) {
		if req.URL.Path == "/library" {
			fmt.Println("This should redirect to gserve servers after catching library in round robin.")
			target := targets[rand.Int()%len(targets)]
			req.URL.Scheme = target.Scheme
			req.URL.Host = target.Host
			req.URL.Path = target.Path
		} else {
			fmt.Println("This should redirect to nginx for all other cases.")
			req.URL.Scheme = "http"
			req.URL.Host = "nginx"
		}
	}
	return &httputil.ReverseProxy{Director: director}
}

func main() {
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
