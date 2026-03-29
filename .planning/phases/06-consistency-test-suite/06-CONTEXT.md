# Phase 6: 一致性测试套件 - Context

**Gathered:** 2026-03-29  
**Status:** Ready for planning  

<domain>
## Phase Boundary

交付 **可重复执行的 Go 测试** 与 **golden/负例向量**，满足 **TEST-01** 与 ROADMAP Phase 6 成功标准：`go test ./...` 可运行；`testdata/`（或等价物）含**多组**代表性帧与负例；覆盖 Phase 1–5 **已冻结**的关键行为（**成帧解析**、**路由/流相关的少量合成帧**、**错误路径**）。  

**不包含：** 新协议能力、Relay 实现、模糊测试/性能基准（除非后续里程碑单列）；不强制引入 **testify**（与当前 `go.mod` 一致）。

讨论结论由用户授权 **「你可自由选择」** —— 下列决策按 **推荐默认** 锁定，供 research / plan 直接使用。

</domain>

<decisions>
## Implementation Decisions

### Golden 向量载体与布局
- **D-01:** 以仓库根下 **`testdata/`** 为权威载体；按子系统分子目录（建议 **`testdata/framing/`**、**`testdata/appenvelope/`**），文件可用 **十六进制文本（`.hex`）** 或 **原始二进制（`.bin`）**；**文件名**表达场景（如 `empty_payload.hex`、`payload_too_large_prefix.hex`）。  
- **D-02:** **`_test.go`** 内保留 **表驱动** 仅用于极短向量或与现有 `decode_test.go` 风格一致的用例；**新增**跨场景 golden **优先落盘到 `testdata/`**，便于 diff 与增量扩展。  
- **D-03:** **`//go:embed`** 不作为 Phase 6 默认；若后续单包分发需要再评估。

### 覆盖范围与分层（相对 TEST-01）
- **D-04:** **第一层（必做）**：**`pkg/framing`** — `ParseFrame` / `AppendFrame` 往返、**半包 `ErrNeedMore`**、**`ErrFrameTooLarge`**、**`ErrProtoVersion`**；**`ErrCode`** 与 **`docs/spec/v1/errors.md`** 数值一致（延伸现有 `errors_test.go`）。  
- **D-05:** **第二层（必做）**：**`pkg/appenvelope`** — 在现有 `split_test.go` 基础上，对 **HAS_APP_ENVELOPE / 边界违反** 增加 **testdata 驱动** 用例（与 `app-envelope.md` 示例对齐者优先）。  
- **D-06:** **第三层（少而精）**：**路由前缀 + 流 opcode** — 使用规范文档中的 **完整帧十六进制**（如 `routing-modes.md` / `streams-lifecycle.md` 示例）做 **少量** 端到端字节级断言（**解析到 payload 后** 可只做长度与前缀检查，避免在本阶段实现完整控制面状态机）。  

### CI 与可重复性
- **D-07:** 增加 **最小 GitHub Actions workflow**（例如 `.github/workflows/go.yml`）：在 **push / pull_request** 上运行 **`go test ./... -count=1`**；**Go 版本**与 **`go.mod`** 的 `go` 指令一致（当前 **1.22**）。  
- **D-08:** 不在 Phase 6 **强制** 覆盖度百分比或 race detector；可作为后续改进。

### 负例与「错误路径」含义
- **D-09:** **优先锁实现已有语义**：解析层错误与 **`ErrCode` 常量表**；**不**在 Phase 6 凭空要求实现 **`PROTOCOL_ERROR` 帧编码 API** —— 若计划阶段为通过测试而引入 **小函数**（如编码控制消息 payload），则测试与实现 **同一提交** 落地。  
- **D-10:** **TLS/TCP 观测** 不做单元测试硬编码（与 **errors.md** 一致）；文档引用即可。

### Claude's Discretion
- 用户表示 **由助手自由选择**；**testdata 文件命名**、向量按 **06-01 / 06-02** 两个计划如何切分、以及是否在 Phase 6 引入 **最小的 `PROTOCOL_ERROR` 载荷编码** 以满足「错误路径」叙述，由 **plan-phase / 执行** 在遵守上述 D-01～D-10 前提下决定。

</decisions>

<canonical_refs>
## Canonical References

**Downstream agents MUST read these before planning or implementing.**

### 路线图与需求
- `.planning/ROADMAP.md` — Phase 6 目标、成功标准、计划条目 06-01 / 06-02  
- `.planning/REQUIREMENTS.md` — **TEST-01**  
- `.planning/PROJECT.md` — 交付顺序（规范 + 一致性测试）、Go 默认栈  

### 规范（帧、路由、流、错误）
- `docs/spec/v1/frame-layout.md` — FRAME-01  
- `docs/spec/v1/version-capability.md` — FRAME-02/03、版本拒绝与 ERR 引用  
- `docs/spec/v1/transport-binding.md` — TRANS-01、半包/粘包  
- `docs/spec/v1/routing-modes.md` — 广播/单播前缀与示例十六进制  
- `docs/spec/v1/streams-lifecycle.md` — 流 opcode、路由前缀衔接  
- `docs/spec/v1/app-envelope.md` — APP-01、边界条件  
- `docs/spec/v1/errors.md` — ERR-01、**ErrCode** 表、**PROTOCOL_ERROR** 布局（若实现编码则对齐）  

### 实现锚点
- `pkg/framing/decode.go` — `ParseFrame` / `AppendFrame`  
- `pkg/framing/errors.go` — `ErrCode`  
- `pkg/appenvelope/` — 信封切分  

</canonical_refs>

<code_context>
## Existing Code Insights

### Reusable Assets
- **`pkg/framing.ParseFrame` / `AppendFrame`** — 已有往返与 hex 示例测试（`decode_test.go`），可迁移或复制 hex 到 **`testdata/`** 后改为从文件读入。  
- **`pkg/framing.ErrCode` + `errors_test.go`** — 已与规范表对齐；新向量主要验证 **解析错误** 与 **边界**。  
- **`pkg/appenvelope`** — 已有 `split_test.go`；可追加文件驱动用例。  

### Established Patterns
- 测试使用 **标准库 `testing`** + **`encoding/hex`**（与现有一致），**不**默认引入 testify。  

### Integration Points
- 向量路径使用 **`testing`** 的相对路径或 `filepath` 相对 **`testdata`**；确保 **`go test ./...`** 在仓库根与子包均可运行。  

</code_context>

<specifics>
## Specific Ideas

- 用户未指定额外产品参考；授权 **「自由选择」** = 采用上文 **推荐默认**（testdata 为主、分三层覆盖、最小 CI、解析层错误优先）。

</specifics>

<deferred>
## Deferred Ideas

- **Property-based / fuzz**（`testing.F` / go-fuzz）— 可作为 Phase 6 之后或 **TEST-01** 扩展项。  
- **testify** — 若未来 `go.mod` 引入再统一风格。  
- **跨语言 golden（非 Go）** — 不属于本仓库 Phase 6 范围。

</deferred>

---

*Phase: 06-consistency-test-suite*  
*Context gathered: 2026-03-29*
