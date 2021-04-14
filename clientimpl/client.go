package clientimpl

import (
	"fmt"
	"goconnect/common"
	"goconnect/serverimpl"
	"log"
	"net"
	"net/http"
	"net/url"
	"strings"
	"sync"
)

type ClientServer struct {
	address string
	clients map[string]*ClientContext

	register   chan *ClientContext
	unregister chan *ClientContext

	port int
	listener *net.Listener

	HttpClient *http.Client
	//TlsCfg
	done chan int

	onNewClientCallback func(c *ClientContext)
}

//NewClient creates new client server
func NewClient(addr string) *ClientServer {
	cs := &ClientServer{
		address: addr,
	}

	cs.OnNewClient(func(c *ClientContext) {})
	return cs
}

// Called right after server starts listening new client
func (s *ClientServer) OnNewClient(callback func(c *ClientContext)) {
	s.onNewClientCallback = callback
}

func (s *ClientServer) GetListeningPort() int {
	s.port = (*s.listener).Addr().(*net.TCPAddr).Port
	return s.port
}

func (s *ClientServer) Close() {
	// ToDo: add correct close of server
	(*s.listener).Close();
}

// Start network server
func (s *ClientServer) Listen() {
	listener, err := net.Listen("tcp", s.address)
	if err != nil {
		log.Fatal("Error starting TCP server.")
	}

	defer listener.Close()
	s.listener = &listener

	for {
		conn, _ := listener.Accept()
		id, _ := common.NewGuid()
		client := &ClientContext{
			Server:    s,
			Conn:      conn,
			done:      make(chan int),
			SessionID: id,
		}

		go client.Server.onNewClientCallback(client)
	}
}

// ClientContext is a default connection ctx
type ClientContext struct {
	SessionID string
	Server    *ClientServer
	Conn      net.Conn

	rawToWeb *serverimpl.ReadPump
	webToRaw *serverimpl.ReadPump

	done      chan int
	waitGroup sync.WaitGroup
}

// StartReadWritePump starts ne pump
func (ctx *ClientContext) StartReadWritePump(rawToWeb, webToRaw *serverimpl.ReadPump) {
	ctx.rawToWeb = rawToWeb
	ctx.webToRaw = webToRaw

	ctx.rawToWeb.StartRawToWeb()
	ctx.webToRaw.StartWebToRaw()

	go func() {
		done := <-ctx.done
		log.Printf("Done signal %d received\n", done)

		r2wWaiter := ctx.rawToWeb.MarkStop()
		w2rWaiter := ctx.webToRaw.MarkStop()
		r2wWaiter.Wait()
		w2rWaiter.Wait()

		log.Println("All done")
	}()
}

func CheckPortHttpReq(client http.Client, host string, port, chekPort uint16) bool {
	shostPort := fmt.Sprintf("%s:%d", host, port)
	wsc := url.URL{Scheme: "https", Host: shostPort, Path: "/api/checkPort"}

	reqStr := fmt.Sprintf(`{"Port":"%v"}`, chekPort)

	resp, err := client.Post(wsc.String(), "application/json", strings.NewReader(reqStr) )
	if err != nil {
		fmt.Println(err)
		return false
	}

	defer resp.Body.Close()
	return resp.StatusCode == http.StatusOK
}