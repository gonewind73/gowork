package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"taptun"
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
	ifce := prepare()
	for {
		conn, err := listener.Accept()
		if err != nil {
			continue
		}
		fmt.Println(conn.RemoteAddr(), conn.LocalAddr())
		// buffer := make([]byte, BUFFERSIZE)
		// blen, err := conn.Read(buffer)
		// checkError(err)
		// fmt.Println(string(buffer[:blen]))
		// daytime := time.Now().String()
		// conn.Write([]byte(daytime)) // don't care about return value
		go exchange(*ifce, conn)
		// time.Sleep(1)
		// conn.Close() // we're finished with this client
	}
	fmt.Println(err)
}

func checkError(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "Fatal error: %s", err)
		os.Exit(1)
	}
}

func session() {

}

func prepare() (iface *taptun.Interface) {
	config := taptun.Config{
		DeviceType: taptun.TAP,
	}

	var (
		self = net.IPv4(192, 168, 1, 85)
		mask = net.IPv4Mask(255, 255, 255, 0)
		// brd  = net.IPv4(10, 0, 42, 255)
	)

	iface, err := taptun.New(config)
	if err != nil {
		log.Println(err, "Unable to allocate TUN interface: ")
		return
	}
	log.Println("Interface allocated: ", iface.Name())

	taptun.Start(net.IPNet{IP: self, Mask: mask}, iface.Name())
	return

}

func conn2chan(conn net.Conn, ch chan<- []byte) {
	// ch = make(chan []byte, 8)
	go func() {
		for {
			buffer := make([]byte, BUFFERSIZE)
			n, err := conn.Read(buffer)
			if err == nil {
				buffer = buffer[:n]
				ch <- buffer
			}
		}
	}()
	return
}

func exchange(iface taptun.Interface, conn net.Conn) {
	ifchan := make(chan []byte, 8)
	iface.ToChan(ifchan)
	connchan := make(chan []byte, 8)
	conn2chan(conn, connchan)
	for {
		select {
		case buffer := <-ifchan:
			fmt.Println(buffer)
			conn.Write(buffer)
		case buffer := <-connchan:
			fmt.Println(buffer)
			iface.Write(buffer)
			// case <-timeout:
			//   t.Fatal("Waiting for broadcast packet timeout")
		}
	}
}
