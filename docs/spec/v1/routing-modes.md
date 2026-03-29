# 路由模式

## 广播（ROUTE-01）

### 范围

本规范定义 **数据面** payload 内 **路由前缀** 的语义与字段布局。逻辑帧外层（10 字节固定帧头、`payload_len`）见 [frame-layout.md](./frame-layout.md)（`FRAME-01`）；**peer_id** 分配与保留值见 [peer-identity.md](./peer-identity.md)（`SESS-03`）；**控制面** `msg_type` 与 `SESSION_*` 见 [session-create-join.md](./session-create-join.md)。

### 数据面与控制面

- 每一帧 **payload** 仍以 **第 1 字节 `msg_type`（`uint8`）** 开头，与 `session-create-join.md` 一致。
- **`0x01`–`0x04`** 保留给 **`SESSION_*`** 控制消息；数据面帧 **MUST NOT** 复用这些取值。
- **`0x10`–`0x12`** 预留给 **流生命周期** opcode（`STREAM_OPEN` / `STREAM_DATA` / `STREAM_CLOSE`），定义见 [streams-lifecycle.md](./streams-lifecycle.md)。
- 承载 **「路由前缀 + 可选流字段 + 应用数据」** 的数据面帧，在示例与互操作说明中 **MUST** 使用 **`msg_type = 0x11`（`STREAM_DATA`）**，以避免与 **`0x10`（`STREAM_OPEN`）** 冲突。

### 路由前缀（固定顺序，大端）

紧跟 **`msg_type`** 之后为 **路由前缀**（共 **17** 字节，与 `msg_type` 合计 **18** 字节起可接流前缀，见 `streams-lifecycle.md`）：

| 字段 | 偏移（相对 payload 起点） | 长度（字节） | 类型 | 说明 |
|------|---------------------------|-------------|------|------|
| `msg_type` | 0 | 1 | `uint8` | 数据面须与流生命周期 opcode 表一致；路由示例见上文。 |
| `routing_mode` | 1 | 1 | `uint8` | 路由模式枚举；见下文。 |
| `src_peer_id` | 2 | 8 | `uint64` BE | 发送方 peer；语义见 [peer-identity.md](./peer-identity.md)。 |
| `dst_peer_id` | 10 | 8 | `uint64` BE | 目标 peer；广播时 **MUST** 为 **`0`**。 |

#### `routing_mode` 取值（本节：广播）

| 值（名称） | 说明 |
|------------|------|
| **`0`** | **保留为非法/未指定**。实现 **MUST** 拒绝解析、丢弃该帧或关闭连接；**`err_code`** 使用 **`ERR_ROUTING_INVALID`**（见 [errors.md](./errors.md)）。 |
| **`1`（`BROADCAST`）** | **广播**：意图为向同一 **session** 内 **除发送者外** 全体成员投递；**`dst_peer_id` MUST 为 `0`**。 |

> 单播（`UNICAST`）见同文件后续章节（`ROUTE-02`）。

### 广播语义与 Relay 行为

当 **`routing_mode = BROADCAST（1）`** 且 **`dst_peer_id = 0`** 时：

1. **语义**：该帧表示发送方请求将 **同一逻辑 payload** 投递给 **同一 session 内除发送该帧的连接外** 的所有 **已加入** 成员。
2. **Relay 转发**：Relay **MUST** 向上述成员各转发 **一帧副本**（可改写与连接相关的实现细节，但 **路由前缀语义** 对下游观察应一致），且 **MUST NOT** 将副本 **回送** 到 **发送该帧的 TLS 连接**（发送者不得通过 Relay 收到自己发出的广播副本）。
3. **`src_peer_id`**：应为发送方已分配 **`peer_id`**（`uint64` BE）；Relay 向其他 peer 转发时 **SHOULD** 原样携带，以便接收方识别来源。

### 完整逻辑帧示例（广播）

下列为一帧：**10 字节帧头** + **payload**。payload 以 **`STREAM_DATA`（`msg_type = 0x11`）** 开头，**广播**路由前缀，**`src_peer_id = 0xab`**（示例），**`dst_peer_id = 0`**；路由前缀之后为 **4 字节 `stream_id` 占位**（`0x00000000`），表示与 [streams-lifecycle.md](./streams-lifecycle.md) 的 **字节级衔接**；应用数据与更多流字段以该文档为准。

