package common

import (
	"crypto/rand"
	"fmt"
	"strconv"
	"strings"

	"github.com/gorilla/websocket"
)

//Command is a command
type Command int

//ScPacketID packet id
type ScPacketID uint64

//known as cmd
const (
	WrongUninitialized Command = iota
	SimpleDataPacket
	AckDataPacket
	CloseBothPacket
	ServerShutdownPacket
	ClientShutdownPacket
	LastEnumValue
)

//ScPacketHeader is a message header
type ScPacketHeader struct {
	Cmd Command // command
	ID  uint64
	Len uint
}

//ScPacket main data packet
type ScPacket struct {
	Header  ScPacketHeader
	Payload []byte
}

// IsData return tru e for data packs
func (sc *ScPacket) IsData() bool {
	return sc.Header.Cmd == SimpleDataPacket
}

// NewDataPacket creates new packet from byte slice. Input data is copied.
func NewDataPacket(data []byte, id uint64) *ScPacket {
	sz := len(data)
	res := &ScPacket{
		Header: ScPacketHeader{
			ID:  id,
			Cmd: SimpleDataPacket,
			Len: uint(sz),
		},
		Payload: make([]byte, len(data)),
	}

	copy(res.Payload, data)
	return res
}

// NewCommandPacket creates new packet from byte array
func NewCommandPacket(cmd Command, id uint64) *ScPacket {
	if cmd == SimpleDataPacket {
		panic("Simple data packet was not expected")
	}

	return &ScPacket{
		ScPacketHeader{
			ID:  id,
			Cmd: cmd,
		},
		nil,
	}
}

//ToCompressedString makes ACK string from vector
func ToCompressedString(packets []ScPacketID) string {
	packetsCount := 0
	var strBuilder strings.Builder
	count := len(packets)
	for i, elem := range packets {
		iNext := i + 1
		notLast := iNext < count
		if notLast && elem+1 == packets[iNext] { // not the last
			if packetsCount == 0 {
				fmt.Fprintf(&strBuilder, "%d", elem)
			}
			packetsCount++
		} else {
			if packetsCount != 0 {
				fmt.Fprintf(&strBuilder, ":%d", packetsCount)
				packetsCount = 0
			} else {
				fmt.Fprintf(&strBuilder, "%d", elem)
			}
			if notLast {
				strBuilder.WriteString(" ")
			}
		}
	}
	return strBuilder.String()
}

//ToScPackets extracts packet ids from comreseed string
func ToScPackets(compressedString string) ([]ScPacketID, error) {
	resArray := make([]ScPacketID, 0, 16)
	for _, elem := range strings.Split(compressedString, " ") {
		pair := strings.Split(elem, ":")
		if len(pair) > 1 {
			offset, err1 := strconv.ParseUint(pair[1], 10, 32)
			base, err2 := strconv.ParseUint(pair[0], 10, 64)
			if err1 != nil || err2 != nil {
				return resArray, fmt.Errorf("Unable to parse %v", pair)
			}

			for j := base; j <= base+offset; j++ {
				resArray = append(resArray, ScPacketID(j))
			}
		} else {
			base, err := strconv.ParseUint(elem, 10, 64)
			if err != nil {
				return resArray, fmt.Errorf("Unable to parse %v", pair)
			}

			resArray = append(resArray, ScPacketID(base))
		}
	}

	return resArray, nil
}

// NewGuid generates new guid
func NewGuid() (string, error) {

	b := make([]byte, 16)
	_, err := rand.Read(b)
	if err != nil {
		fmt.Println("Error: ", err)
		return "", err
	}

	uuid := fmt.Sprintf("%X-%X-%X-%X-%X", b[0:4], b[4:6], b[6:8], b[8:10], b[10:])

	return uuid, nil
}

// Context is a default connection ctx
type Context struct {
	SessionID string
	WebConn   *websocket.Conn
	//TCPConn   net.Conn
}

// Close the default ctx
func (ctx *Context) Close() {
	if ctx.WebConn != nil {
		ctx.WebConn.Close()
		ctx.WebConn = nil
	}
}
