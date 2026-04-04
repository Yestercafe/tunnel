---
phase: "10"
status: passed
verified: 2026-04-04
---

# Phase 10 — 目标验证（VERIFICATION）

## Phase goal

**JOIN_ACK 之后** 对 **STREAM_DATA** 广播（不回送发送者）与单播（按 `dst_peer_id`）；非法/未 JOIN 与 `errors.md` 一致且可测。

## Must-haves

| 要求 | 证据 |
|------|------|
| 广播不回送发送者 | `TestRelay_StreamData_Broadcast` |
| 单播按 `dst_peer_id` | `TestRelay_StreamData_Unicast` |
| 非法路由可测 | `TestRelay_StreamData_UnicastMissingDst` |
| JOIN 前数据面 | 既有 `JoinGate` + Phase 8 client 测试 |

## 自动化

- `go test ./... -count=1` — **通过**

## 结论

**status: passed**
