package main

import (
	"fmt"
	"net/http"
	"os"
)

var my_name string = "Undecided."

func library(w http.ResponseWriter, req *http.Request) {

	fmt.Fprintf(w, "Proudly served by %s\n", my_name)
}

func main() {

	my_name = os.Getenv("gsname")

	http.HandleFunc("/library", library)

	http.ListenAndServe(":80", nil)
}
