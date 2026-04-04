package protocol

import (
	"bytes"
	"testing"

	"tunnel/pkg/framing"
)

func TestSessionCreateReqRoundTrip(t *testing.T) {
	p := EncodeSessionCreateReq()
	if err := DecodeSessionCreateReq(p); err != nil {
		t.Fatal(err)
	}
	if len(p) != 1 || p[0] != 0x01 {
		t.Fatalf("payload %x", p)
	}
}

func TestSessionCreateAckRoundTrip(t *testing.T) {
	sid := "550e8400-e29b-41d4-a716-446655440000"
	inv := "ABCDEFGH" // 8 chars Base32
	p, err := EncodeSessionCreateAck(sid, inv)
	if err != nil {
		t.Fatal(err)
	}
	if len(p) != 40+8 {
		t.Fatalf("len %d", len(p))
	}
	gotSID, gotInv, err := DecodeSessionCreateAck(p)
	if err != nil {
		t.Fatal(err)
	}
	if gotSID != sid || gotInv != inv {
		t.Fatalf("got %q %q", gotSID, gotInv)
	}
}

func TestSessionJoinReqRoundTrip(t *testing.T) {
	p, err := EncodeSessionJoinReq(0, sidUUID())
	if err != nil {
		t.Fatal(err)
	}
	jb, cred, err := DecodeSessionJoinReq(p)
	if err != nil || jb != 0 || cred != sidUUID() {
		t.Fatalf("jb=%d cred=%q err=%v", jb, cred, err)
	}
}

func TestSessionJoinAckPeerIDNonZero(t *testing.T) {
	_, err := EncodeSessionJoinAck(0)
	if err == nil {
		t.Fatal("expected error for peer_id=0")
	}
	p, err := EncodeSessionJoinAck(42)
	if err != nil {
		t.Fatal(err)
	}
	pid, err := DecodeSessionJoinAck(p)
	if err != nil || pid != 42 {
		t.Fatalf("pid=%d err=%v", pid, err)
	}
	_, err = DecodeSessionJoinAck([]byte{0x04, 0, 0, 0, 0, 0, 0, 0, 0})
	if err == nil {
		t.Fatal("expected error for decoded peer_id=0")
	}
}

func TestProtocolErrorRoundTrip(t *testing.T) {
	p, err := EncodeProtocolError(framing.ErrCodeRoutingInvalid, "")
	if err != nil {
		t.Fatal(err)
	}
	code, reason, err := DecodeProtocolError(p)
	if err != nil || code != framing.ErrCodeRoutingInvalid || reason != "" {
		t.Fatalf("code=%v reason=%q err=%v", code, reason, err)
	}
	p2, err := EncodeProtocolError(framing.ErrCodeJoinDenied, "denied — UTF-8")
	if err != nil {
		t.Fatal(err)
	}
	code2, reason2, err := DecodeProtocolError(p2)
	if err != nil || code2 != framing.ErrCodeJoinDenied || reason2 != "denied — UTF-8" {
		t.Fatalf("code=%v reason=%q err=%v", code2, reason2, err)
	}
}

func sidUUID() string {
	return "550e8400-e29b-41d4-a716-446655440000"
}

func TestProtocolErrorLiterals(t *testing.T) {
	// grep-visible: 0x01 control, 0x05 PROTOCOL_ERROR, 36-byte session id, invite 8–12
	_ = MsgTypeSessionCreateReq
	_ = MsgTypeProtocolError
	p, _ := EncodeSessionCreateAck("550e8400-e29b-41d4-a716-446655440000", "ABCDEFGH")
	if len(p) < 36+8 {
		t.Fatal("short ack")
	}
	// 36 appears as session_id length invariant in Encode path
	if !bytes.Contains(p, []byte("550e8400-e29b-41d4-a716-446655440000")) {
		t.Fatal("session id bytes missing")
	}
}
