package main

import (
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"
)

var serviceToken string

func readToken() {
	b, err := ioutil.ReadFile("/var/run/secrets/tokens/api-token")
	if err != nil {
		panic(err)
	}
	serviceToken = string(b)
	log.Print("Refreshing service account token")
}

func handleIndex(w http.ResponseWriter, r *http.Request) {

	// Make a HTTP request to service2
	serviceConnstring := os.Getenv("DATA_STORE_CONNSTRING")
	if len(serviceConnstring) == 0 {
		panic("DATA_STORE_CONNSTRING expected")
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
	// Read the token once at startup first
	readToken()
	// Reload the service account token every 5 minutes
	ticker := time.NewTicker(300 * time.Second)
	done := make(chan bool)
	go func() {
		for {
			select {
			case <-done:
				return
			case <-ticker.C:
				readToken()
			}
		}
	}()

	http.HandleFunc("/", handleIndex)
	listenAddr := os.Getenv("LISTEN_ADDR")
	if len(listenAddr) == 0 {
		panic("LISTEN_ADDR expected")
	}
	http.ListenAndServe(listenAddr, nil)

	// Ideally, we would have a shutdown function to orchestrate the shutdown
	// of the server and stop the ticker
	ticker.Stop()
	done <- true
}
