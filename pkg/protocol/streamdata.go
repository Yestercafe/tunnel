package protocol

import (
	"encoding/binary"
	"errors"
	"fmt"
)

// StreamOpenView is STREAM_OPEN after the 18-byte routing prefix.
type StreamOpenView struct {
	Prefix     RoutingPrefix
	StreamID   uint32
	Metadata   []byte
	RawPayload []byte
}

// StreamDataView is STREAM_DATA after the routing prefix.
type StreamDataView struct {
	Prefix          RoutingPrefix
	StreamID        uint32
	Flags           uint8
	InnerPayloadLen uint16
	ApplicationData []byte
	RawPayload      []byte
}

// StreamCloseView is STREAM_CLOSE after the routing prefix.
type StreamCloseView struct {
	Prefix     RoutingPrefix
	StreamID uint32
	RawPayload []byte
}

// DecodeStreamOpen parses a full STREAM_OPEN payload (including routing prefix).
func DecodeStreamOpen(payload []byte) (StreamOpenView, error) {
	if len(payload) < 24 {
		return StreamOpenView{}, fmt.Errorf("protocol: STREAM_OPEN too short: %d", len(payload))
	}
	if payload[0] != MsgTypeStreamOpen {
		return StreamOpenView{}, fmt.Errorf("protocol: STREAM_OPEN msg_type 0x%02x", payload[0])
	}
	pfx, err := ParseRoutingPrefix(payload)
	if err != nil {
		return StreamOpenView{}, err
	}
	sid := binary.BigEndian.Uint32(payload[18:22])
	mlen := int(binary.BigEndian.Uint16(payload[22:24]))
	if len(payload) < 24+mlen {
		return StreamOpenView{}, errors.New("protocol: truncated STREAM_OPEN metadata")
	}
	meta := payload[24 : 24+mlen]
	return StreamOpenView{
		Prefix:     pfx,
		StreamID:   sid,
		Metadata:   append([]byte(nil), meta...),
		RawPayload: payload,
	}, nil
}

// DecodeStreamData parses STREAM_DATA (streams-lifecycle.md).
func DecodeStreamData(payload []byte) (StreamDataView, error) {
	if len(payload) < 25 {
		return StreamDataView{}, fmt.Errorf("protocol: STREAM_DATA too short: %d", len(payload))
	}
	if payload[0] != MsgTypeStreamData {
		return StreamDataView{}, fmt.Errorf("protocol: STREAM_DATA msg_type 0x%02x", payload[0])
	}
	pfx, err := ParseRoutingPrefix(payload)
	if err != nil {
		return StreamDataView{}, err
	}
	sid := binary.BigEndian.Uint32(payload[18:22])
	flags := payload[22]
	innerLen := binary.BigEndian.Uint16(payload[23:25])
	want := 25 + int(innerLen)
	if len(payload) != want {
		return StreamDataView{}, fmt.Errorf("protocol: STREAM_DATA length %d, want %d", len(payload), want)
	}
	app := payload[25:]
	return StreamDataView{
		Prefix:          pfx,
		StreamID:        sid,
		Flags:           flags,
		InnerPayloadLen: innerLen,
		ApplicationData: app,
		RawPayload:      payload,
	}, nil
}

// DecodeStreamClose parses STREAM_CLOSE.
func DecodeStreamClose(payload []byte) (StreamCloseView, error) {
	if len(payload) < 22 {
		return StreamCloseView{}, fmt.Errorf("protocol: STREAM_CLOSE too short: %d", len(payload))
	}
	if payload[0] != MsgTypeStreamClose {
		return StreamCloseView{}, fmt.Errorf("protocol: STREAM_CLOSE msg_type 0x%02x", payload[0])
	}
	pfx, err := ParseRoutingPrefix(payload)
	if err != nil {
		return StreamCloseView{}, err
	}
	sid := binary.BigEndian.Uint32(payload[18:22])
	if len(payload) != 22 {
		return StreamCloseView{}, fmt.Errorf("protocol: STREAM_CLOSE length %d, want 22", len(payload))
	}
	return StreamCloseView{
		Prefix:     pfx,
		StreamID:   sid,
		RawPayload: payload,
	}, nil
}

// EncodeStreamData builds a STREAM_DATA payload (routing prefix + stream fields + application_data).
func EncodeStreamData(prefix RoutingPrefix, streamID uint32, flags uint8, applicationData []byte) ([]byte, error) {
	if len(applicationData) > 0xffff {
		return nil, errors.New("protocol: application_data too long")
	}
	innerLen := uint16(len(applicationData))
	out := make([]byte, 25+len(applicationData))
	out[0] = MsgTypeStreamData
	out[1] = prefix.RoutingMode
	binary.BigEndian.PutUint64(out[2:10], prefix.SrcPeerID)
	binary.BigEndian.PutUint64(out[10:18], prefix.DstPeerID)
	binary.BigEndian.PutUint32(out[18:22], streamID)
	out[22] = flags
	binary.BigEndian.PutUint16(out[23:25], innerLen)
	copy(out[25:], applicationData)
	return out, nil
}

// EncodeStreamOpen builds STREAM_OPEN with optional metadata.
func EncodeStreamOpen(prefix RoutingPrefix, streamID uint32, metadata []byte) ([]byte, error) {
	if len(metadata) > 0xffff {
		return nil, errors.New("protocol: metadata too long")
	}
	mlen := uint16(len(metadata))
	out := make([]byte, 24+len(metadata))
	out[0] = MsgTypeStreamOpen
	out[1] = prefix.RoutingMode
	binary.BigEndian.PutUint64(out[2:10], prefix.SrcPeerID)
	binary.BigEndian.PutUint64(out[10:18], prefix.DstPeerID)
	binary.BigEndian.PutUint32(out[18:22], streamID)
	binary.BigEndian.PutUint16(out[22:24], mlen)
	copy(out[24:], metadata)
	return out, nil
}

// EncodeStreamClose builds STREAM_CLOSE.
func EncodeStreamClose(prefix RoutingPrefix, streamID uint32) ([]byte, error) {
	out := make([]byte, 22)
	out[0] = MsgTypeStreamClose
	out[1] = prefix.RoutingMode
	binary.BigEndian.PutUint64(out[2:10], prefix.SrcPeerID)
	binary.BigEndian.PutUint64(out[10:18], prefix.DstPeerID)
	binary.BigEndian.PutUint32(out[18:22], streamID)
	return out, nil
}
