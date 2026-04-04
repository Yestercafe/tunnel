package main

import (
	"fmt"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(1)
	}
	switch os.Args[1] {
	case "relay":
		if err := runRelay(os.Args[2:]); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	case "client":
		if len(os.Args) < 3 {
			usage()
			os.Exit(1)
		}
		switch os.Args[2] {
		case "create":
			if err := runClientCreate(os.Args[3:]); err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}
		case "join":
			if err := runClientJoin(os.Args[3:]); err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}
		default:
			usage()
			os.Exit(1)
		}
	default:
		usage()
		os.Exit(1)
	}
}

func usage() {
	fmt.Fprintln(os.Stderr, "usage: tunnel <relay|client> ...")
	fmt.Fprintln(os.Stderr, "  tunnel relay --listen host:port --cert path --key path")
	fmt.Fprintln(os.Stderr, "  tunnel client create --addr host:port [--insecure-skip-verify] [--timeout duration]")
	fmt.Fprintln(os.Stderr, "  tunnel client join --addr host:port (--session <id> | --invite <code>) [--insecure-skip-verify] [--timeout duration]")
}
