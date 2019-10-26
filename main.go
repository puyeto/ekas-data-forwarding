package main

import (
	"fmt"
	"io"
	"math/rand"
	"net"
	"time"
)

func handleConnection(c net.Conn) {
	fmt.Printf("Serving %s\n", c.RemoteAddr().String())
	for {

		bytes := make([]byte, 1024)

		reqLen, err := c.Read(bytes)
		if err != nil {
			if err != io.EOF {
				fmt.Println("End of file error:", err)
			}
			fmt.Println("Error reading:", err.Error(), reqLen)
		}

		fmt.Println(reqLen, bytes)
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
	rand.Seed(time.Now().Unix())

	for {
		c, err := l.Accept()
		if err != nil {
			fmt.Println(err)
			return
		}
		go handleConnection(c)
	}
}
