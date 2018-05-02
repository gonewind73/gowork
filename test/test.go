package main

import (
	// "encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	// "strings"

	// "fastvpn/common"
	"taptun"
	// "github.com/songgao/water"
)

const (
	BUFFERSIZE = 1500
	MTU        = "1500"
)

func checkFatalErr(err error, msg string) {
	if err != nil {
		log.Println(msg)
		log.Fatal(err)
	}
}

func runIP(args ...string) {
	cmd := exec.Command("/sbin/ip", args...)
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout

	err := cmd.Run()

	if err != nil {
		log.Fatal("Error runing /sbin/ip:", err)
	}
}

func main() {
	config := taptun.Config{
		DeviceType: taptun.TAP,
	}

	iface, err := taptun.New(config)
	checkFatalErr(err, "Unable to allocate TUN interface: ")
	log.Println("Interface allocated: ", iface.Name())

	runIP("link", "set", "dev", iface.Name(), "mtu", MTU)
	runIP("addr", "add", "192.168.1.85/24", "dev", iface.Name())
	runIP("link", "set", "dev", iface.Name(), "up")

	packet := make([]byte, BUFFERSIZE)

	for {
		// read packaet
		plen, err := iface.Read(packet)
		// fmt.Printf("readed %x %x %X %X\n", packet[12:14], packet[14], packet[23], packet[26:34])
		if err != nil {
			fmt.Println("erro read")
			fmt.Println(err)
		}

		ippacket := taptun.GetPacketFromFrame(packet[:plen])
		if ippacket[0]&0xf0 == 0x40 {
			dstIP := taptun.IPv4Destination(ippacket)
			srcIP := taptun.IPv4Source(ippacket)
			fmt.Println(srcIP, dstIP, taptun.IPv4Protocol(ippacket))
		} else {
			// fmt.Println("not ipv4 packet")
		}

	}
}
