package serverimpl

import (
	"encoding/gob"
	"fmt"
	"goconnect/common"
	"io"
	"io/ioutil"
	"log"
	"net"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

const (
	// Time allowed to write a message to the peer.
	writeWait = 120 * time.Second

	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer.
	maxBufferSize = 512 * 1024

	// For testing purpose set to 1
	maxDataChanLength = 1
)

// RawReader specifies TCP reader signatures
type RawReader interface {
	RawReader(outChan chan<- *common.ScPacket) (int64, error)
	WebWriter(inChan <-chan *common.ScPacket) (int64, error)
}

// WebReader specifies Web reader Signatures
type WebReader interface {
	WebReader(outChan chan<- *common.ScPacket, outCmdChan chan string) (int64, error)
	RawWriter(inChan <-chan *common.ScPacket) (int64, error)
}

// ClientSession client session on server
type ClientSession struct {
	server    *Server
	sessionId string

	rawToWeb *ReadPump
	webToRaw *ReadPump

	done      chan int
	waitGroup sync.WaitGroup
}

// ReadPump shared by reader writer
type ReadPump struct {
	webConn *websocket.Conn
	tcpConn net.Conn

	waitGroup sync.WaitGroup
	Done      chan int
}

// NewReadPump creates new object
func NewReadPump(webConn *websocket.Conn, tcpConn net.Conn) *ReadPump {
	rp := &ReadPump{
		webConn: webConn,
		tcpConn: tcpConn,
		Done:    make(chan int),
	}

	return rp
}

// MarkStop setting stop signal and retur waiter
func (ts *ReadPump) MarkStop() *sync.WaitGroup {
	ts.Done <- 0 //WebReader
	ts.Done <- 0 //WebWriter

	return &ts.waitGroup
}

// NewClientSession creates new client session
func NewClientSession(s *Server) *ClientSession {
	cs := new(ClientSession)
	cs.server = s

	return cs
}

// WebReader reads packets from web socket
func (ts *ReadPump) WebReader(outChan chan<- *common.ScPacket, outCmdChan chan string) (int64, error) {
	defer func() {
		ts.webConn.Close()
		close(outChan)
		close(outCmdChan)
		ts.waitGroup.Done()
	}()

	var totalRead int64
	var lastSeenID uint64
	var err error
	var messageType int
	var reader io.Reader

	ts.webConn.SetReadLimit(maxBufferSize + maxBufferSize/8)
	ts.webConn.SetReadDeadline(time.Now().Add(pongWait))
	ts.webConn.SetPongHandler(func(string) error { ts.webConn.SetReadDeadline(time.Now().Add(pongWait)); return nil })

	go func() {
		stat := <-ts.Done
		log.Printf("WR: Done signal received %d. Exiting ...", stat)
		ts.webConn.Close()
	}()

	for err == nil {
		messageType, reader, err = ts.webConn.NextReader()
		if err != nil {
			log.Println(err)
		}

		if messageType == websocket.BinaryMessage {
			dec := gob.NewDecoder(reader)
			packet := new(common.ScPacket)

			err = dec.Decode(packet)
			if err != nil {
				log.Printf("decode error: %v", err)
				break
			}

			lastSeenID = packet.Header.ID
			outChan <- packet

			totalRead += int64(packet.Header.Len)
		} else if messageType == websocket.TextMessage {
			var r io.Reader
			var p []byte

			// ToDo: Refactor ReadAll to some other way of reading
			p, err = ioutil.ReadAll(r)
			if err != nil {
				log.Printf("WS: ReadAll %v\n", err)
				break
			}

			outCmdChan <- string(p)
		} else if messageType == websocket.CloseMessage {
			outChan <- common.NewCommandPacket(common.CloseBothPacket, lastSeenID+1)
			log.Println("WS: Received close message")
			break
		} else {
			log.Printf("WS: Received message of unexpected type %v. Ignored", messageType)
		}
	}

	return totalRead, err
}

// WebWriter writes to web socket
func (ts *ReadPump) WebWriter(inChan <-chan *common.ScPacket) (int64, error) {
	defer ts.waitGroup.Done()

	var totalWritten int64
	var err error
	var writer io.WriteCloser
	ticker := time.NewTicker(pingPeriod)
	done := false

	for !done {
		select {
		case packet, ok := <-inChan:
			ts.webConn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				ts.webConn.WriteMessage(websocket.CloseMessage, []byte("Channel closed"))
				done = true
				break
			}

			writer, err = ts.webConn.NextWriter(websocket.BinaryMessage)
			if err != nil {
				log.Printf("Getting next web writer error: %v", err)
				break
			}

			//ts.webConn.WriteMessage()
			// this maybe has to be moved to another processing unit
			dec := gob.NewEncoder(writer)
			err = dec.Encode(*packet)
			if err != nil {
				log.Printf("Encode error: %v", err)
				break
			}

			totalWritten += int64(packet.Header.Len)
			if err = writer.Close(); err != nil {
				log.Printf("Writer close: %v", err)
				break
			}
		case <-ticker.C:
			ts.webConn.SetWriteDeadline(time.Now().Add(writeWait))
			if err = ts.webConn.WriteMessage(websocket.PingMessage, nil); err != nil {
				log.Printf("Ping error: %v", err)
				break
			}

		case status := <-ts.Done:
			log.Printf("Done signal received %d. Exiting ...", status)
			done = true
		}

		done = done || err != nil
	}

	log.Printf("WS: Total written to TCP %d", totalWritten)
	return totalWritten, err
}

