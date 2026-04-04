package main

import (
	"context"
	"crypto/tls"
	"errors"
	"flag"
	"fmt"
	"net"

	"tunnel/pkg/client"
)

func runClientCreate(args []string) error {
	fs := flag.NewFlagSet("create", flag.ContinueOnError)
	addr := fs.String("addr", "", "relay address host:port (required)")
	insecure := fs.Bool("insecure-skip-verify", false, "skip server TLS certificate verification (development only; never silent)")
	timeout := fs.Duration("timeout", 0, "optional overall timeout (e.g. 30s)")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *addr == "" {
		return fmt.Errorf("client create: --addr is required")
	}

	ctx := context.Background()
	if *timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, *timeout)
		defer cancel()
	}

	tlsCfg := tlsConfigForAddr(*addr, *insecure)

	c, err := client.Dial(ctx, *addr, tlsCfg)
	if err != nil {
		return err
	}
	defer c.Close()

	sid, inv, err := c.CreateSession(ctx)
	if err != nil {
		return err
	}
	fmt.Println(sid)
	fmt.Println(inv)
	return nil
}

func runClientJoin(args []string) error {
	fs := flag.NewFlagSet("join", flag.ContinueOnError)
	addr := fs.String("addr", "", "relay address host:port (required)")
	session := fs.String("session", "", "session_id (UUID)")
	invite := fs.String("invite", "", "invite code")
	insecure := fs.Bool("insecure-skip-verify", false, "skip server TLS certificate verification (development only; never silent)")
	timeout := fs.Duration("timeout", 0, "optional overall timeout (e.g. 30s)")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *addr == "" {
		return fmt.Errorf("client join: --addr is required")
	}
	if (*session == "" && *invite == "") || (*session != "" && *invite != "") {
		return errors.New("client join: exactly one of --session or --invite is required")
	}

	ctx := context.Background()
	if *timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, *timeout)
		defer cancel()
	}

	tlsCfg := tlsConfigForAddr(*addr, *insecure)

	c, err := client.Dial(ctx, *addr, tlsCfg)
	if err != nil {
		return err
	}
	defer c.Close()

	var pid uint64
	if *session != "" {
		pid, err = c.JoinSession(ctx, true, *session)
	} else {
		pid, err = c.JoinSession(ctx, false, *invite)
	}
	if err != nil {
		return err
	}
	fmt.Printf("%d\n", pid)
	return nil
}

// tlsConfigForAddr sets RootCAs to the system pool when insecure is false, and only sets
// InsecureSkipVerify when insecure is true. ServerName is the hostname part of addr when not an IP literal.
func tlsConfigForAddr(addr string, insecure bool) *tls.Config {
	cfg := &tls.Config{
		MinVersion: tls.VersionTLS12,
	}
	if insecure {
		cfg.InsecureSkipVerify = true
		return cfg
	}
	// RootCAs nil uses system cert pool (crypto/tls default).
	host, _, err := net.SplitHostPort(addr)
	if err != nil {
		return cfg
	}
	if ip := net.ParseIP(host); ip == nil {
		cfg.ServerName = host
	}
	return cfg
}
