# Feature Research

**Domain:** 公网 TLS 中继隧道（同 session 多 peer，v1 自定义成帧）
**Researched:** 2026-04-04
**Confidence:** HIGH（行为以 `docs/spec/v1/` 为准；与「同类产品」类比为 MEDIUM）

## Feature Landscape

### Table Stakes（本里程碑必须满足 — 否则闭环不成立）

用户/集成方对「最小 Relay + Client」的隐含预期：**在 TCP+TLS 上能开房、拉人、在同一 session 里发广播与私信**。与 v1 规范强绑定；缺失任一项即无法宣称「与 v1 一致」。

| Feature | Why Expected | Complexity | Notes |
|---------|--------------|------------|-------|
| **TLS 上成帧与版本门闸** | 所有控制面/数据面均依赖连续字节流切帧；对端版本不可接受时必须可失败闭合 | MEDIUM | 成帧：`connection-state.md`、`frame-layout.md`；版本/capability：`version-capability.md`；过大帧/坏版本 → `ERR_FRAME_TOO_LARGE` / `ERR_PROTO_VERSION`（`errors.md`） |
| **SESSION_CREATE：分配 session_id + invite_code** | 「开房」是可分享会话的前提 | LOW–MEDIUM | `session-create-join.md`：`SESSION_CREATE_REQ/ACK`；`session_id` 为 36 字符 UUID 串；`invite_code` 8–12 Base32 无填充 |
| **SESSION_JOIN：按 session_id 或 invite_code 加入** | 其它 peer 必须能凭带外凭证加入 | MEDIUM | `join_by` + `credential`；与 `join-credentials.md` 策略一致时可要求 token（本里程碑可实现「无 token」最小路径） |
| **JOIN 成功后分配 peer_id（非 0）** | 路由前缀必须标识源/目标 peer | LOW | `peer-identity.md`：`SESSION_JOIN_ACK` 携带 `uint64` BE；session 内唯一 |
| **数据面前置：仅在有有效 JOIN_ACK 后发送路由帧** | 否则 session 成员与路由语义未定义 | LOW（门禁逻辑） | `session-state.md`：未 ACK 前 **MUST NOT** 发送带路由前缀的数据面帧 |
| **创建者亦须「加入」才能发数据面** | 规范未在 CREATE_ACK 中分配 `peer_id`；数据面 `src_peer_id` 须为已分配身份 | MEDIUM（易误解） | 成功路径上创建者通常 **CREATE 后仍要对本 session 发 JOIN**（或产品层明确扩展；v1 文档以 JOIN_ACK 为门禁）。验证用例应覆盖 |
| **广播（BROADCAST）** | 协作场景默认「除自己外全员可见」 | MEDIUM | `routing-modes.md`：`routing_mode=1` 且 `dst_peer_id=0`；Relay 向同 session 其它成员各发一副本；**不得**回送发送者 |
| **单播（UNICAST）** | 点对点控制/回复 | MEDIUM | `routing_mode=2` 且 `dst_peer_id≠0`；Relay 只投递到目标连接；目标未知/非成员 → 丢弃或 `PROTOCOL_ERROR`（规范允许二选一策略） |
| **数据面使用 STREAM_DATA + 路由前缀 + stream 字段** | v1 示例与互操作以 `msg_type=0x11` 承载路由后数据 | MEDIUM | `routing-modes.md` + `streams-lifecycle.md`：路由前缀 18 字节后接 `stream_id` 等；至少需约定如何用 `STREAM_OPEN`/`STREAM_DATA` 跑通最小流（可固定单流/单 `stream_id` 以降低实现面） |
| **应用层错误：`PROTOCOL_ERROR`（0x05）** | 调用方能区分「找不到 session」「拒绝加入」「路由非法」等 | LOW–MEDIUM | `errors.md`：`err_code` + 可选 `reason`；JOIN 失败常用 `ERR_SESSION_NOT_FOUND`、`ERR_JOIN_DENIED`；路由矛盾 → `ERR_ROUTING_INVALID` |
| **可重复验证（测试或 CLI）** | 里程碑明确要求可演示、可回归 | LOW | 两 peer 同学号 session：双向广播与单播可见；失败路径有断言 |

