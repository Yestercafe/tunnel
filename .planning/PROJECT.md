# Tunnel（传输隧道）

## What This Is

面向公网部署的**中继隧道**：多个 **client** 通过同一服务在同一 **session** 内交换数据。创建者**开房**并获得 `session_id`（或邀请码），其它 **peer** 凭此加入。实现语言为 **Go**。**v1.0** 交付规范与一致性测试；**v1.1（2026-04-14）** 交付**最小可运行** Relay 与 Client（`pkg/relay`、`pkg/client`、`cmd/tunnel`），含同 session 广播/单播与 E2E 负例的 `go test` 路径（非完整生产级部署）。

**面向谁：**需要在公网与内网之间、**进程/服务**之间安全转发数据的开发者与小团队（**少量固定 peer**）。v1 **不以 WebSocket 为承载**；浏览器若需接入，留待 **v2+**（例如 WebSocket 适配层或 WebTransport）；延后需求见各里程碑归档中的 **REQUIREMENTS** 章节。

## Core Value

在 **TLS 由边缘/服务器终止** 的前提下，用**同一套协议**同时支撑：**广播**、**私信（单播）**、**双向流**、**小消息与大块流**；并通过**可选应用信封**让 Web 前端、Copilot 管道等上层复用，而无需各自定义一套私有帧格式。

## Current State（截至 v1.1 交付）

- **v1.0（2026-04-04）：** 规范树、`pkg/framing`、`pkg/appenvelope`、`testdata/`、CI。  
- **v1.1（2026-04-14）：** TCP+TLS 上 **Relay**（监听、Registry、CREATE/JOIN、JOIN 后 `STREAM_DATA` 广播/单播、非法路由与 JOIN 前数据面 `PROTOCOL_ERROR`）；**Client**（Dial、会话、流数据）；**E2E** 见 `pkg/relay/relay_test.go` 与归档需求追溯。  
- **代码规模：** 仓库内 Go 约 **3300+** 行（含测试；粗略 `wc`，随提交变化）。  
- **规划工件：** 当前里程碑需求/路线图已归档至 `.planning/milestones/v1.1-*.md`；根目录 `REQUIREMENTS.md` 将在下一里程碑由 `/gsd-new-milestone` 重建。

## Next Milestone Goals

- 由 **`/gsd-new-milestone`** 收集目标 → 新建 `REQUIREMENTS.md`、更新 `ROADMAP.md`、延续阶段编号（不从 `01` 重置）。  
- （占位）可能的主题方向仍见各归档中的 **v2+ / Out of Scope**（如 WebSocket 承载、运维与规模化）— 以新一轮讨论为准。

## Requirements

### Validated

- ✓ **v1 帧布局、版本/capability、TLS 字节流成帧** — 见 `docs/spec/v1/`；参考实现 `pkg/framing`（Phase 1）
- ✓ **会话创建/加入、peer_id、可选 join token** — `session-create-join.md`、`peer-identity.md`、`join-credentials.md`（Phase 2）
- ✓ **路由（广播/单播）与流生命周期、流内/流间顺序** — `routing-modes.md`、`streams-lifecycle.md`（Phase 3）
- ✓ **可选应用信封（UTF-8 JSON、`HAS_APP_ENVELOPE`、请求/关联 id 边界）** — `app-envelope.md`、`pkg/appenvelope`（Phase 4）
- ✓ **连接级成帧与 session 成员状态** — `connection-state.md`、`session-state.md`（STATE-01，Phase 5）
- ✓ **错误码目录与 `PROTOCOL_ERROR`（`0x05`）** — `errors.md`、`pkg/framing` ErrCode（ERR-01，Phase 5）
- ✓ **TLS 边缘终止与安全假设（v1 无 E2E）** — `security-assumptions.md`（SEC-01，Phase 5）
- ✓ **一致性测试（TEST-01）**：`go test ./...`、CI、`testdata/` golden/负例 — Phase 6
- ✓ **载荷语义层（PROT-01 / PROT-02）** — `pkg/protocol`（Phase 7）
- ✓ **最小 Client（CLNT-01..03）** — `pkg/client` + `internal/fakepeer` + `cmd/tunnel client` — Phase 8
- ✓ **Relay 监听与 Session Registry（RLY-01 / RLY-02）** — `pkg/relay`、`cmd/tunnel relay` — Phase 9
- ✓ **Relay 数据面（RLY-03）** — JOIN 后广播/单播与可测负例 — Phase 10
- ✓ **E2E（E2E-01 / E2E-02）** — `pkg/relay/relay_test.go`；追溯见 `milestones/v1.1-REQUIREMENTS.md` — Phase 11

