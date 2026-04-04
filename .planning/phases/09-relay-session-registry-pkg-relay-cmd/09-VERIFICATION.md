---
phase: "09"
status: passed
verified: 2026-04-04
---

# Phase 9 — 目标验证（VERIFICATION）

## Phase goal

Relay **监听 TCP**、**TLS 终止**、每连接 **成帧循环**；进程内 **Session Registry** 处理 **SESSION_CREATE** / **SESSION_JOIN** 并分配 **`peer_id`**。

## Must-haves（对照 ROADMAP / PLAN）

| 要求 | 证据 |
|------|------|
| RLY-01 监听、TLS、读缓冲、`ParseFrame`/`ErrNeedMore` | `pkg/relay/conn.go`；`pkg/relay/server_test.go` `TestServer_*` |
| RLY-02 Registry、CREATE/JOIN、`peer_id` | `pkg/relay/registry.go`；`pkg/relay/control.go`；`pkg/relay/relay_test.go` `TestRelay_ClientCreateJoin` |
| JOIN 前数据面门禁（PROT-02） | `pkg/relay/control.go` `JoinGateAllowsBusinessDataPlane` |
| 数据面转发非本阶段 | `control.go` 对 JOIN 后 STREAM 返回 phase 10 占位 `PROTOCOL_ERROR` |

## 自动化

- `go test ./... -count=1` — **通过**
- `go test ./pkg/relay/... -count=1` — **通过**
- `go build ./cmd/tunnel` — **通过**

## 结论

**status: passed** — Phase 9 目标已达成（以 `go test` 为准）。
