# 流生命周期与多路复用（STREAM-01、STREAM-02）

## 范围

本规范定义 **数据面** 上 **逻辑流** 的 opcode、payload 布局、`stream_id` 规则，以及 **流内有序 / 流间乱序** 语义。固定 **10 字节帧头** 与 **payload** 边界见 [frame-layout.md](./frame-layout.md)；**路由前缀**（`msg_type` + `routing_mode` + `src_peer_id` + `dst_peer_id`，共 **18** 字节）见 [routing-modes.md](./routing-modes.md)。**控制面** `SESSION_*` 消息 **不** 携带 `stream_id`。

## 数据面 `msg_type` opcode 表

与 [session-create-join.md](./session-create-join.md) 中 **`0x01`–`0x04`**（`SESSION_*`）**无重叠**：

| 值（hex） | 名称 | 说明 |
|-----------|------|------|
| **`0x10`** | **`STREAM_OPEN`** | 打开一条逻辑流，可携带可选元数据。 |
| **`0x11`** | **`STREAM_DATA`** | 在已打开流上承载应用字节；`flags` 见下文（含可选 **FIN**）。 |
| **`0x12`** | **`STREAM_CLOSE`** | 关闭一条逻辑流（显式生命周期终点）。 |

**关闭路径（避免实现分裂）：**

- **主路径**：**`STREAM_CLOSE`（`0x12`）** — 任一端发送即表示该 **`stream_id`** 在双方约定的语义下 **结束**（全关闭），接收方 **MUST** 释放该流的发送/接收状态（实现可定义是否允许同一 `stream_id` 稍后由新的 `STREAM_OPEN` 复用；v1 **建议** 不复用直至连接关闭，由实现文档化）。
- **可选半关闭**：**`STREAM_DATA`** 的 **`flags` 第 0 位（`FIN = 1`）** 表示 **发送方不再在本流上发送后续应用字节**（半关闭）。**优先级**：若 **`STREAM_CLOSE`** 与 **带 `FIN` 的 `STREAM_DATA`** 均可出现，**以 `STREAM_CLOSE` 为权威**：收到 **`STREAM_CLOSE`** 后 **MUST** 按关闭处理；**`FIN`** 仅约束「不再有 DATA」而不替代 **`CLOSE`** 的全部资源语义。

## Payload 布局（紧跟路由前缀之后）

路由前缀 **18** 字节定义见 [routing-modes.md](./routing-modes.md)（`msg_type` @0 … `dst_peer_id` @10–17）。本节字段 **相对 payload 起点** 描述；**`stream_id` 起始偏移为 18**（与下表一致）。

### 通用：`*` @18 — `stream_id`

| 字段 | 偏移 | 长度 | 类型 | 说明 |
|------|------|------|------|------|
| `stream_id` | 18 | 4 | `uint32` BE | 见下文 **STREAM-02**；**`0`** 禁止作为有效数据流 ID。 |

### `STREAM_OPEN`（`msg_type = 0x10`）

| 字段 | 偏移 | 长度 | 类型 | 说明 |
|------|------|------|------|------|
| （路由前缀） | 0–17 | 18 | — | 含 `msg_type = 0x10`。 |
| `stream_id` | 18 | 4 | `uint32` BE | 待打开流的 ID。 |
| `metadata_len` | 22 | 2 | `uint16` BE | 元数据字节数；**可为 `0`**。 |
| `metadata` | 24 | `metadata_len` | 字节序列 | 不透明元数据。 |

### `STREAM_DATA`（`msg_type = 0x11`）

| 字段 | 偏移 | 长度 | 类型 | 说明 |
|------|------|------|------|------|
| （路由前缀） | 0–17 | 18 | — | 含 `msg_type = 0x11`。 |
| `stream_id` | 18 | 4 | `uint32` BE | 目标流。 |
| `flags` | 22 | 1 | `uint8` | **bit 0**：**`FIN`**（`1` = 发送方结束本流上后续应用数据）；其余位保留，发送方 **MUST** 置 **`0`**，接收方 **MUST** 忽略未知位。 |
| `payload_len` | 23 | 2 | `uint16` BE | **紧随其后的** 应用字节长度。 |
| `application_data` | 25 | `payload_len` | 字节序列 | 应用载荷。 |

