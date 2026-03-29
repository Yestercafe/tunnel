package framing

import (
	"encoding/binary"
	"errors"
)

const (
	// HeaderSize is the fixed v1 frame header length in bytes.
	HeaderSize = 10
	// MaxPayloadLen is the maximum allowed payload_len (16 MiB).
	MaxPayloadLen = 16777216
)

// VersionV1 is the only protocol version supported by this package (0x0001).
const VersionV1 = 0x0001

var (
	// ErrNeedMore indicates the buffer does not yet contain a full frame.
	ErrNeedMore = errors.New("need more data")
	// ErrFrameTooLarge corresponds to ERR_FRAME_TOO_LARGE in the spec (ErrCodeFrameTooLarge in docs/spec/v1/errors.md).
	ErrFrameTooLarge = errors.New("ERR_FRAME_TOO_LARGE")
	// ErrProtoVersion corresponds to ERR_PROTO_VERSION in the spec (ErrCodeProtoVersion in docs/spec/v1/errors.md).
	ErrProtoVersion = errors.New("ERR_PROTO_VERSION")
)

// Frame is a parsed v1 logical frame (fixed header + payload).
type Frame struct {
	PayloadLen uint32
	Version    uint16
	Capability uint32
	Payload    []byte
}

// ParseFrame parses one frame from the start of buf.
// On success it returns n = total bytes consumed and a copy of the payload.
// ErrNeedMore means fewer than 10 bytes or fewer than 10+payload_len bytes are available.
func ParseFrame(buf []byte) (n int, f Frame, err error) {
	if len(buf) < HeaderSize {
		return 0, Frame{}, ErrNeedMore
	}
	payloadLen := binary.BigEndian.Uint32(buf[0:4])
	version := binary.BigEndian.Uint16(buf[4:6])
	capability := binary.BigEndian.Uint32(buf[6:10])
	if payloadLen > MaxPayloadLen {
		return 0, Frame{}, ErrFrameTooLarge
	}
	total := HeaderSize + int(payloadLen)
	if len(buf) < total {
		return 0, Frame{}, ErrNeedMore
	}
	if version != VersionV1 {
		return 0, Frame{}, ErrProtoVersion
	}
	f = Frame{
		PayloadLen: payloadLen,
		Version:    version,
		Capability: capability,
		Payload:    append([]byte(nil), buf[HeaderSize:total]...),
	}
	return total, f, nil
}

// AppendFrame encodes a frame. The length field is always len(f.Payload).
func AppendFrame(f Frame) []byte {
	pl := uint32(len(f.Payload))
	out := make([]byte, HeaderSize+int(pl))
	binary.BigEndian.PutUint32(out[0:4], pl)
	binary.BigEndian.PutUint16(out[4:6], f.Version)
	binary.BigEndian.PutUint32(out[6:10], f.Capability)
	copy(out[HeaderSize:], f.Payload)
	return out
}