- **`payload_len = 22`**（仅含下述 payload 字节）。
- **版本 / capability**：`0x0001` / `0x00000000`（示例）。

**十六进制（每行 16 字节）：**

```
00 00 00 16  00 01  00 00 00 00
11 01  00 00 00 00 00 00 00 ab  00 00 00 00 00 00 00 00  00 00 00 00
```

| 片段 | 含义 |
|------|------|
| `00 00 00 16` | `payload_len = 22` |
| `00 01` | `version = 0x0001` |
| `00 00 00 00` | `capability = 0` |
| `11` | `msg_type = STREAM_DATA`（`0x11`） |
| `01` | `routing_mode = BROADCAST` |
| `00 00 00 00 00 00 00 ab` | `src_peer_id`（大端） |
| `00` × 8 | `dst_peer_id = 0`（广播） |
| `00 00 00 00` | `stream_id` 占位；后续应用字节见 `streams-lifecycle.md` |

---

<!-- REQ: ROUTE-01 -->

## 单播（ROUTE-02）

### 枚举与字段

- **`routing_mode = 2`（`UNICAST`）** 与 **`BROADCAST = 1`** 并列；**`routing_mode = 0`** 仍为非法（见上文）。
- **`src_peer_id`**：**MUST** 为发送方 **`uint64` BE**（与广播一致），语义见 [peer-identity.md](./peer-identity.md)。
- **`dst_peer_id`**：**MUST** 为 **非零** `uint64` BE，且表示 **该 session 内** 目标 peer（与 `SESSION_JOIN_ACK` 分配规则一致；**`0`** 保留语义见 `SESS-03`，**不得**作为有效单播目标）。

### 与广播的联合判定

实现 **MUST** 使用 **`routing_mode` 与 `dst_peer_id` 是否为零** 联合判定意图：

- **`routing_mode = BROADCAST` 且 `dst_peer_id = 0`** → 广播语义（见上文）。
- **`routing_mode = UNICAST` 且 `dst_peer_id ≠ 0`** → 单播语义（本节）。
- **禁止**仅凭单一字段推断对端模式（例如不得假设「`dst_peer_id ≠ 0` 即单播」而忽略错误的 `routing_mode`）。

### 发往自身 peer_id

若 **`dst_peer_id`** 与发送方在该 session 已分配的 **`peer_id`** **相等**（发往自身）：

- 实现 **MUST** 采用下列 **之一**（实现 **MUST** 在文档/能力声明中固定所选策略，且 **不得** 将此类帧当作有效投递）：
  1. **静默丢弃**该帧（不转发、不投递、不触发应用数据交付）；或
  2. **协议错误**路径：关闭连接或返回 **`PROTOCOL_ERROR`**，**`err_code`** 见 [errors.md](./errors.md)。

### Relay 行为（单播）

当 **`routing_mode = UNICAST`** 且 **`dst_peer_id`** 为合法非零会话成员时：

1. Relay **MUST** 仅将帧投递到 **`dst_peer_id`** 所绑定的 **前向 TLS 连接**（见 Phase 2 成员表）。
2. 若 **`dst_peer_id`** **未知**、**非成员**或 **尚无已加入连接**：Relay **MUST** **丢弃**该帧 **或** 返回 **`PROTOCOL_ERROR`** / 关闭连接（**`err_code`** 见 [errors.md](./errors.md)）；**MUST NOT** 向其它 peer 误投递。

### 完整逻辑帧示例（单播）

下列为一帧：**10 字节帧头** + **payload**（**`payload_len = 22`**）。payload 以 **`STREAM_DATA`（`msg_type = 0x11`）** 开头，**`routing_mode = UNICAST（0x02）`**，**`src_peer_id = 0xab`**、**`dst_peer_id = 0xcd`**（均为非零示例），随后 **4 字节 `stream_id` 占位**。

```
00 00 00 16  00 01  00 00 00 00
11 02  00 00 00 00 00 00 00 ab  00 00 00 00 00 00 00 cd  00 00 00 00
```

| 片段 | 含义 |
|------|------|
| `11` | `msg_type = STREAM_DATA`（`0x11`） |
| `02` | `routing_mode = UNICAST` |
| `00 00 00 00 00 00 00 ab` | `src_peer_id` |
| `00 00 00 00 00 00 00 cd` | `dst_peer_id`（非零） |
| `00 00 00 00` | `stream_id` 占位（见 `streams-lifecycle.md`） |

---

<!-- REQ: ROUTE-02 -->
