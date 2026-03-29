---
phase: 04-optional-app-envelope
plan: "02"
subsystem: api
tags: [json, documentation, copilot, correlation-id]

requires:
  - phase: 04-optional-app-envelope
    provides: 04-01 冻结的 app-envelope 与键名
provides:
  - app-envelope.md 中两则端到端示例（JSON 请求/响应、Copilot 往返）
affects: [phase-05, phase-06]

tech-stack:
  added: []
  patterns: []

key-files:
  created: []
  modified:
    - docs/spec/v1/app-envelope.md

key-decisions:
  - "示例 A 以相同 request_id 配对应答帧"
  - "示例 B 两帧共享 correlation_id、不同 request_id"

patterns-established: []

requirements-completed: [APP-01]

duration: 15min
completed: 2026-03-29
---

# Phase 4 — 计划 04-02 小结

**在 `app-envelope.md` 增补可抄写的端到端示例 A/B（偏移表与十六进制前缀），覆盖 JSON 请求/响应与 Copilot 关联 id 场景**

## Performance

- **Duration:** ~15 min
- **Tasks:** 2
- **Files modified:** 1

## Task Commits

1. **Task 1: 示例 A** — `5e56527` (docs)
2. **Task 2: 示例 B** — `b3f40d5` (docs)

## Accomplishments

- 示例 A：`flags=0x02`、`application_data` 自 @25 起的分解表与 16 字节十六进制片段
- 示例 B：两帧 `UNICAST` / `STREAM_DATA` / `stream_id` 意图说明、共享 `corr-copilot-01` 与分支 `request_id`、整段 `application_data` 可验证十六进制

## Deviations from Plan

None - plan executed exactly as written

## Issues Encountered

None

## Self-Check: PASSED

---
*Phase: 04-optional-app-envelope*
*Completed: 2026-03-29*
