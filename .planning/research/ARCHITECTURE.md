# Architecture Research

**Domain:** 公网 TLS 字节流上的 v1 中继隧道（最小 Relay + Client）  
**Researched:** 2026-04-04  
**Confidence:** HIGH（以仓库内 `docs/spec/v1/`、`pkg/framing`、`pkg/appenvelope` 与 `PROJECT.md` 为准；goroutine 与并发细节为工程惯例，标为 MEDIUM）

---

## Standard Architecture

### System Overview

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                         Relay 进程（单进程、最小实现）                         │
├─────────────────────────────────────────────────────────────────────────────┤
│  TLS 终止（每连接 *tls.Conn）                                                 │
│  ┌──────────────┐    ┌──────────────┐    ┌──────────────────────────────┐   │
│  │ Conn A       │    │ Conn B       │    │ Session Registry（进程内内存） │   │
│  │ 成帧读循环    │    │ 成帧读循环    │◄──►│ session_id / invite → Session │   │
│  │ + 会话状态机  │    │ + 会话状态机  │    │ peer_id → 可写连接句柄         │   │
│  └──────┬───────┘    └──────┬───────┘    └──────────────────────────────┘   │
│         │                   │                          ▲                     │
│         │  控制面：SESSION_* / PROTOCOL_ERROR           │ 数据面：查表并写帧   │
│         │  数据面：STREAM_* + 路由前缀                   │                     │
│         └───────────────────┴──────────────────────────┘                     │
└─────────────────────────────────────────────────────────────────────────────┘
          ▲ TLS                           ▲ TLS
          │                               │
    ┌─────┴─────┐                   ┌─────┴─────┐
    │ Client A  │                   │ Client B  │
    │ 成帧 +     │                   │ 成帧 +     │
    │ 会话/路由  │                   │ 会话/路由  │
    └───────────┘                   └───────────┘
