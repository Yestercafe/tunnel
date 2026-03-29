---
phase: 6
slug: consistency-test-suite
status: draft
nyquist_compliant: false
wave_0_complete: false
created: 2026-03-29
---

# Phase 6 — Validation Strategy

> Per-phase validation contract for feedback sampling during execution.

---

## Test Infrastructure

| Property | Value |
|----------|-------|
| **Framework** | Go `testing`（标准库） |
| **Config file** | `go.mod`（模块 `tunnel`，Go 1.22） |
| **Quick run command** | `go test ./... -count=1` |
| **Full suite command** | `go test ./... -count=1` |
| **Estimated runtime** | ~10 秒（视机器，全仓库当前包很少） |

---

## Sampling Rate

- **After every task commit:** Run `go test ./... -count=1`
- **After every plan wave:** Run `go test ./... -count=1`
- **Before `/gsd-verify-work`:** Full suite must be green
- **Max feedback latency:** 60 seconds

---

## Per-Task Verification Map

| Task ID | Plan | Wave | Requirement | Test Type | Automated Command | File Exists | Status |
|---------|------|------|-------------|-----------|-------------------|-------------|--------|
| 06-01-01 | 01 | 1 | TEST-01 | 约定 / 目录 | `test -f testdata/README.md` | ⬜ W0 | ⬜ pending |
| 06-01-02 | 01 | 1 | TEST-01 | 文档 / CI | `rg -q 'go test' .github/workflows/go.yml` | ⬜ W0 | ⬜ pending |
| 06-02-01 | 02 | 2 | TEST-01 | 单元 | `go test ./pkg/framing/... -count=1` | ✅ | ⬜ pending |
| 06-02-02 | 02 | 2 | TEST-01 | golden | `go test ./pkg/framing/... -count=1 -run Golden` | ✅ | ⬜ pending |
| 06-02-03 | 02 | 2 | TEST-01 | 单元 | `go test ./pkg/appenvelope/... -count=1` | ✅ | ⬜ pending |
| 06-02-04 | 02 | 2 | TEST-01 | 全量 | `go test ./... -count=1` | ✅ | ⬜ pending |

*Status: ⬜ pending · ✅ green · ❌ red · ⚠️ flaky*

---

## Wave 0 Requirements

- [ ] `testdata/` 目录与 `README` 或注释说明命名约定（Plan 06-01）
- [ ] `.github/workflows/go.yml` 存在且调用 `go test`（Plan 06-01）

*Wave 0：基础设施与约定；测试本体在 Wave 2（06-02）补齐。*

---

## Manual-Only Verifications

| Behavior | Requirement | Why Manual | Test Instructions |
|----------|-------------|------------|-------------------|
| — | — | — | All phase behaviors targeted here are automated via `go test`. |

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
