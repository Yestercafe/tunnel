# Project Research Summary

**Project:** Tunnel（传输隧道）— v1.1 最小 Relay + Client  
**Domain:** 公网 TLS 字节流上的 v1 中继隧道（TCP+TLS、自定义成帧、同 session 多 peer）  
**Researched:** 2026-04-04  
**Confidence:** **HIGH**（综合四份研究：栈与特性以规范与仓库为准；陷阱为 **MEDIUM**，属工程惯例与模式归纳）

## Executive Summary

Tunnel v1.1 是在 **v1.0 已交付的规范与 `pkg/framing` 基线**上，首次落地**可运行的最小 Relay 与 Client**：在 **TCP+TLS** 连续字节流上切帧，完成 **SESSION_CREATE / SESSION_JOIN**、**JOIN_ACK 后门禁**，以及 **STREAM_DATA** 的**广播与单播**转发。业界同类能力通常由「信令 + 数据面路由」拆分实现；本项目的专家做法是把**传输+成帧、会话/成员、路由、流**按 `docs/spec/v1/` **分层**，在进程内用 **Session Registry** 作为 session / peer / 连接映射的唯一真源，**不**引入 WebSocket、gRPC 或第三方成帧库，**保持零新增模块依赖**（Go 1.22 + 标准库 `net` / `crypto/tls` / `context`）。

推荐路径：**先 `pkg/protocol`（payload 与错误码与规范字段级对齐）**，再 **Client 读循环与状态机**（可与 Relay 控制面并行思想验证，但依赖上协议层应先稳定），然后 **Relay 控制面（Registry + CREATE/JOIN）**，再 **数据面路由（广播不回送、单播查表）**，最后 **`cmd/` + TLS + E2E**。这与 FEATURES 中的依赖链（成帧 → 控制面 → peer 表 → JOIN 门禁 → STREAM_DATA + 路由）及 ARCHITECTURE 的 **Suggested Build Order** 一致。

主要风险在于：**流式 Read 无消息边界**（半包/粘包）、**成帧层状态与 session 层状态混写**、**JOIN 前误转数据面**、**Registry 并发与同一 TLS 连接多 goroutine 写交错**、以及**背压缺失导致内存膨胀**。缓解策略：严格「缓冲 → ParseFrame → ErrNeedMore 再读」；成帧与 `joined`/`peer_id` 分层；Relay/Client 双侧门禁；对同一 `Conn` 的 `Write` 串行化并文档化锁顺序；出站路径采用有界队列或明确降级并避免在持全局锁时阻塞写。测试上必须超越「整帧一次 Read」，覆盖分片读、未 JOIN 负例与 `go test -race`。

## Key Findings

### Recommended Stack

研究结论：**保持 Go 1.22（`go.mod`）与标准库**，不新增第三方依赖。TLS 在进程内用 `crypto/tls` 终止，在 `*tls.Conn` 上读写字节流并交给已有 **`pkg/framing`**（`ParseFrame` / `AppendFrame`）。`context` 用于监听关闭、拨号超时与生命周期；`sync` 保护 Registry；测试/演示可用 `testdata/` PEM 或自签证书。**不要**引入 WebSocket、gRPC、额外 TLS 封装或 v1.1 阶段的可观测性全家桶。

**Core technologies:**

- **Go 1.22 + `net`：** `Listen`/`Dial`、`DialContext`、每连接 `net.Conn` — 与 v1 TCP 字节流模型一致。  
- **`crypto/tls`：** `tls.Server`/`tls.Client`、`tls.Config`、SNI/可信根 — v1 假设 TLS 在边缘终止；最小实现即在应用内终止 TLS。  
- **`pkg/framing`（已有）：** 长度前缀成帧、版本/错误码 — **不在此层解析 SESSION/路由**，仅帧边界与 payload。  
- **`context`：** 取消与超时 — 避免 goroutine 泄漏。

详见 [.planning/research/STACK.md](./STACK.md)。

### Expected Features

**Must have（table stakes）：**

- TLS 上成帧 + **版本门闸**；过大帧/坏版本 → `ERR_FRAME_TOO_LARGE` / `ERR_PROTO_VERSION`  
- **SESSION_CREATE**：分配 `session_id` + `invite_code`  
- **SESSION_JOIN**：按 `session_id` 或 `invite_code` 加入；**JOIN_ACK 分配非 0 的 `peer_id`**  
- **JOIN_ACK 前不得发送带数据面路由前缀的帧**；创建者通常仍需在 **CREATE 后再 JOIN** 才获得可路由 `peer_id`  
- **广播（BROADCAST）** 与 **单播（UNICAST）**；广播**不得回送发送者**  
- **STREAM_DATA + 路由前缀 + stream 字段**（可先 **单 `stream_id`** 策略并文档化）  
- **`PROTOCOL_ERROR`** 与规范 `err_code`；可重复验证（测试或 CLI）