```

**分层（与规范一致，禁止混层）：**

| 层 | 规范锚点 | 职责 |
|----|-----------|------|
| 传输 + 成帧 | `transport-binding.md`、`connection-state.md` | TLS 明文字节流上 **10 字节头 + payload** 切分；半包 / 粘包 / 致命错误 |
| 会话 / 成员 | `session-state.md`、`session-create-join.md` | **控制面** `msg_type` 0x01–0x04、JOIN 门禁 |
| 路由 | `routing-modes.md` | **数据面** 路由前缀（`STREAM_DATA` 与 `routing_mode` / `peer_id`） |
| 流 | `streams-lifecycle.md` | `stream_id`、`STREAM_OPEN` / `DATA` / `CLOSE` |
| 可选应用信封 | `app-envelope.md`、`pkg/appenvelope` | `HAS_APP_ENVELOPE` 时拆分 envelope / body |

### 控制面与数据面（边界）

| 平面 | 典型 payload | 处理方 | 规范 |
|------|----------------|--------|------|
| **控制面** | `SESSION_CREATE_REQ/ACK`、`SESSION_JOIN_REQ/ACK`、`PROTOCOL_ERROR` | 修改或查询 **Session Registry**、分配/校验 `peer_id`、返回错误 | `session-create-join.md`、`errors.md` |
| **数据面** | `STREAM_OPEN` / `STREAM_DATA` / `STREAM_CLOSE` + **路由前缀** | **JOIN_ACK 之后**；Relay 按广播/单播**复制或转发**整帧（或等价语义） | `routing-modes.md`、`streams-lifecycle.md` |

`session-state.md` 要求：**未收到 `SESSION_JOIN_ACK` 且 `peer_id` 有效前，不得发送带数据面路由前缀的帧** — 实现上 Client 与 Relay 两侧都应有门禁。

### Goroutine 模型（最小 Relay + Client）

**原则：** 连接数少、无集群；优先 **清晰可测**，而非过早引入队列/actor 框架。

| 角色 | 建议模型 | 说明 |
|------|-----------|------|
| **Relay 每连接** | **1 个读 goroutine**（主循环：`Read` → 追加缓冲 → `framing.ParseFrame` 直到半包；解析成功后按 `msg_type` 分派） | 成帧层 `ErrNeedMore` 与 session 层「未 JOIN」**分开**（对应 `connection-state.md` / `session-state.md`） |
| **Relay 写路径** | 向**本连接**写：可在读循环内同步写，或 **每连接 1 个写 goroutine + channel**；向**其他 peer** 转发：在 Registry 查到 `net.Conn` 后 **串行化** 对该 `Conn` 的 `Write`（避免多 goroutine 交错写同一 TLS 连接） | 最小实现：全局或 per-conn `sync.Mutex` 保护 `Write` 亦可 |
| **Registry** | 后台无独立 goroutine；由读路径 + 控制面处理在持锁下更新 **`session_id` / `invite_code` ↔ Session**、**`peer_id` ↔ `Conn`** | 关闭连接时 **必须** 从 Registry 摘除，避免向已关闭连接写 |
| **Client** | 与 Relay 对称：**1 读循环** + 受控写路径；若仅单线程演示，可同步读写 | E2E 前可先用 `net.Pipe` 测状态机 |

### Component Responsibilities

| 组件 | 职责 | 典型实现要点 |
|------|------|----------------|
| **`pkg/framing`（已有）** | 单帧解析/编码；`version` / `payload_len` 边界 | `ParseFrame` / `AppendFrame`；**不解析** payload 内 `msg_type`（与 `connection-state.md` 一致） |
| **Payload / 协议语义（建议新建 `pkg/protocol` 或 `internal/protocol`）** | 控制面 REQ/ACK、数据面路由前缀、`PROTOCOL_ERROR` 编解码 | 与 `session-create-join.md`、`routing-modes.md`、`streams-lifecycle.md`、`errors.md` 字段级对齐；Relay 与 Client **共用** |
| **Relay：连接句柄** | 每 TLS 连接维护「是否已 JOIN、`peer_id`、所属 `session`」 | 与 Registry 同步更新 |
| **Relay：Session Registry** | `session_id`、`invite_code` ↔ Session；Session 内 `peer_id` → 可写回连接 | 最小：**进程内** `map` + `sync.RWMutex`；创建时生成 UUID 与 invite |
| **Relay：路由器（数据面）** | 解析 `STREAM_*` + 路由前缀；广播复制、单播查表；**不回送发送者** | `src_peer_id` 以 Registry 登记为准覆盖或校验不一致则拒绝（策略固定并文档化） |
| **Client：会话状态机** | CREATE 或 JOIN；仅在 `SESSION_JOIN_ACK` 后发数据面帧 | 与 `session-state.md` 一致 |
| **`pkg/appenvelope`（已有）** | 可选 | 仅 E2E 需演示信封时接入；最小闭环可先 `flags` 无 `HAS_APP_ENVELOPE` |

---

## Recommended Project Structure

```
tunnel/
├── docs/spec/v1/              # 已有 — 权威语义
├── pkg/framing/               # 已有 — 连接级成帧
├── pkg/appenvelope/           # 已有 — STREAM_DATA 可选信封边界
├── pkg/protocol/              # 建议新增 — 控制面/数据面 payload 编解码 + 常量
├── pkg/relay/                 # 建议新增 — Session、Registry、Router、每连接 Run（可单测）
├── pkg/client/                # 建议新增 — 连接、会话 API、发送广播/单播
├── cmd/tunnel-relay/          # 建议新增 — 监听、TLS、main
├── cmd/tunnel-client/         # 建议新增 — CLI 或示例（E2E-DEMO）
└── testdata/                  # 已有 — 可扩展 E2E 向量
```

### Structure Rationale

- **`pkg/framing` 保持「只懂帧头 + payload 切片」**：不把 SESSION 解析塞进 `framing`，符合 `connection-state.md`，避免 Relay/Client 与测试 import 循环。
- **`pkg/protocol` 集中 `msg_type` 与布局**：字段多，单独包便于 golden 测试与双向编码。
- **`pkg/relay` 与 `cmd/tunnel-relay` 分离**：核心逻辑可 `go test`，进程入口只做配置与 `Listen`。

---

## Architectural Patterns

### Pattern 1：每连接「读循环 + 显式写路径」

**What：** 每条 `*tls.Conn` 上一个 goroutine 读入并成帧；写帧时 **对同一 `Conn` 串行化**。

**When to use：** v1.1 最小实现、少量 peer。

**Trade-offs：** 实现简单；后续背压可再拆 `send ch`。

### Pattern 2：Session Registry 为连接 ↔ peer ↔ session 的唯一真源

**What：** 单播目标、广播成员列表只从 Registry 查询；目标不存在时按 `errors.md` 返回（如 `ERR_ROUTING_INVALID` / `ERR_SESSION_NOT_FOUND`）。

**When to use：** 任意多 peer 路由。

**Trade-offs：** 进程内锁在「少量固定 peer」下可接受；集群化 **非 v1.1 范围**。

### Pattern 3：控制面与数据面分派分离

**What：** payload 首字节分派：`0x01`–`0x05` → 控制面；`0x10`–`0x12` → 流与路由（先过 JOIN 门禁）。

**When to use：** 对齐 `routing-modes.md` 与 `session-state.md`。

**Trade-offs：** 代码略多，日志边界清晰。

**概念示例：**

```go
switch msgType := payload[0]; msgType {
case 0x01, 0x02, 0x03, 0x04, 0x05:
    handleControl(conn, payload)
case 0x10, 0x11, 0x12:
    if !conn.joined {
        // 与 session-state.md / errors.md 对齐的策略
        return
    }
    handleStream(conn, payload)
default:
    // 未知 opcode：按 errors 与实现策略处理
}
```

---

## Data Flow

### 控制面（创建 / 加入）

```
Client                         Relay
  |-- 帧: SESSION_CREATE_REQ --------------------------------->|
  |<----------------------- 帧: SESSION_CREATE_ACK ------------|
  |                              Registry：新建 Session，登记创建者
  |                    <-------- SESSION_JOIN_REQ -------------|
  |                    -------- SESSION_JOIN_ACK (peer_id) --->|
  |                              Registry：credential 解析，peer_id 绑定 Conn
