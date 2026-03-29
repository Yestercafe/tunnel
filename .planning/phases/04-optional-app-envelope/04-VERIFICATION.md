---
status: passed
phase: 04-optional-app-envelope
verified: 2026-03-29
---

# Phase 4 — Verification

## Goal（来自 ROADMAP）

**可选应用信封** — content-type、请求/关联 id 与 payload 边界在 v1 规范中可解析；含示例场景与正/反向关联 id 用法。

## Requirement coverage

| REQ-ID | Evidence |
|--------|----------|
| APP-01 | `docs/spec/v1/app-envelope.md`（正文、`<!-- REQ: APP-01 -->`）；`docs/spec/v1/streams-lifecycle.md` 交叉引用；`pkg/appenvelope` 与规范一致的切分 |

## Must-haves（计划 frontmatter）

- `flags` @22：bit0 **FIN**、bit1 **HAS_APP_ENVELOPE**；`HAS_APP_ENVELOPE=1` 时 `application_data` @25 起为 `envelope_len` + `envelope` + `body`。
- v1 互操作 **`envelope`** 为 UTF-8 JSON；键 **`content_type`**、**`request_id`**、**`correlation_id`**；**`envelope_len` ≤ 4096**。
- **至少两个**端到端示例：JSON 请求/响应（**`content_type`** + **`request_id`**）；Copilot 往返（两帧 **`STREAM_DATA`**、共享 **`correlation_id`**、不同 **`request_id`**）；含 **十六进制或偏移表**（自 **@25** / **`application_data`** 起算）。

## Automated

```bash
test -f docs/spec/v1/app-envelope.md
rg -q 'REQ: APP-01' docs/spec/v1/app-envelope.md
rg -q 'HAS_APP_ENVELOPE' docs/spec/v1/streams-lifecycle.md
rg -q '示例 A' docs/spec/v1/app-envelope.md
rg -q '示例 B' docs/spec/v1/app-envelope.md
rg -q 'corr-copilot' docs/spec/v1/app-envelope.md
go test ./...
```

回归：`go test ./...`（含 `pkg/framing`、`pkg/appenvelope`）— 通过。

## Verdict

**passed** — APP-01 规范、`pkg/appenvelope` 与两则端到端示例均满足 Roadmap 与计划 **must_haves**。
