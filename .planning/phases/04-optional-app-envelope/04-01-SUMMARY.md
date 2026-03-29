---
phase: 04-optional-app-envelope
plan: "01"
subsystem: api
tags: [json, tls, streaming, app-envelope]

requires:
  - phase: 03-routing-streams
    provides: STREAM_DATA layout, flags @22, application_data @25
provides:
  - APP-01 规范 docs/spec/v1/app-envelope.md
  - pkg/appenvelope 字节切分与表驱动测试
  - streams-lifecycle 与 README 交叉引用
affects: [phase-05, phase-06]

tech-stack:
  added: []
  patterns: [optional envelope prefix uint16 BE + JSON envelope + body]

key-files:
  created:
    - docs/spec/v1/app-envelope.md
    - pkg/appenvelope/split.go
    - pkg/appenvelope/split_test.go
  modified:
    - docs/spec/v1/streams-lifecycle.md
    - docs/spec/v1/README.md

key-decisions:
  - "v1 互操作信封体为 UTF-8 JSON；TLV 非必选互操作格式"
  - "flags bit1 = HAS_APP_ENVELOPE；application_data 在置位时为 envelope_len + envelope + body"

patterns-established:
  - "SplitApplicationData(flags, applicationData) 仅做边界，不解析 JSON"

requirements-completed: [APP-01]

duration: 25min
completed: 2026-03-29
---

# Phase 4：可选应用信封 — 计划 04-01 小结

**UTF-8 JSON 应用信封（APP-01）、`HAS_APP_ENVELOPE` 位语义，以及 `pkg/appenvelope` 与规范一致的 `application_data` 切分**

## Performance

- **Duration:** ~25 min
- **Started:** 2026-03-29
- **Completed:** 2026-03-29
- **Tasks:** 3
- **Files modified:** 5

## Accomplishments

- 新增 `app-envelope.md`，冻结 JSON 键子集、`envelope_len` 上限与 Relay 不透明语义
- 更新 `streams-lifecycle.md` 中 `STREAM_DATA` 的 `flags` 与 `application_data` 说明并链至专篇
- 实现 `pkg/appenvelope` 表驱动测试，覆盖 FIN|HAS 与截断/上限错误路径

## Task Commits

1. **Task 1: 新建 app-envelope.md** — `a5d4884` (docs)
2. **Task 2: 修订 streams-lifecycle 与 README** — `34c85d3` (docs)
3. **Task 3: pkg/appenvelope** — `a07bbe1` (feat)

## Files Created/Modified

- `docs/spec/v1/app-envelope.md` — APP-01 正文与 REQ 注释
- `docs/spec/v1/streams-lifecycle.md` — STREAM_DATA flags / application_data
- `docs/spec/v1/README.md` — 索引行
- `pkg/appenvelope/split.go` — `SplitApplicationData` 与位常量
- `pkg/appenvelope/split_test.go` — 行为表驱动测试

## Decisions Made

遵循计划：JSON 为 v1 唯一互操作编码；`envelope_len` 上限 4096；错误哨兵与 `fmt.Errorf` 用于超长信封。

## Deviations from Plan

None - plan executed exactly as written

## Issues Encountered

None

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness

- 示例章节（04-02）可在 `app-envelope.md` 上追加端到端实例
- Phase 5/6 可引用 `pkg/appenvelope` 做向量与实现对齐

## Self-Check: PASSED

---
*Phase: 04-optional-app-envelope*
*Completed: 2026-03-29*
