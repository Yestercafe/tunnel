---
status: passed
phase: 03-routing-streams
verified: 2026-03-29
---

# Phase 3 — Verification

## Goal（来自 ROADMAP）

**广播、私信、双向流、流内序与流间乱序** 在 v1 规范中可解析、可互操作且无与 Phase 1/2 opcode 冲突的数据面定义。

## Requirement coverage

| REQ-ID | Evidence |
|--------|----------|
| ROUTE-01 | `docs/spec/v1/routing-modes.md`（广播专节、`REQ: ROUTE-01`） |
| ROUTE-02 | `docs/spec/v1/routing-modes.md`（单播专节、`REQ: ROUTE-02`） |
| STREAM-01 | `docs/spec/v1/streams-lifecycle.md`（双向流、`REQ: STREAM-01`） |
| STREAM-02 | `docs/spec/v1/streams-lifecycle.md`（`stream_id`、顺序、`REQ: STREAM-02`） |

## Must-haves（计划 frontmatter）

- 数据面 `0x10`–`0x12` 与 `SESSION_*` 无重叠；OPEN/DATA/CLOSE 语义与 `routing-modes.md` 路由前缀 **18 字节** 衔接一致。
- `stream_id` 为 `uint32` BE；连接内唯一；**`0`** 禁止作为有效数据流 ID；流内有序、流间乱序条文存在。
- `docs/spec/v1/README.md` 可发现 `routing-modes.md` 与 `streams-lifecycle.md`。

## Automated

```bash
test -f docs/spec/v1/routing-modes.md
test -f docs/spec/v1/streams-lifecycle.md
rg -q 'REQ: ROUTE-01' docs/spec/v1/routing-modes.md
rg -q 'REQ: ROUTE-02' docs/spec/v1/routing-modes.md
rg -q 'REQ: STREAM-01' docs/spec/v1/streams-lifecycle.md
rg -q 'REQ: STREAM-02' docs/spec/v1/streams-lifecycle.md
rg -q 'routing-modes.md' docs/spec/v1/README.md
rg -q 'streams-lifecycle.md' docs/spec/v1/README.md
go test ./...
```

回归：`go test ./...`（含 `pkg/framing` Phase 1）— 通过。

## Verdict

**passed** — 四份 ROUTE/STREAM 要求均有规范正文、REQ 注释与索引行；与 `frame-layout.md` / `session-create-join.md` / `peer-identity.md` 交叉引用一致。
