package main

import (
	"flag"
	"fmt"
	"goconnect/serverimpl"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
)

var addr = flag.String("addr", ":7002", "http service address")
var isTls = flag.Bool("tls", true, "serve TLS")

func startHttpServer(addr *string, tls bool) *http.Server {
	srv := &http.Server{Addr: *addr}
	go func() {
		if tls {
			if err := srv.ListenAndServeTLS("cert.srv.pem", "key.srv.pem"); err != http.ErrServerClosed {
				log.Fatalf("ListenAndServe: '%v'", err)
			}
		} else {
			if err := srv.ListenAndServe(); err != http.ErrServerClosed {
				log.Fatalf("ListenAndServe: '%v'", err)
			}
		}
	}()

	return srv
}

func main() {
	flag.Parse()
	log.SetFlags(log.Lshortfile)

	signals := make(chan os.Signal, 1)
	done := make(chan bool, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		sig := <-signals
		fmt.Println()
		fmt.Println(sig)
		done <- true
	}()

	addrSlice := strings.Split(*addr, ":")
	intAddr, err := strconv.Atoi(addrSlice[1])
	if err != nil {
		log.Printf("Unable to parse port no %s", *addr)
		intAddr = 0
	}

	appServer := serverimpl.NewServer()
	go appServer.RunMsgPump()

	if intAddr >= 400 {
		appServer.Config.HttpsPort = intAddr
	}

	http.Handle("/", appServer.Router)

	srv := startHttpServer(addr, *isTls)

	fmt.Printf("Serving %s\n", strconv.Itoa(appServer.Config.HttpsPort))
	fmt.Println("presss ctrl-c to stop server")

	<-done

	fmt.Println("exiting")

	appServer.Done(0)
	srv.Close()
}
