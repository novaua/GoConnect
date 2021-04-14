package tcp_server

import (
	"bufio"
	"log"
	"net"
	"sync"
)

// Client holds info about connection
type Client struct {
	conn   net.Conn
	Server *server
	Ctx    interface{}
}

// TCP server
type server struct {
	address                  string // Address to open connection: localhost:9999
	onNewClientCallback      func(c *Client)
	onClientConnectionClosed func(c *Client, err error)
	onNewMessage             func(c *Client, message string)
	onNewBinMessage          func(c *Client, message []byte)

	clients map[*Client]int
	waiter  sync.WaitGroup
}

// Read client data from channel
func (c *Client) listen() {
	defer c.Server.waiter.Done()
	c.Server.waiter.Add(1)

	c.Server.onNewClientCallback(c)
	reader := bufio.NewReader(c.conn)
	for {
		message, err := reader.ReadString('\n')
		if err != nil {
			c.conn.Close()
			c.Server.onClientConnectionClosed(c, err)
			return
		}
		c.Server.onNewMessage(c, message)
	}
}

func (c *Client) listenBin() {
	c.Server.onNewClientCallback(c)
	reader := bufio.NewReader(c.conn)
	buff := make([]byte, 8192)
	for {
		lenght, err := reader.Read(buff)
		if err != nil {
			c.conn.Close()
			c.Server.onClientConnectionClosed(c, err)
			return
		}
		c.Server.onNewBinMessage(c, buff[:lenght])
	}
}

// Send text message to client
func (c *Client) Send(message string) error {
	_, err := c.conn.Write([]byte(message))
	return err
}

// Send bytes to client
func (c *Client) SendBytes(b []byte) (int, error) {
	rc, err := c.conn.Write(b)
	return rc, err
}

func (c *Client) Conn() net.Conn {
	return c.conn
}

func (c *Client) Context() interface{} {
	return c.Ctx
}

func (c *Client) SetContext(ctx interface{}) {
	c.Ctx = ctx
}

func (c *Client) Close() error {
	return c.conn.Close()
}

// Called right after server starts listening new client
func (s *server) OnNewClient(callback func(c *Client)) {
	s.onNewClientCallback = callback
}

// Called right after connection closed
func (s *server) OnClientConnectionClosed(callback func(c *Client, err error)) {
	s.onClientConnectionClosed = callback
}

// Called when Client receives new message
func (s *server) OnNewMessage(callback func(c *Client, message string)) {
	s.onNewMessage = callback
}

// Called when Client receives new binary message
func (s *server) OnNewBinMessage(callback func(c *Client, message []byte)) {
	s.onNewBinMessage = callback
}

// Start network server
func (s *server) Listen() {
	listener, err := net.Listen("tcp", s.address)
	if err != nil {
		log.Fatal("Error starting TCP server.")
	}
	defer listener.Close()

	for {
		conn, _ := listener.Accept()
		client := &Client{
			conn:   conn,
			Server: s,
		}
		go client.listen()
	}
}

func (s *server) ListenAsync() net.Listener {
	listener, err := net.Listen("tcp", s.address)
	if err != nil {
		log.Fatal("Error starting TCP server.")
	}

	go func() {
		defer func() {
			listener.Close()
			s.waiter.Done()
		}()

		s.waiter.Add(1)

		count := 0
		for {
			conn, err := listener.Accept()
			if err != nil {
				log.Println("Error accepting connection.")
				break
			}

			client := &Client{
				conn:   conn,
				Server: s,
			}

			count++
			s.clients[client] = count
			go client.listen()
		}
	}()

	return listener
}

func (s *server) CloseAsyncServer() {
	for c := range s.clients {
		c.Close()
	}

	log.Print("All connections closed")
	s.waiter.Wait()
	log.Print("All done")
}

// Start network server
func (s *server) ListenBin() {
	listener, err := net.Listen("tcp", s.address)
	if err != nil {
		log.Fatal("Error starting TCP server.")
	}
	defer listener.Close()

	for {
		conn, _ := listener.Accept()
		client := &Client{
			conn:   conn,
			Server: s,
		}
		go client.listenBin()
	}
}

// New Creates new tcp server instance
func New(address string) *server {
	log.Println("Creating server with address", address)
	server := &server{
		address: address,
	}

	server.OnNewClient(func(c *Client) {})
	server.OnNewMessage(func(c *Client, message string) {})
	server.OnNewBinMessage(func(c *Client, message []byte) {})
	server.OnClientConnectionClosed(func(c *Client, err error) {})
	server.clients = make(map[*Client]int)
	return server
}
