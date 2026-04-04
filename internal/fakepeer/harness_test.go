package fakepeer

import (
	"crypto/tls"
	"testing"

	"tunnel/pkg/framing"
	"tunnel/pkg/protocol"
)

func mustReadFrame(t *testing.T, conn *tls.Conn) framing.Frame {
	t.Helper()
	buf := make([]byte, 0, 16384)
	tmp := make([]byte, 4096)
	for {
		_, f, err := framing.ParseFrame(buf)
		if err == nil {
			return f
		}
		if err != framing.ErrNeedMore {
			t.Fatal(err)
		}
		n, err := conn.Read(tmp)
		if n > 0 {
			buf = append(buf, tmp[:n]...)
		}
		if err != nil {
			t.Fatal(err)
		}
	}
}

func TestHarness_CreateAndJoin(t *testing.T) {
	h := NewHarness(t)
	addr, cleanup := h.Start(t)
	defer cleanup()

	cfg := h.ClientTLS()

	c1, err := tls.Dial("tcp", addr, cfg)
	if err != nil {
		t.Fatal(err)
	}
	defer c1.Close()

	c2, err := tls.Dial("tcp", addr, cfg)
	if err != nil {
		t.Fatal(err)
	}
	defer c2.Close()

	createWire := framing.AppendFrame(framing.Frame{
		Version:    framing.VersionV1,
		Capability: 0,
		Payload:    protocol.EncodeSessionCreateReq(),
	})
	if _, err := c1.Write(createWire); err != nil {
		t.Fatal(err)
	}
	ackF := mustReadFrame(t, c1)
	sid, inv, err := protocol.DecodeSessionCreateAck(ackF.Payload)
	if err != nil {
		t.Fatal(err)
	}
	if sid != h.SessionID() || inv != h.InviteCode() {
		t.Fatalf("session ack mismatch got sid=%q inv=%q", sid, inv)
	}

	join1, err := protocol.EncodeSessionJoinReq(0, sid)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := c1.Write(framing.AppendFrame(framing.Frame{
		Version: framing.VersionV1, Capability: 0, Payload: join1,
	})); err != nil {
		t.Fatal(err)
	}
	jack1 := mustReadFrame(t, c1)
	p1, err := protocol.DecodeSessionJoinAck(jack1.Payload)
	if err != nil {
		t.Fatal(err)
	}

	join2, err := protocol.EncodeSessionJoinReq(1, inv)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := c2.Write(framing.AppendFrame(framing.Frame{
		Version: framing.VersionV1, Capability: 0, Payload: join2,
	})); err != nil {
		t.Fatal(err)
	}
	jack2 := mustReadFrame(t, c2)
	p2, err := protocol.DecodeSessionJoinAck(jack2.Payload)
	if err != nil {
		t.Fatal(err)
	}

	if p1 == 0 || p2 == 0 {
		t.Fatalf("peer_id must be non-zero: %d %d", p1, p2)
	}
	if p1 == p2 {
		t.Fatalf("peer ids must differ: %d", p1)
	}
}
