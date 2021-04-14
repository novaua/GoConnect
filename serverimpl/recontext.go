package serverimpl

import (
	"goconnect/common"
)

type resendResp struct {
	// compressed ack string of received from peer
	req string

	// ordered stream of the rest
	resp chan *common.ScPacket
}

// ReContext contains all requred to store reconnected packets
type ReContext struct {

	//sent to peer but not yet acknowledget
	sendPackets map[uint64]*common.ScPacket

	// received from peer
	ackPackets []common.ScPacketID

	addSent     chan *common.ScPacket
	addReceived chan *common.ScPacketID

	// accepts ack list from remote peer and removes packets from sendPackets
	processAck chan string

	// creates compressed acknowledge for peer
	makeAck chan string

	//negotiate resend
	resender resendResp
}

type tuple struct {
	name  string
	value string
}

/*
func newCtx() *ReContext {
	return &ReContext{
		sendPackets: make(map[string]string),
		ackPackets:       make(chan tuple),
		addSent:    make(chan string),
		addReceived:     make(chan *reqResp),

		processAck : make(Type, size IntegerType),
		makeAck :

	}
}

func (c *ctx) run() {
	for {
		select {
		case kv := <-c.add:
			c.sharedMap[kv.name] = kv.value

		case k := <-c.remove:
			if _, ok := c.sharedMap[k]; ok {
				delete(c.sharedMap, k)
			}
		case check := <-c.check:
			if val, ok := c.sharedMap[check.req]; ok {
				check.resp <- val
			} else {
				check.resp <- "<not found>"
			}
		}
	}
}
*/