### Differentiators（本域常见「加分项」— v1.1 **可不作为必选项**）

与「能用的中继」相比，下列项在本仓库 **已由规范覆盖**，但对 **最小实现** 可分期：先满足 table stakes，再逐步对齐完整流语义与信封。

| Feature | Value Proposition | Complexity | Notes |
|---------|-------------------|------------|-------|
| **双凭证（session_id 与 invite_code）** | 长码精确引用 / 短码人类分享 | LOW（Relay 侧映射表） | 协议已定义；最小实现可两条路径都测一遍 |
| **可选应用信封（JSON、`HAS_APP_ENVELOPE`）** | 请求-响应关联、与 Copilot/HTTP 风格管道对齐 | MEDIUM | `app-envelope.md`；**可与纯 opaque `application_data` 分阶段** |
| **流内有序 / 流间乱序** | 多路复用下的可预期顺序模型 | MEDIUM–HIGH | `streams-lifecycle.md`；最小闭环可用 **单流** 先跑通 |
| **细粒度 STREAM_OPEN/CLOSE 与 FIN** | 背压、半关闭、资源生命周期清晰 | MEDIUM | v1 完整语义；最小实现可先 **固定子集**（文档声明） |
| **Relay 对 `src_peer_id` 校验/改写** | 防伪造来源 | MEDIUM | 规范建议转发时携带 `src_peer_id`；是否强制校验属实现策略，建议最小实现 **校验与连接身份一致** 以避免互测歧义 |

### Anti-Features（本里程碑应 **不** 做或 **不** 冒充已做）

| Feature | Why Requested | Why Problematic | Alternative |
|---------|---------------|-----------------|-------------|
| **集群 / 持久化 session** | 高可用 | 与「最小 Relay」目标冲突，引入 CAP 与运维面 | `PROJECT.md` Out of Scope；后续里程碑 |
| **在 v1 主路径强依赖 WebSocket** | 浏览器友好 | v1 绑定 TCP+TLS 字节流；混用会分裂测试与成帧 | 浏览器接入列 v2+ |
| **宣称端到端加密** | 更强安全叙事 | v1 安全边界为 TLS 边缘终止（`security-assumptions.md`） | 明确文档与握手故事；E2E 另立项 |
| **无错误码静默失败** | 实现省事 | 集成方无法区分「拒收」「未找到」「路由错」 | 对可恢复场景发 `PROTOCOL_ERROR`；纯传输失败走 TLS/TCP 观测 |

## Feature Dependencies

```
[TLS 监听 + 成帧解析]
    └──requires──> [SESSION_CREATE / SESSION_JOIN 控制面]
                           └──requires──> [peer_id 与 session 成员表]
                                              └──requires──> [JOIN_ACK 后门禁]
                                                       └──requires──> [STREAM_DATA + 路由前缀 + 流字段]
                                                                              └──enhances──> [可选 APP_ENVELOPE / 多流]

[ERR_FRAME_TOO_LARGE / ERR_PROTO_VERSION]
    └──requires──> [成帧层先于 session 语义]

[广播 / 单播]
    └──requires──> [同 session 多连接映射 + routing_mode 与 dst 联合判定]
```

### Dependency Notes

- **JOIN_ACK 门禁 → 数据面：** `session-state.md` 规定顺序；路线图应先实现控制面与成员表，再实现转发。
- **CREATE  alone ≠ 可发数据：** 无 `peer_id` 则无法满足路由前缀语义；依赖 **JOIN**（或后续规范扩展，当前未定义）。
- **广播/单播 → 成员与连接映射：** Relay 必须能由 `dst_peer_id` 找到 TLS 连接；单播失败路径依赖该索引是否正确。
- **应用信封：** 依赖 `STREAM_DATA` 与 `flags`；可与「仅 opaque body」并行作为第二阶段。

