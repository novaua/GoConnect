package main

import (
	"fmt"
	"goconnect/common"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strconv"
	"time"

	"github.com/gorilla/websocket"
	gcfg "gopkg.in/gcfg.v1"
)

// ConnectionCfg is a configuration for single secured connection
type ConnectionCfg struct {
	HttpsServer     string
	HttpsServerPort int
	LocalTcpPort    int
	RemoteTcpPort   int
}

type LoggerCfg struct {
	File string
}

type CliConfig struct {
	Connection ConnectionCfg
	Logger     LoggerCfg
}

const defaultConfig = `
    [connection]
    HttpsServer = 127.0.0.1
    HttpsServerPort = 7002
    LocalTcpPort = 7001
    RemoteTcpPort  = 7004
	
	[logger]
	File = conncli.log `

func loadConfiguration(cfgFile string) CliConfig {
	var err error
	var cfg CliConfig

	if cfgFile != "" {
		err = gcfg.ReadFileInto(&cfg, cfgFile)
	} else {
		err = gcfg.ReadStringInto(&cfg, defaultConfig)
	}

	if err != nil {
		panic("Unbale to locad configuration")
	}

	return cfg
}

/*
func generateCertrTmpCode() {
	// generate a new key-pair
	rootKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		log.Fatalf("generating random key: %v", err)
	}

	rootCertTmpl, err := utils.CertTemplate()
	if err != nil {
		log.Fatalf("creating cert template: %v", err)
	}
	// describe what the certificate will be used for
	rootCertTmpl.IsCA = true
	rootCertTmpl.KeyUsage = x509.KeyUsageCertSign | x509.KeyUsageDigitalSignature
	rootCertTmpl.ExtKeyUsage = []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth, x509.ExtKeyUsageClientAuth}
	rootCertTmpl.IPAddresses = []net.IP{net.ParseIP("127.0.0.1")}

	rootCert, rootCertPEM, err := utils.CreateCert(rootCertTmpl, rootCertTmpl, &rootKey.PublicKey, rootKey)
	if err != nil {
		log.Fatalf("error creating cert: %v", err)
	}

	fmt.Println("\n--------")
	fmt.Printf("%s\n", rootCertPEM)
	fmt.Printf("%#x\n", rootCert.Signature) // more ugly binary
}
*/
//var addr = flag.String("addr", "localhost:7002", "http service address")

func main() {
	//flag.Parse()

	log.SetFlags(0)

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	cfg := loadConfiguration("")
	srvAddr := fmt.Sprintf("%s:%d", cfg.Connection.HttpsServer, cfg.Connection.HttpsServerPort)
	fmt.Printf("Server %s:%d \nLogger %s", cfg.Connection.HttpsServer, cfg.Connection.HttpsServerPort, cfg.Logger.File)
	fmt.Println()

	wsc := url.URL{Scheme: "ws", Host: srvAddr, Path: "/api/wsTransfer"}
	log.Printf("connecting to %s", wsc.String())

	strPort := strconv.Itoa(cfg.Connection.RemoteTcpPort)

	header := &http.Header{}
	header.Add("Port", strPort)

	id, _ := common.NewGuid()

	header.Add("SessionId", id)

	c, _, err := websocket.DefaultDialer.Dial(wsc.String(), *header)
	if err != nil {
		log.Fatal("dial:", err)
	}
	defer c.Close()

	done := make(chan struct{})

	go func() {
		defer close(done)
		for {
			_, message, err := c.ReadMessage()
			if err != nil {
				log.Println("read:", err)
				return
			}
			log.Printf("recv: %s", message)
		}
	}()

	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-done:
			return
		case t := <-ticker.C:
			err := c.WriteMessage(websocket.BinaryMessage, []byte(fmt.Sprintf("%s\n", t.String())))
			if err != nil {
				log.Println("write:", err)
				return
			}
		case <-interrupt:
			log.Println("interrupt")

			// Cleanly close the connection by sending a close message and then
			// waiting (with timeout) for the server to close the connection.
			err := c.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
			if err != nil {
				log.Println("write close:", err)
				return
			}
			select {
			case <-done:
			case <-time.After(time.Second):
			}
			return
		}
	}
}
