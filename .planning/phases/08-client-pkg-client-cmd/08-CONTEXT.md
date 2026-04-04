# Phase 8: Client（pkg/client + cmd）- Context

**Gathered:** 2026-04-04  
**Status:** Ready for planning  

<domain>
## Phase Boundary

交付 **`pkg/client`** 与 **`cmd`** 入口，使 Client 在 **TCP+TLS** 上完成 **SESSION_CREATE**、**SESSION_JOIN**，在 **SESSION_JOIN_ACK** 之后能发送并接收带路由前缀的 **STREAM_DATA**，并至少各有一条**可重复的广播与单播**路径；`stream_id` 等策略须**文档化**。  

本阶段**不**实现完整 Relay；**两 Client + 真 Relay 的可重复 E2E 演示**属于 **Phase 11（E2E-01）** 的验收面。产品经理仅关注最终 E2E 效果时，以 **Phase 11** 为「一眼能看懂」的里程碑；Phase 8 负责交付**可被 E2E 直接复用**的客户端库与**可自动化**的验证路径。

讨论方式：**用户将 Phase 8 全部实现取向委托给实现方**，下列决策为锁定结论，供 research / plan / execute 直接使用。

</domain>

<decisions>
## Implementation Decisions

### E2E 叙事与阶段分工
- **D-01:** **对外可展示的「完整 E2E」**（两 peer、同 session、广播 + 单播、CI 可跑）以 **Phase 11** 为正式验收；Phase 8 **不**以此作为完成标准，但必须避免阻塞：公共 API 与测试策略须让 Phase 9–11 **无需推翻 Client 外形**即可接 Relay 与 E2E。
- **D-02:** Phase 8 的 **CLNT-01..03** 证明方式：在仓库内使用 **测试专用 minimal peer**（fake / harness，**非**生产 Relay），在 **`go test`** 中与真实 **TLS + 成帧 + `pkg/protocol`** 路径一致，行为对齐 **`docs/spec/v1/`**。

### 集成与验证策略（无 Relay 阶段）
- **D-03:** **主验证** = `go test ./...`，**不**引入 Docker、不依赖外部预置二进制；fake 仅实现本阶段所需的**最小控制面/数据面子集**（SESSION_CREATE、SESSION_JOIN、JOIN_ACK、STREAM_DATA 广播/单播所需字段）。
- **D-04:** Fake 的放置与命名由 plan-phase 决定（例如 `internal/...` 或 `pkg/client/..._test.go` 旁），但必须**可读、可维护**，避免与 Phase 9 的 `pkg/relay` **包名冲突**（不得叫 `relay` 冒充生产实现）。

### CLI / 可执行入口
- **D-05:** 单二进制入口（**一个 `main` 包**），子命令结构由 plan-phase 定稿；**Phase 8** 仅覆盖 **client 相关子命令**（例如 `tunnel client create` / `tunnel client join` 或等价命名），用于**人工冒烟**与**脚本**；**规范验收以测试为准**，CLI 为辅助。
- **D-06:** 必备标志概念：`--addr`（或环境变量，与实现一致）、TLS：**默认系统根 CA** + **开发用跳过校验**（见下）；超时与取消通过 **`context`** 语义在库层支持。

### TLS 与信任锚
- **D-07:** 默认使用 **系统证书池** 校验服务端证书；提供 **显式开发开关**（如 `--insecure-skip-verify` 或等价命名），**禁止**静默跳过校验。
- **D-08:** **mTLS / 客户端证书** 不在 Phase 8；与 **`docs/spec/v1/security-assumptions.md`** 一致。

### `pkg/client` API 与错误模型
- **D-09:** 面向调用方暴露 **`Client`（或等价）类型**；连接与会话操作为 **同步 API**，**`context.Context`** 用于 **Dial / 读写 / 会话握手** 的取消与超时；与 **`net.Conn` 生命周期**清晰可分。
- **D-10:** 错误返回须能**关联到协议语义**：至少能区分 **TLS/IO**、**成帧**、**控制面/PROTOCOL_ERROR（ErrCode）**；具体类型命名由实现定，但**不得**吞掉 **`pkg/framing.ErrCode`** / **`pkg/protocol`** 可给出的信息。
- **D-11:** **广播**与**单播**演示各使用**固定 `stream_id`**（建议：**广播 `1`、单播 `2`**，若与现有示例冲突可在实现时调整唯一约束：**必须写进包注释 + 本仓库 `docs/` 一处**），满足路线图「策略已文档化」。

