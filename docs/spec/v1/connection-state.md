# 连接级成帧状态（STATE-01 — 连接 / 成帧层）

## 范围声明

- 本文件仅描述：**TLS 应用数据已可读** 之后，在 **明文字节流**（与 [transport-binding.md](./transport-binding.md) 前提一致）上进行的 **成帧解析状态机**。
- **不包含**：session 成员关系、`SESSION_*` 控制消息语义、数据面路由或流生命周期 —— 那些属于 [session-state.md](./session-state.md) 与 [streams-lifecycle.md](./streams-lifecycle.md)。
- 成帧层从 **空接收缓冲区** 开始，按与 `transport-binding.md` **相同的编号步骤**（读入 → 至少 10 字节头 → `payload_len` → 收满整帧 → 交付）循环运行；下文用 **状态 / 事件** 视角复述，避免与 session 层状态混写。

## 与 transport-binding 的对应关系

| 概念 | 说明 |
|------|------|
| **半包（需更多数据）** | 缓冲区长度不足 **10** 字节（头半包），或已有头但不足 **10 + payload_len**（payload 半包）。**不是**协议错误；实现 **等待** 更多字节，保留缓冲区。 |
| **已交付一帧** | 缓冲区足以切出长度 **10 + payload_len** 的连续字节；**从缓冲区移除**该段，将 **整段原始字节**（含 10 字节头 + payload）作为 **一帧** 交给上层。成帧层 **不解析** payload 内 `msg_type` 语义 —— 仅交付 **payload 字节序列**（对上层为不透明字节，除非上层自行解析）。 |
| **粘包** | 一次读入后，若缓冲区在消费后仍 **≥ 10** 字节，**MUST** **循环** 重复解析，直到不足以构成下一帧的至少 10 字节头（与 `transport-binding.md` 一致）。 |
| **致命协议错误** | **`ERR_FRAME_TOO_LARGE`**：`payload_len` 超过允许最大值（见 [frame-layout.md](./frame-layout.md)）。**`ERR_PROTO_VERSION`**（及 capability 校验失败等）：`version` / capability 不满足 [version-capability.md](./version-capability.md)。占位名与 **数值错误码** 在 **ERR-01**（[errors.md](./errors.md)，若已发布）中统一；本层进入错误状态后 **MUST** 按规范关闭连接或进入错误处理，**不得**与「半包」混淆。 |

## 状态机（摘要）

从实现角度，可视为在下列结果间转移（**不**与 session 的「未加入」等状态共用同一枚举）：

1. **Accumulating** — 缓冲区不足以判定完整帧（半包或等待头）。
2. **FrameReady** — 已切出一帧，交付上层；缓冲区移除已消费字节。
3. **FatalError** — 不可恢复协议错误（如 `ERR_FRAME_TOO_LARGE`、`ERR_PROTO_VERSION`），终止正常解析。

**非错误**：**Accumulating** 因数据不足而持续等待 **不是** `FatalError`。

## 边界：成帧层 vs payload 语义层

- 成帧层 **只产出**「一帧的 **payload 字节**（长度 = `payload_len`）」及帧头元数据（`version`、`capability`），供 **session / 流** 等上层解析 `msg_type` 与路由。
- 将 **「半包」**（字节不够）与 **「未加入 session」**（session 层门禁）混在同一状态集合是 **错误** 建模；后者见 [session-state.md](./session-state.md)。

## 实现注意（非规范性）

- 接收缓冲区可为 **环形缓冲** 或 **线性可收缩缓冲**；**语义**上须等价于「追加新字节 → 按上文循环解析」。
- **首帧**与 **后续帧**使用同一套规则；不要求单独的「连接级握手帧」类型（会话语义由 `SESSION_*` 等上层消息定义）。

---

<!-- REQ: STATE-01 -->
