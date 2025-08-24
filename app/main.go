package main

import (
	"fmt"
	"net"
	"os"
	"io"
)

// Ensures gofmt doesn't remove the "net" and "os" imports in stage 1 (feel free to remove this!)
var _ = net.Listen
var _ = os.Exit

func main() {
	fmt.Println("Logs from your program will appear here!")

	l, err := net.Listen("tcp", "0.0.0.0:6379")
	if err != nil {
		fmt.Println("Failed to bind to port 6379")
		os.Exit(1)
	}

	for {
		conn, err := l.Accept()
		if err != nil {
			fmt.Println("Error accepting connection: ", err.Error())
			os.Exit(1)
		}

		go handleConnection(conn)
	}

}


func handleConnection(conn net.Conn) {
	defer conn.Close()
	fmt.Println("Handling Connection")

	buf := make([]byte, 128)

	for {
		n, _ := conn.Read(buf)

		fmt.Printf("Command Received: %s\n", buf[:n])
		
		conn.Write([]byte("+PONG\r\n"))
		
	}

	// scanner := bufio.NewScanner(conn)

	data, err := io.ReadAll(conn)
	if err != nil {
		fmt.Println("Error reading from connection:", err)
		return
	}

	fmt.Printf("Data Recieved: %s", data)

	fmt.Fprint(conn, "+PONG\r\n")
	
	
}