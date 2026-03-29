# Phase 4 — Technical Research

**Phase:** 4 — 可选应用信封（optional-app-envelope）  
**Researched:** 2026-03-29  
**Requirement IDs:** APP-01  
**Confidence:** MEDIUM（编码选型需在 04-01 规范中拍板；与 Phase 3 字节布局的衔接已基于现有 `docs/spec/v1/` 交叉核对）

<phase_requirements>
## Phase Requirements

| ID | Description | Research Support |
|----|-------------|------------------|
| APP-01 | 定义**可选应用信封**（至少 **content-type**、**请求 id**、**关联 id** 等字段占位），并说明与帧 **payload** 的边界 | 见下文「与 Phase 3 数据面的衔接」「出现条件与无信封语义」「编码方案比选」；成功标准 2 对应「示例与关联 id」小节 |
</phase_requirements>

## User Constraints

**None** — 本阶段无 discuss-phase 的 `CONTEXT.md`。规划与撰文须对齐现有 `docs/spec/v1/`（尤其 `streams-lifecycle.md` 中 `STREAM_DATA` 布局、`routing-modes.md` 路由前缀）以及 `.planning/REQUIREMENTS.md` 中 **APP-01**。

### Project Constraints（来自 `.cursor/rules/gsd-project.md`）

- **语言**：实现默认 **Go**；交付顺序：**规范 + 一致性测试** 优先。  
- **传输**：v1 **不采用 WebSocket**；主路径为 **TLS 上的自定义成帧**。  
- **顺序模型**：**流内有序、流间可乱序**（信封不改变 STREAM-02）。  
- **安全边界**：机密性主要依赖 **TLS**；信封为应用元数据，**不**承担 E2E 加密（与 SEC-01 后续文档一致）。

## Objective

回答：**在 Phase 3 已冻结的「`msg_type` + 路由前缀 + `stream_id` + `flags` + `payload_len` + `application_data`」数据面之上，如何定义可选应用信封，使 HTTP 式请求/响应元数据与 Copilot 式关联 id 能共存，且「无信封」路径与 Phase 3 语义一致、可测试。**

## Findings

### 与 Phase 3 数据面的衔接（边界）

- **外层不变**：10 字节帧头 + **整段 payload** 的切分仍由 `frame-layout.md` / `transport-binding.md` 定义（**FRAME-01、TRANS-01**）。  
- **数据面不变**：`STREAM_DATA`（`msg_type = 0x11`）在 `streams-lifecycle.md` 中：**路由前缀 18 字节**（`routing-modes.md`）→ **`stream_id` @18** → **`flags` @22** → **`payload_len` @23（uint16 BE，紧随其后的应用字节长度）** → **`application_data` @25**。  
- **信封位置（建议锁定的叙述）**：**应用信封（若存在）必须完全落在 `application_data` 所覆盖的字节范围内**，或与「`application_data` = 信封区 + 应用主体」的分解式定义等价地写清；**不得**向帧头或路由前缀插入字段（避免破坏 ROUTE/STREAM 已发布布局）。  
- **与「帧 payload」的边界（APP-01 条文）**：在文档中应显式区分三层：  
  1. **TLS 明文上的逻辑帧** = 帧头 + **payload**；  
  2. **payload** 内 **STREAM_DATA** 已解析部分 = 路由 + 流字段 + **`application_data` 原始字节**；  
  3. **可选应用信封** = 对 **`application_data`** 的进一步结构化（仅当规范声明存在时）。

### 出现条件与「无信封」语义（成功标准 3）

若不在规范中引入 **显式开关**，则「在 `application_data` 前加 `envelope_len`」会与 Phase 3 中「`application_data` 为不透明字节序列」冲突：旧实现会把信封长度误读为应用数据的一部分。

**推荐（HIGH confidence for design intent，须在 04-01 规范中写死）：**

- 使用 **`flags` 第 1 位**（**`HAS_APP_ENVELOPE`**；**bit 0** 已为 **FIN**，见 `streams-lifecycle.md`）表示本帧 **`application_data` 是否携带应用信封前缀**。  
- **`HAS_APP_ENVELOPE = 0`（默认）**：`application_data` 的语义 **与 Phase 3 完全一致** —— 全长 **`payload_len`** 字节均为应用 opaque 数据；**无额外前缀、无长度字段**。  
- **`HAS_APP_ENVELOPE = 1`**：`application_data` 解释为：**`envelope_len`（uint16 BE）** + **`envelope`（`envelope_len` 字节）** + **`body`（剩余 `payload_len - 2 - envelope_len` 字节）**；并规定 **`2 + envelope_len ≤ payload_len`**，否则 **协议错误**（具体码与 Phase 5 **ERR_*** 对齐）。  

