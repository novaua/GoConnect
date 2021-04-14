package main

import (
	"crypto/tls"
	"crypto/x509"
	"flag"
	"fmt"
	"github.com/gorilla/websocket"
	"goconnect/clientimpl"
	"goconnect/serverimpl"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
)

const (
	localCertFile = "cert.pem"
)

var (
	lport = flag.String("lport", "7001", "Local port")
	sport = flag.String("sport", "7002", "Http server's port")
	shost = flag.String("shost", "127.0.0.7", "Http server's host")
	rport = flag.String("rport", "7004", "Http server's port")
	help = flag.Bool("help", false, "Prints usage")
)

func setupCert(insecure bool)  *tls.Config{
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

func main() {
	var usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [options]\n", os.Args[0])
		flag.PrintDefaults()
	}

	flag.Parse()
	if *help {
		usage()
		return
	}

	tlsConfig:=setupCert(true)

	tr := &http.Transport{TLSClientConfig: tlsConfig}
	httpsClient := &http.Client{Transport: tr}

	sporti, err := strconv.Atoi(*sport)
	rporti, err1 := strconv.Atoi(*rport)

	if err!=nil || err1 !=nil {
		log.Fatal("Unable to parse remote port: %s \n", sporti)
	}

	if !clientimpl.CheckPortHttpReq(*httpsClient, *shost, uint16(sporti), uint16(rporti)) {
		fmt.Printf("Unable to connect to remote port: %d\n", rporti)
		return
	}

	addr := fmt.Sprintf(":%s", *lport)
	srv := clientimpl.NewClient(addr)
	srv.HttpClient = httpsClient

	srv.OnNewClient(func(c *clientimpl.ClientContext) {
		log.Print("New cli connected")
		shostPort := fmt.Sprintf("%s:%s", *shost, *sport)
		wsc := url.URL{Scheme: "wss", Host: shostPort, Path: "/api/wsTransfer"}
		log.Printf("connecting to secured %s", wsc.String())

		header := &http.Header{}
		header.Add("Port", *rport)
		header.Add("SessionId", c.SessionID)

		websocket.DefaultDialer.TLSClientConfig = setupCert(true)

		wc, _, err := websocket.DefaultDialer.Dial(wsc.String(), *header)
		if err != nil {
			log.Fatal("dial:", err)
		}

		c.StartReadWritePump(serverimpl.NewReadPump(wc, c.Conn), serverimpl.NewReadPump(wc, c.Conn))
	})

	log.Printf("Listening for clients on %s", *lport)

	srv.Listen()
}