### 测试风格
- **D-12:** 延续 Phase 6：**标准库 `testing`**，**不**默认引入 **testify**；表驱动 + 必要时 **`testdata/`**。

### Claude's Discretion
- **子命令字面量**、**包路径**（`internal/` 划分）、**fake server 的具体结构体名**、**是否提供极简 `stdin/stdout` demo 子命令**，由 plan-phase / 执行在遵守 **D-01～D-12** 前提下决定。

</decisions>

<canonical_refs>
## Canonical References

**Downstream agents MUST read these before planning or implementing.**

### 路线图与需求
- `.planning/ROADMAP.md` — Phase 8 目标、成功标准、Depends on Phase 7  
- `.planning/REQUIREMENTS.md` — **CLNT-01, CLNT-02, CLNT-03**  
- `.planning/PROJECT.md` — v1.1 范围、TCP+TLS、交付顺序  

### 会话与控制面
- `docs/spec/v1/session-create-join.md` — SESSION_CREATE / SESSION_JOIN / ACK 语义  
- `docs/spec/v1/join-credentials.md` — `session_id`、邀请码  
- `docs/spec/v1/peer-identity.md` — `peer_id`  
- `docs/spec/v1/connection-state.md` — 连接与会话状态  
- `docs/spec/v1/session-state.md` — JOIN_ACK 前后行为（与 **join_gate** 一致）  

### 数据面与路由
- `docs/spec/v1/routing-modes.md` — 广播 / 单播前缀  
- `docs/spec/v1/streams-lifecycle.md` — `stream_id`、`STREAM_DATA`  

### 传输与错误
- `docs/spec/v1/transport-binding.md` — TCP+TLS 字节流、半包  
- `docs/spec/v1/frame-layout.md` — 帧布局（与 `pkg/framing` 边界）  
- `docs/spec/v1/errors.md` — **PROTOCOL_ERROR**、`ErrCode`  

### 安全假设
- `docs/spec/v1/security-assumptions.md` — TLS 边缘终止、威胁模型  

### 实现锚点（已有代码）
- `pkg/framing/` — 成帧解析与 **`ErrCode`**  
- `pkg/protocol/` — PROT-01 / PROT-02 **`join_gate`**  

</canonical_refs>

<code_context>
## Existing Code Insights

### Reusable Assets
- **`pkg/framing`**：`ParseFrame` / `AppendFrame`、半包 **`ErrNeedMore`**。  
- **`pkg/protocol`**：控制面/数据面 payload 视图、**`JoinGateAllowsBusinessDataPlane`**（JOIN_ACK 前数据面门禁）。  

### Established Patterns
- **测试**：标准库 **`testing`**，Phase 6 起 **testdata 驱动** 与表驱动；**不**默认 testify。  
- **模块**：`go 1.22`，模块路径 **`tunnel`**。  

### Integration Points
- Client 在字节流上叠 TLS 后送入 **`framing` 循环**；payload 交给 **`protocol`** 解码再进入会话状态机。  
- Phase 9 **`pkg/relay`** 将替换测试 fake，**`pkg/client` 公共 API 应保持稳定**。  

</code_context>

<specifics>
## Specific Ideas

- 产品经理仅关注 **E2E 最终效果**；Phase 8 的细粒度工程权衡由实现方按本文 **`decisions`** 执行，**无需再向产品经理确认实现细节**。  

</specifics>

<deferred>
## Deferred Ideas

- **两 Client + 生产级 Relay + CI E2E** — Phase 11（**E2E-01**），不在 Phase 8 范围。  

### Reviewed Todos (not folded)

- （本阶段 **todo match-phase** 无匹配项。）  

</deferred>

---

*Phase: 08-client-pkg-client-cmd*  
*Context gathered: 2026-04-04*  