这样 **无信封** 与 **有信封** 可通过 **同字段表 + 分支** 描述；测试侧可对 **相同路由/流上下文** 下 **`flags` 仅差 `HAS_APP_ENVELOPE`** 的两帧做对比（成功标准 3）。

> **注意**：若规划选择「始终带 2 字节 `envelope_len`（0 表示无信封）」而不使用 `flags`，则每帧多 2 字节开销且需在规范中说明与 Phase 3 文档的 **版本/次版本** 或 **新 msg_type** 关系 —— **不推荐** 作为默认方案，除非产品强制要求「单布局无分支」。

### 字段：content-type、request id、correlation id

| 字段 | 角色 | 建议 |
|------|------|------|
| **content-type** | 描述 `body` 的媒体类型（如 HTTP `Content-Type` 类比） | 字符串；可为 **MIME** 或产品约定 token；**可为空** 表示「未指定/二进制」由应用约定 |
| **request id** | 单次请求/操作在会话或连接内的标识 | **字符串或 uint64** 二选一写死；若跨实现互操作，**UTF-8 字符串**更稳妥 |
| **correlation id** | 关联多跳或异步应答（Copilot 管道、分布式 trace） | 与 **request id** 正交：**correlation** 贯穿多帧；**request** 标识单次调用；规范应用 **示例** 说明二者同时出现时的优先级 |

至少 **占位** 满足 APP-01；具体 **必选/可选** 在 04-01 规范表中列清。

### 编码方案比选（04-01 须择一）

| 方案 | 优点 | 缺点 | 与 v1 一致性 |
|------|------|------|----------------|
| **A. 二进制 TLV**（`tag` + `len` + `value`，大端） | 解析确定、无转义、易做 **最大信封长度** | 需维护 **tag 注册表**；人类可读差 | 与现有 **uint16/uint32 BE** 风格一致 |
| **B. JSON 对象**（UTF-8） | **HTTP/REST** 心智一致；示例易读 | 需 **边界**（最大长度、禁止嵌套深度）；`encoding/json` 解析成本 | 与帧级二进制并存，仍常见 |
| **C. 定长头 + 变长字符串** | 实现极快 | 扩展字段需改版 | 适合极简 v1，但扩展性差 |

**规划建议：** 在 **04-01** 中 **选定一种** 作为 **v1 唯一互操作编码**；若选 **JSON**，须规定 **schema 子集**（键名、类型）与 **最大 `envelope_len`**（例如 ≤ 4096，与帧最大 16 MiB 兼容）。若选 **TLV**，须在文档中给出 **至少三条 tag**（对应 content-type / request id / correlation id）及 **未知 tag 跳过规则**（与 capability 未知位「可忽略」哲学类似，但作用于信封内）。

### 示例场景（成功标准 2）

1. **JSON 请求/响应式**：一帧 **`STREAM_DATA`**，`HAS_APP_ENVELOPE=1`，`body` 为 **JSON**；`content-type` 为 `application/json`；**request id** 标识一次 RPC；响应帧复用 **同一 request id** 或成对字段，由规范给出一组 **十六进制** 或 **表格分解** 示例。  
2. **Copilot 往返式（correlation id）**：**A → B** 与 **B → A** 两帧共享 **correlation id**，**request id** 可区分分支；说明 **Relay 不解释信封**（透明转发），仅端点消费 —— 与 ROUTE 行为一致。

### 与 FRAME-03 capability

- 当前 **v1 capability 全 0**（`version-capability.md`）。**可选信封** 可通过 **`HAS_APP_ENVELOPE` 标志** 表达，**不必**新增帧头 capability 位即可互通；若未来要求「对端必须理解信封语义」，可再引入 **必须理解** 位（与 FRAME-03 扩展流程一致）。**本阶段建议不依赖新 capability**，降低 Phase 4 与 Phase 1–3 文档冲突面。

### 实现侧（Go）与「不要手搓」

| 问题 | 避免 | 建议 |
|------|------|------|
| JSON 信封 | 自写字符串拼接与转义 | **`encoding/json`** 解码到 `struct` 或 `map`，并设 **`Decoder` 限制**（或先校验 `envelope_len` 上限） |
| 数值字段 | 混用大小端 | **`encoding/binary.BigEndian`** 与规范一致 |
| 分帧 | 在信封层解决粘包 | **TRANS-01** 已解决 TLS 字节流成帧；信封仅解析 **已切出的 `application_data`** |

### 建议新增的 `docs/spec/v1/` 落点