**Should have（差异化 / 规范内可分期）：**

- 双凭证路径、可选 **应用信封**（`pkg/appenvelope`）、完整流 OPEN/CLOSE/FIN、`src_peer_id` 与连接身份校验策略  

**Defer（v2+ / Out of Scope）：**

- 集群/持久化 session、主路径 WebSocket、宣称 E2E 加密、无错误码静默失败  

详见 [.planning/research/FEATURES.md](./FEATURES.md)。

### Architecture Approach

单进程 **Relay**：每 `*tls.Conn` **读循环**（缓冲 → `ParseFrame` → 分派）；**Session Registry**（内存 map + 锁）维护 session ↔ peer ↔ 可写连接；控制面更新 Registry，数据面查表转发。建议新建 **`pkg/protocol`**（控制面/数据面 payload、`PROTOCOL_ERROR`）、**`pkg/relay`**、**`pkg/client`**，**`cmd/tunnel-relay`** / **`cmd/tunnel-client`**；**保持 `pkg/framing` 仅帧级语义**。写路径：**同一连接上的 `Write` 必须串行化**（每连接 mutex 或单写 goroutine + channel）。

**Major components:**

1. **`pkg/framing`（已有）** — 帧解析/编码，不解析 payload 内会话语义  
2. **`pkg/protocol`（建议新建）** — `msg_type`、路由前缀、与规范一致的编解码；Relay 与 Client 共用  
3. **`pkg/relay`** — Registry、每连接状态、路由器（广播/单播）  
4. **`pkg/client`** — 会话状态机、CREATE→JOIN、门禁后发送数据面  

详见 [.planning/research/ARCHITECTURE.md](./ARCHITECTURE.md)。

### Critical Pitfalls

1. **把 TLS `Read` 当消息边界** — 必须持久化读缓冲、循环解析直到 `ErrNeedMore` 解决后再读；集成测用分片写。  
2. **成帧「半包」与 session「未 JOIN」混在同一状态机** — 先交付完整帧，再解析 `msg_type` 与门禁。  
3. **JOIN_ACK 前路由或盲信对端 `src_peer_id`** — 仅在对连接记录有效 JOIN 后处理 `STREAM_*`；以 Registry 为准校验或覆盖 `src_peer_id`。  
4. **Registry 无锁或锁顺序混乱** — 文档化锁顺序；连接关闭必须从表摘除；对 `-race` 敏感。  
5. **背压缺失** — 避免无界队列与持锁内阻塞写；v1.1 至少文档化有界或断开策略。  
6. **TLS 捷径上线** — `InsecureSkipVerify` 仅测试；生产路径校验 SAN/证书链。  
7. **测试与协议脱节** — E2E 需 TCP+TLS、双 peer、负例与半包/粘包，非仅 framing 单测。

详见 [.planning/research/PITFALLS.md](./PITFALLS.md)。

## Implications for Roadmap

基于依赖链与 ARCHITECTURE **Suggested Build Order**，建议路线图按下述顺序组织（可与 PROJECT.md 中 RELAY-IMPL / CLIENT-IMPL / E2E-DEMO 交叉对齐：**协议层先立，再分别充实 Client 与 Relay，最后粘合 TLS 与 E2E**）。

### Phase 1: 协议载荷层（`pkg/protocol`）

**Rationale：** 控制面与数据面编解码可与网络解耦，便于 golden 与 `net.Pipe` 单测；避免在 Relay/Client 中散落魔数。  
**Delivers：** SESSION_* / PROTOCOL_ERROR / STREAM_DATA + 路由前缀（最小子集）的编码与解析，与 `docs/spec/v1/` 字段对齐。  
**Addresses：** FEATURES 中成帧之上的控制面/数据面语义、错误码路径。  
**Avoids：** 在 `pkg/framing` 内解析会话（PITFALLS 反模式 1 / Anti-Pattern 1）。

### Phase 2: Client 核心（`pkg/client`）

**Rationale：** 验证「CREATE →（同连接）JOIN → 门禁 → STREAM_DATA」闭环，对称于 Relay，便于对照测试。  
**Delivers：** 在 `net.Conn` 上的读循环、`SESSION_CREATE`/`JOIN`、JOIN 后发送广播/单播视图。  
**Uses：** `pkg/framing`、`pkg/protocol`。  
**Avoids：** JOIN 前发数据面（Pitfall 3）；半包误判（Pitfall 1–2）。

### Phase 3: Relay 控制面（`pkg/relay`：Registry + SESSION_*）

**Rationale：** 先有正确 session/peer 映射与 JOIN 失败错误路径，再谈转发。  
**Delivers：** 进程内 Registry、CREATE_ACK、JOIN_ACK、`ERR_SESSION_NOT_FOUND` / `ERR_JOIN_DENIED` 等。  
**Implements：** ARCHITECTURE 中 Session Registry、连接状态。  
**Avoids：** 未登记 peer 即参与路由；邀请码与 `session_id` 索引不一致（Integration Gotchas）。

