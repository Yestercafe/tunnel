# Peer 身份（SESS-03）

## 范围

本规范定义 **peer_id** 的分配、唯一性范围、与 **TLS 连接** 的绑定关系，以及在协议中的**语义引用**。**二进制帧头**与 **payload** 外层格式见 [frame-layout.md](./frame-layout.md)；**加入 session** 的控制消息见 [session-create-join.md](./session-create-join.md)。

## peer_id 类型与字节序

- **类型**：`uint64`（无符号 64 位整数）。
- **字节序**：在任意 **固定宽度字段** 中编码时均为 **大端（big-endian，网络字节序）**。
- **保留值**：**`0`** 保留，表示 **未分配** 或 **无效** 的 peer 标识。成功响应（如 `SESSION_JOIN_ACK`）中 **MUST NOT** 使用 **`0`** 作为已分配 `peer_id`。

## 分配者与时机

- **分配者**：**Relay**（或等价会话协调者）。
- **时机**：在 peer **成功加入 session** 之后（即 Relay 已接受 `SESSION_JOIN_REQ` 并完成策略校验），在 **`SESSION_JOIN_ACK`**（`msg_type = 0x04`）中携带新分配的 **`peer_id`**。
- **唯一性范围**：在 **同一 session** 内 **唯一**：任意时刻，两个不同 **已加入** 的 peer **不得** 共享同一非零 `peer_id`。

## 与连接的绑定

- **一条 TLS 连接**（Phase 1 成帧语义下的单字节流）在 **同一 session** 内 **至多对应一个** `peer_id`。
- 若实现允许同一物理连接重新协商 session，则以实现为准；v1 规范建议 **单连接单 session 单 peer_id**，直至连接关闭。

## 在后续协议中的引用（预告）

后续 Phase（路由与流）可在 **帧内子头** 或 **payload 前缀** 中携带：

- **源** `peer_id`（`uint64` BE）
- **目标** `dst_peer_id`（`uint64` BE），用于单播等

本文件 **仅定义语义**（分配、唯一性、保留值）；**具体出现在帧的哪一偏移** 由 **路由/流** 规范定义，本阶段不重复帧级布局。

## 与 SESSION_JOIN_ACK 的一致性

[session-create-join.md](./session-create-join.md) 中 **`SESSION_JOIN_ACK`** 消息体的 **`peer_id`（8 字节，`uint64` BE）** 字段 **即为本规范所定义的 peer 标识**，实现 MUST 遵守上述保留值与唯一性规则。

---

<!-- REQ: SESS-03 -->
