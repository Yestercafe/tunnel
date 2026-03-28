# Phase 1: 协议基础 — 帧与 TLS 字节流承载 - Discussion Log

> **Audit trail only.** Do not use as input to planning, research, or execution agents.  
> Decisions are captured in `01-CONTEXT.md`.

**Date:** 2026-03-29  
**Phase:** 01-tls  
**Modes:** 非交互补全（用户未逐项选择灰区；由编排者根据 RESEARCH + REQUIREMENTS 写入默认决策）

---

## 会话说明

1. **已有计划：** Phase 1 在运行本讨论前已有 `01-01`～`01-03` PLAN。按工作流，用户应选择「先补 CONTEXT 再 **replan**」或取消。本次假定意图为 **补全 CONTEXT 并后续可 replan**。  
2. **灰区讨论：** 未进行多轮 conversational Q&A；**Implementation Decisions** 以 `01-RESEARCH.md` 与项目约束为准，在 `01-CONTEXT.md` 中列为 **D-01～D-10**。  
3. **用户可修订：** 若需魔数、总长语义、最大帧长等变更，请直接编辑 `01-CONTEXT.md` 并运行 `/gsd-plan-phase 1`（可加 `--skip-research`）刷新计划。

---

## 帧布局与长度（对应灰区：长度语义、最大帧）

| Option | Description | Selected |
|--------|-------------|----------|
| Payload 长度 | `uint32` 表示固定头之后 payload | ✓ |
| 总长度 | 含头总长 |  |
| 最大 16MiB | 与 RESEARCH 一致 | ✓ |

**User's choice:** Payload 长度 + 16MiB 上限（见 CONTEXT D-03、D-04）  
**Notes:** 用户自述技术细节可暂缓，采用工程常见默认。

---

## 版本与 Capability（对应灰区：位宽、未知位）

| Option | Description | Selected |
|--------|-------------|----------|
| uint16 version + uint32 cap | 与 RESEARCH 一致 | ✓ |
| 未知 capability 位忽略 | forward compat | ✓ |

**User's choice:** 见 CONTEXT D-06～D-08  

---

## 文档结构

| Option | Description | Selected |
|--------|-------------|----------|
| 三文件 + README | 对齐已有 PLAN | ✓ |

---

## Claude's Discretion

- 固定头字段偏移、具体示例十六进制、是否增加魔数 — 见 CONTEXT「Claude's Discretion」。

## Deferred Ideas

- 浏览器承载、会话语义、E2E — 已记入 CONTEXT `<deferred>`  
