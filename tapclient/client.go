package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"taptun"
)

const BUFFERSIZE = 1522

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintf(os.Stderr, "Usage: %s host:port ", os.Args[0])
		os.Exit(1)
	}
	service := os.Args[1]
	tcpAddr, err := net.ResolveTCPAddr("tcp4", service)
	checkError(err)
	conn, err := net.DialTCP("tcp", nil, tcpAddr)
	checkError(err)
	// _, err = conn.Write([]byte("HEAD / HTTP/1.0\r\n\r\n"))
	checkError(err)
	if err == nil {
		ifce := prepare()
		exchange(*ifce, conn)
	}

	// buffer := make([]byte, 1522)
	// for {
	//
	// 	n, err := conn.Read(buffer)
	// 	checkError(err)
	// 	fmt.Println(buffer[:n])
	// }
	// result, err := ioutil.ReadAll(conn)

	// fmt.Println(string(result))
	os.Exit(0)
}
func checkError(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "Fatal error: %s", err)
		os.Exit(1)
	}
}

func prepare() (iface *taptun.Interface) {
	config := taptun.Config{
		DeviceType: taptun.TAP,
	}

	var (
		self = net.IPv4(192, 168, 1, 86)
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
			fmt.Println("local", buffer)
			conn.Write(buffer)
		case buffer := <-connchan:
			fmt.Println("peer", buffer)
			iface.Write(buffer)
			// case <-timeout:
			//   t.Fatal("Waiting for broadcast packet timeout")
		}
	}
}
