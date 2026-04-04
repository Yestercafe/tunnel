---
phase: "07"
status: passed
verified: 2026-04-04
---

# Phase 7 — 目标验证（VERIFICATION）

## Phase goal

Relay 与 Client 共用的 **v1 载荷语义层**：控制面/数据面 `msg_type`、`PROTOCOL_ERROR` 与路由前缀 **`STREAM_*`** 视图与规范字段对齐；**JOIN_ACK 前** 不得将带数据面路由的 `STREAM_*` 业务路径视为合法（共享 `join_gate`）。

## Must-haves（对照 ROADMAP / PLAN）

| 要求 | 证据 |
|------|------|
| PROT-01 编解码与 `pkg/framing` 仅共享 `ErrCode`、不解析帧头 | `pkg/protocol/control.go`、`streamdata.go`；`rg ParseFrame pkg/protocol` 无匹配 |
| ROUTE / STREAM 偏移与 golden | `pkg/protocol/streamdata_test.go`、`testdata/protocol/stream_data_min.hex` |
| PROT-02 JOIN 前门禁 | `pkg/protocol/join_gate.go`、`join_gate_test.go`（含 `0x05` 与 `0x11` 区分） |
| 全量测试 | `go test ./... -count=1` |

## 自动化

- `go test ./... -count=1` — **通过**
- `go test ./pkg/protocol/... -run JoinGate -count=1` — **通过**

## 人工 / 文档审阅

- 无额外人工项；规范对齐以 spec 与表驱动测试为主。

## 结论

**status: passed** — Phase 7 目标已达成。
