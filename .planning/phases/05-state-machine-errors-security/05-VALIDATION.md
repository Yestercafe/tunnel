---
phase: 5
slug: state-machine-errors-security
status: draft
nyquist_compliant: false
wave_0_complete: false
created: 2026-03-29
---

# Phase 5 — Validation Strategy

> Per-phase validation contract for feedback sampling during execution.

---

## Test Infrastructure

| Property | Value |
|----------|-------|
| **Framework** | go test |
| **Config file** | none — `go.mod` at repo root |
| **Quick run command** | `go test ./...` |
| **Full suite command** | `go test ./... -count=1` |
| **Estimated runtime** | ~5 seconds |

---

## Sampling Rate

- **After every task commit:** Run `go test ./...`
- **After every plan wave:** Run `go test ./... -count=1`
- **Before `/gsd-verify-work`:** Full suite must be green
- **Max feedback latency:** 30 seconds

---

## Per-Task Verification Map

| Task ID | Plan | Wave | Requirement | Test Type | Automated Command | File Exists | Status |
|---------|------|------|-------------|-----------|-------------------|-------------|--------|
| 05-01-* | 01 | 1 | STATE-01 | docs + optional rg | `rg 'STATE-01|connection-state|session-state' docs/spec/v1/` | ⬜ pending |
| 05-02-* | 02 | 2 | ERR-01 | unit + rg | `go test ./pkg/...` ; `rg 'ERR_' docs/spec/v1/errors.md` | ⬜ pending |
| 05-03-* | 03 | 3 | SEC-01 | rg | `rg 'SEC-01|security-assumptions' docs/spec/v1/` | ⬜ pending |

*Status: ⬜ pending · ✅ green · ❌ red · ⚠️ flaky*

---

## Wave 0 Requirements

- [ ] `docs/spec/v1/errors.md` — stubs for ERR-01 (if not created in wave 1)
- Existing infrastructure: `pkg/framing` covers frame decode errors; Phase 5 extends catalog alignment

*Wave 0: documentation-first phase; primary verification is spec grep + `go test ./...`.*

---

## Manual-Only Verifications

| Behavior | Requirement | Why Manual | Test Instructions |
|----------|-------------|------------|---------------------|
| State diagrams match narrative | STATE-01 | Diagram review | Compare tables in spec to ROADMAP success criteria |

---

## Validation Sign-Off

- [ ] All tasks have `<automated>` verify or Wave 0 dependencies
- [ ] Sampling continuity: no 3 consecutive tasks without automated verify
- [ ] Wave 0 covers all MISSING references
- [ ] No watch-mode flags
- [ ] Feedback latency < 30s
- [ ] `nyquist_compliant: true` set in frontmatter

**Approval:** pending
