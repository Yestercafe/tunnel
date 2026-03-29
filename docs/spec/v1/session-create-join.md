# Session 创建与加入（SESS-01、SESS-02）

## 范围

本规范定义 **控制面** 消息：在已建立的 **TLS 字节流** 上，通过 Phase 1 定义的 **逻辑帧** 传送；**payload** 的边界与长度语义见 [frame-layout.md](./frame-layout.md)（`FRAME-01`）。本文件只定义 **payload 内部** 的 **msg_type** 与消息体结构。

## Payload 与控制消息的关系

- 每一帧的 **payload**（`payload_len` 字节，紧跟 10 字节固定帧头）是**不透明字节序列**。
- **控制消息**（本规范及后续会话相关规范）将 payload 解析为：**第 1 字节 `msg_type`（`uint8`）** + **消息体（message body）**。
- **多字节整数字段**除特别声明外均为 **大端（big-endian）**，与 `frame-layout.md` 一致。

```
payload 布局（控制消息）:

  偏移 0     |  偏移 1 …
  -----------+----------------------------------
  msg_type   |  message body（依 msg_type 解析）
  (uint8)    |
```

## 标识符格式

### session_id

- **类型**：UTF-8 编码的 **UUID 字符串**（canonical textual representation）。
- **长度**：**36** 个 ASCII 字符，形式 `8-4-4-4-12` 十六进制数字与小写连字符，例如 `550e8400-e29b-41d4-a716-446655440000`。
- **约束**：与下列正则等价（实现可用等价校验）：

  ```regex
  ^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$
  ```

- **语义**：由 **Relay** 在 **SESSION_CREATE** 成功时分配，在 **同一 Relay 部署** 内应可唯一引用一个 **session**（具体唯一性范围以实现为准；对 client 表现为不透明字符串）。

### invite_code（邀请码）

- **类型**：仅含 **Base32 字母表**（RFC 4648）**无填充** 的 ASCII 字符串；实现 MUST NOT 包含 `=` 填充。
- **长度**：**8～12** 个字符（含端点）。
- **语义**：由 Relay 在创建 session 时生成，与 `session_id` **一一映射**（在 Relay 策略允许公开短码的前提下供分享）；与 `session_id` 二选一或同时使用由产品策略决定，但协议层两种都可用于 **加入**。

## Opcode 表（msg_type）

| 值（hex） | 名称 | 方向 | 说明 |
|-----------|------|------|------|
| `0x01` | `SESSION_CREATE_REQ` | Client → Relay | 请求创建 session |
| `0x02` | `SESSION_CREATE_ACK` | Relay → Client | 创建成功，返回 `session_id` 与 `invite_code` |
| `0x03` | `SESSION_JOIN_REQ` | Client → Relay | 请求加入已有 session |
| `0x04` | `SESSION_JOIN_ACK` | Relay → Client | 加入成功，分配 **peer_id**（见 [peer-identity.md](./peer-identity.md)，`SESS-03`） |
| `0x05` | `PROTOCOL_ERROR` | 双向 | **应用层错误指示**（载荷见下表；权威定义见 [errors.md](./errors.md)，**ERR-01**） |

保留：`0x00` 不得作为有效控制消息类型（可与实现约定为「未初始化」探测，不在本规范强制）；**`0x06`–`0x0F`** 预留给后续控制消息扩展（**`0x05`** 已分配给 **`PROTOCOL_ERROR`**，见 [errors.md](./errors.md)）。

## 消息体字段表

### SESSION_CREATE_REQ (`0x01`)

| 字段 | 长度（字节） | 类型 | 说明 |
|------|-------------|------|------|
| （空） | 0 | — | v1 最小实现体可为空；扩展字段由后续版本定义 |

### SESSION_CREATE_ACK (`0x02`)

| 字段 | 长度（字节） | 类型 | 说明 |
|------|-------------|------|------|
| `session_id_len` | 2 | `uint16` BE | **必须**为 **36** |
| `session_id` | `session_id_len` | UTF-8 | **必须**符合上文 **session_id** 格式 |
| `invite_code_len` | 1 | `uint8` | **必须**在 **8～12** |
| `invite_code` | `invite_code_len` | ASCII | **必须**符合上文 **invite_code** 格式 |

