---
phase: 09
slug: relay-session-registry-pkg-relay-cmd
status: draft
nyquist_compliant: false
wave_0_complete: true
created: 2026-04-04
---

# Phase 9 — Validation Strategy

> 本阶段验证契约：Go 实现、无独立 Wave 0（沿用仓库既有 `go test` 与 `internal/fakepeer` TLS 模式）。

---

## Test Infrastructure

| Property | Value |
|----------|-------|
| **Framework** | Go `testing`（`go test`） |
| **Config file** | none — 使用 `testdata/*.pem` 或 `internal/fakepeer` TLS 辅助 |
| **Quick run command** | `go test ./pkg/relay/... -count=1` |
| **Full suite command** | `go test ./... -count=1` |
| **Estimated runtime** | ~15–60 seconds（视机器） |

---

## Sampling Rate

- **After every task commit:** Run `go test` on the touched package (see each PLAN task `<automated>`).
- **After every plan wave:** `go test ./pkg/relay/... -count=1`
- **Before `/gsd-verify-work`:** `go test ./... -count=1` must be green
- **Max feedback latency:** 60 seconds (full suite)

---

## Per-Task Verification Map

| Task ID | Plan | Wave | Requirement | Test Type | Automated Command | File Exists | Status |
|---------|------|------|-------------|-----------|-------------------|-------------|--------|
| 09-01-01 | 01 | 1 | RLY-01 | unit | `go test ./pkg/relay -run TestNonExistent -count=1` | ✅ | ⬜ pending |
| 09-01-02 | 01 | 1 | RLY-01 | unit | `go test ./pkg/relay -count=1` | ✅ | ⬜ pending |
| 09-02-01 | 02 | 2 | RLY-02 | integration | `go test ./pkg/relay -count=1` | ✅ | ⬜ pending |

*Status: ⬜ pending · ✅ green · ❌ red · ⚠️ flaky*

---

## Wave 0 Requirements

Existing infrastructure covers phase requirements: **`pkg/framing`**, **`pkg/protocol`**, **`pkg/client`**, **`internal/fakepeer`**. No new framework install.

---

## Manual-Only Verifications

| Behavior | Requirement | Why Manual | Test Instructions |
|----------|-------------|------------|---------------------|
| None | — | — | All phase behaviors targeted by automated `go test` in PLAN tasks. |

---

## Validation Sign-Off

- [ ] All tasks have `<automated>` verify or Wave 0 dependencies
- [ ] Sampling continuity: no 3 consecutive tasks without automated verify
- [ ] Wave 0 covers all MISSING references
- [ ] No watch-mode flags
- [ ] Feedback latency < 60s
- [ ] `nyquist_compliant: true` set in frontmatter

**Approval:** pending
