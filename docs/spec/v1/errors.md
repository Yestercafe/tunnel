# 错误码与应用层错误指示（ERR-01）

## 范围

本文件为 v1 **应用层协议错误** 的 **唯一权威** 目录：**稳定符号名**（`ERR_*`）与 **`uint16` 数值码** 双轨；**控制消息** **`PROTOCOL_ERROR`**（`msg_type = 0x05`）的 **线格式**；以及 **传输层观测**（TLS、`TCP`）与 **协议层 `ERR_*`** 的关系说明。

- **连接级成帧**占位名见 [connection-state.md](./connection-state.md)、[transport-binding.md](./transport-binding.md)。
- **会话失败路径**叙述见 [session-state.md](./session-state.md)。
- **版本不支持**时的关闭义务见 [version-capability.md](./version-capability.md)（`docs/spec/v1/version-capability.md`），与本文件的 **`ERR_PROTO_VERSION`** / **`PROTOCOL_ERROR`** 配合使用。

## 符号名与 `err_code`（uint16）表

| 符号名 | `err_code`（hex） | 说明（摘要） |
|--------|-------------------|--------------|
| **`ERR_FRAME_TOO_LARGE`** | `0x0001` | 帧 `payload_len` 超过允许最大值（见 [frame-layout.md](./frame-layout.md)） |
| **`ERR_PROTO_VERSION`** | `0x0002` | 帧头 `version` 不支持（见 [version-capability.md](./version-capability.md)） |
| **`ERR_JOIN_DENIED`** | `0x0003` | JOIN 被拒绝（凭证/token/策略等，见 [join-credentials.md](./join-credentials.md)） |
| **`ERR_SESSION_NOT_FOUND`** | `0x0004` | 请求的 session 不存在或已失效 |
| **`ERR_ROUTING_INVALID`** | `0x0005` | 数据面路由非法或自相矛盾（如 `routing_mode = 0`，见 [routing-modes.md](./routing-modes.md)） |
| **`ERR_ENVELOPE_INVALID`** | `0x0006` | 应用信封或 `application_data` 边界违反（见 [app-envelope.md](./app-envelope.md)） |

发送方与接收方 **MUST** 使用上表 **数值** 作为 **`PROTOCOL_ERROR` 载荷** 中的 **`err_code`**（**大端**）。新增错误时须更新本表与 **`pkg/framing`** 常量（见下文）。

## `PROTOCOL_ERROR` 控制消息（`msg_type = 0x05`）

### Payload 布局（无数据面路由前缀）

**`PROTOCOL_ERROR`** 为 **控制消息**：payload 以 **`msg_type = 0x05`** 开头，**不** 使用 [routing-modes.md](./routing-modes.md) 中的 **18 字节路由前缀**（与 **`STREAM_*` 数据面** 不同）。

| 字段 | 偏移（相对 payload 起点） | 长度（字节） | 类型 | 说明 |
|------|---------------------------|-------------|------|------|
| `msg_type` | 0 | 1 | `uint8` | **固定 `0x05`**（`PROTOCOL_ERROR`） |
| `err_code` | 1 | 2 | `uint16` BE | 上表中的 **`err_code`** |
| `reason_len` | 3 | 2 | `uint16` BE | 可选 **`reason`** 的 UTF-8 字节数；**`0`** 表示无 `reason` |
| `reason` | 5 | `reason_len` | UTF-8 | 人类可读原因；**`reason_len = 0` 时省略** |

**约束：** `reason` **MUST** 为有效 UTF-8；`reason_len` **MUST** 与实际字节数一致。

### Opcode 冲突（MUST）

| `msg_type`（hex） | 名称 | 层 |
|-------------------|------|-----|
| **`0x05`** | **`PROTOCOL_ERROR`** | **控制面**（本文件） |
| **`0x10`** | **`STREAM_OPEN`** | **数据面**（见 [streams-lifecycle.md](./streams-lifecycle.md)，`docs/spec/v1/streams-lifecycle.md`） |

**MUST NOT** 将 **`PROTOCOL_ERROR`** 分配为 **`0x10`** — **`0x10`** 已用于 **`STREAM_OPEN`**。

## 关闭顺序与 TLS

- 当接收方已能发送 **有效逻辑帧** 时，对 **`ERR_PROTO_VERSION`** 等可恢复为 **单帧响应** 的场景，接收方 **SHOULD** 先发送一帧 **`PROTOCOL_ERROR`**（`err_code` 填上表对应值），再 **`Close` TLS**；与 **`docs/spec/v1/version-capability.md`**「未知版本」路径一致。
- 若 **尚无法** 构造合法帧（例如缓冲区不足、连接已不可用），实现 **MAY** **直接关闭 TCP**，**无需** 先发 **`PROTOCOL_ERROR`**。

## 传输层 vs 协议层（观测关系）

| 观测 | 说明 |
|------|------|
| **TLS** | 实现可能将 **`crypto/tls.AlertError`** 等作为 **TLS 层** 失败；**不要**在规范中写死 **某 TLS alert 字节 ↔ 某 `err_code`** 的一一映射。 |
| **TCP** | 对端关闭、`EOF`、读超时等属于 **传输** 语义；与 **`ERR_*`** **并列** 描述即可。 |
| **协议 `ERR_*`** | 出现在 **`PROTOCOL_ERROR`** 的 **`err_code`** 或实现日志/指标中，表示 **本协议** 判定的错误类别。 |

---

<!-- REQ: ERR-01 -->
