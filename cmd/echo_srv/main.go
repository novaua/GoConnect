package main

import (
	"flag"
	"fmt"
	"goconnect/tcp_server"
	"log"
	"os"
	"os/signal"
	"syscall"
)

var (
	port = flag.String("port", "7004", "Listening port")
)

func main() {
	flag.Parse()

	signals := make(chan os.Signal, 1)
	done := make(chan bool, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		sig := <-signals
		fmt.Println()
		fmt.Println(sig)
		done <- true
	}()

	addr := fmt.Sprintf(":%s", *port)
	srv := tcp_server.New(addr)

	srv.OnNewClient(func(c *tcp_server.Client) {
		log.Print("New cli connected")
	})

	srv.OnNewMessage(func(c *tcp_server.Client, message string) {
		log.Printf("Message received: %s", message)

		err := c.Send(message)
		if err != nil {
			log.Print("Write failed")
		}
	})

	srv.OnClientConnectionClosed(func(c *tcp_server.Client, err error) {
		log.Printf("Bye %s", c.Conn().RemoteAddr().String())
	})

	listener := srv.ListenAsync()

	fmt.Println("Press Ctrl-c to close server")
	<-done

	listener.Close()
	srv.CloseAsyncServer()
}
