package serverimpl

import (
	"encoding/json"
	"fmt"
	"goconnect/utils"
	"io"
	"log"
	"net"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
)

// Server serves all application http
// Using architecture inspired by https://medium.com/statuscode/how-i-write-go-http-services-after-seven-years-37c208122831
type Server struct {
	Router *mux.Router

	register   chan *ClientSession
	unregister chan string

	exists chan *existChekPair
	done   chan int

	sessions map[string]*ClientSession
	Config   Config
}

type existChekPair struct {
	id      string
	session chan *ClientSession
}

// NewServer returns a new server instance.
func NewServer() *Server {
	srv := Server{
		Router:   mux.NewRouter(),
		sessions: make(map[string]*ClientSession),
		Config:   NewConfig(false),

		register:   make(chan *ClientSession),
		unregister: make(chan string),

		exists: make(chan *existChekPair),
		done:   make(chan int),
	}

	srv.routes()
	return &srv
}

// Register registers new sessio
func (srv *Server) Register(s *ClientSession) {
	srv.register <- s
}

// Unregister client by id
func (srv *Server) Unregister(id string) {
	srv.unregister <- id
}

// GetRegistered checks if client is registred
func (srv *Server) GetRegistered(id string) *ClientSession {
	obj := &existChekPair{
		id:      id,
		session: make(chan *ClientSession),
	}

	srv.exists <- obj
	outClient := <-obj.session
	return outClient
}

// RunMsgPump process server's messages. Todo: this probably has to be private
func (srv *Server) RunMsgPump() {
	for {
		select {
		case client := <-srv.register:
			srv.sessions[client.sessionId] = client

		case clientId := <-srv.unregister:
			if _, ok := srv.sessions[clientId]; ok {
				delete(srv.sessions, clientId)
			}

		case checkRequest := <-srv.exists:
			if _, ok := srv.sessions[checkRequest.id]; ok {
				checkRequest.session <- srv.sessions[checkRequest.id]
			} else {
				checkRequest.session <- nil
			}

		case stat := <-srv.done:
			fmt.Printf("Done signal received: %d\n", stat)
			for k, v := range srv.sessions {
				fmt.Printf("Closing session: %s\n", k)
				v.done <- stat
			}

			break
		}
	}
}

func (s *Server) Done(status int) {
	s.done <- status
}

func (s *Server) routes() {
	s.Router.HandleFunc("/api/time", s.handleTime(time.RFC3339))
	s.Router.HandleFunc("/api/checkPort", s.handleCheckPort())
	s.Router.HandleFunc("/api/wsTransfer", s.handleWsTransfer())
	s.Router.HandleFunc("/", s.handleHome())
}

func (s *Server) handleTime(timeFormat string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		if r.URL.Path != "/api/time" {
			http.Error(w, "Not found.", http.StatusNotFound)
			return
		}

		if r.Method != "GET" {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		t := time.Now()
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		fmt.Fprintf(w,
			`<html><head><title>HTTPTimeServer powered by Go Libraries</title>
			 <meta http-equiv="refresh" content="1"></head>
			 <body><p style="text-align: center; font-size: 48px;">
			%s </p></body></html>`, t.Format(timeFormat))

	}
}

func checkPort(port int) bool {
	err := utils.Ping("tcp", fmt.Sprintf(":%d", port), 2*time.Second)
	fmt.Printf("Port %d is %s\n ", port, utils.IfSelect(err == nil, "listening", "closed"))
	return err == nil
}

func (s *Server) handleCheckPort() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/checkPort" {
			http.Error(w, "Not found.", http.StatusNotFound)
			return
		}

		var result map[string]interface{}

		json.NewDecoder(r.Body).Decode(&result)

		iport := result["Port"]
		if iport == nil {
			http.Error(w, "Port not found.", http.StatusNotFound)
			return
		}

		i, err := strconv.Atoi(iport.(string))

		if err != nil {
			fmt.Print(err)
			http.Error(w, "Port not found.", http.StatusNotFound)
			return
		}

		if checkPort(i) {
			w.WriteHeader(http.StatusOK)
		} else {
			w.WriteHeader(http.StatusExpectationFailed)
		}
	}
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  maxBufferSize,
	WriteBufferSize: maxBufferSize,
}

func (s *Server) handleWsTransfer() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/wsTransfer" {
			http.Error(w, "Not found.", http.StatusNotFound)
			return
		}

		fmt.Printf("Request method was '%v'\n", r.Method)
		portStr := r.Header.Get("Port")
		sid := r.Header.Get("SessionId")
		fmt.Printf("Port method was %s\n", portStr)
		fmt.Printf("Session Id was %s\n", sid)

		tcpConn, err := net.Dial("tcp", fmt.Sprintf("%s:%s", "127.0.0.7", portStr))
		if err != nil {
			http.Error(w, fmt.Sprintf("Unable to dial port: %s", portStr), http.StatusNotFound)
			return
		}

		fmt.Print("Upgrading connection...")
		wsConn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			tcpConn.Close()
			log.Println(err)
			return
		}

		wc := &ClientSession{
			server:    s,
			sessionId: sid,
			rawToWeb: &ReadPump{
				webConn: wsConn,
				tcpConn: tcpConn,
			},

			webToRaw: &ReadPump{
				webConn: wsConn,
				tcpConn: tcpConn,
			},

			done: make(chan int),
		}

		s.register <- wc

		go wc.transferHandler()
	}
}

func (s *Server) handleHome() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.Error(w, "Not found.", http.StatusNotFound)
			return
		}

		if r.Method != "GET" {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		io.WriteString(w, fmt.Sprintf(
			"<html><body>Secure Server is listening on %d port</body></html>",
			s.Config.HttpsPort))
	}
}
