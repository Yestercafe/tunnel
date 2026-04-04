package protocol

import (
	"encoding/binary"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"unicode/utf8"

	"tunnel/pkg/framing"
)

var (
	sessionIDRe = regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)
	// RFC 4648 alphabet (no padding).
	base32Chars = "ABCDEFGHIJKLMNOPQRSTUVWXYZ234567"
)

func validInviteCode(s string) bool {
	if len(s) < 8 || len(s) > 12 {
		return false
	}
	for _, r := range s {
		if strings.IndexRune(base32Chars, r) < 0 {
			return false
		}
	}
	return true
}

// EncodeSessionCreateReq returns the minimal SESSION_CREATE_REQ payload (msg_type only).
func EncodeSessionCreateReq() []byte {
	return []byte{MsgTypeSessionCreateReq}
}

// DecodeSessionCreateReq validates a SESSION_CREATE_REQ payload.
func DecodeSessionCreateReq(payload []byte) error {
	if len(payload) != 1 {
		return fmt.Errorf("protocol: SESSION_CREATE_REQ length %d, want 1", len(payload))
	}
	if payload[0] != MsgTypeSessionCreateReq {
		return fmt.Errorf("protocol: SESSION_CREATE_REQ msg_type 0x%02x, want 0x%02x", payload[0], MsgTypeSessionCreateReq)
	}
	return nil
}

// EncodeSessionCreateAck builds SESSION_CREATE_ACK (msg_type + body per session-create-join.md).
func EncodeSessionCreateAck(sessionID, inviteCode string) ([]byte, error) {
	if !sessionIDRe.MatchString(sessionID) {
		return nil, fmt.Errorf("protocol: invalid session_id")
	}
	if len(sessionID) != 36 {
		return nil, fmt.Errorf("protocol: session_id length %d, want 36", len(sessionID))
	}
	if !validInviteCode(inviteCode) {
		return nil, fmt.Errorf("protocol: invalid invite_code")
	}
	l := 1 + 2 + 36 + 1 + len(inviteCode)
	out := make([]byte, l)
	out[0] = MsgTypeSessionCreateAck
	binary.BigEndian.PutUint16(out[1:3], 36)
	copy(out[3:39], sessionID)
	out[39] = uint8(len(inviteCode))
	copy(out[40:], inviteCode)
	return out, nil
}

// DecodeSessionCreateAck parses and validates SESSION_CREATE_ACK.
func DecodeSessionCreateAck(payload []byte) (sessionID, inviteCode string, err error) {
	if len(payload) < 40 {
		return "", "", fmt.Errorf("protocol: SESSION_CREATE_ACK too short: %d", len(payload))
	}
	if payload[0] != MsgTypeSessionCreateAck {
		return "", "", fmt.Errorf("protocol: SESSION_CREATE_ACK msg_type 0x%02x", payload[0])
	}
	sidLen := binary.BigEndian.Uint16(payload[1:3])
	if sidLen != 36 {
		return "", "", fmt.Errorf("protocol: session_id_len %d, want 36", sidLen)
	}
	if len(payload) < 3+int(sidLen)+1 {
		return "", "", errors.New("protocol: truncated SESSION_CREATE_ACK")
	}
	sessionID = string(payload[3 : 3+sidLen])
	if !sessionIDRe.MatchString(sessionID) {
		return "", "", fmt.Errorf("protocol: session_id format invalid")
	}
	inviteLen := int(payload[3+sidLen])
	if inviteLen < 8 || inviteLen > 12 {
		return "", "", fmt.Errorf("protocol: invite_code_len %d, want 8–12", inviteLen)
	}
	end := 40 + inviteLen
	if len(payload) != end {
		return "", "", fmt.Errorf("protocol: SESSION_CREATE_ACK length %d, want %d", len(payload), end)
	}
	inviteCode = string(payload[40:40+inviteLen])
	if !validInviteCode(inviteCode) {
		return "", "", fmt.Errorf("protocol: invite_code format invalid")
	}
	return sessionID, inviteCode, nil
}

