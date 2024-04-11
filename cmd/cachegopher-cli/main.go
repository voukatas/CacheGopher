package main

import (
	"fmt"
	"log"

	"github.com/voukatas/CacheGopher/pkg/client"
)

func main() {
	newPool := client.NewConnPool(3, "localhost:31337")
	newClient, err := client.NewClient(newPool)

	if err != nil {

		log.Fatalf("failed to establish connection %s", err)
	}

	//defer newClient.Close()

	resp, err := newClient.Set("testKey", "testValue\\n1111")
	if err != nil {

		fmt.Printf("failed to SET key, error: %s", err)

	} else {
		fmt.Printf("Response from cache server: %s\n", resp)
	}

	resp, err = newClient.Get("testKey")
	if err != nil {

		fmt.Printf("failed to GET key, error: %s\n", err)
	} else {

		fmt.Printf("Response from cache server: %s\n", resp)
	}

	/*
	   // tests
	   resp, err = newClient.Set("testKey1", "testValue1\\n1111")
	   if err != nil {

	   	fmt.Printf("failed to SET key, error: %s", err)

	   	} else {
	   		fmt.Printf("Response from cache server: %s\n", resp)
	   	}

	   resp, err = newClient.Get("testKey1")
	   if err != nil {

	   		fmt.Printf("failed to GET key, error: %s", err)
	   	} else {

	   		fmt.Printf("Response from cache server: %s\n", resp)
	   	}

	   resp, err = newClient.Set("testKey2", "testValue2\\n1111")
	   if err != nil {

	   	fmt.Printf("failed to SET key, error: %s", err)

	   	} else {
	   		fmt.Printf("Response from cache server: %s\n", resp)
	   	}

	   resp, err = newClient.Get("testKey2")
	   if err != nil {

	   		fmt.Printf("failed to GET key, error: %s", err)
	   	} else {

	   		fmt.Printf("Response from cache server: %s\n", resp)
	   	}
	*/
}
