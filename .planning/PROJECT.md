# Tunnel（传输隧道）

## What This Is

面向公网部署的**中继隧道**：多个 **client** 通过同一服务在同一 **session** 内交换数据。创建者**开房**并获得 `session_id`（或邀请码），其它 **peer** 凭此加入。协议优先：先交付**帧格式、状态机、错误码**与**一致性测试**；实现语言为 **Go**，但第一步不强制完成端到端产品实现。

**面向谁：**需要在公网与内网之间、**进程/服务**之间安全转发数据的开发者与小团队（**少量固定 peer**）。v1 **不以 WebSocket 为承载**；浏览器若需接入，留待 **v2+**（例如 WebSocket 适配层或 WebTransport），见 REQUIREMENTS。

## Core Value

在 **TLS 由边缘/服务器终止** 的前提下，用**同一套协议**同时支撑：**广播**、**私信（单播）**、**双向流**、**小消息与大块流**；并通过**可选应用信封**让 Web 前端、Copilot 管道等上层复用，而无需各自定义一套私有帧格式。

## Requirements

### Validated

- ✓ **v1 帧布局、版本/capability、TLS 字节流成帧** — 见 `docs/spec/v1/`；参考实现 `pkg/framing`（Phase 1）
- ✓ **会话创建/加入、peer_id、可选 join token** — `session-create-join.md`、`peer-identity.md`、`join-credentials.md`（Phase 2）
- ✓ **路由（广播/单播）与流生命周期、流内/流间顺序** — `routing-modes.md`、`streams-lifecycle.md`（Phase 3）
- ✓ **可选应用信封（UTF-8 JSON、`HAS_APP_ENVELOPE`、请求/关联 id 边界）** — `app-envelope.md`、`pkg/appenvelope`（Phase 4）

### Active

- [ ] 协议规范：**会话/连接状态机**、**完整错误码目录**（路线图 Phase 2–5）
- [ ] 一致性测试：可机器执行的用例/向量，覆盖规范中的关键行为
- [x] 会话语义：创建者创建 session；成员凭 `session_id`/邀请码加入；**同 session 内默认广播**（除发送者外均收到）；支持**私信/单播**至指定 peer — Phase 2–3 已规范会话成员与路由投递
- [x] 传输语义：**双向流**；**按流（或逻辑通道）内有序**，**流之间允许乱序** — Phase 3 已写入 `streams-lifecycle.md`
- [x] 分层：帧之上**可选应用信封**（如 content-type、请求 id、关联 id）— Phase 4 已写入 `app-envelope.md` 与 `streams-lifecycle.md`
- [ ] 成员与连接：协议内会话与成员逻辑；可配合**短 token**；不将端到端加密作为 v1 必选项
- [x] **传输承载**：v1 规范以 **TLS 之上的字节流（典型为 TCP + TLS）** 为参考路径；**成帧、粘包与流边界**在规范中写清（不依赖 WebSocket）— Phase 1 已文档化

### Out of Scope

- **v1 不要求**：完整生产级服务端/客户端实现（可在规范稳定后再做）
- **v1 不要求**：端到端加密（可后续阶段设计；当前信任模型为 TLS 在边缘/服务器）
- **未承诺**：大规模集群、万级并发 session（当前假设少量固定 peer）

## Context

- 典型路径：sender → 公网中继 → 同 session 内其它 peer；接收侧可将数据接入 Copilot 等再写回，依赖**应用信封**关联请求/响应。
- **传输**：v1 对齐 **TCP + TLS** 上的连续字节流；逻辑帧通过**长度与解析规则**从字节流切分，与是否使用 WebSocket **解耦**。

## Constraints

- **语言**：实现默认 **Go**（规范与测试可先于完整实现落地）。
- **交付顺序**：先**规范 + 一致性测试**，再展开实现与运维化。
- **传输**：v1 **不采用 WebSocket**；主路径为 **TLS（如基于 TCP）上的自定义成帧**。浏览器侧非 v1 目标（见 REQUIREMENTS v2）。
- **安全边界**：机密性与完整性主要依赖 **TLS（边缘/服务器）**；协议负责会话、成员与消息语义。
- **顺序模型**：**流内有序、流间可乱序**（在规范中用流 ID 或逻辑通道精确定义）。

## Key Decisions

| Decision | Rationale | Outcome |
|----------|-----------|---------|
| 协议先行（规范 + 一致性测试） | 降低后期互操作与实现返工 | — Pending |
| 会话模型：创建者开房 + 邀请加入 | 简单可解释，适合固定小团队 | — Pending |
| 默认广播 + 私信 | 兼顾协作广播与点对点控制/回复 | — Pending |
| 流内有序、流间乱序 | 适配多路复用与浏览器/异步 IO | — Pending |
| 可选应用信封 | 让 HTTP 风格与 Copilot 管道共享隧道 | v1：`app-envelope.md`（JSON）、`HAS_APP_ENVELOPE`；Phase 4 |
| TLS 在边缘、协议内会话/成员 + 可选短 token | 与「公网中继」部署方式一致 | — Pending |
| 帧头预留版本与 capability | 为未来扩展（传输、加密策略等）留余地 | — Pending |
| Go 实现 | 静态编译、并发与网络生态成熟 | — Pending |
| v1 承载为 TCP+TLS 字节流，不用 WebSocket | 简化成帧与一致性测试；与 WS message 边界解耦 | — Pending |

## Evolution

本文档在**阶段切换**与**里程碑结束**时更新。

**每次阶段切换**（通过 `/gsd-transition`）：

1. 需求是否失效？→ 移入 Out of Scope 并注明原因  
2. 需求是否已验证？→ 移入 Validated 并注明阶段/版本  
3. 是否出现新需求？→ 加入 Active  
4. 是否有新决策？→ 记入 Key Decisions  
5. 「What This Is」是否仍准确？→ 若漂移则更新  

**每次里程碑**（通过 `/gsd-complete-milestone`）：

1. 全文回顾  
2. 核对 Core Value 是否仍为最高优先级  
3. 审计 Out of Scope 的理由是否仍成立  
4. 用当前状态更新 Context  

---

*Last updated: 2026-03-29 after Phase 4 execution*
