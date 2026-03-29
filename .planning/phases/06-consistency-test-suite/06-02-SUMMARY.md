---
phase: 06-consistency-test-suite
plan: "02"
subsystem: testing
tags: [go, golden, testdata]

requires:
  - phase: 06-consistency-test-suite
    provides: Plan 06-01 testdata 约定与 CI
provides:
  - framing 与 appenvelope 的 testdata 驱动测试
  - routing-modes 广播整帧 golden
affects: []

tech-stack:
  added: []
  patterns: [mustDecodeHexFile 跳过 # 注释；appenvelope TAB 行格式]

key-files:
  created:
    - testdata/framing/*.hex
    - testdata/appenvelope/split_cases.hextxt
  modified:
    - pkg/framing/decode_test.go
    - pkg/framing/errors_test.go
    - pkg/appenvelope/split_test.go
    - testdata/README.md

key-decisions:
  - "广播 golden 十六进制与 routing-modes.md 示例逐字节对齐"

patterns-established:
  - "pkg/framing 测试从 ../../testdata/framing 读入 .hex"

requirements-completed: [TEST-01]

duration: 45min
completed: 2026-03-29
---

# Phase 6 — Plan 02 Summary

**在 06-01 约定下补齐成帧与信封切分的 testdata 驱动用例，并加入 routing-modes 广播整帧 golden，`go test ./... -count=1` 全绿。**

## Performance

- **Duration:** ~45 min
- **Completed:** 2026-03-29
- **Tasks:** 4
- **Files modified:** 多文件（见上）

## Accomplishments

- `pkg/framing`：往返 hex 落盘；`ErrNeedMore` / `ErrFrameTooLarge` / `ErrProtoVersion` 文件用例；`TestGolden_RoutingModes_Broadcast` 对齐 `routing-modes.md` 广播示例
- `pkg/appenvelope`：`split_cases.hextxt` + `TestSplitApplicationData_FileDriven`；`testdata/README` 补充 `.hextxt` 说明
- `errors_test.go`：标注 Phase 6 与 errors.md 无新增码

## Task Commits

（本计划以单次提交交付，任务在提交说明中可追溯。）

## Files Created/Modified

- `testdata/framing/*.hex` — 成帧 golden / 负例
- `testdata/appenvelope/split_cases.hextxt` — 信封切分行用例
- `pkg/framing/decode_test.go`、`pkg/appenvelope/split_test.go`、`pkg/framing/errors_test.go`
- `testdata/README.md`

## Decisions Made

与计划一致；广播帧 hex 经逐字节核对 `src_peer_id = 0xab`。

## Deviations from Plan

None - plan executed exactly as written

## Issues Encountered

广播示例十六进制在初次写入时多写零；已按文档逐字段修正为 64 个十六进制字符（32 字节帧）。

## User Setup Required

None

## Next Phase Readiness

TEST-01 在本里程碑的自动化验证已闭环；后续若新增 `ErrCode` 需同步 `errors_test.go`。

---
*Phase: 06-consistency-test-suite*
*Completed: 2026-03-29*
