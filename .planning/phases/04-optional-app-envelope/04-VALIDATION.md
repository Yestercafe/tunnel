---
phase: 4
slug: optional-app-envelope
status: draft
nyquist_compliant: false
wave_0_complete: false
created: 2026-03-29
---

# Phase 4 — Validation Strategy

> Per-phase validation contract for feedback sampling during execution.

---

## Test Infrastructure

| Property | Value |
|----------|-------|
| **Framework** | go test (stdlib) |
| **Config file** | none — `go.mod` at repo root |
| **Quick run command** | `go test ./...` |
| **Full suite command** | `go test ./...` |
| **Estimated runtime** | ~5–30 seconds |

---

## Sampling Rate

- **After every task commit:** Run `go test ./...`
- **After every plan wave:** Run `go test ./...`
- **Before `/gsd-verify-work`:** Full suite must be green
- **Max feedback latency:** 60 seconds

---

## Per-Task Verification Map

| Task ID | Plan | Wave | Requirement | Test Type | Automated Command | File Exists | Status |
|---------|------|------|-------------|-----------|-------------------|-------------|--------|
| 4-01-01 | 01 | 1 | APP-01 | unit + doc grep | `go test ./...`；`rg 'REQ: APP-01|HAS_APP_ENVELOPE' docs/spec/v1/` | ⬜ W0 | ⬜ pending |
| 4-02-01 | 02 | 2 | APP-01 | doc grep + go test | `rg '示例 A|request_id|application/json|0x02' docs/spec/v1/app-envelope.md`；`go test ./... -count=1` | ⬜ W0 | ⬜ pending |
| 4-02-02 | 02 | 2 | APP-01 | doc grep + go test | `rg '示例 B|correlation_id|corr-copilot|Relay' docs/spec/v1/app-envelope.md`；`go test ./... -count=1` | ⬜ W0 | ⬜ pending |

*Status: ⬜ pending · ✅ green · ❌ red · ⚠ flaky*

---

## Wave 0 Requirements

- [ ] `docs/spec/v1/app-envelope.md` — stubs or full text for APP-01 (created in wave 1 per plans)
- [ ] Existing `pkg/framing` tests remain green after any parser additions

*If none: "Existing infrastructure covers all phase requirements."*

---

## Manual-Only Verifications

| Behavior | Requirement | Why Manual | Test Instructions |
|----------|-------------|------------|-------------------|
| Relay transparency for envelope | APP-01 | Relay behavior is normative text | Read `app-envelope.md` Relay bullet; spot-check against `routing-modes.md` |

*If none: "All phase behaviors have automated verification."*

---

## Validation Sign-Off

- [ ] All tasks have `<automated>` verify or Wave 0 dependencies
- [ ] Sampling continuity: no 3 consecutive tasks without automated verify
- [ ] Wave 0 covers all MISSING references
- [ ] No watch-mode flags
- [ ] Feedback latency < 60s
- [ ] `nyquist_compliant: true` set in frontmatter

**Approval:** pending