## MVP Definition

### Launch With（v1.1 最小闭环）

- [ ] **Relay：** TCP 监听、TLS、`SESSION_CREATE` / `SESSION_JOIN`、session ↔ peer 映射、**STREAM_DATA** 的广播与单播转发（含不回送广播给自己）。
- [ ] **Client：** 建连、**CREATE →（同连接）JOIN**、记录 `peer_id`、发送/接收带路由前缀的 **STREAM_DATA**（可先 **单 stream_id** 策略并文档化）。
- [ ] **错误路径：** JOIN 失败返回 `PROTOCOL_ERROR`（`ERR_SESSION_NOT_FOUND` / `ERR_JOIN_DENIED` 等）；路由非法 → `ERR_ROUTING_INVALID` 或丢弃/关连（与规范允许策略一致）。
- [ ] **验证：** `go test` 和/或 CLI：两 peer 演示广播与单播。

### Add After Validation（v1.1 之后 — 仍属 v1 规范内、可排期）

- [ ] 完整流生命周期（OPEN/CLOSE/FIN 组合）、多流并发与顺序断言 — **当** 单流路径稳定。
- [ ] 应用信封与 golden 扩展 — **当** 需要与上层请求关联 id 对接。
- [ ] Join token 强制策略 — **当** Relay 部署需要防扫号/滥用。

### Future Consideration（v2+）

- [ ] WebSocket/WebTransport 承载、浏览器 — `PROJECT.md` 已列。
- [ ] 生产级观测、配额、多租户 — 非本里程碑。

## Feature Prioritization Matrix

| Feature | User Value | Implementation Cost | Priority |
|---------|------------|---------------------|----------|
| CREATE + JOIN + peer_id | HIGH | MEDIUM | P1 |
| JOIN 后门禁与错误码 | HIGH | LOW | P1 |
| 广播转发（不回送） | HIGH | MEDIUM | P1 |
| 单播转发与「目标不存在」策略 | HIGH | MEDIUM | P1 |
| 成帧 + 版本错误路径 | HIGH | MEDIUM | P1 |
| STREAM_OPEN/CLOSE 全语义 | MEDIUM | HIGH | P2（最小可先子集） |
| 应用信封 | MEDIUM | MEDIUM | P2 |
| 集群/持久化 | LOW（对本里程碑） | HIGH | P3 |

**Priority key:**

- P1：本里程碑必须交付，否则无法宣称「最小 Relay + Client」与 v1 一致。
- P2：规范已有，实现可分阶段。
- P3：明确不在 v1.1 范围。

## Competitor Feature Analysis

| Feature | 典型 SFU/聊天中继 | 典型信令服务器 | Our Approach（v1.1） |
|---------|-------------------|----------------|----------------------|
| 会话创建 | Room ID / token | 房间名或频道 | `SESSION_CREATE_ACK` 分配 `session_id` + `invite_code`（`session-create-join.md`） |
| 加入校验 | Ticket、权限 | Token、密码 | `SESSION_JOIN_REQ` + 可选 join token（`join-credentials.md`）；最小实现可先无 token |
| 消息模式 | DataChannel、fan-out | Topic、DM | 显式 `routing_mode` + `dst_peer_id` 联合判定（`routing-modes.md`） |
| 错误语义 | 多为应用自定义 | HTTP/自定义 close code | 统一 `PROTOCOL_ERROR` + `err_code`（`errors.md`） |

## Sources

- 权威：`docs/spec/v1/session-create-join.md`、`routing-modes.md`、`session-state.md`、`errors.md`、`connection-state.md`、`streams-lifecycle.md`、`peer-identity.md`、`join-credentials.md`
- 项目范围：`.planning/PROJECT.md`（v1.1 里程碑与 Out of Scope）

---
*Feature research for: Tunnel — Minimal Relay + Client（v1.1）*
*Researched: 2026-04-04*
