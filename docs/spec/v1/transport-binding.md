# TLS 字节流绑定（TRANS-01）

## 前提

- 连接已为 **TLS** 保护下的 **双向字节流**（典型实现：TCP 上 `crypto/tls`，Go 中为 `*tls.Conn` 的 `Read`/`Write`）。
- 以下描述的是 **TLS 应用数据** 解密后送入协议层的 **明文字节序列**。

## 解析状态机（从空缓冲区开始）

1. 将新读入的字节 **追加** 到接收缓冲区末尾。
2. 若缓冲区长度 **不足 10 字节**：**等待** 更多字节（**半包**），不解析帧。
3. 读取 **固定帧头**（10 字节），按 [frame-layout.md](./frame-layout.md) 解析 `payload_len`、`version`、`capability`（大端）。
4. 若 `payload_len` **大于 16 777 216**：**协议错误**，进入 **`ERR_FRAME_TOO_LARGE`**（占位名；与 Phase 5 错误码表对齐），**MUST** 关闭连接或进入错误处理状态。
5. 若缓冲区长度 **不足 10 + payload_len 字节**：**半包**，保留缓冲区，等待更多字节。
6. 否则从缓冲区 **切出** 长度 `10 + payload_len` 的字节序列作为 **一帧**，交付上层；缓冲区移除已消费字节。
7. 回到步骤 2（**粘包**：若缓冲区仍有 ≥10 字节，可继续解析下一帧）。

## 半包（partial read）

- **头半包**：不足 10 字节 → 不解析，继续读。
- **payload 半包**：已有完整头，但 `payload` 未读完 → 不交付帧，继续读。

## 粘包（multiple frames）

- 一次 `Read` 可能包含 **多帧**；实现 **MUST** 在单次读入后 **循环** 执行状态机，直到缓冲区不足以构成下一帧的至少 10 字节头。

## 最大帧长

- 与 frame-layout 一致：**最大 payload 16 MiB**；超限即 **`ERR_FRAME_TOO_LARGE`**。

## 版本与 capability

- 每帧解析出头后，按 [version-capability.md](./version-capability.md) 校验 **`version`**（v1 仅 `0x0001`）并解释 **`capability`**（未知位忽略）。
- **首帧**与 **后续每一帧**均使用相同规则；不要求单独「握手帧」类型（会话语义见后续阶段）。

## 错误命名（占位）

| 占位名 | 条件 |
|--------|------|
| `ERR_FRAME_TOO_LARGE` | `payload_len` 超过允许最大值 |
| `ERR_PROTO_VERSION` | `version` 不支持（见 version-capability） |

与 TLS alert / TCP 断开的对应关系在 Phase 5 文档化。

---

**TRANS-01**
