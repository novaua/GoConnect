package tests

import (
	"fmt"
	"goconnect/serverimpl"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
)

func Test_HelloRouter(t *testing.T) {
	appServer := serverimpl.NewServer()
	ts := httptest.NewServer(appServer.Router)
	defer ts.Close()

	res, err := http.Get(ts.URL)
	if err != nil {
		t.Fatal(err)
	}

	body, err := ioutil.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		t.Fatal(err)
	}

	exp := "Secure Server"
	strbody := string(body)

	if !strings.Contains(strbody, exp) {
		t.Fatalf("Expected %s got %s", exp, body)
	}
}

func Test_CheckPort(t *testing.T) {
	appServer := serverimpl.NewServer()
	ts := httptest.NewServer(appServer.Router)
	defer ts.Close()

	reqStr := fmt.Sprintf(`{"Port":"%s"}`, "80")
	resp, err := http.Post(ts.URL+"/api/checkPort", "application/json", strings.NewReader(reqStr))
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, resp.StatusCode, http.StatusOK, fmt.Sprintf("Status was %s", resp.Status))
}

//ToDo this shuld work fully one day
func Test_WebSocketBasic(t *testing.T) {
	var err error

	appServer := serverimpl.NewServer()
	s := httptest.NewServer(appServer.Router)
	defer s.Close()

	d := websocket.Dialer{}
	dialAddr := "ws://" + s.Listener.Addr().String() + "/api/wsTransfer"
	c, resp, err := d.Dial(dialAddr, nil)
	if err != nil {
		t.Fatal(err)
	}

	if got, want := resp.StatusCode, http.StatusSwitchingProtocols; got != want {
		t.Errorf("resp.StatusCode = %q, want %q", got, want)
	}

	err = c.WriteJSON("test")
	if err != nil {
		t.Fatal(err)
	}
}

func echoServe(t *testing.T, url string) {
	l, err := net.Listen("tcp", url)
	if err != nil {
		t.Fatal(err)
	}
	defer l.Close()
	for {
		conn, err := l.Accept()
		if err != nil {
			t.Fatal(err)
		}

		defer conn.Close()

		buf, err := ioutil.ReadAll(conn)
		if err != nil {
			t.Fatal(err)
		}

		_, err = fmt.Fprintf(conn, "Reply: %s", string(buf))
		if err != nil {
			t.Fatal(err)
		}

		if strings.Contains(string(buf), "bye") {
			break
		}
	}
}

func openServerSocketAndSendReceive(t *testing.T, url string, sendList []string) (receiveList []string) {
	l, err := net.Listen("tcp", url)
	if err != nil {
		t.Fatal(err)
	}

	defer l.Close()
	for {
		conn, err := l.Accept()
		if err != nil {
			t.Fatal(err)
		}

		defer conn.Close()
		receiveList = make([]string, len(sendList))

		i := 0
		for msg := range sendList {
			_, err = fmt.Fprint(conn, msg)
			if err != nil {
				t.Fatal(err)
			}

			buf, err := ioutil.ReadAll(conn)
			if err != nil {
				t.Fatal(err)
			}

			receiveList[i] = string(buf)
		}
	}
}

func Test_WebSocketEcho(t *testing.T) {
	var wsUri = "wss://echo.websocket.org/"

	// to make this manual test run ncat -lk 7008
	var tcpUrl = "localhost:7008"

	d := websocket.Dialer{}
	c, resp, err := d.Dial(wsUri, nil)
	if err != nil {
		t.Fatal(err)
	}

	if got, want := resp.StatusCode, http.StatusSwitchingProtocols; got != want {
		t.Errorf("resp.StatusCode = %q, want %q", got, want)
	}

	defer c.Close()
	tcpConn, err := net.Dial("tcp4", tcpUrl)
	if err != nil {
		t.Fatal(err)
	}

	defer tcpConn.Close()

	rp := serverimpl.NewReadPump(c, tcpConn)

	rp.StartRawToWeb()
	rp.StartWebToRaw()
	log.Print("slipping")

	time.Sleep(time.Duration(time.Minute / 3))

	log.Print("exiting")

	waiter := rp.MarkStop()
	waiter.Wait()

	log.Print("wait done")
}
