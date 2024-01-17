package main

import (
	"fmt"
	"net"
	"sync"
	"time"
)

func recevier(wg *sync.WaitGroup) {
	defer wg.Done()
	addr, err := net.ResolveUDPAddr("udp", "0.0.0.0:20009")
	if err != nil {
		panic(err)
	}
	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	fmt.Println("Listening for UDP packets")

	buffer := make([]byte, 1024)
	for {
		n, remoteAddr, err := conn.ReadFromUDP(buffer)
		if err != nil {
			fmt.Println("Error receving data: ", err)
			continue
		}
		fmt.Println("Recived %d bytes from %s: %s \n", n, remoteAddr, string(buffer[:n]))
	}
}

func sender(wg *sync.WaitGroup) {
	defer wg.Done()
	for {
		serverAddr, _ := net.ResolveUDPAddr("udp", "255.255.255.255:20009")

		conn, _ := net.DialUDP("udp", nil, serverAddr)
		defer conn.Close()

		message := []byte("Hello to you UDP")
		conn.Write(message)
		fmt.Println("Sendt message to server:", string(message))
		time.Sleep(1 * time.Second)
	}
}

func main() {
	var wg sync.WaitGroup
	wg.Add(1)
	go recevier(&wg)

	wg.Add(1)
	go sender(&wg)

	wg.Wait()
}