### Phase 4: Relay 数据面（`pkg/relay`：路由与转发）

**Rationale：** 依赖完整成员表与 JOIN 门禁。  
**Delivers：** `STREAM_DATA` 解析、广播（不回送）、单播查表；可选校验 `src_peer_id`。  
**Addresses：** FEATURES 中广播/单播、路由模式。  
**Avoids：** 广播回环；盲信 `src_peer_id`（Anti-Pattern 3）。

### Phase 5: 进程入口与 E2E（`cmd/tunnel-relay`、`cmd/tunnel-client` + 集成测试）

**Rationale：** TLS 与配置最后接合，减少过早绑定；满足「可演示、可回归」。  
**Delivers：** TCP 监听、TLS、双 peer 同 session 广播/单播演示；分片读与负例；`go test -race` 覆盖核心路径。  
**Avoids：** TLS 误配置（Pitfall 6）；测试策略脱节（Pitfall 7）。

### Phase 6（可选）：应用信封深化

**Rationale：** 最小闭环可先无 `HAS_APP_ENVELOPE`；稳定后再接 `pkg/appenvelope`。  
**Delivers：** 与上层请求 id 关联的演示或测试。

### Phase Ordering Rationale

- **依赖：** 成帧（已有）→ 协议语义 → Client/Relay 状态机 → 转发 → TLS + E2E。  
- **分组：** 「编解码」与「运行时路由」分离，符合分层与可测性。  
- **风险：** 前半段消除 payload 与门禁错误；后半段集中 TLS、并发与背压。

### Research Flags

**规划期可能需要深入调研或设计拍板的阶段：**

- **Relay 数据面 / E2E：** 背压策略（有界队列 vs 断开）、慢消费者语义是否与 v1「流内有序」表述完全一致 — 若规范未逐字规定，需在实现或文档中**显式定稿**。  
- **`pkg/protocol`：** `STREAM_DATA` 与路由前缀相对 `streams-lifecycle.md` 的**字节偏移** — 建议对照规范做一次字段级 checklist（不必整阶段 `/gsd-research-phase`，但 PLAN 中应列出验证项）。

**可沿用标准模式、通常不必单独开研究阶段的阶段：**

- **TLS + 标准库：** `crypto/tls` 与测试证书夹具模式成熟。  
- **成帧循环 + `ParseFrame`：** 仓库已有 `pkg/framing` 与一致性测试基线。

## Confidence Assessment

| Area | Confidence | Notes |
|------|------------|-------|
| Stack | **HIGH** | Go 官方文档与仓库 `go.mod`/framing；零第三方为明确决策 |
| Features | **HIGH** | 以 `docs/spec/v1/` 与 FEATURES 调研表为准 |
| Architecture | **HIGH**（组件边界）；**MEDIUM**（goroutine/锁细节为惯例建议） | 与规范及现有包边界一致 |
| Pitfalls | **MEDIUM** | 与规范及 Go 实践一致，部分为行业模式非官方枚举 |

**Overall confidence:** **HIGH**（路线图与栈/特性）；**MEDIUM**（并发与背压的具体数值与策略需在实现时验证）

### Gaps to Address

- **背压与资源上限：** v1.1「最小」下采用何种有界策略、是否在规范或 README 中**写死可观察行为**，需在执行阶段决定。  
- **创建者是否必须显式 JOIN：** 规范叙述以 JOIN_ACK 为门禁；若产品层希望「CREATE 即含隐式 peer」，属规范外扩展，**未**纳入本研究结论。  
- **go1.22.x 补丁号：** 以 [Go Release History](https://go.dev/doc/devel/release.html) 当前 1.22 系列为准，不在此文档写死补丁号。

## Sources

### Primary（HIGH confidence）

- 仓库：`docs/spec/v1/`（`connection-state.md`、`session-state.md`、`session-create-join.md`、`routing-modes.md`、`streams-lifecycle.md`、`errors.md`、`security-assumptions.md`、`join-credentials.md`）  
- 仓库：`pkg/framing`、`pkg/appenvelope`、`go.mod`、`.planning/PROJECT.md`  
- [Go Release History](https://go.dev/doc/devel/release.html)、[crypto/tls - pkg.go.dev](https://pkg.go.dev/crypto/tls)

### Secondary（MEDIUM confidence）

- Go TLS/流式 `Read` 分片行为、背压与锁内 IO — 通用服务模式与社区实践

### 各研究文件

- [.planning/research/STACK.md](./STACK.md)  
- [.planning/research/FEATURES.md](./FEATURES.md)  
- [.planning/research/ARCHITECTURE.md](./ARCHITECTURE.md)  
- [.planning/research/PITFALLS.md](./PITFALLS.md)

---
*Research completed: 2026-04-04*  
*Ready for roadmap: yes*