### Active（下一里程碑）

- （待 `/gsd-new-milestone` 填写）

### Out of Scope

- **仍不要求（除非新里程碑明确纳入）**：完整生产级 Relay（集群、持久化、观测栈、配额与滥用防护等）
- **v1 规范阶段已说明**：完整生产级服务端/客户端曾列为非 v1 必选项；v1.1 仅交付**最小**可运行实现
- **v1 不要求**：端到端加密（可后续阶段设计；当前信任模型为 TLS 在边缘/服务器）
- **未承诺**：大规模集群、万级并发 session（当前假设少量固定 peer）

<details>
<summary>v1.1 里程碑内「实现中」清单（已交付，仅作存档）</summary>

- [x] **RELAY-IMPL：** 最小 Relay — TCP+TLS、CREATE/JOIN、**STREAM_DATA** 广播/单播路由 — Phase 9–10  
- [x] **CLIENT-IMPL：** 最小 Client — 开房/加入、收发广播与单播 — Phase 8  

</details>

## Context

- 典型路径：sender → 公网中继 → 同 session 内其它 peer；接收侧可将数据接入 Copilot 等再写回，依赖**应用信封**关联请求/响应。
- **传输**：v1 对齐 **TCP + TLS** 上的连续字节流；逻辑帧通过**长度与解析规则**从字节流切分。
- **v1.1 已落地**：最小 Relay + Client；浏览器与 WebSocket 承载仍为后续里程碑候选。

## Constraints

- **语言**：实现默认 **Go**（规范与测试可先于完整实现落地）。
- **交付顺序**：先**规范 + 一致性测试**，再展开实现与运维化。
- **传输**：v1 **不采用 WebSocket**；主路径为 **TLS（如基于 TCP）上的自定义成帧**。
- **安全边界**：机密性与完整性主要依赖 **TLS（边缘/服务器）**；协议负责会话、成员与消息语义。
- **顺序模型**：**流内有序、流间可乱序**（在规范中用流 ID 或逻辑通道精确定义）。

## Key Decisions

| Decision | Rationale | Outcome |
|----------|-----------|---------|
| 协议先行（规范 + 一致性测试） | 降低后期互操作与实现返工 | ✓ v1.0：规范目录 + `testdata` + CI |
| 会话模型：创建者开房 + 邀请加入 | 简单可解释，适合固定小团队 | ✓ v1.0：`session-create-join.md` 等 |
| 默认广播 + 私信 | 兼顾协作广播与点对点控制/回复 | ✓ v1.0：`routing-modes.md` |
| 流内有序、流间乱序 | 适配多路复用与浏览器/异步 IO | ✓ v1.0：`streams-lifecycle.md` |
| 可选应用信封 | 让 HTTP 风格与 Copilot 管道共享隧道 | ✓ v1.0：`app-envelope.md`（JSON）、`HAS_APP_ENVELOPE`；Phase 4 |
| TLS 在边缘、协议内会话/成员 + 可选短 token | 与「公网中继」部署方式一致 | ✓ v1.0：`join-credentials.md`、`security-assumptions.md` |
| 帧头预留版本与 capability | 为未来扩展（传输、加密策略等）留余地 | ✓ v1.0：帧头与 capability 语义成文 |
| Go 实现 | 静态编译、并发与网络生态成熟 | ✓ v1.0：`pkg/framing`、`pkg/appenvelope`、测试 |
| v1 承载为 TCP+TLS 字节流，不用 WebSocket | 简化成帧与一致性测试；与 WS message 边界解耦 | ✓ v1.0：TRANS-01 与成帧文档 |
| 进程内 Session Registry（v1.1） | 最小闭环、无外部存储 | ✓ v1.1：`pkg/relay` Registry + CREATE/JOIN |
| JOIN 后数据面路由与可测负例（v1.1） | 与 `errors.md` 一致、CI 可重复 | ✓ v1.1：`relay_test` + `JoinGate` / `PROTOCOL_ERROR` |

## Evolution

本文档在**阶段切换**与**里程碑结束**时更新。

**每次里程碑**（通过 `/gsd-complete-milestone`）：

1. 全文回顾  
2. 核对 Core Value 是否仍为最高优先级  
3. 审计 Out of Scope 的理由是否仍成立  
4. 用当前状态更新 Context 与 Validated  

---

*Last updated: 2026-04-14 — v1.1 里程碑归档完成*
