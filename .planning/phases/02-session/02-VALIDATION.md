---
phase: 2
slug: session
status: draft
nyquist_compliant: false
wave_0_complete: false
created: 2026-03-29
---

# Phase 2 — Validation Strategy

## Test Infrastructure

| Property | Value |
|----------|-------|
| **Framework** | go test（若引入解析包） |
| **Quick run command** | `go test ./...` |
| **Full suite command** | `go test ./... -count=1` |

## Per-Task Verification Map

| Task | Plan | Wave | Requirement | Check |
|------|------|------|-------------|-------|
| 2-01 | 02-01 | 1 | SESS-01,02 | `test -f docs/spec/v1/session-create-join.md` |
| 2-02 | 02-02 | 2 | SESS-03 | `rg SESS-03 docs/spec/v1/peer-identity.md` |
| 2-03 | 02-03 | 3 | SESS-04 | `rg SESS-04 docs/spec/v1/join-credentials.md` |

## Validation Sign-Off

- [ ] 所有 SESS-* 在文档中有 REQ 注释或章节标题  
- [ ] `nyquist_compliant: true`（执行完成后）  

**Approval:** pending