| 文件（建议） | 主要内容 | REQ 标记 |
|--------------|----------|----------|
| `app-envelope.md`（新） | `HAS_APP_ENVELOPE`、`envelope_len`、编码（TLV 或 JSON 其一）、字段表、与 `application_data` / `body` 边界、无信封等价语义 | **APP-01** |
| `streams-lifecycle.md`（修订） | `flags` 位定义扩展（bit1）；`STREAM_DATA` 行表更新 | 引用 **APP-01** |
| `README.md`（修订） | 索引新文档 | — |

### 规划者检查清单（供 PLAN.md 使用）

1. **`flags`**：bit0 **FIN** 与 bit1 **`HAS_APP_ENVELOPE`** 无歧义；保留位发送 **MUST 0**、接收 **MUST 忽略未知位**（与现有 `flags` 规则一致）。  
2. **`payload_len` 与信封**：`payload_len` 仍为 **`application_data` 总长度**；信封+body 必须 **适配**该长度。  
3. **无信封**：`HAS_APP_ENVELOPE=0` 时，**与 Phase 3 示例帧字节级兼容**（回归测试可复用 Phase 3 向量）。  
4. **至少 2 个端到端示例**（JSON 风格 + correlation 往返），含 **分解表或十六进制**。  
5. **错误路径**：`envelope_len` 越界、JSON 无效等 —— 指向 Phase 5 **ERR_*** 占位或本文件 **临时占位名**。

## Validation Architecture

> `workflow.nyquist_validation` 在 `.planning/config.json` 中为 **true**，本阶段保留验证架构说明。

### REQ 标记（grep）

- 在 `docs/spec/v1/` 内为 **APP-01** 增加与 Phase 1 一致的 HTML 注释，例如：  
  `<!-- REQ: APP-01 -->`  
- **检索命令示例**：  
  `rg 'APP-01|REQ: APP-01' docs/spec/v1/`  
  `rg 'HAS_APP_ENVELOPE|app-envelope' docs/spec/v1/`

### 文档路径

- 新规范：建议 `docs/spec/v1/app-envelope.md`；修订 `docs/spec/v1/streams-lifecycle.md`、`docs/spec/v1/README.md`。  
- 需求追溯：`.planning/REQUIREMENTS.md` 中 **APP-01** 在规范落点后由执行阶段改为 Complete。

### 代码与测试

- **回归**：`go test ./...`（当前仓库含 `pkg/framing`；Phase 4 若在 `pkg/` 增加信封解析辅助，应对 **无信封 `STREAM_DATA`** 与 **有信封** 各补 **表驱动测试**）。  
- **Golden**：可在 Phase 6 统一收束；Phase 4 至少保证 **规范内示例** 可被抄成 **`[]byte` 常量** 做解码断言（与 **TEST-01** 衔接）。  
- **REQ 与代码**：若实现解析函数，建议在测试文件中以注释引用 **APP-01** 或规范章节，便于 Nyquist 追溯。

### Wave 0 缺口（相对 Phase 4）

- [ ] `docs/spec/v1/app-envelope.md` — 承载 **APP-01** 正文。  
- [ ] `streams-lifecycle.md` — **`flags` / `application_data`** 行与 **APP-01** 一致。  
- [ ] （可选）`pkg/...` 最小解析单元 + `go test` — **非** Phase 4 硬性前提（规范先行），但若 roadmap 04-02 含「实现示例」，则补齐。

## Sources

### Primary（HIGH）

- `docs/spec/v1/streams-lifecycle.md` — `STREAM_DATA` 偏移表、`flags`、`application_data`。  
- `docs/spec/v1/routing-modes.md` — 路由前缀与 `STREAM_DATA` 选用约定。  
- `docs/spec/v1/version-capability.md` — capability 与未知位规则。  
- `.planning/REQUIREMENTS.md` — APP-01 条文。

### Secondary（MEDIUM）

- Go 1.22+：`encoding/binary`、`encoding/json` 标准库（与 `go.mod` / 项目栈一致）。

### Tertiary（LOW）

- 业界「TLV vs JSON 元数据」无单一标准；**以 v1 互操作与实现成本为准** 在 04-01 拍板。

## Environment Availability

| Dependency | Required By | Available | Version | Fallback |
|------------|-------------|-----------|---------|----------|
| Go toolchain | `go test`、参考实现 | ✓ | go1.23.4（本机 `go version`）；`go.mod` 声明 1.22 | — |
| golangci-lint | 静态检查（项目栈提及） | ✗ | — | 依赖 CI/本地安装；**不阻塞** Phase 4 规范与 `go test` |

**无外部网络服务依赖** — Phase 4 规范与单元测试可在离线环境完成。

## RESEARCH COMPLETE
