// Package protocol implements v1 payload semantics (control + data plane) without parsing framing headers.
package protocol

// Control-plane message types (session-create-join.md).
const (
	MsgTypeSessionCreateReq  uint8 = 0x01
	MsgTypeSessionCreateAck  uint8 = 0x02
	MsgTypeSessionJoinReq    uint8 = 0x03
	MsgTypeSessionJoinAck    uint8 = 0x04
	MsgTypeProtocolError     uint8 = 0x05
)

// Data-plane stream opcodes (streams-lifecycle.md).
const (
	MsgTypeStreamOpen  uint8 = 0x10
	MsgTypeStreamData  uint8 = 0x11
	MsgTypeStreamClose uint8 = 0x12
)