```

### 数据面（广播）

```
Peer A                         Relay                         Peer B, C
  |-- STREAM_DATA, BROADCAST, src=A ------------------------->|
  |                              复制帧 → B、C（不写回 A）
```

### 数据面（单播）

```
Peer A                         Relay                         Peer B
  |-- STREAM_DATA, UNICAST, dst=B --------------------------->|
  |                              Registry 查表 → 仅写 B
```

### Key Data Flows

1. **成帧 → 语义：** `pkg/framing` 产出 `Frame.Payload` → `pkg/protocol` 读 `msg_type` → 控制面或 `STREAM_*`。
2. **JOIN 门禁：** 未 `SESSION_JOIN_ACK` 前不得发数据面路由帧。
3. **错误：** 统一 `PROTOCOL_ERROR`（`pkg/framing` 中 `ErrCode` 与 `errors.md` 对齐）。

---

## Scaling Considerations

| 规模 | 架构调整 |
|------|-----------|
| 演示 / 少量 peer（v1.1） | 单进程、内存 Registry、每连接 1～2 个 goroutine |
| 同机数百连接 | 缓冲池、Registry 分片锁或 per-session 锁 |
| 多机 / HA | **超出 v1.1** |

### Scaling Priorities

1. **第一瓶颈：** 广播 O(n) 复制 — 最小实现可接受。
2. **第二瓶颈：** 全局大锁 — peer 数极小时不必过早拆分。

---

## Anti-Patterns

### Anti-Pattern 1：在成帧层解析会话与路由

**现象：** 在 `pkg/framing` 内识别 SESSION 或路由。

**问题：** 违反 `connection-state.md`；职责混杂。

**应：** 成帧仅 `ParseFrame`；会话与路由在 `pkg/protocol` + Relay/Client。

### Anti-Pattern 2：用同一状态机表示「半包」与「未 JOIN」

**现象：** `ErrNeedMore` 与「未 ACK」混在同一枚举。

**问题：** `connection-state.md` 与 `session-state.md` 明确分层。

**应：** 成帧返回 `ErrNeedMore`；会话层单独 `joined` / `peerID`。

### Anti-Pattern 3：盲信对端 `src_peer_id`

**问题：** 伪造来源。

**应：** 以 Registry 为准重写或拒绝不一致帧（策略固定）。

---

## Integration Points

### 与现有仓库

| 集成点 | 新建 / 修改 | 说明 |
|--------|-------------|------|
| `pkg/framing` | **复用为主** | `ParseFrame` / `AppendFrame` / `ErrCode`；缺 `PROTOCOL_ERROR` payload 辅助时优先放在 **`pkg/protocol`**，避免污染成帧层 |
| `pkg/appenvelope` | **可选** | 最小路径可先无信封；演示 HTTP 风格关联时再接 |
| `docs/spec/v1/*.md` | **规范源**（一般不随 v1.1 改语义） | 实现与字段/opcodes/错误码不一致视为 bug |
| `testdata/` | **扩展** | E2E 或负例向量 |

### Internal Boundaries

| Boundary | 通信方式 | 备注 |
|----------|----------|------|
| `framing` ↔ `protocol` | `[]byte` payload、`Frame` | `protocol` 不依赖 TLS |
| `protocol` ↔ Relay Router | 结构体或固定偏移视图 | Router 只关心 session 内成员集合与目标 peer |
| Registry ↔ `Conn` | 查表后 `Write` | 连接关闭必须摘除 |

### New vs Modified（显式）

| 类别 | 路径 / 组件 |
|------|-------------|
| **已有、尽量少改** | `pkg/framing`、`pkg/appenvelope`、`docs/spec/v1/`（仅互斥 bug 时修文档或实现） |
| **新建（推荐）** | `pkg/protocol`、`pkg/relay`、`pkg/client`、`cmd/tunnel-relay`、`cmd/tunnel-client` |
| **测试** | `pkg/*_test.go`、扩展 `testdata`；`go test ./...` |

---

## Suggested Build Order（依赖优先）

1. **`pkg/protocol`** — 控制面、`PROTOCOL_ERROR`、路由前缀 + `STREAM_DATA` 基础视图（对齐 `streams-lifecycle.md` 偏移）。依赖：可仅用 payload 测，或用 `framing.AppendFrame` 拼整帧。
2. **`pkg/client` 核心** — `net.Conn` 上成帧循环、CREATE/JOIN、JOIN 后门禁、发送 `STREAM_DATA`。依赖：`pkg/framing`、`pkg/protocol`；可用 `net.Pipe` 单测。
3. **`pkg/relay` 控制面** — Registry、CREATE→ACK、JOIN→ACK、`ERR_SESSION_NOT_FOUND` / `ERR_JOIN_DENIED`。
4. **`pkg/relay` 数据面** — 广播（不回送发送者）、单播（Registry 查 `dst_peer_id`）。
5. **`cmd/tunnel-relay` + `cmd/tunnel-client` + E2E** — TLS、测试证书、**E2E-DEMO**（`PROJECT.md`）。
6. **（可选）`pkg/appenvelope`** — 在稳定 `STREAM_DATA` 上演示 `HAS_APP_ENVELOPE`。

**排序理由：** 协议编解码与 Client 状态机可脱离真实网络先测；Relay 先正确会话与错误路径，再加转发；进程入口与 TLS 最后接合，避免过早绑在配置上。

---

## Sources

- [`.planning/PROJECT.md`](../../PROJECT.md)
- [`docs/spec/v1/transport-binding.md`](../../docs/spec/v1/transport-binding.md)
- [`docs/spec/v1/connection-state.md`](../../docs/spec/v1/connection-state.md)
- [`docs/spec/v1/session-state.md`](../../docs/spec/v1/session-state.md)
- [`docs/spec/v1/session-create-join.md`](../../docs/spec/v1/session-create-join.md)
- [`docs/spec/v1/routing-modes.md`](../../docs/spec/v1/routing-modes.md)
- [`docs/spec/v1/streams-lifecycle.md`](../../docs/spec/v1/streams-lifecycle.md)
- [`pkg/framing/decode.go`](../../pkg/framing/decode.go)、[`pkg/framing/errors.go`](../../pkg/framing/errors.go)

---

*Architecture research for: Tunnel v1.1 最小 Relay + Client*  
*Researched: 2026-04-04*
