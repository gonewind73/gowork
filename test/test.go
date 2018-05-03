package main

import (
	"fmt"
	"log"
	"net"
	"taptun"
)

const (
	BUFFERSIZE = 1500
	MTU        = "1500"
)

func main() {
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

	frame := make([]byte, BUFFERSIZE)

	for {
		// read packaet
		flen, err := iface.Read(frame)
		// fmt.Printf("readed %x %x %X %X\n", packet[12:14], packet[14], packet[23], packet[26:34])
		if err != nil {
			log.Println(err, "error in read. ")
		}

		ippacket := taptun.GetPacketFromFrame(frame[:flen])
		if ippacket[0]&0xf0 == 0x40 {
			dstIP := taptun.IPv4Destination(ippacket)
			srcIP := taptun.IPv4Source(ippacket)
			fmt.Println(srcIP, dstIP, taptun.IPv4Protocol(ippacket))
		} else {
			// log.Println("not ipv4 packet")
		}

	}
}
