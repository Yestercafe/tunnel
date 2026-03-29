# Phase 5 — 技术调研：状态机、错误与安全假设

**Phase:** 5 — 状态机、错误与安全假设（STATE-01、ERR-01、SEC-01）  
**调研日期:** 2026-03-29  
**置信度:** **MEDIUM–HIGH**（状态机与既有 `docs/spec/v1/` 交叉核对充分；**应用层错误帧**的 **opcode / 载荷布局**须在 05-02 规范中拍板；TLS 威胁模型为文档结论，非形式化验证）

<phase_requirements>
## Phase Requirements

| ID | 描述 | 调研如何支撑实现 |
|----|------|-------------------|
| **STATE-01** | 文档化**连接级**与 **session 级**状态机（状态、事件、转移） | 下文「分层状态机模型」将 **TRANS-01 成帧解析**、**TLS 连接**、**控制面会话**、**数据面流**拆成可组合的机；与 `session-create-join.md`、`streams-lifecycle.md`、`transport-binding.md` 对齐 |
| **ERR-01** | **错误码目录**与**连接/会话中止原因**；与 TLS alert、TCP 断开的关系可简述 | 下文「错误码合并清单」「与 TLS/TCP 的映射」汇总既有占位名并给出 Go `crypto/tls` 侧可查行为；**线格式**待 05-02 固化 |
| **SEC-01** | 文档化 **TLS 在边缘/服务器终止**；协议内**不负责 v1 E2E 加密** | 下文「SEC-01 威胁模型与部署假设」显式信任边界与**不提供**的保证，与 `.planning/REQUIREMENTS.md` / `PROJECT.md` 一致 |
</phase_requirements>

## User Constraints

**无** — 本阶段无 discuss-phase 的 `*CONTEXT.md`。须遵守 `.cursor/rules/gsd-project.md` 与既有规范。

### Project Constraints（来自 `.cursor/rules/gsd-project.md`）

- **语言**：实现默认 **Go**；交付顺序：**规范 + 一致性测试**优先。  
- **传输**：v1 **不采用 WebSocket**；主路径为 **TLS（如基于 TCP）上的自定义成帧**。  
- **安全边界**：机密性与完整性主要依赖 **TLS（边缘/服务器）**；协议负责会话、成员与消息语义。  
- **顺序模型**：**流内有序、流间可乱序**。  
- **GSD**：重大改动宜经 GSD 工作流（规则内述）。

## Findings

### STATE-01：连接 / 会话 / 流三层正交

- **连接级**：`transport-binding.md` 的**成帧解析循环**（半包/粘包/最大帧长）与 **TLS 已建立、应用数据可读**的前提分开写。  
- **Session 级**：`SESSION_CREATE_*` / `SESSION_JOIN_*` 与 `join-credentials.md` 的成功/失败路径；**未 JOIN_ACK 前**不得发送数据面路由帧（与现有控制面/数据面分离一致）。  
- **流级**：`streams-lifecycle.md` 的 OPEN / DATA+FIN / CLOSE；**不得**在 SESSION 控制消息中带 `stream_id`。

### ERR-01：占位名收敛与传输层关系

