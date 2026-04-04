---
phase: 7
slug: pkg-protocol
status: draft
nyquist_compliant: false
wave_0_complete: false
created: "2026-04-04"
---

# Phase 7 — Validation Strategy

> Per-phase validation contract for feedback sampling during execution.

---

## Test Infrastructure

| Property | Value |
|----------|-------|
| **Framework** | Go `testing` |
| **Config file** | none |
| **Quick run command** | `go test ./pkg/protocol/... -count=1` |
| **Full suite command** | `go test ./... -count=1` |
| **Estimated runtime** | ~10–30 seconds |

---

## Sampling Rate

- **After every task commit:** Run `go test ./pkg/protocol/... -count=1`
- **After every plan wave:** Run `go test ./... -count=1`
- **Before `/gsd-verify-work`:** Full suite must be green
- **Max feedback latency:** 60 seconds

---

## Per-Task Verification Map

| Task ID | Plan | Wave | Requirement | Test Type | Automated Command | File Exists | Status |
|---------|------|------|-------------|-----------|-------------------|-------------|--------|
| 07-01 T1 | 01 | 1 | PROT-01 | unit | `go test ./pkg/protocol/... -run 'Test.*Session\|Test.*ProtocolError' -count=1` | ❌ W0 | ⬜ pending |
| 07-01 T2 | 01 | 1 | PROT-01 | unit | `go test ./pkg/protocol/... -run 'Stream\|Routing' -count=1` | ❌ W0 | ⬜ pending |
| 07-02 T1 | 02 | 2 | PROT-02 | unit | `go test ./pkg/protocol/... -run JoinGate -count=1` | ❌ W0 | ⬜ pending |
| 07-02 T2 | 02 | 2 | PROT-02 | unit | `go test ./... -count=1` | — | ⬜ pending |

*Status: ⬜ pending · ✅ green · ❌ red · ⚠️ flaky*

---

## Wave 0 Requirements

- [ ] `pkg/protocol/control_test.go` — SESSION_* / PROTOCOL_ERROR 覆盖 PROT-01
- [ ] `pkg/protocol/join_gate_test.go` — PROT-02 门禁矩阵
- [ ] `testdata/protocol/` — 可选 hex 向量（对齐 Phase 6 `testdata/framing/`）

---

## Manual-Only Verifications

| Behavior | Requirement | Why Manual | Test Instructions |
|----------|-------------|------------|---------------------|
| — | — | — | All phase behaviors target automated verification via `go test`. |

---

## Validation Sign-Off

- [ ] All tasks have `<automated>` verify or Wave 0 dependencies
- [ ] Sampling continuity: no 3 consecutive tasks without automated verify
- [ ] Wave 0 covers all MISSING references
- [ ] No watch-mode flags
- [ ] Feedback latency < 60s
- [ ] `nyquist_compliant: true` set in frontmatter

**Approval:** pending
