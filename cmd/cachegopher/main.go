package main

import (
	"fmt"
	"log"
	"net"

	"github.com/voukatas/CacheGopher/pkg/cache"
)

func main() {

	localCache := cache.NewCache()
	// numOfKeys := cache.Set("mykey", "1")
	// fmt.Println("numOfKeys: ", numOfKeys)
	// cache.Set("mykey2", "2")
	// item1, exists := cache.Get("mykey2")
	// if exists {
	// 	fmt.Println("item: ", item1)
	// }

	listener, err := net.Listen("tcp", "localhost:31337")
	if err != nil {
		log.Fatal(err)
	}

	defer listener.Close()

	fmt.Println("Server is running on localhost:31337")

	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("Error accepting connection: ", err)
			continue
		}
		go cache.HandleConnection(conn, localCache)
	}

}
