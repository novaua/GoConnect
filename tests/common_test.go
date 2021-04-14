package tests

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"goconnect/common"
	"goconnect/utils"
	"reflect"
	"strings"
	"sync"
	"testing"
)

// TestToCompressedString is the first test
func TestToCompressedString(t *testing.T) {
	packs := make([]common.ScPacketID, 1)
	packs[0] = common.ScPacketID(44)

	res := common.ToCompressedString(packs)

	if res != fmt.Sprintf("%d", packs[0]) {
		t.Errorf("result was wrong %s", res)
	}
}

// TestToCompressedString is the first test
func TestToCompressedStringSparse(t *testing.T) {
	packs := make([]common.ScPacketID, 4)
	var expected strings.Builder
	for i := 0; i < len(packs); i++ {
		packs[i] = common.ScPacketID((i + 1) * 2)
		expected.WriteString(fmt.Sprintf("%d", packs[i]))
		if i+1 != len(packs) {
			expected.WriteString(" ")
		}
	}

	res := common.ToCompressedString(packs)

	if res != expected.String() {
		t.Errorf("result was %s", res)
	}
}

// TestToCompressedString is the first test
func TestToCompressedStringCont(t *testing.T) {
	packs := make([]common.ScPacketID, 10)
	for i := 0; i < 5; i++ {
		packs[i] = common.ScPacketID(i)
	}

	for i := 5; i < 10; i++ {
		packs[i] = common.ScPacketID(33 - 5 + i)
	}

	res := common.ToCompressedString(packs)

	if res != "0:4 33:4" {
		t.Errorf("result was %s", res)
	}
}

func TestToCompressedDecompress(t *testing.T) {
	packs := make([]common.ScPacketID, 10)
	for i := 0; i < 5; i++ {
		packs[i] = common.ScPacketID(i)
	}

	for i := 5; i < 10; i++ {
		packs[i] = common.ScPacketID(33 - 5 + i)
	}

	res := common.ToCompressedString(packs)
	restoredPack, err := common.ToScPackets(res)

	if err != nil {
		t.Error("Parse error!")
	}

	if !reflect.DeepEqual(packs, restoredPack) {
		t.Error("Restored was not equal")
	}
}

func TestGuid(t *testing.T) {
	g1, err1 := common.NewGuid()
	g2, err2 := common.NewGuid()

	if err2 != nil {
		t.Errorf("Guids geneartion failed %v", err2)
	}

	if err1 != nil {
		t.Errorf("Guids geneartion failed %v", err1)
	}

	if g1 == g2 {
		t.Errorf("Guids were equal %s == %s", g1, g2)
	}
}

func TestSyncMap(t *testing.T) {
	sm := common.NewSyncMap()
	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		sm.Store("foo", 1)
		sm.Store("bar", 1)
	}()

	go func() {
		defer wg.Done()
		sm.Store("koo", 1)
		sm.Store("bar", 1)
	}()

	sm.Store("koo", 1)

	wg.Wait()

	if _, ok := sm.Load("bar"); !ok {
		t.Error("bar failed")
	}

	if l := sm.Lenght(); l != 3 {
		t.Error("len failed")
	}
}

func responder(c chan string, done chan bool) {
	for {
		select {
		case str := <-c:
			if str == "boo" {
				c <- "yes boo!"
			} else {
				c <- str
			}
		case <-done:
			break
		}
	}
}

func TestChanel(t *testing.T) {
	strChan := make(chan string)
	done := make(chan bool)
	go responder(strChan, done)

	tm := "Hello chan"
	strChan <- tm
	resp := <-strChan

	if resp != tm {
		t.Errorf("Expected %s, received %s", tm, resp)
	}

	strChan <- "boo"
	resp = <-strChan

	t.Log(resp)

	done <- true
}

func TestGobEncoding(t *testing.T) {
	var network bytes.Buffer        // Stand-in for a network connection
	enc := gob.NewEncoder(&network) // Will write to network.
	dec := gob.NewDecoder(&network) // Will read from network.

	// Encode (send) the value.
	dp := common.NewDataPacket([]byte("Test one pack"), 33)

	err := enc.Encode(*dp)
	if err != nil {
		t.Fatalf("encode error: %v", err)
	}

	// Decode (receive) the value.
	var dpd common.ScPacket
	err = dec.Decode(&dpd)
	if err != nil {
		t.Fatalf("decode error: %v", err)
	}

	if !reflect.DeepEqual(*dp, dpd) {
		t.Error("Decoded missmatch")
	}
}

func TestPanicOnWrongCmd(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("The code did not panic")
		}
	}()

	common.NewCommandPacket(common.SimpleDataPacket, 0)
}

func TestCreateCmd(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("The code did panic")
		}
	}()

	dp := common.NewCommandPacket(common.ServerShutdownPacket, 0)

	if dp.Header.Cmd != common.ServerShutdownPacket {
		t.Error("Failed")
	}
}

func asChan(vs ...int) <-chan int {
	c := make(chan int)
	go func() {
		for _, v := range vs {
			c <- v
			//time.Sleep(time.Duration(rand.Intn(1000)) * time.Millisecond)
		}
		close(c)
	}()
	return c
}

func TestDuplicateChan(t *testing.T) {
	src := asChan(1, 2, 3, 4, 5)
	out1, out2 := utils.DublicateChanInt(src)

	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		defer wg.Done()

		for r := range out1 {
			t.Logf("Out 1 %d", r)
		}
	}()

	for r := range out2 {
		t.Logf("Out 2 %d", r)
	}

	wg.Wait()
}
