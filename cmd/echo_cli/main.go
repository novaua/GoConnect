package main

import (
	"flag"
	"fmt"
	"net"
	"os"
)

var (
	host = flag.String("host", "127.0.0.7", "Connection host")
	port = flag.String("port", "7004", "Connection port")
	connCount = flag.Int("con", 1, "Connection count")
	msgCount = flag.Int("timeout", 10, "Comunication timeout in seconds")
	help = flag.Bool("help", false, "Prints usage")
)

func usage() {
	fmt.Fprintf(os.Stderr, "Usage: %s [options]\n", os.Args[0])
	flag.PrintDefaults()
}

func main() {
	flag.Parse()

	if *help {
		usage()
		return
	}

	connections:= make([]net.Conn, 0)
	for i:=0; i<*connCount;i++ {
		conn, err:= net.Dial("tcp", fmt.Sprintf("%s:%v", *host, *port) )
		if err!=nil {
			fmt.Println(err)
			return
		}

		connections= append(connections, conn)
	}

	// ToDo: For each connection send messages until timeout is reached
	for i:=0; i<*connCount;i++ {

	}

	//ToDo: Close all connections
}
