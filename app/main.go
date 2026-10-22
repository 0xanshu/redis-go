package main

import (
	"fmt"
	"net"
	"os"
	"strings"
)

var _ = net.Listen
var _ = os.Exit

func handleParsing(request string) []string {
	stringData := strings.Split(request, "\r\n")
	return stringData
}

func handleConn(conn net.Conn) {
	defer conn.Close()

	buf := make([]byte, 1024)
	for {
		n, err := conn.Read(buf)
		if err != nil {
			fmt.Println("is the redis-cli up?: ", err.Error())
			return
		}

		msg := string(buf[:n])
		fmt.Printf("received: %q\n", msg)

		parsedString := handleParsing(msg)
		fmt.Println(parsedString)

		switch strings.ToUpper(parsedString[2]) {
		case "PING":
			conn.Write([]byte("+PONG\r\n"))
		case "ECHO":
			if len(parsedString) > 4 && parsedString[4] != "" {
				conn.Write([]byte("+" + parsedString[4] + "\r\n"))
			} else {
				conn.Write([]byte("*0\r\n"))
			}
		default:
			conn.Write([]byte("*0\r\n"))
		}
	}
}

func main() {
	fmt.Println("Logs..")

	l, err := net.Listen("tcp", "0.0.0.0:6379")
	if err != nil {
		fmt.Println("Failed to bind to port 6379")
		os.Exit(1)
	}
	defer l.Close()

	for {
		conn, err := l.Accept()
		if err != nil {
			fmt.Println("Error accepting connection: ", err.Error())
			os.Exit(1)
		}

		go handleConn(conn)
	}
}
