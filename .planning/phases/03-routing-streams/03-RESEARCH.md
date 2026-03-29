# Phase 3 — Technical Research

**Phase:** 3 — 路由与多路流（routing-streams）  
**Researched:** 2026-03-29  
**Requirement IDs:** ROUTE-01, ROUTE-02, STREAM-01, STREAM-02

## User Constraints

**None** — 本阶段无 discuss-phase 的 `CONTEXT.md`；规划与撰文须**对齐**现有 `docs/spec/v1/`（帧头、版本/capability、传输绑定、session/加入、peer、join 凭证）以及 `.planning/REQUIREMENTS.md` 中 **ROUTE-01、ROUTE-02、STREAM-01、STREAM-02** 的条文。

## Objective

回答：**为在单条 TLS 字节流上实现广播、私信、双向流，并在规范中无歧义地定义流 ID 与「流内有序、流间乱序」**，规划阶段需要预先锁定哪些**数据面布局、枚举、生命周期与文档切分**，以便与 Phase 1 帧头、Phase 2 会话/peer 语义一致地撰写 PLAN 与 v1 规范。

## Findings

### 广播（ROUTE-01）

- **语义**：同一 **session** 内，Relay 将一帧投递给**除发送者外**的**所有已加入成员**；规范 MUST 写明「**不**回送给发送连接」——这是 **Relay 行为**，不是仅靠字段能表达的，但字段应能表达「意图为广播」以便实现与抓包一致。
- **建议帧级（payload 前缀）字段**（与 `peer-identity.md` 预告一致，放在**数据面** payload 内、固定顺序）：
  - `routing_mode`：`uint8` 枚举；**`BROADCAST = 1`**（示例；`0` 保留为「未指定/非法」由规范定死）。
  - `src_peer_id`：`uint64` BE，**发送方** peer（对 client 发出帧：即本连接已分配 id；Relay 向其他 peer 转发时可**原样携带**以便接收方识别来源）。
  - `dst_peer_id`：`uint64` BE；在广播语义下 **MUST 为 `0`**（表示「无单一目标」），避免与单播混淆。
- **文档落点**：独立小节说明「`routing_mode=BROADCAST` 且 `dst_peer_id=0`」的十六进制示例帧（整帧 = 10 字节头 + payload），并引用 **ROUTE-01**。

### 私信 / 单播（ROUTE-02）

- **语义**：`dst_peer_id` 为**会话内**某一非零 peer；**仅**该 peer 的前向连接收到（Relay 路由）；发送者通常可从连接上下文得知，但仍建议在负载前缀带 **`src_peer_id`** 以便日志与多跳扩展一致。
- **建议字段**：
  - `routing_mode`：**`UNICAST = 2`**。
  - `dst_peer_id`：**非零** `uint64` BE，且 **MUST NOT** 等于发送者自身 `peer_id`（若相等 → 协议错误或静默丢弃，由规范二选一并与 Phase 5 ERR 对齐）。
- **与广播的互斥**：由 `routing_mode` + `dst_peer_id==0` vs 非零 联合判定，避免仅看一位导致实现分歧。
- **文档落点**：**ROUTE-02** 示例（单播一帧 payload 布局 + 至少一段十六进制）。

### 双向流（STREAM-01）

- **语义**：在**单条 peer↔Relay TLS 连接**上，允许多个**逻辑流**各自持续收发；**双向**指：同一 `stream_id` 上 client 与 Relay（及对端 peer）可交替发送 **DATA** 类帧，直至**显式或半显式**关闭。
- **建议生命周期**（规划可三选一或组合，须在规范中只选一种以免实现分裂）：
  1. **显式 OPEN**：`msg_type = STREAM_OPEN`，体中含 `stream_id`、可选元数据长度；之后 `STREAM_DATA`。
  2. **隐式打开**：首个携带**新** `stream_id` 的 `STREAM_DATA` 即视为打开（实现简单，规范 MUST 定义「新」的判定：此前未在该连接上关闭过该 id，或 id 自增永不复用）。
  3. **关闭**：`msg_type = STREAM_CLOSE` **或** `STREAM_DATA` + **`FIN` 标志位**（单 bit 在 flags 字节）；规范 MUST 定义 half-close vs 全关闭是否支持。
- **与 Phase 2 关系**：流属于 **已加入 session 之后**的数据面；**不**在 `SESSION_*` 控制消息中携带 `stream_id`（避免与控制面混淆）。
- **文档落点**：`STREAM-01` 绑定到「数据面 opcode 表 + 状态表（谁可发、何时关）」；若仅文档、可先 **文字状态机** + 与 Phase 5 大状态机交叉引用占位。

### 流 ID（STREAM-02）

- **类型建议**：**`stream_id`：`uint32` BE**（与 `uint64` peer 区分清晰；2³²-1 个逻辑流 per 连接足够 v1）。若担心扩展，可预留 `uint64`，但会增加每条数据前缀长度；**规划阶段优先 uint32** 并在规范注明「仅在本连接内唯一」。
- **分配规则**（须选一种写死）：
  - **连接级唯一**：发送方为每个新流分配未使用 `stream_id`（实现可用单调递增）；**对端**在该连接上引用同一 id 即视为同一条流。
  - **禁止**：`stream_id = 0` 作为数据流（建议保留给「无流/控制」或禁止出现在 DATA 中，由规范二选一）。
