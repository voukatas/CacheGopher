package main

import (
	"fmt"
	"log"

	"github.com/voukatas/CacheGopher/pkg/client"
)

func main() {
	newClient, err := client.NewClient("localhost:31337")

	if err != nil {

		log.Fatalf("failed to establish connection %s", err)
	}

	defer newClient.Close()

	resp, err := newClient.Set("testKey", "testValue")
	if err != nil {

		fmt.Printf("failed to SET key, error: %s", err)

	} else {
		fmt.Printf("Response from cache server: %s", resp)
	}

	resp, err = newClient.Get("testKey")
	if err != nil {

		fmt.Printf("failed to GET key, error: %s", err)
	} else {

		fmt.Printf("Response from cache server: %s", resp)
	}

}
