package main

import (
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
)

var serviceToken string

func readToken() {
	b, err := ioutil.ReadFile("/var/run/secrets/kubernetes.io/serviceaccount/token")
	if err != nil {
		panic(err)
	}
	serviceToken = string(b)
}

func handleIndex(w http.ResponseWriter, r *http.Request) {

	// Make a HTTP request to service2
	serviceConnstring := os.Getenv("SECRET_STORE_CONNSTRING")
	if len(serviceConnstring) == 0 {
		panic("SERVICE2_CONNSTRING expected")
	}
	client := &http.Client{}
	req, err := http.NewRequest("GET", serviceConnstring, nil)
	if err != nil {
		panic(err)
	}
	// Identity self to service 2 using service account token
	req.Header.Add("X-Client-Id", serviceToken)
	resp, err := client.Do(req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	} else {
		defer resp.Body.Close()
		body, _ := ioutil.ReadAll(resp.Body)
		io.WriteString(w, string(body))
	}
}

func main() {
	// Read the token at startup
	readToken()
	http.HandleFunc("/", handleIndex)
	listenAddr := os.Getenv("LISTEN_ADDR")
	if len(listenAddr) == 0 {
		panic("LISTEN_ADDR expected")
	}
	http.ListenAndServe(listenAddr, nil)
}
