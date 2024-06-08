package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"strings"
)

func main() {

	ip := flag.String("host", "", "Set the ip of the host")
	port := flag.String("port", "", "Set the port of the host")

	flag.Parse()

	address := fmt.Sprintf("%s:%s", *ip, *port)

	conn, err := net.Dial("tcp", address)
	if err != nil {
		log.Fatal("Failed to connect: ", err.Error())
	}
	defer conn.Close()

	scanner := bufio.NewScanner(os.Stdin)
	respScanner := bufio.NewScanner(conn)
	//fmt.Print(">> ")

	for {

		fmt.Print(">> ")
		scanner.Scan()
		//fmt.Println("You wrote: ", scanner.Text())

		if strings.Contains(strings.ToUpper(scanner.Text()), "EXIT") {
			fmt.Println("Bye!!")
			return
		}
		//fmt.Print(">> ")
		//fmt.Println("I will send: ", scanner.Text())

		_, err := conn.Write([]byte(scanner.Text() + "\n"))
		if err != nil {
			fmt.Println("Write error:", err)
			continue
		}

		//fmt.Println("Waiting for response...")

		if respScanner.Scan() {
			fmt.Println(respScanner.Text())
		}
		if err := respScanner.Err(); err != nil {
			fmt.Println(err.Error())
		}

	}
}