- **REQ**：**STREAM-02** 在条文中明确：**顺序保证仅针对同一 `(connection, stream_id)` 上的帧序列**。

### 顺序：流内有序 vs 流间乱序（STREAM-02）

- **流内有序**：对同一 `stream_id`，接收方处理 **DATA** 的**应用字节序** MUST 与发送顺序一致（在单连接 TCP+TLS 上通常自然成立；规范仍应写明「若实现重排则违反协议」以防中间层未来引入并行）。
- **流间乱序**：不同 `stream_id` 的帧**可能**以任意交错顺序到达；接收方 **MUST NOT** 假设 `stream_id=1` 的某帧早于 `stream_id=2` 的某帧被处理。
- **与 TRANS-01**：TLS 字节流上帧顺序即线序；**乱序**指**逻辑多路**维度，不是 TCP 乱序；避免读者误解。
- **文字用例（成功标准 3）**：规划至少收录一则**非实现**场景（例如：peer A 同时向 peer B 打开 `stream_id=3` 传文件块、`stream_id=4` 传心跳；B 侧观测到两类 DATA 帧交错到达，但各自流内字节序保持不变）。

### 与现有帧头、Phase 2 session/peer 文档的衔接

- **固定 10 字节帧头**（`frame-layout.md`）：Phase 3 **不**修改帧头；**路由与流字段全部在 `payload`** 中，紧跟 Phase 2 已采用的 **`msg_type`（uint8）** 模式时，需区分：
  - **控制面**（已定义：`SESSION_*`）继续 `msg_type` + 体。
  - **数据面**（本阶段）：建议 **`msg_type` 新取值**（如 `0x10` DATA、`0x11` STREAM_OPEN…）后接**统一「路由+流」前缀**，再接应用数据；**或**单一 `DATA` opcode，由前缀 `routing_mode` 区分广播/单播。
- **`peer_id`**（`peer-identity.md`）：帧内 **`src_peer_id` / `dst_peer_id`** 均为 **`uint64` BE**；与 `SESSION_JOIN_ACK` 中 peer 分配一致；**0** 保留语义在路由中用于「无 dst（广播）」时已占用，**不得**与「有效 peer 0」混淆（Phase 2 已规定成功分配非 0）。
- **`session_id`**：路由帧**不**必每帧重复携带（session 已由加入流程绑定连接）；若未来多 session 复连，属 v2+，v1 保持单连接单 session。
- **capability**（`version-capability.md`）：若某路由特性需「必须理解」，应预留 **capability 位**并在 Phase 3 文档声明；当前 v1 可全部为 **数据面默认可用**，不强制新 capability。

### 建议新增的 `docs/spec/v1/` 文件与 README 索引

| 文件（建议） | 主要内容 | REQ 标记 |
|--------------|----------|----------|
| `routing-modes.md` | `routing_mode` 枚举、`src_peer_id`/`dst_peer_id` 规则；广播与单播字段表；各至少一帧十六进制示例 | ROUTE-01, ROUTE-02 |
| `streams-lifecycle.md` | `stream_id` 宽度与唯一性范围；OPEN/DATA/FIN/CLOSE 的 opcode 与标志；双向流语义；流内/流间顺序；并发多流文字用例 | STREAM-01, STREAM-02 |
| （可选合并） | 若希望减少文件数：可将上述合并为 `routing-and-streams.md`，但需两簇 REQ 标记分段清晰 | 同上 |

**`docs/spec/v1/README.md`**：须增加上表路径与状态列，与 Phase 1/2 表格风格一致。

### 规划者检查清单（供 PLAN.md 使用）

1. 数据面 `msg_type` 取值表与 Phase 2 已占用 opcode **无冲突**（`0x01`–`0x04` 已用；从 `0x10` 起定义数据面为宜）。  
2. 每一类投递（广播/单播）均有**字段级**说明 + **完整帧**示例（含 10 字节头）。  
3. `stream_id`、`routing_mode`、`src`/`dst` 在 payload 中的**偏移与长度表**单页可grep。  
4. **Relay 行为**（广播不回送、未知 peer、非成员）在规范中有**可执行**句子；错误码细节可指向 Phase 5 占位名。  
5. 成功标准中的**并发多流用例**为独立小节，不依赖 Go 代码。

## Validation Architecture

- **REQ 标记（grep）**：在 `docs/spec/v1/` 内可检索 `<!-- REQ: ROUTE-01 -->`、`ROUTE-02`、`STREAM-01`、`STREAM-02`（或与 Phase 1 一致的 `**` 尾注风格，但须与既有 `frame-layout.md` 注释风格统一）。  
- **文档路径**：`rg 'ROUTE-0|STREAM-0'` 于 `docs/spec/v1/`；`README.md` 索引列出新文与 REQ 对应关系。  
- **代码**：Phase 3 若以规范为主，可不强制改 `pkg/framing`；若增加数据面解析示例或 `testdata`，则运行 **`go test ./...`** 作为回归。  
- **Nyquist / 可追溯**：`.planning/REQUIREMENTS.md` 中 Phase 3 四条在规范落点后，将状态从 Pending 改为 Complete（在 execute 收尾，非本研究文件职责）。

## RESEARCH COMPLETE
