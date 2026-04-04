package relay_test

import (
	"context"
	"crypto/tls"
	"testing"

	"tunnel/internal/fakepeer"
	"tunnel/pkg/client"
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
