# Phase 1: 协议基础 — 帧与 TLS 字节流承载 - Context

**Gathered:** 2026-03-29  
**Status:** Ready for planning  

<domain>
## Phase Boundary

交付 **v1 二进制帧** 的**可实现的布局**、**协议版本与 capability 语义**、以及在 **TLS（TCP）连续字节流**上的**成帧/解析规则**（含粘包/半包）。对应 REQUIREMENTS：**FRAME-01、FRAME-02、FRAME-03、TRANS-01**。  
不包含：session 路由、流 ID、应用信封、完整错误码表（后续阶段）。

讨论范围仅澄清 **本阶段规范如何写、写到多细**；不新增能力。

</domain>

<decisions>
## Implementation Decisions

### 字节序与数值编码
- **D-01:** 全帧多字节整数均为 **big-endian（网络字节序）**；与 Go `encoding/binary.BigEndian` 对齐，便于实现与文档对照。
- **D-02:** **不强制** 4 字节魔数；若后续在规范中增加魔数，须放在帧最前且与 TLS 解密后首字节对齐说明。**v1 草案可先长度前缀为主**，魔数列为可选扩展（capability 位控制）。

### 长度字段与帧边界（FRAME-01 / TRANS-01）
- **D-03:** **长度字段表示「固定头之后的 payload 字节数」**（不含头）；固定头长度在规范中单点定义，避免「总长」与「payload」混用。
- **D-04:** 长度字段类型为 **uint32**；**最大单帧 payload** 上限 **16 MiB（16777216 字节）**；超过则 **协议错误**，连接策略在 Phase 5 细化，本阶段仅命名占位错误（如 `FRAME_TOO_LARGE`）。
- **D-05:** 解析器在 **半包** 时保留缓冲；**粘包** 时循环解析直至缓冲区不足一帧；与 `01-RESEARCH.md` 一致。

### 版本（FRAME-02）
- **D-06:** 使用 **单一 `uint16` protocol version**（单调递增或语义分段由规范表定义）；**不支持的主版本/区间** → **必须断开**并带错误码占位；同一大版本内 **次版本**差异：规范写「忽略未知扩展」策略。

### Capability（FRAME-03）
- **D-07:** **uint32 capability 位图**；**未定义且被对端置位的位** → **必须忽略**（forward compatibility），除非该位在规范中标记为「必须理解」（v1 不设此类位）。
- **D-08:** v1 **不分配**具体业务 capability 位亦可，但位图字段**占位**并在文档中列「保留，必须为 0」。

### 规范文档形态（与现有 3 份 PLAN 对齐）
- **D-09:** 规范拆为 **`docs/spec/v1/`** 下三文件：**`frame-layout.md`**、**`version-capability.md`**、**`transport-binding.md`**，并由 **`README.md`** 索引；与 `01-01`～`01-03-PLAN.md` 一致。
- **D-10:** 文档正文语言 **中文为主**，字段名、类型名、错误码标识符 **英文**（便于代码与互操作）；示例十六进制可中英注释。

### Claude's Discretion
- 固定头各字段的**精确字节偏移**、示例帧的**具体十六进制**、是否增加 **4 字节魔数**、**uint16 版本**的取值表（仅 0x0001 等）由规划/撰写规范时细化，只要不违反上述决策。

</decisions>

<canonical_refs>
## Canonical References

**Downstream agents MUST read these before planning or implementing.**

### 项目与阶段范围
- `.planning/PROJECT.md` — 传输假设（TCP+TLS、无 WebSocket v1）、交付顺序  
- `.planning/REQUIREMENTS.md` — FRAME-01～03、TRANS-01 条文  
- `.planning/ROADMAP.md` — Phase 1 目标与成功标准  

### 本阶段研究
- `.planning/phases/01-tls/01-RESEARCH.md` — 成帧、版本、capability 推荐与验证思路  

### 已存在的计划（待按本 CONTEXT 重规划时可覆盖）
- `.planning/phases/01-tls/01-01-PLAN.md`  
- `.planning/phases/01-tls/01-02-PLAN.md`  
- `.planning/phases/01-tls/01-03-PLAN.md`  

### 外部标准（引用即可，非仓库内文件）
- IETF RFC 8446（TLS 1.3）— 传输安全上下文  
- Go `encoding/binary` / `crypto/tls` — 参考实现侧惯例  

</canonical_refs>

<code_context>
## Existing Code Insights

- **无 Go 源码**：仓库尚无 `go.mod` / 实现包；Phase 1 以 **Markdown 规范**为主，可选 `pkg/framing` 与 `01-03-PLAN` 一致。  
- **可复用资产：** 无既有组件。  

### Established Patterns
- 无  

### Integration Points
- 未来 `internal/` 或 `pkg/framing` 应消费 `docs/spec/v1/*.md` 中定义的帧布局。  

</code_context>

<specifics>
## Specific Ideas

- 用户此前表示对技术方向 **暂不深入**，接受 **默认工程化决策**；若后续要魔数、或总长而非 payload 长度，可改 **D-02/D-03** 并 **replan**。  

</specifics>

<deferred>
## Deferred Ideas

- **WebSocket / WebTransport 适配** — REQUIREMENTS v2 **TRANS-02**  
- **Session、路由、流 ID** — Phase 2–3  
- **完整错误码与 TLS alert 映射** — Phase 5  

### Reviewed Todos (not folded)
- 无（`todo match-phase` 无匹配）  

</deferred>

---

*Phase: 01-tls*  
*Context gathered: 2026-03-29*
