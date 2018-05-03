package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"time"
)

const (
	BUFFERSIZE = 1522
	MTU        = "1500"
)

func main() {
	service := ":7777"
	tcpAddr, err := net.ResolveTCPAddr("tcp4", service)
	checkError(err)
	log.Println("listen on", tcpAddr)
	listener, err := net.ListenTCP("tcp", tcpAddr)
	checkError(err)
	for {
		conn, err := listener.Accept()
		if err != nil {
			continue
		}
		fmt.Println(conn.RemoteAddr(), conn.LocalAddr())
		buffer := make([]byte, BUFFERSIZE)
		_, err = conn.Read(buffer)
		checkError(err)
		fmt.Println(string(buffer))
		daytime := time.Now().String()
		conn.Write([]byte(daytime)) // don't care about return value
		time.Sleep(1)
		conn.Close() // we're finished with this client
	}
	fmt.Println(err)
}

func checkError(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "Fatal error: %s", err)
		os.Exit(1)
	}
}
