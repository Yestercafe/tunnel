---
phase: 05-state-machine-errors-security
plan: "02"
subsystem: spec
tags: [errors, PROTOCOL_ERROR, framing, ERR-01]

requires:
  - phase: "05"
    provides: STATE-01 文档基线
provides:
  - docs/spec/v1/errors.md 权威错误码与 0x05 线格式
  - pkg/framing ErrCode 与规范数值对齐
affects: [SEC-01]

tech-stack:
  added: []
  patterns:
    - "PROTOCOL_ERROR 使用 0x05；数据面 STREAM_OPEN 保持 0x10"

key-files:
  created:
    - docs/spec/v1/errors.md
    - pkg/framing/errors.go
    - pkg/framing/errors_test.go
  modified:
    - docs/spec/v1/session-create-join.md
    - docs/spec/v1/version-capability.md
    - docs/spec/v1/join-credentials.md
    - docs/spec/v1/routing-modes.md
    - docs/spec/v1/app-envelope.md
    - docs/spec/v1/README.md
    - pkg/framing/decode.go

requirements-completed: [ERR-01]

duration: —
completed: 2026-03-29
---

# Phase 5：05-02 小结

**建立 ERR-01 唯一权威 `errors.md`**，固化 **`PROTOCOL_ERROR`（`msg_type = 0x05`）** 载荷，并将 **`pkg/framing.ErrCode`** 与 **`uint16` 表**对齐、表驱动测试锁定。

## Task Commits

1. **errors.md** — `43e8dc6`
2. **规范交叉引用** — `847a5be`
3. **framing ErrCode** — `ad4a568`

## Self-Check: PASSED