// EncodeSessionJoinReq builds SESSION_JOIN_REQ.
func EncodeSessionJoinReq(joinBy uint8, credential string) ([]byte, error) {
	if joinBy > 1 {
		return nil, fmt.Errorf("protocol: join_by %d, want 0 or 1", joinBy)
	}
	cred := []byte(credential)
	if len(cred) > 0xffff {
		return nil, errors.New("protocol: credential too long")
	}
	out := make([]byte, 1+1+2+len(cred))
	out[0] = MsgTypeSessionJoinReq
	out[1] = joinBy
	binary.BigEndian.PutUint16(out[2:4], uint16(len(cred)))
	copy(out[4:], cred)
	return out, nil
}

// DecodeSessionJoinReq parses SESSION_JOIN_REQ.
func DecodeSessionJoinReq(payload []byte) (joinBy uint8, credential string, err error) {
	if len(payload) < 4 {
		return 0, "", fmt.Errorf("protocol: SESSION_JOIN_REQ too short: %d", len(payload))
	}
	if payload[0] != MsgTypeSessionJoinReq {
		return 0, "", fmt.Errorf("protocol: SESSION_JOIN_REQ msg_type 0x%02x", payload[0])
	}
	joinBy = payload[1]
	if joinBy > 1 {
		return 0, "", fmt.Errorf("protocol: join_by %d, want 0 or 1", joinBy)
	}
	clen := int(binary.BigEndian.Uint16(payload[2:4]))
	if len(payload) != 4+clen {
		return 0, "", fmt.Errorf("protocol: SESSION_JOIN_REQ length %d, want %d", len(payload), 4+clen)
	}
	return joinBy, string(payload[4:]), nil
}

// EncodeSessionJoinAck builds SESSION_JOIN_ACK with non-zero peer_id.
func EncodeSessionJoinAck(peerID uint64) ([]byte, error) {
	if peerID == 0 {
		return nil, errors.New("protocol: peer_id must not be 0 in SESSION_JOIN_ACK")
	}
	out := make([]byte, 9)
	out[0] = MsgTypeSessionJoinAck
	binary.BigEndian.PutUint64(out[1:9], peerID)
	return out, nil
}

// DecodeSessionJoinAck parses SESSION_JOIN_ACK.
func DecodeSessionJoinAck(payload []byte) (peerID uint64, err error) {
	if len(payload) != 9 {
		return 0, fmt.Errorf("protocol: SESSION_JOIN_ACK length %d, want 9", len(payload))
	}
	if payload[0] != MsgTypeSessionJoinAck {
		return 0, fmt.Errorf("protocol: SESSION_JOIN_ACK msg_type 0x%02x", payload[0])
	}
	peerID = binary.BigEndian.Uint64(payload[1:9])
	if peerID == 0 {
		return 0, errors.New("protocol: peer_id must not be 0")
	}
	return peerID, nil
}

// EncodeProtocolError builds PROTOCOL_ERROR (no routing prefix).
func EncodeProtocolError(err framing.ErrCode, reason string) ([]byte, error) {
	if !utf8.ValidString(reason) {
		return nil, errors.New("protocol: reason must be valid UTF-8")
	}
	r := []byte(reason)
	if len(r) > 0xffff {
		return nil, errors.New("protocol: reason too long")
	}
	out := make([]byte, 5+len(r))
	out[0] = MsgTypeProtocolError
	binary.BigEndian.PutUint16(out[1:3], uint16(err))
	binary.BigEndian.PutUint16(out[3:5], uint16(len(r)))
	copy(out[5:], r)
	return out, nil
}

// DecodeProtocolError parses PROTOCOL_ERROR after msg_type byte (full payload).
func DecodeProtocolError(payload []byte) (framing.ErrCode, string, error) {
	if len(payload) < 5 {
		return 0, "", fmt.Errorf("protocol: PROTOCOL_ERROR too short: %d", len(payload))
	}
	if payload[0] != MsgTypeProtocolError {
		return 0, "", fmt.Errorf("protocol: PROTOCOL_ERROR msg_type 0x%02x", payload[0])
	}
	code := framing.ErrCode(binary.BigEndian.Uint16(payload[1:3]))
	rlen := int(binary.BigEndian.Uint16(payload[3:5]))
	if len(payload) != 5+rlen {
		return 0, "", fmt.Errorf("protocol: PROTOCOL_ERROR length %d, want %d", len(payload), 5+rlen)
	}
	reason := string(payload[5:])
	if !utf8.ValidString(reason) {
		return 0, "", errors.New("protocol: invalid UTF-8 in reason")
	}
	return code, reason, nil
}
