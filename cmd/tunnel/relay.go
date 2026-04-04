package main

import (
	"context"
	"crypto/tls"
	"flag"
	"fmt"

	"tunnel/pkg/relay"
)

func runRelay(args []string) error {
	fs := flag.NewFlagSet("relay", flag.ContinueOnError)
	listen := fs.String("listen", "", "listen address host:port (required)")
	certPath := fs.String("cert", "", "TLS certificate PEM path (required)")
	keyPath := fs.String("key", "", "TLS private key PEM path (required)")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *listen == "" || *certPath == "" || *keyPath == "" {
		return fmt.Errorf("relay: --listen, --cert, and --key are required")
	}

	cert, err := tls.LoadX509KeyPair(*certPath, *keyPath)
	if err != nil {
		return err
	}
	tlsCfg := &tls.Config{
		Certificates: []tls.Certificate{cert},
		MinVersion:   tls.VersionTLS12,
	}

	srv := &relay.Server{
		ListenAddr: *listen,
		TLSConfig:  tlsCfg,
	}
	if err := srv.Listen(); err != nil {
		return err
	}
	fmt.Println(srv.Addr().String())

	if err := srv.Serve(context.Background()); err != nil {
		return err
	}
	return nil
}
