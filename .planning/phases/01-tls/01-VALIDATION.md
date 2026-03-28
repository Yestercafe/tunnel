---
phase: 1
slug: tls
status: draft
nyquist_compliant: false
wave_0_complete: false
created: 2026-03-29
---

# Phase 1 — Validation Strategy

> Per-phase validation contract for feedback sampling during execution.

---

## Test Infrastructure

| Property | Value |
|----------|-------|
| **Framework** | go test（Go 1.22+） |
| **Config file** | 无 — 随 `go.mod` 引入 |
| **Quick run command** | `go test ./...` |
| **Full suite command** | `go test ./... -count=1` |
| **Estimated runtime** | ~5–30 秒（视是否含编解码包） |

---

## Sampling Rate

- **After every task commit:** `go test ./...`（若存在 `_test.go`）  
- **After every plan wave:** 同上 + 文档内 REQ-ID 与字段表 grep 检查  
- **Before `/gsd-verify-work`:** 全绿 + 规范章节齐全  
- **Max feedback latency:** 60 秒  

---

## Per-Task Verification Map

| Task ID | Plan | Wave | Requirement | Test Type | Automated Command | File Exists | Status |
|---------|------|------|-------------|-----------|-------------------|-------------|--------|
| 1-01-* | 01-01 | 1 | FRAME-01 | doc + grep | `test -f docs/spec/v1/frame-layout.md` | ⬜ | ⬜ pending |
| 1-02-* | 01-02 | 2 | FRAME-02, FRAME-03 | doc + grep | `rg -q FRAME-02 docs/spec/` | ⬜ | ⬜ pending |
| 1-03-* | 01-03 | 3 | TRANS-01 | doc + grep + optional go test | `rg -q TRANS-01 docs/spec/` | ⬜ | ⬜ pending |

*Status: ⬜ pending · ✅ green · ❌ red · ⚠️ flaky*

---

## Wave 0 Requirements

- [ ] `go.mod` — 若 PLAN 引入 `pkg/framing` 或 `internal/framing`  
- [ ] `docs/spec/v1/` — 规范目录存在  
- [ ] 若尚无 Go 代码：Wave 0 可为「仅文档」；Phase 6 再补测试  

---

## Manual-Only Verifications

| Behavior | Requirement | Why Manual | Test Instructions |
|----------|-------------|------------|-------------------|
| 两实现互操作 | FRAME/TRANS | 本阶段可能仅单仓库 | Phase 6 或 pair 实现对照 |

---

## Validation Sign-Off

- [ ] All tasks have `<automated>` verify or Wave 0 dependencies  
- [ ] Sampling continuity: no 3 consecutive tasks without automated verify  
- [ ] Wave 0 covers all MISSING references  
- [ ] No watch-mode flags  
- [ ] Feedback latency < 60s  
- [ ] `nyquist_compliant: true` set in frontmatter  

**Approval:** pending
