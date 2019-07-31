package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"
)

func main() {
	//get Configuration
	address := flag.String("server", "http://localhost:8080", "HTTP gateway url, e.g. http://localhost:8080")
	flag.Parse()

	t := time.Now().In(time.UTC)
	pfx := t.Format(time.RFC3339Nano)

	var body string

	//Call Create
	resp, err := http.Post(*address+"/v1/tasq", "application/json", strings.NewReader(fmt.Sprintf(`
	{
		"api":"v1",
		"toDo":{
			"title":"title (%s)",
			"description":"description (%s)",
			"reminder":"%s"
		}
	}
	`, pfx, pfx, pfx)))
	if err != nil {
		log.Fatal("failed to call Create method: %v", err)
	}
	bodyBytes, err := ioutil.ReadAll(resp.Body)
	resp.Body.Close()

	if err != nil {
		body = fmt.Sprintf("failed to read Create response body: %v", err)
	} else {
		body = string(bodyBytes)
	}
	log.Printf("Create response: Code=%d, Body=%s\n\n", resp.StatusCode, body)

	//Parse ID of created ToDo
	var created struct {
		API string `json:"api"`
		ID  string `json:"id"`
	}
	err = json.Unmarshal(bodyBytes, &created)
	if err != nil {
		log.Fatalf("failed to unmarshal JSON response of Create method: %v", err)
	}

	//Call Read
	resp, err = http.Get(fmt.Sprintf("%s%s/%s", *address, "/v1/tasq", created.ID))
	if err != nil {
		log.Fatalf("failed to call Read method: %v", err)
	}
	bodyBytes, err = ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		body = fmt.Sprintf("failed to read Read response body: %v", err)
	} else {
		body = string(bodyBytes)
	}
	log.Printf("Read response: Code=%d, Body=%s\n\n", resp.StatusCode, body)
}
