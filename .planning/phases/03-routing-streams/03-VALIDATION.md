---
phase: 3
slug: routing-streams
status: draft
nyquist_compliant: false
wave_0_complete: false
created: 2026-03-29
---

# Phase 3 — Validation Strategy

> Per-phase validation contract for feedback sampling during execution.

---

## Test Infrastructure

| Property | Value |
|----------|-------|
| **Framework** | go test |
| **Config file** | none — existing `go.mod` |
| **Quick run command** | `go test ./...` |
| **Full suite command** | `go test ./...` |
| **Estimated runtime** | ~5 seconds |

---

## Sampling Rate

- **After every task commit:** Run `go test ./...`
- **After every plan wave:** Run `go test ./...`
- **Before `/gsd-verify-work`:** Full suite must be green
- **Max feedback latency:** 30 seconds

---

## Per-Task Verification Map

| Task ID | Plan | Wave | Requirement | Test Type | Automated Command | File Exists | Status |
|---------|------|------|-------------|-----------|-------------------|-------------|--------|
| 03-01-01 | 01 | 1 | ROUTE-01 | doc grep | `rg -q 'ROUTE-01' docs/spec/v1/` | ⬜ pending | ⬜ pending |
| 03-02-01 | 02 | 1 | ROUTE-02 | doc grep | `rg -q 'ROUTE-02' docs/spec/v1/` | ⬜ pending | ⬜ pending |
| 03-03-01 | 03 | 1 | STREAM-01, STREAM-02 | doc grep | `rg -q 'STREAM-01' docs/spec/v1/` | ⬜ pending | ⬜ pending |

*Status: ⬜ pending · ✅ green · ❌ red · ⚠️ flaky*

---

## Wave 0 Requirements

- [ ] Existing `pkg/framing` tests remain green after any parser changes
- [ ] New spec files under `docs/spec/v1/` include `<!-- REQ: ... -->` or equivalent traceability markers consistent with Phase 1

*If none: "Existing infrastructure covers all phase requirements."*

---

## Manual-Only Verifications

| Behavior | Requirement | Why Manual | Test Instructions |
|----------|-------------|------------|-------------------|
| Relay broadcast does not echo to sender | ROUTE-01 | Relay behavior, not single-client grep | Read routing-modes section + scenario narrative |

*If none: "All phase behaviors have automated verification."*

---

## Validation Sign-Off

- [ ] All tasks have `<automated>` verify or Wave 0 dependencies
- [ ] Sampling continuity: no 3 consecutive tasks without automated verify
- [ ] Wave 0 covers all MISSING references
- [ ] No watch-mode flags
- [ ] Feedback latency < 30s
- [ ] `nyquist_compliant: true` set in frontmatter

**Approval:** pending