// RawReader reads packets from TCP connection and wraps to ScPackets, send-only chan
func (ts *ReadPump) RawReader(outChan chan<- *common.ScPacket) (int64, error) {
	defer func() {
		ts.tcpConn.Close()
		close(outChan)

		ts.waitGroup.Done()
	}()

	var totalRead int64
	var nid uint64

	buf := make([]byte, maxBufferSize)

	for {
		length, err := ts.tcpConn.Read(buf)
		if err != nil {
			log.Println(err)
			outChan <- common.NewCommandPacket(common.CloseBothPacket, nid)
			return totalRead, err
		}
		nid++

		// makes copy of input buffer so buf can be reused
		outChan <- common.NewDataPacket(buf[:length], nid)
	}
}

//RawWriter writes to tcp
func (ts *ReadPump) RawWriter(inChan <-chan *common.ScPacket) (int64, error) {
	defer ts.tcpConn.Close()
	defer ts.waitGroup.Done()

	var totalWrite int64
	var lastErr error

	// range detects can close
	for packet := range inChan {
		plen := int(packet.Header.Len)
		tw := 0

		for tw < plen {
			n, err := ts.tcpConn.Write(packet.Payload[tw : plen-tw])
			if err != nil {
				log.Printf("RW: WS write failed '%s'\n", err)
				lastErr = err
				break
			}

			tw += n
		}

		totalWrite += int64(tw)
	}

	log.Printf("WS: Total read from TCP %d", totalWrite)
	return totalWrite, lastErr
}

// StartWebToRaw starts serving web to tcp bypass
func (ts *ReadPump) StartWebToRaw() {
	web2RawChan := make(chan *common.ScPacket, maxDataChanLength)
	webControl := make(chan string)

	ts.waitGroup.Add(3)

	go ts.RawWriter(web2RawChan)
	go ts.WebReader(web2RawChan, webControl)

	go func() {
		defer ts.waitGroup.Done()

		for msg := range webControl {
			fmt.Print("Web msg received:", msg)
		}
	}()
}

// StartRawToWeb starts serving web to tcp bypass
func (ts *ReadPump) StartRawToWeb() {
	packetChan := make(chan *common.ScPacket, maxDataChanLength)

	ts.waitGroup.Add(2)

	go ts.RawReader(packetChan)
	go ts.WebWriter(packetChan)
}

func (ts *ClientSession) transferHandler() {

	ts.rawToWeb.StartRawToWeb()
	ts.webToRaw.StartWebToRaw()

	done := <-ts.done
	log.Printf("Done signal %d received\n", done)

	r2wWaiter := ts.rawToWeb.MarkStop()
	w2rWaiter := ts.webToRaw.MarkStop()
	r2wWaiter.Wait()
	w2rWaiter.Wait()

	log.Println("All done")
}
