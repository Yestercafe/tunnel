---
phase: 8
slug: client-pkg-client-cmd
status: draft
nyquist_compliant: false
wave_0_complete: false
created: 2026-04-04
---

# Phase 8 — Validation Strategy

> Per-phase validation contract for feedback sampling during execution.

---

## Test Infrastructure

| Property | Value |
|----------|-------|
| **Framework** | Go `testing` (stdlib) |
| **Config file** | none — `go test ./...` |
| **Quick run command** | `go test ./pkg/client/... -count=1` |
| **Full suite command** | `go test ./... -count=1` |
| **Estimated runtime** | ~30–120 seconds（视集成测试规模） |

---

## Sampling Rate

- **After every task commit:** Run `go test ./pkg/client/... -count=1`
- **After every plan wave:** Run `go test ./... -count=1`
- **Before `/gsd-verify-work`:** Full suite must be green
- **Max feedback latency:** 120 seconds

---

## Per-Task Verification Map

| Task ID | Plan | Wave | Requirement | Test Type | Automated Command | File Exists | Status |
|---------|------|------|-------------|-----------|-------------------|-------------|--------|
| TBD | TBD | TBD | CLNT-01 | integration | `go test ./pkg/client -run TestClient_CreateSession -count=1` | ⬜ W0 | ⬜ pending |
| TBD | TBD | TBD | CLNT-02 | integration | `go test ./pkg/client -run TestClient_JoinSession -count=1` | ⬜ W0 | ⬜ pending |
| TBD | TBD | TBD | CLNT-03 | integration | `go test ./pkg/client -run TestClient_StreamData -count=1` | ⬜ W0 | ⬜ pending |

*Status: ⬜ pending · ✅ green · ❌ red · ⚠️ flaky*

*Planner 将把具体 task ID / plan 编号填入上表。*

---

## Wave 0 Requirements

- [ ] `pkg/client/*_test.go` — TLS fake + CLNT-01..03 集成路径
- [ ] `testdata/` 或测试内动态证书 — 共享 TLS 夹具
- [ ] `cmd/...` — 可选；规范验收以 `pkg/client` 测试为准（见 08-CONTEXT D-05）

*Existing infrastructure: `pkg/protocol`、`pkg/framing` 已有表驱动测试；本阶段新增 `pkg/client` 与 fake peer。*

---

## Manual-Only Verifications

| Behavior | Requirement | Why Manual | Test Instructions |
|----------|-------------|------------|-------------------|
| CLI 冒烟 | D-05（辅助） | 非 CLNT 规范主路径 | `tunnel client create` / `join` 对 fake 或本地 addr 手跑 |

*Canonical proof for CLNT-01..03 is automated per 08-CONTEXT D-03.*

---

## Validation Sign-Off

- [ ] All tasks have `<automated>` verify or Wave 0 dependencies
- [ ] Sampling continuity: no 3 consecutive tasks without automated verify
- [ ] Wave 0 covers all MISSING references
- [ ] No watch-mode flags
- [ ] Feedback latency < 120s
- [ ] `nyquist_compliant: true` set in frontmatter

**Approval:** pending
