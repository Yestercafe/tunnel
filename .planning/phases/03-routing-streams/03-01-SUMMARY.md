---
phase: 03-routing-streams
plan: "01"
subsystem: docs
tags: [tunnel, v1, routing, broadcast, ROUTE-01]

requires:
  - phase: 02-session
    provides: session control opcodes, peer_id semantics
provides:
  - docs/spec/v1/routing-modes.md 广播路由前缀与 Relay 行为
  - README 索引行指向 ROUTE-01
affects: [03-02, 03-03, phase-4-envelope]

tech-stack:
  added: []
  patterns:
    - "数据面 payload：msg_type + 18 字节路由前缀（与 streams-lifecycle 衔接）"

key-files:
  created:
    - docs/spec/v1/routing-modes.md
  modified:
    - docs/spec/v1/README.md

key-decisions:
  - "数据面示例统一使用 msg_type=0x11（STREAM_DATA）承载路由前缀，避免与 0x10 OPEN 混淆"

patterns-established:
  - "routing_mode=0 非法；BROADCAST=1 且 dst_peer_id=0；Relay 不回送发送连接"

requirements-completed: [ROUTE-01]

duration: 15min
completed: 2026-03-29
---

# Phase 3：路由与多路流 — 计划 01 小结

**v1 数据面广播路由前缀（`routing_mode`/`src`/`dst`）与 Relay 不回送规则已写入 `routing-modes.md`，README 可发现。**

## Performance

- **Duration:** ~15 min
- **Started:** 2026-03-29
- **Completed:** 2026-03-29
- **Tasks:** 2
- **Files modified:** 2

## Accomplishments

- 新建 `routing-modes.md`：广播专节、偏移表、完整帧示例、`REQ: ROUTE-01`
- 更新 `docs/spec/v1/README.md` 索引

## Task Commits

1. **Task 1: 新建 routing-modes.md（仅广播）** — `a1a73ab` (docs)
2. **Task 2: 更新 README 索引** — `300b1c1` (docs)

## Files Created/Modified

- `docs/spec/v1/routing-modes.md` — ROUTE-01 广播与路由前缀
- `docs/spec/v1/README.md` — 索引

## Decisions Made

与计划一致：保留 `0x01`–`0x04` 给 SESSION_*，`0x10`–`0x12` 流生命周期；示例用 `0x11` + 路由前缀 + `stream_id` 占位衔接后续文档。

## Deviations from Plan

None - plan executed as written

## Self-Check: PASSED

- `test -f docs/spec/v1/routing-modes.md`
- `go test ./...` ok

## Issues Encountered

None

## User Setup Required

None

## Next Phase Readiness

可在同文件扩展 ROUTE-02（单播），并撰写 `streams-lifecycle.md`（STREAM-01/02）。

---
*Phase: 03-routing-streams*
*Completed: 2026-03-29*
