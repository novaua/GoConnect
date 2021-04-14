package main

import "C"

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"github.com/gorilla/websocket"
	"goconnect/clientimpl"
	"goconnect/serverimpl"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
)

const (
	localCertFile = "cert.pem"
)

var handles map[int]*clientimpl.ClientServer
var nextId int

func init() {
	handles = make(map[int]*clientimpl.ClientServer)
}

//export create_sc_client
func create_sc_client(
	bind_ip string,
	bind_port *C.ushort,
	err_code *C.int,
	sc_ip string,
	sc_port uint16,
	tgt_ip string,
	tgt_port uint16,
	login, pwd string,
	reconnectTimeout int32) int {

	tlsConfig := setupCert(true)
	tr := &http.Transport{TLSClientConfig: tlsConfig}
	httpsClient := &http.Client{Transport: tr}

	fmt.Printf("inside dll bind host: %s sc_ip: %s tgt port: %v bind port: %v\n", bind_ip, sc_ip, tgt_port, *bind_port)
	if !clientimpl.CheckPortHttpReq(*httpsClient, sc_ip, sc_port, tgt_port) {
		fmt.Printf("Unable to connect to remote port: %d\n", tgt_port)
		*err_code = 1
		return 0
	}

	addr := fmt.Sprintf(":%v", *bind_port)
	fmt.Printf("Local address %s \n", addr)
	srv := clientimpl.NewClient(addr)

	srv.OnNewClient(func(c *clientimpl.ClientContext) {
		log.Print("New cli connected")
		shostPort := fmt.Sprintf("%s:%v", tgt_ip, tgt_port)
		wsc := url.URL{Scheme: "ws", Host: shostPort, Path: "/api/wsTransfer"}
		log.Printf("connecting to secured %s", wsc.String())

		header := &http.Header{}
		header.Add("Port", fmt.Sprint("%v", tgt_port))

		header.Add("SessionId", c.SessionID)

		wc, _, err := websocket.DefaultDialer.Dial(wsc.String(), *header)
		if err != nil {
			log.Fatal("dial:", err)
		}

		c.StartReadWritePump(serverimpl.NewReadPump(wc, c.Conn), serverimpl.NewReadPump(wc, c.Conn))
	})

	go srv.Listen()

	log.Printf("Listening for clients on %v", *bind_port)
	nextId++
	handles[nextId] = srv
	return nextId
}

//export add_tcp_server
func add_tcp_server(sc int, bid_ip string, bind_port *uint16, tgt_ip string, tgt_port uint16) {
	log.Println("Not implemented")
}

//export delete_sc_client
func delete_sc_client(sc_handle int) {
	srv, ok := handles[sc_handle]
	if ok {
		srv.Close()
	}

	log.Printf("Closed server  %d", sc_handle)
	delete(handles, sc_handle)
}

//export check_sc_error
func check_sc_error(sc_handle int, error_text *string) int {
	*error_text = "No error"
	return 0
}

func setupCert(insecure bool) *tls.Config {
	// Get the SystemCertPool, continue with an empty pool on error
	rootCAs, _ := x509.SystemCertPool()
	if rootCAs == nil {
		rootCAs = x509.NewCertPool()
	}

	// Read in the cert file
	certs, err := ioutil.ReadFile(localCertFile)
	if err != nil {
		log.Fatalf("Failed to append %q to RootCAs: %v", localCertFile, err)
	}

	// Append our cert to the system pool
	if ok := rootCAs.AppendCertsFromPEM(certs); !ok {
		log.Println("No certs appended, using system certs only")
	}

	// Trust the augmented cert pool in our client
	config := &tls.Config{
		InsecureSkipVerify: insecure,
		RootCAs:            rootCAs,
	}

	return config
}

func main() {}
