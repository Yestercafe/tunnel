package relay_test

import (
	"context"
	"crypto/tls"
	"testing"
	"time"

	"tunnel/internal/fakepeer"
	"tunnel/pkg/client"
	"tunnel/pkg/framing"
	"tunnel/pkg/protocol"
	"tunnel/pkg/relay"
)

func TestRelay_ClientCreateJoin(t *testing.T) {
	srvCfg, pool := fakepeer.LocalhostTLSConfig(t)
	srv := &relay.Server{
		ListenAddr: "127.0.0.1:0",
		TLSConfig:  srvCfg,
		Registry:   relay.NewSessionRegistry(),
	}
	if err := srv.Listen(); err != nil {
		t.Fatal(err)
	}
	addr := srv.Addr().String()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go func() { _ = srv.Serve(ctx) }()
	defer srv.Close()

	tlsClient := &tls.Config{
		RootCAs:    pool,
		ServerName: "localhost",
		MinVersion: tls.VersionTLS12,
	}

	c1, err := client.Dial(context.Background(), addr, tlsClient)
	if err != nil {
		t.Fatal(err)
	}
	defer c1.Close()

	sid, inv, err := c1.CreateSession(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if sid == "" || inv == "" {
		t.Fatal("empty session or invite")
	}

	p1, err := c1.JoinSession(context.Background(), true, sid)
	if err != nil {
		t.Fatal(err)
	}
	if p1 == 0 {
		t.Fatal("peer_id 0")
	}

	c2, err := client.Dial(context.Background(), addr, tlsClient)
	if err != nil {
		t.Fatal(err)
	}
	defer c2.Close()

	p2, err := c2.JoinSession(context.Background(), true, sid)
	if err != nil {
		t.Fatal(err)
	}
	if p2 == 0 {
		t.Fatal("peer_id 0")
	}
	if p1 == p2 {
		t.Fatalf("peer ids must differ: both %d", p1)
	}
}

// E2E-01: two clients, same session, broadcast
func TestRelay_StreamData_Broadcast(t *testing.T) {
	srvCfg, pool := fakepeer.LocalhostTLSConfig(t)
	srv := &relay.Server{
		ListenAddr: "127.0.0.1:0",
		TLSConfig:  srvCfg,
		Registry:   relay.NewSessionRegistry(),
	}
	if err := srv.Listen(); err != nil {
		t.Fatal(err)
	}
	addr := srv.Addr().String()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go func() { _ = srv.Serve(ctx) }()
	defer srv.Close()

	tlsClient := &tls.Config{
		RootCAs:    pool,
		ServerName: "localhost",
		MinVersion: tls.VersionTLS12,
	}

	c1, err := client.Dial(context.Background(), addr, tlsClient)
	if err != nil {
		t.Fatal(err)
	}
	defer c1.Close()

	sid, _, err := c1.CreateSession(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if _, err := c1.JoinSession(context.Background(), true, sid); err != nil {
		t.Fatal(err)
	}

	c2, err := client.Dial(context.Background(), addr, tlsClient)
	if err != nil {
		t.Fatal(err)
	}
	defer c2.Close()

	if _, err := c2.JoinSession(context.Background(), true, sid); err != nil {
		t.Fatal(err)
	}

	ctx2, cancel2 := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel2()

	want := []byte("broadcast-payload")
	if err := c1.SendStreamData(ctx2, protocol.RoutingModeBroadcast, 0, 1, 0, want); err != nil {
		t.Fatal(err)
	}

	f, err := c2.ReadFrame(ctx2)
	if err != nil {
		t.Fatal(err)
	}
	v, err := protocol.DecodeStreamData(f.Payload)
	if err != nil {
		t.Fatal(err)
	}
	if string(v.ApplicationData) != string(want) {
		t.Fatalf("got %q, want %q", v.ApplicationData, want)
	}
}

// E2E-01: two clients, same session, unicast
func TestRelay_StreamData_Unicast(t *testing.T) {
	srvCfg, pool := fakepeer.LocalhostTLSConfig(t)
	srv := &relay.Server{
		ListenAddr: "127.0.0.1:0",
		TLSConfig:  srvCfg,
		Registry:   relay.NewSessionRegistry(),
	}
	if err := srv.Listen(); err != nil {
		t.Fatal(err)
	}
	addr := srv.Addr().String()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go func() { _ = srv.Serve(ctx) }()
	defer srv.Close()

	tlsClient := &tls.Config{
		RootCAs:    pool,
		ServerName: "localhost",
		MinVersion: tls.VersionTLS12,
	}

	c1, err := client.Dial(context.Background(), addr, tlsClient)
	if err != nil {
		t.Fatal(err)
	}
	defer c1.Close()

	sid, _, err := c1.CreateSession(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if _, err := c1.JoinSession(context.Background(), true, sid); err != nil {
		t.Fatal(err)
	}

	c2, err := client.Dial(context.Background(), addr, tlsClient)
	if err != nil {
		t.Fatal(err)
	}
	defer c2.Close()

	p2, err := c2.JoinSession(context.Background(), true, sid)
	if err != nil {
		t.Fatal(err)
	}

	ctx2, cancel2 := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel2()

	want := []byte("unicast-payload")
	if err := c1.SendStreamData(ctx2, protocol.RoutingModeUnicast, p2, 2, 0, want); err != nil {
		t.Fatal(err)
	}

	f, err := c2.ReadFrame(ctx2)
	if err != nil {
		t.Fatal(err)
	}
	v, err := protocol.DecodeStreamData(f.Payload)
	if err != nil {
		t.Fatal(err)
	}
	if string(v.ApplicationData) != string(want) {
		t.Fatalf("got %q, want %q", v.ApplicationData, want)
	}
}

// E2E-02: STREAM_DATA before JOIN_ACK → PROTOCOL_ERROR (ErrCodeRoutingInvalid)
func TestRelay_StreamData_BeforeJoinAck(t *testing.T) {
	srvCfg, pool := fakepeer.LocalhostTLSConfig(t)
	srv := &relay.Server{
		ListenAddr: "127.0.0.1:0",
		TLSConfig:  srvCfg,
		Registry:   relay.NewSessionRegistry(),
	}
	if err := srv.Listen(); err != nil {
		t.Fatal(err)
	}
	addr := srv.Addr().String()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go func() { _ = srv.Serve(ctx) }()
	defer srv.Close()

	tlsClient := &tls.Config{
		RootCAs:    pool,
		ServerName: "localhost",
		MinVersion: tls.VersionTLS12,
	}

	c1, err := client.Dial(context.Background(), addr, tlsClient)
	if err != nil {
		t.Fatal(err)
	}
	defer c1.Close()

	if _, _, err := c1.CreateSession(context.Background()); err != nil {
		t.Fatal(err)
	}

	payload, err := protocol.EncodeStreamData(protocol.RoutingPrefix{
		RoutingMode: protocol.RoutingModeBroadcast,
		SrcPeerID:   0,
		DstPeerID:   0,
	}, 1, 0, []byte("prejoin"))
	if err != nil {
		t.Fatal(err)
	}
	wire := framing.AppendFrame(framing.Frame{
		Version:    framing.VersionV1,
		Capability: 0,
		Payload:    payload,
	})

	ctx2, cancel2 := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel2()
	if d, ok := ctx2.Deadline(); ok {
		_ = c1.UnderlyingTLSConn().SetWriteDeadline(d)
	}
	if _, err := c1.UnderlyingTLSConn().Write(wire); err != nil {
		t.Fatal(err)
	}

	f, err := c1.ReadFrame(ctx2)
	if err != nil {
		t.Fatal(err)
	}
	if len(f.Payload) == 0 || f.Payload[0] != protocol.MsgTypeProtocolError {
		t.Fatalf("expected PROTOCOL_ERROR, got %#v", f.Payload[:minLen(4, len(f.Payload))])
	}
	code, reason, err := protocol.DecodeProtocolError(f.Payload)
	if err != nil {
		t.Fatal(err)
	}
	if code != framing.ErrCodeRoutingInvalid {
		t.Fatalf("got err_code %v, want ErrCodeRoutingInvalid", code)
	}
	if reason != "ERR_ROUTING_INVALID" {
		t.Fatalf("got reason %q, want ERR_ROUTING_INVALID", reason)
	}
}

// E2E-02: illegal unicast dst → PROTOCOL_ERROR (ErrCodeRoutingInvalid)
func TestRelay_StreamData_UnicastMissingDst(t *testing.T) {
	srvCfg, pool := fakepeer.LocalhostTLSConfig(t)
	srv := &relay.Server{
		ListenAddr: "127.0.0.1:0",
		TLSConfig:  srvCfg,
		Registry:   relay.NewSessionRegistry(),
	}
	if err := srv.Listen(); err != nil {
		t.Fatal(err)
	}
	addr := srv.Addr().String()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go func() { _ = srv.Serve(ctx) }()
	defer srv.Close()

	tlsClient := &tls.Config{
		RootCAs:    pool,
		ServerName: "localhost",
		MinVersion: tls.VersionTLS12,
	}

	c1, err := client.Dial(context.Background(), addr, tlsClient)
	if err != nil {
		t.Fatal(err)
	}
	defer c1.Close()

	sid, _, err := c1.CreateSession(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if _, err := c1.JoinSession(context.Background(), true, sid); err != nil {
		t.Fatal(err)
	}

	ctx2, cancel2 := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel2()

	if err := c1.SendStreamData(ctx2, protocol.RoutingModeUnicast, 999, 2, 0, []byte("x")); err != nil {
		t.Fatal(err)
	}

	f, err := c1.ReadFrame(ctx2)
	if err != nil {
		t.Fatal(err)
	}
	if len(f.Payload) == 0 || f.Payload[0] != protocol.MsgTypeProtocolError {
		t.Fatalf("expected PROTOCOL_ERROR, got %#v", f.Payload[:minLen(4, len(f.Payload))])
	}
	code, _, err := protocol.DecodeProtocolError(f.Payload)
	if err != nil {
		t.Fatal(err)
	}
	if code != framing.ErrCodeRoutingInvalid {
		t.Fatalf("got err_code %v, want ErrCodeRoutingInvalid", code)
	}
}

func minLen(a, b int) int {
	if a < b {
		return a
	}
	return b
}
