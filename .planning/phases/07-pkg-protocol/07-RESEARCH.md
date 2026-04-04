# Phase 7: 协议载荷层（pkg/protocol）- Research

**Researched:** 2026-04-04  
**Domain:** v1 控制面/数据面 payload 编解码、JOIN 前门禁、与 `pkg/framing` 分层  
**Confidence:** HIGH（规范与现有代码已对齐；实现细节以 PLAN 为准）

## Summary

Phase 7 在 **Phase 6 已交付的成帧基线**（`pkg/framing`：`ParseFrame` / `AppendFrame`、帧头 10 字节、`ErrNeedMore` 与连接级致命错误分离）之上，新增 **`pkg/protocol`**：对 **帧 payload**（`Frame.Payload`）做 **会话语义** 解析与构造，使 Relay 与 Client **共用**同一套与 `docs/spec/v1/` **字段级一致**的类型与函数。

**PROT-01** 要求：`SESSION_*`、`PROTOCOL_ERROR`、`STREAM_*` 数据面（含 **18 字节路由前缀** + `streams-lifecycle.md` 中 `stream_id` / `flags` / `application_data` 等）的编解码；**`ErrCode`** 数值与 **`pkg/framing/errors.go`**、`errors.md`（ERR-01）一致，**不在成帧层**重复定义错误码表。`PROTOCOL_ERROR` 载荷 **无** 路由前缀（见 `errors.md`）。

**PROT-02** 要求：在 **未收到 `SESSION_JOIN_ACK` 且 `peer_id` 有效** 前，**不得**将带 **`routing-modes.md` 所定义数据面路由前缀** 的帧当作合法业务路径；实现上应对 **`msg_type ∈ {0x10, 0x11, 0x12}`**（`STREAM_OPEN` / `STREAM_DATA` / `STREAM_CLOSE`）与规范一致地拦截或标记为非法，供 Client/Relay **复用同一 helper**（见 `session-state.md` STATE-01）。

**Primary recommendation：** 新建 `pkg/protocol`，仅依赖 **`encoding/binary`**、**`unicode/utf8`**（校验 `reason`）等标准库；与 `pkg/framing` 的边界为：**framing 不解析 payload 内 `msg_type`**；**protocol 不解析 10 字节帧头**。

<phase_requirements>
## Phase Requirements

| ID | Description | Research Support |
|----|-------------|------------------|
| PROT-01 | 控制面/数据面 payload、`msg_type`、路由前缀与 `PROTOCOL_ERROR` 编解码；与 v1 规范字段对齐；与 `pkg/framing` 边界清晰 | 见下文 Standard Stack、Architecture Patterns、`docs/spec/v1/session-create-join.md`、`routing-modes.md`、`streams-lifecycle.md`、`errors.md` |
| PROT-02 | JOIN_ACK 前不得将带数据面路由的帧视为合法业务路径；共享门禁 helper | `session-state.md` 门禁 + 数据面 opcode 表（`0x10`–`0x12`）；见 Architecture Patterns「JOIN 前门禁」 |
</phase_requirements>

## Project Constraints（from .cursor/rules/）

- **语言**：实现默认 **Go**；v1 **不采用 WebSocket**；主路径 **TLS 上自定义成帧**。
- **交付顺序**：规范与一致性测试优先；本阶段延续 **testdata / `go test`** 风格。
- **GSD**：重大实现仍建议经 `/gsd-plan-phase` / execute 工作流；本文档供 PLAN 消费。

## Standard Stack

### Core

| Library | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| Go | `go 1.22+`（仓库 `go.mod` 为 1.22；本机可更新） | 协议实现 | 项目锁定 |
| `encoding/binary` | stdlib | `uint16`/`uint32`/`uint64` 大端编解码 | 与 `frame-layout.md` / 各 payload 表一致 |
| `unicode/utf8` | stdlib | 可选：校验 `PROTOCOL_ERROR` 的 `reason` 为合法 UTF-8 | `errors.md` 约束 |
| `pkg/framing` | 仓库内 | 帧级：`ParseFrame`、`AppendFrame`、`ErrCode` | Phase 6 基线；protocol **调用**而非复制 |

### Supporting

