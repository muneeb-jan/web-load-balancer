package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
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

// Encode the form data
func encoder(unencodedJSON []byte) string {
	// get go object from json byte
	var unencodedRows RowsType
	json.Unmarshal(unencodedJSON, &unencodedRows)

	//  encode all fields value of go object , return EncRowsType
	encodedRows := unencodedRows.encode()

	// convert to json byte[] from go object (EncRowsType)
	encodedJSON, _ := json.Marshal(encodedRows)

	return string(encodedJSON)
}

// Decode response from HBASE
func decoder(encodedJSON []byte) string {

	// get go object from json byte
	var encodedRows EncRowsType
	fmt.Println("From decoder test print: ", string(encodedJSON))
	json.Unmarshal(encodedJSON, &encodedRows)
	fmt.Println("From decoder first: ", encodedRows)

	//  decode all fields value of go object , return RowsType
	decodedRows, err := encodedRows.decode()
	if err != nil {
		fmt.Println("%+v", err)
	}
	fmt.Println("From decoder second: ", decodedRows)
	// convert to json byte[] from go object (RowsType)
	deCodedJSON, _ := json.Marshal(decodedRows)

	//fmt.Println("From decoder method: ", string(deCodedJSON))
	return string(deCodedJSON)
}

func getValue() string {

	hbase_url := "http://hbase:8080/se2:library/fakerow/*"
	//my_request, _ := http.NewRequest("GET", hbase_url, nil)
	//my_client := &http.Client{}
	//hbase_response, err := my_client.Do(my_request)
	hbase_response, err := http.Get(hbase_url)
	if err != nil {
		fmt.Printf("Error during getValue method: %v\n", err)
	}

	encoded_response, ioerr := ioutil.ReadAll(hbase_response.Body)
	if ioerr != nil {
		fmt.Printf("Error during IO util read of hbase response. %v", ioerr)
	}

	final_response := decoder(encoded_response)
	return final_response
}

// Function to post the application body into HBASE
func postValue(in_req string) {
	hbase_url := "http://hbase:8080/se2:library/fakerow"
	post_response, err := http.Post(hbase_url, "application/json", bytes.NewBuffer([]byte(in_req)))
	//post_response, err := http.NewRequest("PUT", hbase_url, )
	if err != nil {
		fmt.Printf("Error from postValue function: %v\n", err)
	}

	fmt.Printf("Data has been posted: %s", post_response.Status)
}

func library(w http.ResponseWriter, req *http.Request) {

	if req.Method == "GET" {
		// If Request is GET, send to getValue to retrieve response.
		response_value := getValue()
		fmt.Fprintf(w, "The Response to your request is below:\n\n %s\n", response_value)

	} else if req.Method == "POST" || req.Method == "PUT" {

		//If request for POST, read the body, and send to postValue function.
		incoming_req, err := ioutil.ReadAll(req.Body)
		if err != nil {
			fmt.Printf("Error while reading new incoming request: %v\n", err)
		}

		encoded_data := encoder(incoming_req)
		//req.Header.Set("Content-type", "application/json")

		postValue(string(encoded_data))

	} else {

		//In all OTHER cases, no action
		fmt.Fprint(w, "This HTTP method is not served by the application. Try GET, PUT or POST.")
	}

	fmt.Fprintf(w, "\nProudly served by %s\n", my_name)
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
