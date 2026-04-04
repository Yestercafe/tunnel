# Phase 8: Client（pkg/client + cmd）- Discussion Log

> **Audit trail only.** Do not use as input to planning, research, or execution agents.  
> Decisions are captured in `08-CONTEXT.md` — this log preserves the session narrative.

**Date:** 2026-04-04  
**Phase:** 8 — Client（pkg/client + cmd）  
**Areas discussed:** E2E 叙事与阶段分工、集成与验证策略、CLI、TLS、`pkg/client` API、测试风格（委托决策）

---

## Session summary

用户角色为**产品经理**，声明**只关心 E2E 效果**，并将 Phase 8 所列工程灰区（集成策略、CLI 形态、TLS、API 与错误模型等）**全部委托给实现方决定**。

实现方据此采用与 **`08-CONTEXT.md`** 一致的锁定结论；未使用交互式逐项选项表（无逐项 A/B 点击记录）。

---

## Delegated scope

| Topic | User stance | Captured in CONTEXT |
|-------|-------------|---------------------|
| E2E vs 本阶段验收 | 关注最终 E2E；本阶段细节不介入 | D-01, D-02，`domain` |
| 无 Relay 时如何验证 | 委托 | D-03, D-04 |
| CLI | 委托 | D-05, D-06 |
| TLS | 委托 | D-07, D-08 |
| 公共 API / 错误 | 委托 | D-09, D-10, D-11 |
| 测试库与风格 | 委托 | D-12 |
| 剩余细节 | 委托（Claude's Discretion） | `08-CONTEXT.md` § Claude's Discretion |

---

## Claude's Discretion

- 用户未指定子命令命名、目录结构、fake 包名；由 **plan-phase / 执行** 在 **`08-CONTEXT.md`** 约束内确定。

---

## Deferred Ideas

- 完整跨进程 E2E（两 Client + 真 Relay）— **Phase 11**，见 `08-CONTEXT.md` `<deferred>`。
