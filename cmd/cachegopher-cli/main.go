package main

import (
	"fmt"
	"log"

	"github.com/voukatas/CacheGopher/pkg/client"
)

func main() {
	newPool := client.NewConnPool(5, "localhost:31337")
	newClient, err := client.NewClient(newPool)

	if err != nil {

		log.Fatalf("failed to establish connection %s", err)
	}

	//defer newClient.Close()

	resp, err := newClient.Set("testKey", "testValue\\n1111")
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
