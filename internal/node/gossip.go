package node

import (
	nodeProto "github.com/aminsalami/repartido/proto/node"
	"github.com/hashicorp/memberlist"
	"google.golang.org/protobuf/proto"
)

// gossip implement the memberlist.Delegate interface
type gossip struct {
	service *nodeService
}

func (g *gossip) NodeMeta(limit int) []byte {
	return []byte("? to be implemented ?")
}

func (g *gossip) extractEvent(bytes []byte) *nodeProto.Event {
	// parse the received message
	event := nodeProto.Event{}
	err := proto.Unmarshal(bytes, &event)
	if err != nil {
		logger.Errorw("Invalid broadcast message", "event", string(bytes))
		return nil
	}
	if !g.service.witnessClock(event.LClock) {
		logger.Debugw("event is duplicate", "Origin", event.Origin, "Id", event.Id, "LClock", event.LClock)
		return nil
	}
	logger.Debugw("new event", "Origin", event.Origin, "Id", event.Id, "LClock", event.LClock)
	return &event
}

// NotifyMsg is called whenever a broadcast message is received
func (g *gossip) NotifyMsg(bytes []byte) {
	if len(bytes) == 0 {
		return
	}
	logger.Debugw("-> new broadcast received")

	event := g.extractEvent(bytes)
	if event == nil {
		return
	}

	rebroadcast := false
	switch payload := event.Payload.(type) {
	case *nodeProto.Event_RingState:
		g.service.importRing(payload.RingState)
		rebroadcast = true
	case nil:
		logger.Errorw("broadcast message with invalid payload", "event", string(bytes))
		return
	}

	// gossip the received message
	if rebroadcast {
		logger.Debugw("re-broadcasting the event", "event.Id", event.Id, "event.LClock", event.LClock)
		newBytes := make([]byte, len(bytes))
		copy(newBytes, bytes)
		g.service.queueBroadcast(&broadcast{newBytes})
	}
}

func (g *gossip) GetBroadcasts(overhead, limit int) [][]byte {
	messages := g.service.getBroadcastMessages(overhead, limit)
	if len(messages) == 0 {
		return [][]byte{}
	}
	size := 0
	for _, buf := range messages {
		size = size + len(buf)
	}
	return messages
}

// LocalState is used for a TCP Push/Pull. This is sent to
// the remote side in addition to the membership information. Any
// data can be sent here. See MergeRemoteState as well. The `join`
// boolean indicates this is for a join instead of a push/pull.
func (g *gossip) LocalState(join bool) []byte {
	if !join {
		return nil
	}
	// send my status, in this case, the ring status as a new event
	toSendBuf := g.service.makeRingEvent()
	return toSendBuf
}

func (g *gossip) MergeRemoteState(buf []byte, join bool) {
	logger.Debugw("-> MergeRemoteState", "join-value", join, "buf-size", len(buf))
	if join {
		event := g.extractEvent(buf)
		if event == nil {
			return
		}
		rs := event.Payload.(*nodeProto.Event_RingState)
		g.service.importRing(rs.RingState)
	}
}

// ----------------------------------

type broadcast struct {
	msg []byte
}

func (b *broadcast) Invalidates(msg memberlist.Broadcast) bool {
	// TODO: implement invalidates
	// Based on: https://github.com/hashicorp/memberlist/issues/10
	// TransmitLimitedQueue allows for custom invalidation behavior. For example, if I'm
	// broadcasting "A is dead", and A comes back alive, I want to invalidate the previous message, and
	// start broadcasting "A is alive" instead)
	//
	return false
}

func (b *broadcast) Message() []byte {
	return b.msg
}

func (b *broadcast) Finished() {

}