- 既有占位：**`ERR_FRAME_TOO_LARGE`**、**`ERR_PROTO_VERSION`**（`pkg/framing` 已对齐符号）、**`ERR_JOIN_DENIED`**、**`ERR_SESSION_NOT_FOUND`**；路由/信封等需**新增**表项。  
- **TLS/TCP**：Go 中可对 `Read`/`Close` 错误 **unwrap `tls.AlertError`**（[pkg.go.dev/crypto/tls#AlertError](https://pkg.go.dev/crypto/tls#AlertError)）；**不要**在规范里写死「某 ERR ↔ 某 alert 字节」。**TCP**：对端关闭 → `EOF` / 读错误；与**应用层错误帧**并列描述。  
- **版本策略**：`version-capability.md` 已要求不支持版本时 **先发错误指示（线格式 Phase 5）再关 TLS**；无法实现时允许直接关 TCP。

### SEC-01：TLS 边缘终止与无 E2E

- **机密性/完整性**在 **TLS 段**内由 **边缘/服务器证书域**保证；**Relay 进程内**解析的是**明文逻辑帧**。  
- **v1 不提供**端到端加密；若需防中继窃听，走 **部署隔离 / mTLS / 应用层加密** 或 **v2 SEC-02**，在 `security-assumptions.md` 中逐条列出「提供 / 不提供」。

## Summary

Phase 5 的目标是把 **Phase 1–4 已分散描述的行为**收敛为三类可交付物：**（1）** 覆盖主路径与失败路径的**状态机叙述**（至少区分「TLS/字节流成帧」「会话成员关系」「逻辑流」三层）；**（2）** 将各文档中的 **`ERR_*` 占位名**升格为可测试、可引用的**错误码目录**，并说明**中止**时是否先发应用层错误指示、再关 TLS/TCP，以及与 **TLS alert**、**TCP 半关闭/全关闭**、**对端 EOF** 的观测关系；**（3）** 用短文档写清 **TLS 在边缘终止**时的信任边界、**中继可见明文**的含义、以及 v1 **不做 E2E** 的产品/部署含义。

**首要建议：** 规划时把 **「成帧解析状态机」（TRANS-01）**、**「会话/成员状态」**、**「流状态」**分开成图或表，再在 `docs/spec/v1/` 增加一篇 **总览**（或拆分 `connection-state.md` / `errors.md` / `security-assumptions.md`），避免单一大图无法维护；错误码采用 **稳定符号名 + 规范内数值枚举（uint16 等）** 双轨，便于 `pkg/framing` 与规范同步。

## Standard Stack

### Core

| 组件 | 版本 / 依据 | 用途 | 为何是默认栈 |
|------|----------------|------|----------------|
| **Go** | `go.mod` 声明 **1.22**；本机验证 **go1.23.4** | 参考实现与测试 | 项目约束 |
| **`crypto/tls` + `net`** | 标准库 | TLS 上字节流、`Conn.Close` / `CloseWrite`、握手与告警 | 与 TRANS-01「TLS 解密后成帧」一致 |
| **`encoding/binary`** | 标准库 | 帧与控制字段大端编解码 | 与规范一致 |

### Supporting（与 Phase 6 衔接）

| 组件 | 用途 | 何时引入 |
|------|------|----------|
| `github.com/stretchr/testify` | 断言（`.cursor/rules` 中 STACK 推荐） | **当前 `go.mod` 未声明**；可在 **TEST-01 / Phase 6** 引入，Phase 5 以**规范与常量对齐**为主 |
| `github.com/google/go-cmp` | 结构体深度比较 | 复杂状态机测试时可选 |

### Alternatives Considered

| 而非 | 可选用 | 权衡 |
|------|--------|------|
| 纯文档错误名、无数值码 | **uint16 应用层错误码** + 符号名 | 互操作与 golden 向量需要稳定二进制语义 |
| 将 TLS alert 暴露给应用业务 | **分层**：TLS 错误留在传输观测；**协议 ERR_** 表示应用层可恢复/可记录语义 | 避免与具体 TLS 库版本强绑定 |

**版本核实：** `go version` → go1.23.4；`go.mod` → `go 1.22`。

## Architecture Patterns

### 推荐：三层正交状态机

**What：** 不把「连接」与「session」混在同一枚举里；规范中分别定义：

1. **字节流 / 成帧层（连接级）** — 对应 `transport-binding.md`：缓冲区从空开始 → 读入 → 解析 10 字节头 → 按 `payload_len` 收满 → 交付帧 → 循环；失败进入 **致命错误**（如 `ERR_FRAME_TOO_LARGE`）或 **需更多数据**（非错误）。  
2. **会话 / 成员层（session 级）** — 对应 `session-create-join.md`、`join-credentials.md`：在**已建立 TLS** 上，peer 是否已完成 **SESSION_JOIN_ACK**、是否绑定 **session_id**、**peer_id**；失败路径含 **`ERR_JOIN_DENIED`**、**`ERR_SESSION_NOT_FOUND`** 等。  
3. **流层（stream 级）** — 对应 `streams-lifecycle.md`：每条 `stream_id` 从 **OPEN →（DATA/FIN）→ CLOSE**；与 **SESSION_*** 控制消息**不得**混用 `stream_id`。

**When to use：** 所有 Phase 5 状态图/表应标明「本表属于哪一层」，避免把 **「半包」**与 **「未加入 session」**画在同一状态集合。

### 模式：协议错误指示 vs 直接关连接

**What：** `version-capability.md` 已要求：不支持的 **version** 时，**SHOULD** 发**协议错误指示**（线格式 Phase 5 定义）、**MUST** 关 TLS；若无法发帧可**直接关 TCP**。  
**When to use：** ERR-01 应统一「哪些错误**必须**先发错误帧再关」「哪些允许直接关」（例如已解析头前即发现版本错误，可能尚无有效 payload 通道）。

### Anti-Patterns to Avoid

- **把 WebSocket 关闭码搬进 v1：** 项目明确 v1 不采用 WebSocket；**不要**引入 WS 的 close code 语义，仅用 **TCP/TLS** 与 **应用层 ERR_** 描述。  
- **在规范中承诺「映射到某一 TLS alert 字节」：** 实现依赖具体栈；应描述为 **「可能观测到 `AlertError` / 连接关闭」**，而非写死 alert 编号与协议 ERR 的一一映射（除非团队锁死某一 TLS 版本与实现）。  
- **单一全局状态枚举：** 会导致 session 与 stream 生命周期无法单独测试。

## Don't Hand-Roll

| 问题 | 不要自造 | 应用什么 | 原因 |
|------|----------|----------|------|
| TLS 错误类型判断 | 字符串匹配 `"tls:"` | `errors.As` / unwrap 查 **`tls.AlertError`**（Go 1.21+ 类型）；或记录 `Close()` / `Read()` 返回的 error | 与标准库演进兼容 |
| v1 机密性 | 在协议里加「简易 E2E」 | **SEC-01**：写清不做 E2E；若需要走 **v2 SEC-02** 或应用层加密 | 密钥管理与身份绑定超出当前范围 |
| 错误码散落 | 各文档随意起名 | **单一错误码表**（05-02）收录所有占位名并分配数值 | `pkg/framing` 已有 `ERR_FRAME_TOO_LARGE`、`ERR_PROTO_VERSION` 字符串常量，需与数值表对齐 |

**要点：** 中继与客户端之间的 **TLS 仅保护到边缘**时，**Relay 进程内协议解析看到的是明文帧** — 这是**部署与信任模型**问题，不是再套一层自定义 XOR 能解决的。

## Runtime State Inventory

本阶段为**规范收敛与文档化**，不涉及仓库级重命名或数据迁移。**不适用** — 已跳过。

## Common Pitfalls

### 1. 混淆「TLS 连接失败」与「应用层协议错误」

**现象：** 握手失败、证书错误、对端 `alert` 与 **`ERR_PROTO_VERSION`** 混在同一用户提示。  
**根因：** 分层未写清。  
**避免：** ERR-01 中分栏：**传输层（TLS/TCP）** vs **成帧/版本** vs **会话/路由/信封**。  
**信号：** 握手未完成时 `Read` 返回的错误不应解读为 `ERR_*` 帧。

### 2. `CloseWrite` 与「半关闭」语义滥用

**现象：** 试图用 TLS `CloseWrite` 表达「session 结束但连接保持」。  
**根因：** v1 会话生命周期**未**要求与 TCP 半关闭一一对应。  
**避免：** 会话离开用 **控制面消息 + ERR/原因** 描述；**`CloseWrite`** 仅作实现优化时说明（**MEDIUM** — 见 Go 文档「Most callers should just use Conn.Close」）。  
**来源：** [pkg.go.dev/crypto/tls#Conn.CloseWrite](https://pkg.go.dev/crypto/tls#Conn.CloseWrite)

### 3. 错误帧与「仍可解析的后续字节」

**现象：** 发送错误后未排空缓冲区或未关闭读侧，导致实现继续解析垃圾帧。  
**避免：** 规范中写清 **进入致命错误状态后 MUST NOT** 再交付应用帧；实现宜 **关闭连接**。

## Code Examples

### 观测 TLS 告警类型（Go）

```go
// 思路来源：https://pkg.go.dev/crypto/tls#AlertError
// AlertError 为 TLS alert 的整数码包装；实际错误链需按项目日志策略 unwrap。
var alert tls.AlertError
if errors.As(err, &alert) {
    _ = uint8(alert) // 调试记录；与协议 ERR_* 无强制数值对应
}
```

### 成帧错误与现有参考实现

```21:24:pkg/framing/decode.go
	// ErrFrameTooLarge corresponds to ERR_FRAME_TOO_LARGE in the spec.
	ErrFrameTooLarge = errors.New("ERR_FRAME_TOO_LARGE")
	// ErrProtoVersion corresponds to ERR_PROTO_VERSION in the spec.
	ErrProtoVersion = errors.New("ERR_PROTO_VERSION")
```

Phase 5 为上述符号名增加**规范级数值**与**会话/路由错误**时，保持 **同一命名前缀** 便于 `grep` 与 Nyquist 追溯。

## State of the Art

| 旧做法 | 当前建议（v1） | 说明 |
|--------|----------------|------|
| 各文档独立 `ERR_*` 占位 | **统一错误码表 + 关闭语义** | 满足 ERR-01 |
| 仅叙述 TRANS-01 循环 | **连接 / 会话 / 流** 三机并列 | 满足 STATE-01 |
| 「TLS 就够了」一句话 | **边缘终止 + 中继明文 + 无 E2E** 短威胁列表 | 满足 SEC-01 |

**过时：** 依赖 WebSocket 关闭码描述连接结束 — **v1 排除**（见 REQUIREMENTS v2 TRANS-02）。

## Open Questions

1. **应用层「协议错误指示」帧的 opcode 与载荷是否独立一节？**  
   - 已知：`version-capability.md` 承诺存在此类指示，**线格式未定**。  
   - 建议：05-02 增加 **`ERROR` 或带 `msg_type` 的控制消息**，体为 **`uint16 err_code` + 可选 `reason_len/reason`**（**UTF-8**），大端与现有控制消息一致。  
   - **置信度：MEDIUM**（需与 `session-create-join.md` opcode 表预留位协调）。

2. **路由非法（如 `routing_mode=0`）是否单独 ERR 名？**  
   - 已知：`routing-modes.md` 引用 Phase 5 **ERR**。  
   - 建议：在错误表中新增 **`ERR_ROUTING_INVALID`** 或与「协议违反」共用母码 + 子原因 — **规划阶段二选一**。

3. **应用信封解码失败（`app-envelope.md`）是否与 `ERR_FRAME_*` 区分？**  
   - 建议： **`ERR_ENVELOPE_*`** 或归入 **`ERR_PAYLOAD_INVALID`**，避免与成帧层混淆。

## Environment Availability

| 依赖 | 用途 | 可用 | 版本 | 回退 |
|------|------|------|------|------|
| Go | `go test`、实现 | ✓ | 1.23.4（本机）；mod 1.22 | — |
| golangci-lint | 静态检查 | ✗ | — | CI/本地可选；**不阻塞** Phase 5 文档 |
| 外部网络服务 | 无 | — | — | 本阶段无需 |

**无阻塞缺失项。**

## Validation Architecture

> `.planning/config.json` 中 `workflow.nyquist_validation` 为 **true**，保留验证架构说明。

### REQ 标记（grep）

规范落点后建议：

```bash
rg 'STATE-01|ERR-01|SEC-01|REQ: STATE-01|REQ: ERR-01|REQ: SEC-01' docs/spec/v1/ .planning/
rg 'ERR_FRAME_TOO_LARGE|ERR_PROTO_VERSION|ERR_JOIN_DENIED|ERR_SESSION_NOT_FOUND' docs/spec/v1/ pkg/
```

### 文档路径（建议新增或修订）

| 文件（建议） | 内容 | REQ |
|--------------|------|-----|
| `docs/spec/v1/connection-state.md`（新）或并入 `transport-binding.md` 扩展章 | TRANS-01 成帧状态机 + TLS 上「应用数据已可用」前提 | STATE-01（连接部分） |
| `docs/spec/v1/session-state.md`（新）或 `session-create-join.md` 扩展 | CREATE/JOIN 成功/失败路径、成员关系 | STATE-01（session 部分） |
| `docs/spec/v1/errors.md`（新） | 错误码数值表、关闭顺序、与 TLS/TCP 关系 | ERR-01 |
| `docs/spec/v1/security-assumptions.md`（新） | TLS 边缘终止、中继信任、无 E2E | SEC-01 |
| `docs/spec/v1/README.md` | 索引上述文档 | — |
| `.planning/REQUIREMENTS.md` | STATE/ERR/SEC 勾选 Complete | 执行阶段 |

### 代码与 `go test`

- **现有：** `pkg/framing` — `ParseFrame` 与 `ERR_FRAME_TOO_LARGE` / `ERR_PROTO_VERSION` 对齐；Phase 5 若引入 **`pkg/.../errors` 或代码生成常量**，应保持**符号名与 `errors.md` 一致**。  
- **采样：** 每任务/波次 `go test ./...`；Phase 6 将扩大 golden 覆盖。  
- **追溯：** 测试注释引用 **ERR-01** 条目或规范章节链接。

### Phase Requirements → 测试映射（预览）

| Req ID | 行为 | 测试类型 | 自动化命令 | Phase 5 结束时文件 |
|--------|------|----------|------------|-------------------|
| STATE-01 | 文档内状态转移与引用一致 | 文档/审查 + 可选表驱动 | `rg` + 人工 | 以规范为主 |
| ERR-01 | 错误码与 `pkg/` 常量一致 | 单元 | `go test ./pkg/... -run Error -count=1` | 若新增常量则补测试 |
| SEC-01 | 威胁模型文档存在且与 REQUIREMENTS 一致 | 文档 | `rg SEC-01 security-assumptions` | — |

### Wave 0 缺口

- [ ] `docs/spec/v1/errors.md`（或等价）— **ERR-01**  
- [ ] 状态机终稿落点 — **STATE-01**  
- [ ] `security-assumptions.md` — **SEC-01**  
- [ ] （可选）统一 `pkg/framing` 或新包内 **数值码** 与规范同步  

## Sources

### Primary（HIGH）

- `docs/spec/v1/transport-binding.md` — TRANS-01、占位 `ERR_FRAME_TOO_LARGE`  
- `docs/spec/v1/version-capability.md` — `ERR_PROTO_VERSION`、关连接义务  
- `docs/spec/v1/session-create-join.md` — 控制面序列、Phase 5 错误应答占位  
- `docs/spec/v1/join-credentials.md` — `ERR_JOIN_DENIED`、`ERR_SESSION_NOT_FOUND`  
- `docs/spec/v1/routing-modes.md` — 非法路由与 Phase 5 ERR  
- `docs/spec/v1/streams-lifecycle.md` — 流状态、STREAM_CLOSE / FIN  
- `docs/spec/v1/app-envelope.md` — 信封错误与 Phase 5 `ERR_*`  
- `.planning/REQUIREMENTS.md` — STATE-01、ERR-01、SEC-01  
- `pkg/framing/decode.go` — 现有错误符号  

### Secondary（MEDIUM）

- [Go `crypto/tls` — `Conn.Close`、`Conn.CloseWrite`、`AlertError`](https://pkg.go.dev/crypto/tls) — 传输层关闭与告警类型  

### Tertiary（LOW）

- 具体 **TLS alert 编号 ↔ 协议 ERR_** 一一映射 — **不推荐**在 v1 写死（跨实现差异）

## Metadata

**置信度分解：**  
- 标准栈：**HIGH**（Go 标准库 + 现有代码）  
- 架构（三层状态机）：**HIGH**（与现有规范结构吻合）  
- 错误帧线格式：**MEDIUM**（待 05-02 拍板）  
- 威胁模型条目：**HIGH**（与 PROJECT/REQUIREMENTS 已锁定方向一致）

**调研日期:** 2026-03-29  
**建议复审:** 规范冻结后 **30** 天内或 TLS 相关依赖大版本升级时  

## RESEARCH COMPLETE
