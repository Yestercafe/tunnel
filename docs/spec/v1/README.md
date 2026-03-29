# Tunnel — v1 协议规范

本目录收录 **Tunnel** 传输隧道 **v1** 的二进制协议规范文档（Phase 1：帧与 TLS 字节流承载；Phase 2：会话创建与加入等）。

## 文档索引

| 文档 | 说明 | 状态 |
|------|------|------|
| [frame-layout.md](./frame-layout.md) | 固定帧头、长度字段、payload 边界（FRAME-01） | 已发布 |
| [version-capability.md](./version-capability.md) | 协议版本与 capability 位图（FRAME-02、FRAME-03） | 已发布 |
| [transport-binding.md](./transport-binding.md) | TLS 字节流上的成帧与解析（TRANS-01） | 已发布 |
| [session-create-join.md](./session-create-join.md) | session 创建/加入、session_id 与邀请码、控制消息 opcode（SESS-01、SESS-02） | 已发布 |
| [peer-identity.md](./peer-identity.md) | peer_id（uint64）分配、会话内唯一性、与 JOIN 响应一致（SESS-03） | 已发布 |
| [join-credentials.md](./join-credentials.md) | 可选 join token 位置、长度与 ERR 占位（SESS-04） | 已发布 |
| [routing-modes.md](./routing-modes.md) | 数据面路由前缀；广播（ROUTE-01）、单播（ROUTE-02） | 已发布 |
| [streams-lifecycle.md](./streams-lifecycle.md) | 流 opcode（OPEN/DATA/CLOSE）、`stream_id`、双向流与顺序（STREAM-01、STREAM-02） | 已发布 |
| [app-envelope.md](./app-envelope.md) | 可选应用信封、`HAS_APP_ENVELOPE`、`envelope` JSON 子集（APP-01） | 已发布 |

## 字节序

除特别声明外，多字节整数均为 **大端（big-endian，网络字节序）**，与 Go `encoding/binary.BigEndian` 一致。

## 版本

规范版本与 `frame-layout.md` 中的 **protocol version** 字段配合使用；当前正文对应 **v1** 帧布局。
