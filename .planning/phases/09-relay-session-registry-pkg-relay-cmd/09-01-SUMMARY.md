---
phase: 09-relay-session-registry-pkg-relay-cmd
plan: "01"
subsystem: relay
tags: [go, tls, framing]

requires: []
provides:
  - pkg/relay Server with TLS listen and per-connection ParseFrame loop
  - cmd/tunnel relay subcommand
affects: [09-02]

tech-stack:
  added: []
  patterns: [read buffer + framing.ParseFrame loop]

key-files:
  created:
    - pkg/relay/doc.go
    - pkg/relay/server.go
    - pkg/relay/conn.go
    - pkg/relay/server_test.go
    - cmd/tunnel/relay.go
  modified:
    - cmd/tunnel/main.go

key-decisions:
  - "Server field ListenAddr (string) vs Addr() method to avoid name clash"

patterns-established:
  - "serveConn accumulates buffer until ParseFrame succeeds or framing error closes connection"

requirements-completed:
  - RLY-01

duration: 0min
completed: 2026-04-04
---

# Phase 9 — Plan 01 Summary

**Delivered `pkg/relay` TLS TCP listener, per-connection framing loop with PROTOCOL_ERROR on frame errors, and `tunnel relay --listen --cert --key`.**

## Performance

- **Tasks:** 4
- **Files modified:** 6

## Accomplishments

- `Server.Listen` / `Serve` accept loop with TLS termination.
- `serveConn` implements `ErrNeedMore` buffering and `writeFramingError` for `ErrProtoVersion` / `ErrFrameTooLarge`.
- Tests `TestServer_ProtoVersionError`, `TestServer_NeedMore` against localhost TLS.

## Task Commits

1. **Wave 1 (RLY-01)** — single integration commit with registry wiring (see repo history).

## Files Created/Modified

- `pkg/relay/*.go` — relay server core and tests
- `cmd/tunnel/main.go`, `cmd/tunnel/relay.go` — CLI relay entry
