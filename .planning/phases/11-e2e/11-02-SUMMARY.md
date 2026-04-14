---
phase: 11-e2e
plan: "02"
subsystem: testing
tags: [go, relay, client, tls, e2e]

requires:
  - phase: 10
    provides: Relay 数据面 STREAM_DATA 与负例基础
provides:
  - `client.UnderlyingTLSConn()` 供集成测试绕过 API 门禁写原始帧
  - `TestRelay_StreamData_BeforeJoinAck`（E2E-02 JOIN 前数据面）
affects: [phase-11-docs]

tech-stack:
  added: []
  patterns: [raw framed write on TLS for negative E2E]

key-files:
  created: []
  modified:
    - pkg/client/client.go
    - pkg/relay/relay_test.go

key-decisions:
  - "UnderlyingTLSConn 仅用于测试场景文档化，不改变 SendStreamData 行为"

patterns-established:
  - "CreateSession 后、JoinSession 前经 TLS 直接写 STREAM_DATA，断言 Relay PROTOCOL_ERROR"

requirements-completed: [E2E-02]

duration: 15min
completed: 2026-04-14
---

# Phase 11 e2e — Plan 02 Summary

**在真实 `relay.Server` 上补齐 JOIN_ACK 前 `STREAM_DATA` 负例，与 `control.go` 门禁一致（`ErrCodeRoutingInvalid`）。**

## Performance

- **Duration:** ~15 min
- **Tasks:** 2
- **Files modified:** 2

## Accomplishments

- 导出 `UnderlyingTLSConn()` 供测试写入 `framing.AppendFrame` 线格式。
- 新增 `TestRelay_StreamData_BeforeJoinAck`，断言 `PROTOCOL_ERROR` 与 `ERR_ROUTING_INVALID`。

## Task Commits

1. **Task: UnderlyingTLSConn** — `996e496` (feat)
2. **Task: BeforeJoinAck test** — `edd7d60` (test)

## Self-Check: PASSED

- `go test ./pkg/relay/... -run TestRelay_StreamData_BeforeJoinAck -count=1` 通过
- `go test ./... -count=1` 通过
