---
phase: 11
slug: e2e
status: draft
nyquist_compliant: false
wave_0_complete: false
created: 2026-04-05
---

# Phase 11 — Validation Strategy

> Per-phase validation contract for feedback sampling during execution.

---

## Test Infrastructure

| Property | Value |
|----------|--------|
| **Framework** | Go `testing`（stdlib） |
| **Config file** | 无独立配置；入口为 `go.mod` |
| **Quick run command** | `go test ./pkg/relay/... -count=1` |
| **Full suite command** | `go test ./... -count=1` |
| **Estimated runtime** | ~1–3 分钟（全仓库） |

---

## Sampling Rate

- **After every task commit:** `go test ./pkg/relay/... -count=1`（或任务涉及的包）
- **After every plan wave:** `go test ./... -count=1`
- **Before `/gsd-verify-work`:** 全绿
- **Max feedback latency:** 180 秒（全量 `./...`）

---

## Per-Task Verification Map

| Task ID | Plan | Wave | Requirement | Test Type | Automated Command | File Exists | Status |
|---------|------|------|---------------|-----------|-------------------|-------------|--------|
| 11-01-01 | 01 | 2 | E2E-01 | 集成 | `go test ./pkg/relay/... -run TestRelay_StreamData_Broadcast -count=1` | ✅ | ⬜ pending |
| 11-01-02 | 01 | 2 | E2E-01 | 集成 | `go test ./pkg/relay/... -run TestRelay_StreamData_Unicast -count=1` | ✅ | ⬜ pending |
| 11-01-03 | 01 | 2 | E2E-02 | 集成 | `go test ./pkg/relay/... -run TestRelay_StreamData_UnicastMissingDst -count=1` | ✅ | ⬜ pending |
| 11-02-01 | 02 | 1 | E2E-02 | 集成 | `go test ./pkg/relay/... -run TestRelay_StreamData_BeforeJoinAck -count=1` | ❌ W0 | ⬜ pending |

*Status: ⬜ pending · ✅ green · ❌ red · ⚠️ flaky*

---

## Wave 0 Requirements

- [ ] 新增 `pkg/relay`（或约定路径）中 **JOIN_ACK 前**发送 `STREAM_DATA` 的集成测试，经 TLS 写到 **真实 `relay.Server`**，断言 `PROTOCOL_ERROR` / `ErrCodeRoutingInvalid`（与 `control.go` 一致）。
- [ ] `REQUIREMENTS.md` Traceability 或 Phase 11 `VERIFICATION.md` 中将 E2E-01 / E2E-02 与测试函数名显式对齐。

*已有基础设施：`fakepeer.LocalhostTLSConfig`、`relay_test` 双 Client 模式。*

---

## Manual-Only Verifications

| Behavior | Requirement | Why Manual | Test Instructions |
|----------|-------------|------------|-------------------|
| — | — | — | 本阶段目标为全自动 `go test`；无手工步骤。 |

*If none: "All phase behaviors have automated verification."*

---

## Validation Sign-Off

- [ ] All tasks have `<automated>` verify or Wave 0 dependencies
- [ ] Sampling continuity: no 3 consecutive tasks without automated verify
- [ ] Wave 0 covers all MISSING references
- [ ] No watch-mode flags
- [ ] Feedback latency < 180s
- [ ] `nyquist_compliant: true` set in frontmatter

**Approval:** pending
