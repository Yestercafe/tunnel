package protocol

import (
	"encoding/binary"
	"errors"
	"fmt"
)

// Routing modes (routing-modes.md ROUTE-01 / ROUTE-02).
const (
	RoutingModeInvalid    uint8 = 0
	RoutingModeBroadcast  uint8 = 1
	RoutingModeUnicast    uint8 = 2
)

// RoutingPrefix is the 18-byte data-plane prefix (msg_type + routing_mode + src + dst).
type RoutingPrefix struct {
	MsgType      uint8
	RoutingMode  uint8
	SrcPeerID    uint64
	DstPeerID    uint64
}

// ParseRoutingPrefix parses the first 18 bytes of a payload.
func ParseRoutingPrefix(payload []byte) (RoutingPrefix, error) {
	if len(payload) < 18 {
		return RoutingPrefix{}, fmt.Errorf("protocol: routing prefix needs 18 bytes, got %d", len(payload))
	}
	return RoutingPrefix{
		MsgType:     payload[0],
		RoutingMode: payload[1],
		SrcPeerID:   binary.BigEndian.Uint64(payload[2:10]),
		DstPeerID:   binary.BigEndian.Uint64(payload[10:18]),
	}, nil
}

// ValidateRoutingIntent enforces joint routing_mode + dst_peer_id rules.
func ValidateRoutingIntent(mode uint8, dst uint64) error {
	switch mode {
	case RoutingModeInvalid:
		return errors.New("protocol: routing_mode 0 is invalid")
	case RoutingModeBroadcast:
		if dst != 0 {
			return fmt.Errorf("protocol: BROADCAST requires dst_peer_id 0, got %d", dst)
		}
	case RoutingModeUnicast:
		if dst == 0 {
			return errors.New("protocol: UNICAST requires non-zero dst_peer_id")
		}
	default:
		return fmt.Errorf("protocol: unknown routing_mode %d", mode)
	}
	return nil
}
