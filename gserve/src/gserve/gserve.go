package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"time"
	"strings"

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
	// convert JSON to Go objects
	var unencodedRows RowsType
	json.Unmarshal(unencodedJSON, &unencodedRows)

	// encode fields in Go objects
	encodedRows := unencodedRows.encode()

	// convert encoded Go objects to JSON
	encodedJSON, _ := json.Marshal(encodedRows)

	return string(encodedJSON)
}

// Decode data from HBASE
func decoder(encodedJSON []byte) string {

	// convert JSON to Go objects
	var encodedRows EncRowsType
	json.Unmarshal(encodedJSON, &encodedRows)

	// decode fields in Go objects
	decodedRows, _ := encodedRows.decode()

	// convert encoded Go objects to JSON
	decodedJSON, _ := json.Marshal(decodedRows)

	return string(decodedJSON)
}


// GET EVERYTHING FROM HBASE
func getValue(my_path string) string {

	hbase_url := "http://hbase:8080/se2:library/scanner/"
	//hbase_url := "http://hbase:8080" + my_path
	
	//GET SCANNER URL
	my_request, _ := http.NewRequest("PUT", hbase_url, bytes.NewBuffer([]byte("<Scanner batch=\"10\"/>")))
	my_client := &http.Client{}
	my_request.Header.Add("Accept", "text/plain")
	my_request.Header.Add("Content-Type", "text/xml")
	scanner_response, sc_err := my_client.Do(my_request)
	if sc_err != nil {
		fmt.Printf("Error during getValue method: %v\n", sc_err)
	}
	var scanner_value string
	for key,value := range scanner_response.Header {
		if strings.Contains(key,"Location") {
			scanner_value = value[0]
		}
	}

	// GET THE MAIN TABLE DATA
	hbase_request, _ := http.NewRequest("GET", scanner_value, nil)
	hbase_request.Header.Add("Accept", "application/json")
	hbase_response, err := my_client.Do(hbase_request)
	if err != nil {
		fmt.Printf("Error during scanner get method: %v\n", err)
	}
	final_response, ioerr := ioutil.ReadAll(hbase_response.Body)
	if ioerr != nil {
		fmt.Printf("Error during IO util read of hbase response. %v", ioerr)
	}	
	fmt.Println("The response: ", string(final_response))
	decoded_response := decoder(final_response) // DECODE THE RECEIVED RESPNSE BODY

	// DELETE THE SCANNER OBJ
	del_req, _ := http.NewRequest("DELETE", scanner_value, nil)
	del_req.Header.Add("Accept","text/plain")
	del_response, d_err := my_client.Do(del_req)
	if d_err != nil {
		fmt.Printf("Error during Scanner delete operation: %v", d_err)
	}
	fmt.Printf("Scanner delete status: %s", del_response.Status)


	return decoded_response
}

// Function to post the application body into HBASE
func postValue(in_req string) {
	hbase_url := "http://hbase:8080/se2:library/fakerow"
	post_response, err := http.Post(hbase_url, "application/json", bytes.NewBuffer([]byte(in_req)))
	if err != nil {
		fmt.Printf("Error from postValue function: %v\n", err)
	}
	fmt.Println("Data has been posted: ", post_response.Status)
}


// MAIN HANDLER FUNCTION
func library(w http.ResponseWriter, req *http.Request) {

	if req.Method == "GET" {
		// If Request is GET, send to getValue to retrieve response.
		my_path := req.URL.Path
		response_value := getValue(my_path)
		fmt.Fprintf(w, "\nThe Response to your request is below:\n\n %s\n", response_value)

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

	gserv_flag := int32(zk.FlagEphemeral)
	gserv_ac := zk.WorldACL(zk.PermAll)
	gserve_eph, err := conn.Create("/grproxy/"+my_name, []byte(my_name), gserv_flag, gserv_ac)
	if err != nil {
		fmt.Printf("Error while creating the ephemeral node: %v\n", err)
	}
	fmt.Println("Z-Node created: ", gserve_eph)

	http.HandleFunc("/", library)
	http.ListenAndServe(":80", nil)
}