| Library | Version | Purpose | When to Use |
|---------|---------|---------|-------------|
| `pkg/appenvelope` | 仓库内 | `SplitApplicationData`（`STREAM_DATA` 内 `application_data`） | 若测试或示例需拆信封；**非** PROT-01 最小子集必需 |
| `github.com/stretchr/testify` | — | 断言 | **当前仓库未使用**；新测优先延续 **`testing` + 表驱动**（与 `pkg/framing` 一致） |

### Alternatives Considered

| Instead of | Could Use | Tradeoff |
|------------|-----------|----------|
| 手写结构体 + `Append` | `struct`-tag 反射编解码库 | v1 布局固定、字段少；手写更贴近规范表格、便于 golden 对照 |
| 在 `pkg/framing` 内解析 `msg_type` | 独立 `pkg/protocol` | 违反 `connection-state.md` 分层；会话语义污染成帧层 |

**Installation：**

```bash
# 无新增第三方模块时无需安装；仅标准库 + 本仓库包
go test ./pkg/protocol/...
```

**Version verification：** 仓库当前 **无 `go.sum`**（无第三方依赖）。若 PLAN 引入 `testify` / `go-cmp`，须在 `go.mod` 中显式添加并记录版本。

## Architecture Patterns

### 推荐包内布局

```
pkg/protocol/
├── msgtype.go          # MsgType 常量（0x01–0x05 控制面，0x10–0x12 数据面）
├── control.go          # SESSION_* / PROTOCOL_ERROR 编解码
├── routing.go          # 路由前缀 18 字节视图 + 数据面分类
├── streamdata.go       # STREAM_DATA（及可选 OPEN/CLOSE）payload 视图
├── join_gate.go        # JOIN_ACK 前门禁（PROT-02）
└── *_test.go           # 表驱动 + testdata（对齐 Phase 6）
```

### Pattern 1：成帧 → 协议两层流水线

**What：** `ParseFrame` 成功后，仅将 `f.Payload` 交给 `protocol.DecodeX`；发送时 `protocol.EncodeX` 产出 `[]byte` 再 `AppendFrame(Frame{Payload: ...})`。

**When：** 所有 Client/Relay 读写路径。

**Example（概念）：**

```go
// 成帧层：connection-state.md — 不解析 payload 内语义
n, frame, err := framing.ParseFrame(buf)
// 协议层：session-create-join.md / routing-modes.md
msgType := frame.Payload[0]
switch msgType {
case MsgTypeSessionCreateReq:
    // ...
}
```

### Pattern 2：数据面路由前缀（18 字节）

**What：** `routing-modes.md` + `streams-lifecycle.md`：`msg_type` @0，`routing_mode` @1，`src_peer_id` @2–9，`dst_peer_id` @10–17；`stream_id` **起始于偏移 18**。

**When：** 解析/构造 `STREAM_OPEN`（0x10）、`STREAM_DATA`（0x11）、`STREAM_CLOSE`（0x12）。

### Pattern 3：JOIN 前门禁（PROT-02）

**What：** `session-state.md`：未 `SESSION_JOIN_ACK` 前 **MUST NOT** 发送带数据面路由前缀的帧。共享 helper 建议语义：

- 输入：`joined bool`（或等价：已收到 ACK 且 `peer_id != 0`）、**完整 payload**（或至少首字节 `msg_type`）。
- 若 **`!joined`** 且 **`msg_type` 为 `0x10`/`0x11`/`0x12`** → **非合法业务路径**（与规范中「数据面」opcode 表一致；**不**将 `0x05` `PROTOCOL_ERROR` 算作数据面）。

**When：** Client 发送前校验；Relay 在 Registry 确认成员身份前对入站帧做一致判定。

### Anti-Patterns to Avoid

- **混淆 `ErrNeedMore` 与门禁失败：** 半包属成帧层；JOIN 门禁仅在 **完整 payload** 解析之后适用（见 `.planning/research/PITFALLS.md`）。
- **在 protocol 包内读写 `net.Conn`：** 保持纯编解码 + 纯函数门禁，便于单测与复用。
- **手写与 `errors.md` 冲突的 `err_code`：** 必须使用 `pkg/framing.ErrCode` 或从同一表生成的常量。

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| 帧切分与版本门闸 | 自定义读循环替代 `ParseFrame` | `pkg/framing` | 已与 `connection-state.md`、TEST-01 对齐 |
| 错误码数值 | 第二套 `uint16` 枚举 | `framing.ErrCode` + `errors.md` | ERR-01 唯一权威 |
| UTF-8 校验 | 逐字节 ad-hoc | `utf8.ValidString` / `Valid` | 与 `PROTOCOL_ERROR` 规范一致 |
| 应用信封边界 | 在 protocol 重复解析 | `pkg/appenvelope.SplitApplicationData` | APP-01 已落地 |

