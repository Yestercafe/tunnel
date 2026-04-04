package client

import (
	"errors"
	"fmt"

	"tunnel/pkg/framing"
)

// ErrNotJoined is returned when sending data-plane traffic before SESSION_JOIN_ACK.
var ErrNotJoined = errors.New("client: not joined")

// ProtocolError wraps PROTOCOL_ERROR from the wire (ErrCode + UTF-8 reason).
type ProtocolError struct {
	Code   framing.ErrCode
	Reason string
}

func (e *ProtocolError) Error() string {
	return fmt.Sprintf("client: protocol error 0x%04x: %s", uint16(e.Code), e.Reason)
}
