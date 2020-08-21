package main

import (
	"io"
	"io/ioutil"
	"net/http"
	"os"
)

func handleIndex(w http.ResponseWriter, r *http.Request) {
	// Read token and send it as a method to identify self
	// to service 2
	b, err := ioutil.ReadFile("/var/run/secrets/kubernetes.io/serviceaccount/token")
	if err != nil {
		panic(err)
	}
	saToken := string(b)

	// Make a HTTP request to service2
	serviceConnstring := os.Getenv("SERVICE2_CONNSTRING")
	if len(serviceConnstring) == 0 {
		panic("SERVICE2_CONNSTRING expected")
	}
	client := &http.Client{}
	req, err := http.NewRequest("GET", serviceConnstring, nil)
	if err != nil {
		panic(err)
	}
	req.Header.Add("X-Client-Id", saToken)
	resp, err := client.Do(req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)
	io.WriteString(w, string(body))
}

func main() {

	http.HandleFunc("/", handleIndex)
	listenAddr := os.Getenv("LISTEN_ADDR")
	if len(listenAddr) == 0 {
		panic("LISTEN_ADDR expected")
	}
	http.ListenAndServe(listenAddr, nil)
}
