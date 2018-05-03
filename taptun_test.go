package taptun

import (
	"fmt"
	"log"
	"net"
	"taptun"
	"testing"
	"time"

	"github.com/songgao/water/waterutil"
)

const BUFFERSIZE = 1522

func startRead(ch chan<- []byte, ifce *Interface) {
	go func() {
		for {
			buffer := make([]byte, BUFFERSIZE)
			n, err := ifce.Read(buffer)
			if err == nil {
				buffer = buffer[:n:n]
				ch <- buffer
			}
		}
	}()
}

func TestTAP(t *testing.T) {
	var (
		self = net.IPv4(10, 0, 42, 1)
		mask = net.IPv4Mask(255, 255, 255, 0)
		brd  = net.IPv4(10, 0, 42, 255)
	)
	fmt.Println("Start to test TAP!")
	ifce, err := New(Config{DeviceType: TAP})
	if err != nil {
		t.Fatalf("creating TAP error: %v\n", err)
	}

	setupIfce(t, net.IPNet{IP: self, Mask: mask}, ifce.Name())
	startBroadcast(t, brd)

	dataCh := make(chan []byte, 8)
	startRead(dataCh, ifce)

	timeout := time.NewTimer(5 * time.Second).C

readFrame:
	for {
		select {
		case buffer := <-dataCh:
			ethertype := waterutil.MACEthertype(buffer)
			if ethertype != waterutil.IPv4 {
				continue readFrame
			}

			if !waterutil.IsBroadcast(waterutil.MACDestination(buffer)) {
				continue readFrame
			}
			packet := waterutil.MACPayload(buffer)
			if !waterutil.IsIPv4(packet) {
				continue readFrame
			}
			if !waterutil.IPv4Source(packet).Equal(self) {
				continue readFrame
			}
			if !waterutil.IPv4Destination(packet).Equal(brd) {
				continue readFrame
			}
			if waterutil.IPv4Protocol(packet) != waterutil.ICMP {
				continue readFrame
			}
			// t.Logf("received broadcast frame: %#v\n", buffer)
			fmt.Println("received broadcast frame: ", buffer)
			break readFrame
		case <-timeout:
			t.Fatal("Waiting for broadcast packet timeout")
		}
	}
}

func TestTUN(t *testing.T) {
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
