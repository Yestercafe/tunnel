package client

import (
	"context"
	"testing"
	"time"

	"tunnel/internal/fakepeer"
	"tunnel/pkg/framing"
	"tunnel/pkg/protocol"
)

func TestClient_CreateSession(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	h := fakepeer.NewHarness(t)
	addr, cleanup := h.Start(t)
	defer cleanup()

	c, err := Dial(ctx, addr, h.ClientTLS())
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()

	sid, inv, err := c.CreateSession(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if sid != h.SessionID() || inv != h.InviteCode() {
		t.Fatalf("session mismatch sid=%q inv=%q", sid, inv)
	}
}

func TestClient_JoinSession(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	h := fakepeer.NewHarness(t)
	addr, cleanup := h.Start(t)
	defer cleanup()

	c, err := Dial(ctx, addr, h.ClientTLS())
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()

	sid, inv, err := c.CreateSession(ctx)
	if err != nil {
		t.Fatal(err)
	}
	_ = inv

	pid, err := c.JoinSession(ctx, true, sid)
	if err != nil {
		t.Fatal(err)
	}
	if pid == 0 {
		t.Fatal("peer_id must be non-zero")
	}
}

func TestClient_StreamData(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	h := fakepeer.NewHarness(t)
	addr, cleanup := h.Start(t)
	defer cleanup()

	cfg := h.ClientTLS()

	ca, err := Dial(ctx, addr, cfg)
	if err != nil {
		t.Fatal(err)
	}
	defer ca.Close()
	cb, err := Dial(ctx, addr, cfg)
	if err != nil {
		t.Fatal(err)
	}
	defer cb.Close()

	sid, inv, err := ca.CreateSession(ctx)
	if err != nil {
		t.Fatal(err)
	}

	if _, err := ca.JoinSession(ctx, true, sid); err != nil {
		t.Fatal(err)
	}
	if _, err := cb.JoinSession(ctx, false, inv); err != nil {
		t.Fatal(err)
	}

	if err := ca.SendStreamData(ctx, protocol.RoutingModeBroadcast, 0, 1, 0, []byte("bc")); err != nil {
		t.Fatal(err)
	}

	fb, err := readUntilStreamData(ctx, cb)
	if err != nil {
		t.Fatal(err)
	}
	sd, err := protocol.DecodeStreamData(fb.Payload)
	if err != nil {
		t.Fatal(err)
	}
	if sd.StreamID != 1 || string(sd.ApplicationData) != "bc" {
		t.Fatalf("broadcast payload got stream=%d app=%q", sd.StreamID, sd.ApplicationData)
	}

	if err := cb.SendStreamData(ctx, protocol.RoutingModeUnicast, ca.PeerID(), 2, 0, []byte("u")); err != nil {
		t.Fatal(err)
	}

	fa, err := readUntilStreamData(ctx, ca)
	if err != nil {
		t.Fatal(err)
	}
	su, err := protocol.DecodeStreamData(fa.Payload)
	if err != nil {
		t.Fatal(err)
	}
	if su.StreamID != 2 || string(su.ApplicationData) != "u" {
		t.Fatalf("unicast payload got stream=%d app=%q", su.StreamID, su.ApplicationData)
	}
	if su.Prefix.DstPeerID != ca.PeerID() {
		t.Fatalf("dst peer want %d got %d", ca.PeerID(), su.Prefix.DstPeerID)
	}
}

func readUntilStreamData(ctx context.Context, c *Client) (framing.Frame, error) {
	for {
		f, err := c.ReadFrame(ctx)
		if err != nil {
			return framing.Frame{}, err
		}
		if len(f.Payload) > 0 && f.Payload[0] == protocol.MsgTypeStreamData {
			return f, nil
		}
	}
}

func TestClient_SendStreamData_NotJoined(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	h := fakepeer.NewHarness(t)
	addr, cleanup := h.Start(t)
	defer cleanup()

	c, err := Dial(ctx, addr, h.ClientTLS())
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()

	if _, _, err := c.CreateSession(ctx); err != nil {
		t.Fatal(err)
	}

	err = c.SendStreamData(ctx, protocol.RoutingModeBroadcast, 0, 1, 0, []byte("x"))
	if err != ErrNotJoined {
		t.Fatalf("want ErrNotJoined got %v", err)
	}
}
