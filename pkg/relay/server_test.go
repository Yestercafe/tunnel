package relay_test

import (
	"context"
	"crypto/tls"
	"testing"

	"tunnel/internal/fakepeer"
	"tunnel/pkg/framing"
	"tunnel/pkg/protocol"
	"tunnel/pkg/relay"
)

func TestServer_ProtoVersionError(t *testing.T) {
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

	conn, err := tls.Dial("tcp", addr, &tls.Config{
		RootCAs:    pool,
		ServerName: "localhost",
		MinVersion: tls.VersionTLS12,
	})
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()

	bad := framing.AppendFrame(framing.Frame{
		Version:    0x9999,
		Capability: 0,
		Payload:    []byte{protocol.MsgTypeSessionCreateReq},
	})
	if _, err := conn.Write(bad); err != nil {
		t.Fatal(err)
	}

	buf := make([]byte, 4096)
	n, err := conn.Read(buf)
	if err != nil {
		t.Fatal(err)
	}
	_, f, err := framing.ParseFrame(buf[:n])
	if err != nil {
		t.Fatal(err)
	}
	code, _, err := protocol.DecodeProtocolError(f.Payload)
	if err != nil {
		t.Fatal(err)
	}
	if code != framing.ErrCodeProtoVersion {
		t.Fatalf("got err_code %v, want ErrCodeProtoVersion", code)
	}
}

func TestServer_NeedMore(t *testing.T) {
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

	conn, err := tls.Dial("tcp", addr, &tls.Config{
		RootCAs:    pool,
		ServerName: "localhost",
		MinVersion: tls.VersionTLS12,
	})
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()

	wire := framing.AppendFrame(framing.Frame{
		Version:    framing.VersionV1,
		Capability: 0,
		Payload:    protocol.EncodeSessionCreateReq(),
	})
	for i := 0; i < len(wire); i++ {
		if _, err := conn.Write(wire[i : i+1]); err != nil {
			t.Fatal(err)
		}
	}

	buf := make([]byte, 8192)
	n, err := conn.Read(buf)
	if err != nil {
		t.Fatal(err)
	}
	_, f, err := framing.ParseFrame(buf[:n])
	if err != nil {
		t.Fatal(err)
	}
	if len(f.Payload) == 0 || f.Payload[0] != protocol.MsgTypeSessionCreateAck {
		t.Fatalf("expected SESSION_CREATE_ACK, got first byte %v", f.Payload)
	}
}
