---
phase: "08"
status: passed
verified: 2026-04-04
---

# Phase 8 — 目标验证（VERIFICATION）

## Phase goal

Client 在 **TCP+TLS** 上完成 **SESSION_CREATE**、**SESSION_JOIN**，在 **JOIN_ACK** 之后收发带路由前缀的 **STREAM_DATA**；至少各有一条可重复的**广播**与**单播**路径；`stream_id` 策略已文档化。

## Must-haves（对照 ROADMAP / PLAN）

| 要求 | 证据 |
|------|------|
| CLNT-01 TCP+TLS、SESSION_CREATE、`session_id` + `invite_code` | `pkg/client/client_test.go` `TestClient_CreateSession`；`internal/fakepeer` 应答 `EncodeSessionCreateAck` |
| CLNT-02 SESSION_JOIN、JOIN_ACK 后非 0 `peer_id` | `TestClient_JoinSession` |
| CLNT-03 JOIN 后 STREAM_DATA 广播/单播 | `TestClient_StreamData`；`docs/client-stream-ids.md`（stream_id 1/2） |
| PROT-02 发送路径 JoinGate | `TestClient_SendStreamData_NotJoined`；`grep JoinGateAllowsBusinessDataPlane pkg/client/client.go` |
| 无 `package relay` 假实现 | `internal/fakepeer` 使用 `package fakepeer` |
| CLI 冒烟（辅助） | `cmd/tunnel` `client create` / `client join`，`--insecure-skip-verify` 显式 |

## 自动化

- `go test ./... -count=1` — **通过**
- `go test ./pkg/client/... -count=1` — **通过**
- `go test ./internal/fakepeer/... -count=1` — **通过**
- `go build ./cmd/tunnel` — **通过**

## 结论

**status: passed** — Phase 8 目标已达成（规范验收以 `go test` 为准；CLI 为辅助）。