### SESSION_JOIN_REQ (`0x03`)

| 字段 | 长度（字节） | 类型 | 说明 |
|------|-------------|------|------|
| `join_by` | 1 | `uint8` | **0** = 按 `session_id` 加入；**1** = 按 `invite_code` 加入 |
| `credential_len` | 2 | `uint16` BE | 下列凭证字节数 |
| `credential` | `credential_len` | UTF-8 或 ASCII | `join_by=0` 时为 **session_id** 字符串；`join_by=1` 时为 **invite_code** |

可选 **join token**（若 Relay 策略要求）不在本表展开，见 [join-credentials.md](./join-credentials.md)（`SESS-04`）。

### SESSION_JOIN_ACK (`0x04`)

| 字段 | 长度（字节） | 类型 | 说明 |
|------|-------------|------|------|
| `peer_id` | 8 | `uint64` BE | Relay 分配的 **peer 标识**；**0** 保留，ACK 中 MUST NOT 为 **0**（语义见 `SESS-03`） |

### PROTOCOL_ERROR (`0x05`)

与 [errors.md](./errors.md) 一致：payload 第 1 字节为 **`msg_type = 0x05`**，随后为 **`err_code`（uint16 BE）**、**`reason_len`（uint16 BE）**、可选 **`reason`（UTF-8）**；**无** 数据面路由前缀。

成功路径下，client 在收到 `SESSION_JOIN_ACK` 后即视为 **已加入 session**，可使用后续 Phase 定义的数据面消息。

## Relay 最小成员表（实现推导）

Relay 为维持 session 与连接映射，**至少**需能索引下列字段（名称逻辑，非 SQL 约束）：

| 字段 | 说明 |
|------|------|
| `session_id` | 会话主键（字符串） |
| `invite_code` | 与 session 映射的短码（若启用） |
| `peer_id` | 成员在 **该 session 内** 的唯一 `uint64`（见 `SESS-03`） |
| 连接句柄 | 与 **TLS 连接** 的绑定（同一连接在同一 session 内对应一个 `peer_id`） |

## 创建与加入序列（最小往返）

下列步骤省略 TLS 握手与成帧细节（见 [transport-binding.md](./transport-binding.md)）。

```
创建者 (A)                    Relay                    加入者 (B)
    |                            |                            |
    |-- 帧: SESSION_CREATE_REQ --->|                            |
    |<-- 帧: SESSION_CREATE_ACK ---|                            |
    |   (session_id, invite_code)  |                            |
    |                            |                            |
    | ....  通过带外方式分享 session_id 或 invite_code ....    |
    |                            |                            |
    |                            |<-- 帧: SESSION_JOIN_REQ ---|
    |                            |    (credential)            |
    |                            |--- 帧: SESSION_JOIN_ACK -->|
    |                            |    (peer_id)               |
```

**编号步骤：**

1. **A** 发送一帧，payload 以 `msg_type = 0x01`（`SESSION_CREATE_REQ`）开头。
2. **Relay** 分配 `session_id` 与 `invite_code`，回复 `msg_type = 0x02`（`SESSION_CREATE_ACK`），体字段见上表。
3. **A** 将 `session_id` 或 `invite_code` 交给 **B**（带外）。
4. **B** 发送 `SESSION_JOIN_REQ`（`0x03`），`join_by` 与 `credential` 与所选标识一致。
5. **Relay** 校验通过后回复 `SESSION_JOIN_ACK`（`0x04`），携带 **`peer_id`**；**B** 记录后加入完成。

**会话层状态转移**（未加入 / 已 ACK、数据面前置、失败路径占位错误名）的 **终稿** 见 **[session-state.md](./session-state.md)**。

错误应答与关闭原因见 **[errors.md](./errors.md)**（**ERR-01**）：实现 **SHOULD** 使用 **`PROTOCOL_ERROR`（`0x05`）** 携带 **`err_code`**（如 **`ERR_JOIN_DENIED`**、**`ERR_SESSION_NOT_FOUND`**）。

---

<!-- REQ: SESS-01 -->
<!-- REQ: SESS-02 -->