**Key insight：** 本阶段价值在于 **与规范逐字段对齐的可测试类型**，而非新抽象层。

## Common Pitfalls

### Pitfall 1：将「广播/单播」与 `routing_mode` 单一字段绑定

**What goes wrong：** 仅看 `dst_peer_id` 推断单播，违反 `routing-modes.md`「**必须**联合判定」。

**Why：** 规范要求 `routing_mode` 与 `dst_peer_id` 组合语义。

**How to avoid：** 提供 `RoutingIntent` 或联合校验函数，单元测试覆盖矛盾组合（如 `BROADCAST` + 非零 `dst`）。

### Pitfall 2：`PROTOCOL_ERROR` 误加 18 字节路由前缀

**What goes wrong：** 与 `STREAM_DATA` 布局混淆，破坏与 `errors.md` 的互操作性。

**Why：** `errors.md` 明确 **无** 数据面路由前缀。

**How to avoid：** 单独类型 `ProtocolErrorPayload`；编码路径不经过 `RoutingPrefix` 结构。

### Pitfall 3：`payload` 内第二层 `payload_len`（STREAM_DATA）与帧级 `payload_len` 混淆

**What goes wrong：** `streams-lifecycle.md` 中 `STREAM_DATA` 在偏移 23 处有 **16 位** `payload_len`（应用数据长度），与帧头 32 位 `payload_len` 名称碰撞。

**Why：** 同名不同层。

**How to avoid：** 命名如 `FramePayloadLen` vs `ApplicationDataLen`；注释引用规范偏移表。

## Code Examples

### 路由前缀读取（概念，大端）

```go
// Source: docs/spec/v1/routing-modes.md
// payload 至少 18 字节
routingMode := payload[1]
srcPeerID := binary.BigEndian.Uint64(payload[2:10])
dstPeerID := binary.BigEndian.Uint64(payload[10:18])
```

### PROTOCOL_ERROR 布局（控制面，无路由前缀）

```go
// Source: docs/spec/v1/errors.md — err_code @1, reason_len @3, reason @5
// msg_type @0 = 0x05
```

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| 仅在 `pkg/framing` 测 payload 前缀字节 | 独立 `pkg/protocol` 完整类型与往返 | Phase 7 | Client/Relay 共享语义层 |
| 规范分散 | ERR-01 / SESS / ROUTE / STREAM 交叉引用 | v1.0 已发布 | 实现须对照多文档偏移表 |

**Deprecated/outdated：** 无（v1 仍为当前里程碑）。

## Open Questions

1. **未知 `msg_type`（非控制、非数据面表内）在 JOIN 前的处理**
   - **已知：** 规范未要求 Phase 7 实现完整「未知类型」状态机；E2E-02 覆盖非法路由等。
   - **不清：** 是否需在 protocol 暴露 `MsgTypeUnknown` 统一分支。
   - **建议：** PLAN 中列为实现选项；至少 **不**将未知类型当作合法数据面路径。

2. **创建者是否必须先 JOIN 再发数据面**
   - **已知：** `session-state.md` 以 JOIN_ACK 为门禁；`FEATURES.md` 指出创建者通常需对同 session 再发 JOIN。
   - **建议：** 协议层不特殊 case「CREATE 即成员」；门禁只认 **`joined` 标志**。

## Environment Availability

| Dependency | Required By | Available | Version | Fallback |
|------------|-------------|-----------|---------|----------|
| Go toolchain | 编译与测试 | ✓ | go1.26.1（本机探测）；`go.mod` 1.22 | — |
| 外部 TCP/TLS 对端 | 本阶段 | — | — | **不需要**：payload 级单测即可 |

**Step 2.6 说明：** 本阶段以 **纯 Go 编解码 + 表驱动测试** 为主，无数据库/云服务硬依赖。

**Missing dependencies with no fallback：** 无。

## Validation Architecture

> 面向 Nyquist **维度 8（测试/验证策略）**：本阶段「反馈信号」为 **`go test` 退出码** + 覆盖率足够的表驱动用例；无 UI。

### Test Framework