### `STREAM_CLOSE`（`msg_type = 0x12`）

| 字段 | 偏移 | 长度 | 类型 | 说明 |
|------|------|------|------|------|
| （路由前缀） | 0–17 | 18 | — | 含 `msg_type = 0x12`。 |
| `stream_id` | 18 | 4 | `uint32` BE | 要关闭的流。 |

### 偏移汇总（从 payload 起点）

| 字段 | 起始偏移（字节） |
|------|-------------------|
| `msg_type` | 0 |
| `routing_mode` | 1 |
| `src_peer_id` | 2 |
| `dst_peer_id` | 10 |
| `stream_id` | 18 |
| `STREAM_DATA`：`flags` | 22 |
| `STREAM_DATA`：`payload_len` | 23 |
| `STREAM_DATA`：`application_data` | 25 |
| `STREAM_OPEN`：`metadata_len` | 22 |
| `STREAM_OPEN`：`metadata` | 24 |

若 **`STREAM_OPEN`** 带非零 `metadata`，下游字段偏移随 `metadata_len` 顺延；**`stream_id` @18** 不变。

## `stream_id` 规则（STREAM-02）

- **类型**：`uint32` BE。
- **唯一性范围**：**仅在单条 peer↔Relay TLS 连接上唯一**（发送方为每条新流选择未使用或非冲突的 id；对端在同一连接上引用同一 `stream_id` 即视为同一条逻辑流）。
- **`stream_id = 0`**：**禁止**作为有效数据流标识 — 实现 **MUST NOT** 在 **`STREAM_OPEN` / `STREAM_DATA`** 的有效流语义中使用 **`0`**（保留；可用于实现内部「无流」哨兵，但不作为协议数据流）。

## 双向流（STREAM-01）

在 **单条** peer↔Relay **TLS 连接**上可同时存在多条逻辑流。对同一 **`stream_id`**，**client 与 Relay（及对端 peer）** 可交替发送 **`STREAM_DATA`**，直至 **`STREAM_CLOSE`**（及可选 **`FIN`**）完成关闭语义。**不得**在 **`SESSION_*`** 控制消息中携带 **`stream_id`**（与控制面分离）。

## 顺序：流内有序与流间乱序（STREAM-02）

- **与 TLS 线序的关系**：在 **单连接 TCP+TLS** 上，**帧的字节流顺序** 即传输顺序；本规范所说的 **「乱序」** 指 **逻辑多路**（不同 `stream_id`）维度上 **应用处理顺序** 的互不保证，**不是** TCP 字节乱序。
- **流内有序**：对同一 **`(TLS 连接, stream_id)`**，接收方对 **`STREAM_DATA` 所承载的应用字节序列** 的处理顺序 **MUST** 与发送顺序 **一致**。
- **流间乱序**：不同 **`stream_id`** 的帧 **MAY** 以任意交错顺序到达；接收方 **MUST NOT** 假定「流 A 的某一 `STREAM_DATA` 早于流 B 的某一 `STREAM_DATA`」在应用层必然成立。

## 并发多流用例（非规范性示例）

**Peer A** 向 **Peer B** 同时通过 Relay 使用两条逻辑流：**`stream_id = 3`** 传输文件分块，**`stream_id = 4`** 传输心跳。**B** 侧可能观测到两类 **`STREAM_DATA`** 帧 **交错**到达；但 **在 `stream_id = 3` 上** 的应用字节顺序 **保持发送序**，**`stream_id = 4`** 同理。不得假设文件块一定先于某次心跳被处理。

---

<!-- REQ: STREAM-01 -->

<!-- REQ: STREAM-02 -->
