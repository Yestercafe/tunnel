---
phase: "11"
status: passed
verified: 2026-04-14
---

# Phase 11 — 目标验证（VERIFICATION）

## Phase goal

自动化证明**两 Client** 同 Relay、**同 session** 的 **广播**与**单播**；并覆盖 **JOIN_ACK 前数据面**与**非法路由**负例（`PROTOCOL_ERROR` / `ErrCodeRoutingInvalid`）。

## Must-haves

| 要求 | 证据 |
|------|------|
| E2E-01 广播/单播 | `TestRelay_StreamData_Broadcast`, `TestRelay_StreamData_Unicast` |
| E2E-02 非法单播路由 | `TestRelay_StreamData_UnicastMissingDst` |
| E2E-02 JOIN 前数据面 | `TestRelay_StreamData_BeforeJoinAck`（经 `client.UnderlyingTLSConn` 原始写帧） |

## Requirement IDs（PLAN 追溯）

- **E2E-01** — `11-01-PLAN.md`、`REQUIREMENTS.md` Traceability Complete  
- **E2E-02** — `11-01-PLAN.md`、`11-02-PLAN.md`、`REQUIREMENTS.md` Traceability Complete  

## 自动化

- `go test ./... -count=1` — **通过**

## 结论

**status: passed**
