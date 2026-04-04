---
phase: 09-relay-session-registry-pkg-relay-cmd
plan: "02"
subsystem: relay
tags: [go, session, registry]

requires:
  - phase: 09-01
    provides: TLS framing server
provides:
  - SessionRegistry with CreateSession / JoinSession
  - SESSION_CREATE_ACK / SESSION_JOIN_ACK dispatch
  - pkg/client integration test against relay
affects: [Phase 10 RLY-03]

tech-stack:
  added: [github.com/google/uuid]
  patterns: [in-memory session maps by id and invite]

key-files:
  created:
    - pkg/relay/registry.go
    - pkg/relay/control.go
    - pkg/relay/relay_test.go
  modified:
    - pkg/relay/conn.go
    - pkg/relay/server.go

key-decisions:
  - "CREATE then JOIN on same connection for first peer (matches harness semantics)"

patterns-established:
  - "JoinSession registers FrameWriter (tls.Conn) per peer for future RLY-03 routing"

requirements-completed:
  - RLY-02

duration: 0min
completed: 2026-04-04
---

# Phase 9 — Plan 02 Summary

**Delivered process-local Session Registry, SESSION_CREATE / SESSION_JOIN handling with non-zero peer_id, JOIN 前 STREAM_DATA 拒绝，以及 `pkg/client` 对真实 Relay 的集成测试。**

## Performance

- **Tasks:** 3
- **Files modified:** 5

## Accomplishments

- `SessionRegistry` with UUID session_id and random invite codes validated via `protocol.EncodeSessionCreateAck`.
- `dispatchFrame` in `control.go` handles control plane and data-plane gate before JOIN.
- `TestRelay_ClientCreateJoin` asserts two distinct peer_ids on same session.

## Files Created/Modified

- `pkg/relay/registry.go` — registry
- `pkg/relay/control.go` — message dispatch
- `pkg/relay/relay_test.go` — client integration test
