# Session 级成员状态（STATE-01 — 会话 / 成员层）

## 范围

- 本文件描述：在 **单条 peer↔Relay TLS 连接** 上，与 **session 成员关系** 相关的 **控制面** 状态与转移（**非**成帧半包、**非**单条 `stream_id` 的 OPEN/CLOSE）。
- **TLS 与成帧**前提见 [transport-binding.md](./transport-binding.md) 与 [connection-state.md](./connection-state.md)；**CREATE/JOIN 消息体与 opcode** 见 [session-create-join.md](./session-create-join.md)；**凭证与 ERR 占位** 见 [join-credentials.md](./join-credentials.md)。

## 主路径（成功）

| 阶段 | 说明 |
|------|------|
| **未创建 / 未加入** | 连接已建立 TLS 且成帧可解析，但本连接 **尚未** 完成 **SESSION_CREATE** 成功路径，或 **尚未** 完成 **SESSION_JOIN** 成功路径（依角色：创建者 vs 加入者）。 |
| **已创建 session（创建者）** | 收到 **`SESSION_CREATE_ACK`**（`msg_type = 0x02`），获得 **`session_id`**（及可选 `invite_code`）。 |
| **已加入 session** | 收到 **`SESSION_JOIN_ACK`**（`msg_type = 0x04`），且 **`peer_id` 有效**（非 0，见 [peer-identity.md](./peer-identity.md)）。此后本连接在该 session 内具备 **peer 身份**。 |

### 数据面前置条件（门禁）

- 在 **收到 `SESSION_JOIN_ACK` 且 `peer_id` 有效** 之前，**MUST NOT** 发送带 [routing-modes.md](./routing-modes.md) 所定义 **数据面路由前缀** 的帧（即「未 ACK 前不得发送数据面路由帧」）。
- 流上的 OPEN/DATA/CLOSE 语义见 [streams-lifecycle.md](./streams-lifecycle.md)；**未 JOIN_ACK 前** 不得将 **数据面** `STREAM_*` 与 **session 成员** 语义混用为「已可路由」。

## 失败路径（占位错误名）

| 占位名 | 语义（摘要） |
|--------|----------------|
| **`ERR_JOIN_DENIED`** | 凭证 / token 拒绝、策略不允许加入等（细节见 [join-credentials.md](./join-credentials.md) 与 [session-create-join.md](./session-create-join.md)）。 |
| **`ERR_SESSION_NOT_FOUND`** | 请求的 `session_id` **或** `invite_code` 在 Relay 侧无法解析为有效 session。 |

数值码与 on-wire 错误指示在 **Phase 5** **[errors.md](./errors.md)**（ERR-01）统一；本文件保留 **符号名** 与 **会话层语义**。

## 与流层的关系

- **`SESSION_*` 控制消息**（`0x01`–`0x04`）**不得** 携带 `stream_id` 或数据面路由前缀中的流字段 —— 见 [session-create-join.md](./session-create-join.md) 与 [streams-lifecycle.md](./streams-lifecycle.md)。
- **流** 生命周期（`STREAM_OPEN` / `STREAM_DATA` / `STREAM_CLOSE`）**仅** 在 **streams-lifecycle.md** 中定义；**session 状态** 与 **单条流状态** 分属不同层，**不得** 在同一枚举中混用 `stream_id`。

## 与路由模式的关系

- 数据面帧的 **`routing_mode` / `src_peer_id` / `dst_peer_id`** 语义见 [routing-modes.md](./routing-modes.md)；发送此类帧前须满足上文 **JOIN_ACK** 门禁，否则接收方 / Relay **可** 拒绝或按实现策略关闭连接（具体错误码见 **errors.md**）。

---

<!-- REQ: STATE-01 -->