| Property | Value |
|----------|-------|
| Framework | Go `testing`（与现有 `pkg/framing` 一致） |
| Config file | 无；`go test ./pkg/protocol/...` |
| Quick run command | `go test ./pkg/protocol/... -count=1` |
| Full suite command | `go test ./... -count=1` |

### Phase Requirements → Test Map

| Req ID | Behavior | Test Type | Automated Command | File Exists? |
|--------|----------|-----------|-------------------|--------------|
| PROT-01 | `SESSION_CREATE_ACK` / `SESSION_JOIN_REQ/ACK` 字段往返与边界（`session_id` 36 字符、`invite_code` 长度等） | unit | `go test ./pkg/protocol/... -run Session -count=1` | ❌ Wave 0 新增 |
| PROT-01 | `PROTOCOL_ERROR`：`err_code` 大端、`reason_len`、`reason` UTF-8；过短/截断返回明确错误 | unit | `go test ./pkg/protocol/... -run ProtocolError -count=1` | ❌ Wave 0 |
| PROT-01 | `STREAM_DATA`：18 字节前缀 + `stream_id` @18 + `flags` + 内层 `payload_len` + `application_data` 与 `routing-modes.md` / `streams-lifecycle.md` 示例一致 | unit + golden | `go test ./pkg/protocol/... -count=1` + 可选 `testdata/protocol/*.hex` | ❌ Wave 0 |
| PROT-02 | `joined==false` 时 `msg_type∈{0x10,0x11,0x12}` 判定为非法业务路径；`joined==true` 时允许 | unit | `go test ./pkg/protocol/... -run JoinGate -count=1` | ❌ Wave 0 |

### Sampling Rate

- **Per task commit：** `go test ./pkg/protocol/... -count=1`
- **Per wave merge：** `go test ./... -count=1`
- **Phase gate：** 全量测试绿后再进入 Phase 8 集成

### Wave 0 Gaps

- [ ] `pkg/protocol/control_test.go` — SESSION_* / PROTOCOL_ERROR 覆盖 PROT-01
- [ ] `pkg/protocol/join_gate_test.go` — PROT-02 门禁矩阵
- [ ] `testdata/protocol/` — 可选：从 `routing-modes.md` / `errors.md` 复制的 hex 向量（对齐 Phase 6 `testdata/framing/` 约定）
- [ ] 文档注释中引用 **REQ-ID** 与 `docs/spec/v1/*.md` 锚点（可追溯）

### Nyquist Dimension 8（本阶段要点）

- **信号源：** 编译成功 + 单测通过；PR 上与 Phase 6 相同 **`go test ./...`** CI。
- **负例必备：** 半包 **不属于** protocol 测程（属 `pkg/framing`）；protocol 测「**完整 payload** 过短」「JOIN 前数据面 opcode」。
- **可追溯：** 每个 golden 注释对应 `docs/spec/v1/` 章节（与 `06-RESEARCH.md` 第三层一致）。

## Sources

### Primary（HIGH confidence）

- `docs/spec/v1/session-create-join.md` — SESS-01/02；控制面 opcode 与消息体
- `docs/spec/v1/errors.md` — ERR-01；`PROTOCOL_ERROR` 线格式
- `docs/spec/v1/routing-modes.md` — ROUTE-01/02；18 字节路由前缀
- `docs/spec/v1/streams-lifecycle.md` — STREAM-01/02；`STREAM_DATA` 偏移表
- `docs/spec/v1/session-state.md` — STATE-01；JOIN 前门禁
- `docs/spec/v1/connection-state.md` — 成帧层 vs payload 语义层
- `pkg/framing/decode.go`、`pkg/framing/errors.go` — 帧边界与 `ErrCode`

### Secondary（MEDIUM confidence）

- `.planning/research/PITFALLS.md`、`ARCHITECTURE.md`、`FEATURES.md` — 项目内已达成共识的坑与分层描述
- `.planning/phases/06-consistency-test-suite/06-RESEARCH.md` — testdata 与 golden 策略

## Metadata

**Confidence breakdown:**

- Standard stack: **HIGH** — 规范与仓库代码可核对
- Architecture: **HIGH** — 分层在 `connection-state.md` / `session-state.md` 已定义
- Pitfalls: **MEDIUM–HIGH** — 部分边界情况需在实现中补 fuzz/向量

**Research date:** 2026-04-04  
**Valid until:** ~30 天（协议稳定；仅当规范修订时重研）

## RESEARCH COMPLETE
