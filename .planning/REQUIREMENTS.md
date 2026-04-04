# Requirements: Tunnel（传输隧道）

**Defined:** 2026-04-04  
**Milestone:** v1.1 — 最小 Relay 与 Client  
**Core Value:** 在 TLS 由边缘/服务器终止的前提下，用同一套 v1 协议支撑广播、私信、流与可选应用信封；本里程碑交付**可运行的最小** Relay 与 Client。

## v1.1 Requirements（本里程碑）

### 共享协议层（`pkg/protocol` 或等价命名）

- [ ] **PROT-01**：控制面/数据面 payload、`msg_type`、路由前缀与 **`PROTOCOL_ERROR`** 的编解码与 **v1 规范字段级对齐**（与 `pkg/framing` 边界清晰：仅帧级，不解析会话语义）。
- [ ] **PROT-02**：**JOIN_ACK 前**不得将带**数据面路由**的帧视为合法业务路径；该门禁在 Client/Relay 两侧可一致实现或共享同一判定逻辑。

### Client（`pkg/client` + `cmd/...`）

- [ ] **CLNT-01**：Client 可通过 **TCP+TLS** 连接 Relay，并完成 **SESSION_CREATE**，获得 `session_id` 与 `invite_code`（与规范一致）。
- [ ] **CLNT-02**：Client 可 **SESSION_JOIN**（凭 `session_id` 或邀请码），在 **SESSION_JOIN_ACK** 后获得非 0 的 **`peer_id`**。
- [ ] **CLNT-03**：在 **JOIN_ACK 之后**，Client 可发送并接收 **`STREAM_DATA`**（含路由前缀）；至少验证 **广播** 与 **单播** 各一条可重复路径（可先固定 `stream_id` 策略并文档化）。

### Relay（`pkg/relay` + `cmd/...`）

- [ ] **RLY-01**：Relay **监听 TCP** 并以 **TLS** 终止连接；每条连接上维护**读缓冲**与成帧循环（`ParseFrame` / `ErrNeedMore`），与 `pkg/framing` 一致。
- [ ] **RLY-02**：Relay 维护进程内 **Session Registry**（session ↔ peer ↔ 可写连接），处理 **SESSION_CREATE** / **SESSION_JOIN**，并分配 **`peer_id`**。
- [ ] **RLY-03**：在 **JOIN_ACK 之后**，Relay 对 **STREAM_DATA** 执行**广播**（不回送发送者）与**单播**（按 `dst_peer_id` 查表）；对非法路由或未 JOIN 连接的行为与 `errors.md` / 实现注释一致且可测。

### 验证与可重复演示

- [ ] **E2E-01**：**自动化测试**（或同等 CI 可运行脚本）：至少 **两个 Client** 连接同一 Relay、**同一 session**，可验证 **广播** 与 **单播** 各至少一条用例。
- [ ] **E2E-02**：**负例或门禁**：至少覆盖 **JOIN_ACK 前发送数据面帧** 与 **非法/未知路由** 之一（期望 `PROTOCOL_ERROR`、丢弃或断连 — 与实现选定策略一致，并在测试中断言）。

## v2+ Requirements（延后）

### 承载与生态

- **TRANS-02**：WebSocket 等浏览器友好承载（见历史 `PROJECT.md` / 规范中的 v2 方向）。

### 产品化

- 集群与高可用、持久化 session、完整观测栈、配额与滥用防护。

## Out of Scope（本里程碑明确不做）

| 能力 | 原因 |
|------|------|
| 生产级集群 / 持久化 / 万级并发 | v1.1 为最小可运行实现 |
| 端到端加密（非 TLS） | 与 v1 安全假设一致，延后 |
| 规范未要求的静默失败 | 须与 `errors.md` 与测试策略一致 |

## Traceability

由路线图创建阶段时更新：每条需求映射到**恰好一个**阶段。

| Requirement | Phase | Status |
|-------------|-------|--------|
| PROT-01 | — | Pending |
| PROT-02 | — | Pending |
| CLNT-01 | — | Pending |
| CLNT-02 | — | Pending |
| CLNT-03 | — | Pending |
| RLY-01 | — | Pending |
| RLY-02 | — | Pending |
| RLY-03 | — | Pending |
| E2E-01 | — | Pending |
| E2E-02 | — | Pending |

**Coverage:**

- v1.1 requirements: 10 total  
- Mapped to phases: 0（待路线图）  
- Unmapped: 10 ⚠️  

---

*Requirements defined: 2026-04-04*  
*Last updated: 2026-04-04 — milestone v1.1 new-milestone*
