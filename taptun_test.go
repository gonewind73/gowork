package taptun

import (
	"fmt"
	"net"
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
	// startBroadcast(t, brd)

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
	var (
		self = net.IPv4(10, 0, 43, 1)
		mask = net.IPv4Mask(255, 255, 255, 0)
		// brd  = net.IPv4(10, 0, 42, 255)
	)
	fmt.Println("Start to test TUN!")
	ifce, err := New(Config{DeviceType: TAP})
	if err != nil {
		t.Fatalf("creating TUN error: %v\n", err)
	}

	setupIfce(t, net.IPNet{IP: self, Mask: mask}, ifce.Name())
	startPing(t, self)

	dataCh := make(chan []byte, 8)
	startRead(dataCh, ifce)

	timeout := time.NewTimer(20 * time.Second).C

readFrame:
	for {
		select {
		case buffer := <-dataCh:
			ethertype := waterutil.MACEthertype(buffer)
			if ethertype != waterutil.IPv4 {
				continue readFrame
			}

			// if !waterutil.IsBroadcast(waterutil.MACDestination(buffer)) {
			// 	continue readFrame
			// }
			packet := waterutil.MACPayload(buffer)
			// case packet := <-dataCh:
			// fmt.Println("received packet: ", packet)
			// fmt.Println("received : ", packet[0]>>4)

			if !waterutil.IsIPv4(packet) {
				continue readFrame
			}
			// if !waterutil.IPv4Source(packet).Equal(self) {
			// 	continue readFrame
			// }
			// if !waterutil.IPv4Destination(packet).Equal(brd) {
			// 	continue readFrame
			// }
			if waterutil.IPv4Protocol(packet) != waterutil.UDP {
				continue readFrame
			}
			if waterutil.IPv4Protocol(packet) != waterutil.UDP {
				continue readFrame
			}
			// t.Logf("received broadcast frame: %#v\n", packet)
			fmt.Println("received packet ", packet)
			continue readFrame
		case <-timeout:
			t.Fatal("Waiting for  packet timeout")
		}
	}
}
