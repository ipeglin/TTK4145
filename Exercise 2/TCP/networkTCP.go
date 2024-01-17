package main

import (
	"fmt"
	"net"
	"sync"
	"time"
)

func reader(conn net.Conn) {
	for {
		buffer := make([]byte, 1024)
		_, err := conn.Read(buffer)
		if err != nil {
			fmt.Println("Error reading from server", err)
			return
		}
		fmt.Println("Server response: ", string(buffer))
	}
}

func client(wg *sync.WaitGroup) {

	conn, _ := net.Dial("tcp", "10.100.23.129:33546")
	defer conn.Close()
	defer wg.Done()
	go reader(conn)

	data := append([]byte("Connect to: 10.100.23.19:33546"), 0)
	_, err := conn.Write(data)
	if err != nil {
		fmt.Println("Write error: ", err)
		return
	}

	for i := 0; i < 3; i++ {
		time.Sleep(1 * time.Second)
		data := append([]byte("Hello World from Group 9"), 0)
		_, err := conn.Write(data)
		if err != nil {
			fmt.Println("Write error: ", err)
			return
		}
	}
}

func handleClient(conn net.Conn) {
	defer conn.Close()

	buffer := make([]byte, 1024)

	for {
		n, err := conn.Read(buffer)
		if err != nil {
			fmt.Println("Read error:", err)
			return
		}

		fmt.Println("Received from server:", buffer[:n])
	}
}

func server(wg *sync.WaitGroup) {

	listener, err := net.Listen("tcp", "10.100.23.19:33546")
	if err != nil {
		fmt.Println("Server error:", err)
	}
	defer listener.Close()
	defer wg.Done()
	fmt.Println("Server is listeneing on port: 33546")

	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("Accept error:", err)
			continue
		}

		go handleClient(conn)
	}
}

func main() {
	var wg sync.WaitGroup
	wg.Add(1)
	go client(&wg)

	wg.Add(1)
	go server(&wg)

	wg.Wait()
}
