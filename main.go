package main

import (
	"fmt"
	"io"
	"math/rand"
	"net"
	"time"
)

func handleConnection(c net.Conn, conn net.Conn) {
	clientJobs := make(chan []byte)
	go sendData(clientJobs, conn)

	fmt.Printf("Serving %s\n", c.RemoteAddr().String())
	for {

		bytes := make([]byte, 70)

		reqLen, err := c.Read(bytes)
		if err != nil {
			if err != io.EOF {
				fmt.Println("End of file error:", err)
				return
			}
			fmt.Println("Error reading:", err.Error(), reqLen)
		}

		clientJobs <- bytes
	}
	c.Close()
}

func main() {
	fmt.Println("Listening to port 6033 ....")
	PORT := ":6033"
	l, err := net.Listen("tcp4", PORT)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer l.Close()

	conn, err := net.Dial("tcp", "157.230.203.114:6033")
	if err != nil {
		fmt.Println(err)
	}
	defer conn.Close()

	rand.Seed(time.Now().Unix())

	for {
		c, err := l.Accept()
		if err != nil {
			fmt.Println(err)
			return
		}
		go handleConnection(c, conn)
	}
}

func sendData(data chan []byte, conn net.Conn) {
	// connect to this socket
	fmt.Println("Connecting as client")

	for {
		data := <-data
		n, err := conn.Write(data)
		if err != nil {
			if err == io.EOF {
				fmt.Println("Client ", conn.LocalAddr(), " disconnected")
				conn.Close()
			} else {
				fmt.Println("Failed writing bytes to conn: ", conn, " with error ", err)
				conn.Close()
			}
		}
		fmt.Println("Wrote bytes", n, " to connection ", conn.RemoteAddr())

		// fmt.Fprint(conn, data)
		// listen for reply

		fmt.Println(data)
	}

}
