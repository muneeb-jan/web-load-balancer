package main

import (
    "fmt"
    "net/http"
)

func library(w http.ResponseWriter, req *http.Request) {

    fmt.Fprintf(w, "Welcome to Gserve!. Have a nice day.\n")
}

func main() {

    http.HandleFunc("/library", library)

    http.ListenAndServe(":80", nil)
}
