---
phase: 10
slug: relay
status: draft
nyquist_compliant: false
wave_0_complete: true
created: 2026-04-04
---

# Phase 10 — Validation Strategy

> Go `testing`；沿用仓库既有 TLS（`internal/fakepeer`）与 **`pkg/client`**。

## Test Infrastructure

| Property | Value |
|----------|-------|
| **Framework** | Go `testing` |
| **Quick run command** | `go test ./pkg/relay/... -count=1` |
| **Full suite command** | `go test ./... -count=1` |

## Per-Task Verification Map

| Task ID | Plan | Wave | Requirement | Automated Command | Status |
|---------|------|------|-------------|-------------------|--------|
| 10-01-* | 01 | 1 | RLY-03 | `go test ./pkg/relay -count=1` | ⬜ |
| 10-02-* | 02 | 2 | RLY-03 | `go test ./pkg/relay -run TestRelay_ -count=1` | ⬜ |

## Wave 0

沿用 Phase 9 的 relay 测试基础设施；无新增框架安装。

## Manual-Only Verifications

无 — 行为以自动化测试断言为准。

## Validation Sign-Off

**Approval:** pending
