# Join 凭证（SESS-04）

## 范围

本规范描述 **可选** 的 **join token**（加入凭证）：何时需要、在 **`SESSION_JOIN_REQ`** 消息体中的位置与长度限制，以及校验失败时使用的 **占位错误名**（数值码与线格式见 [errors.md](./errors.md)，**ERR-01**）。**加入 session 的主路径**（`join_by`、`credential`、成功/失败概览）见 [session-create-join.md](./session-create-join.md)。

## 何时需要 token

- **join token 为可选**：Relay 是否在某一 session 上要求 token，由 **创建 session 时** 的 **策略**（实现/产品配置）决定；本协议不强制所有 session 都使用 token。
- **Client 行为**：若 Relay 策略要求 token，client **MUST** 在 **`SESSION_JOIN_REQ`** 中携带符合本规范的 **`join_token`**；若策略不要求，client **不得**因未发送 token 而被拒绝（除非实现另有带外约定）。

## JOIN 消息体中的字段

`SESSION_JOIN_REQ`（`msg_type = 0x03`）的 **前缀字段**（`join_by`、`credential_len`、`credential`）见 [session-create-join.md](./session-create-join.md)。

在 **`credential` 字节序列之后**，可 **追加** 下列 **可选** 字段以携带 **join token**：

| 字段 | 长度（字节） | 类型 | 说明 |
|------|-------------|------|------|
| `join_token_len` | 2 | `uint16` BE | **`join_token` 的 UTF-8 字节长度**；**`0`** 表示未携带 token（与「未发送」等价，由策略决定是否允许） |
| `join_token` | `join_token_len` | UTF-8 | 凭证字符串本身 |

**约束：**

- **`join_token` 的 UTF-8 编码长度**（即 `join_token_len`）**MUST NOT** 超过 **256**。
- 字段名在文档层称为 **`join_token`**；在二进制布局中由 **`join_token_len` + 原始字节** 表示。

若 **未追加** `join_token_len` 与 `join_token`（即消息体在 `credential` 处结束），则视为 **未提供** join token。

## 校验失败与 session 不存在

下列为 **占位错误名**；与 **`uint16 err_code`** 及 **`PROTOCOL_ERROR`** 线格式的对应关系见 [errors.md](./errors.md)。

| 情形 | 占位错误名 | 说明 |
|------|-------------|------|
| Relay 策略要求 join token，但缺失、错误或校验不通过 | **`ERR_JOIN_DENIED`** | 拒绝本次 **JOIN**；不分配 `peer_id` |
| `credential` 指向的 **session** 不存在或已失效（与 `join_by` 模式一致） | **`ERR_SESSION_NOT_FOUND`** | 与 Phase 5 会话查找语义对齐 |

实现 **SHOULD** 在拒绝 **JOIN** 时向 client 返回 **`PROTOCOL_ERROR`**，**`err_code`** 使用上表对应符号（见 [errors.md](./errors.md)）。

## 成功路径

成功路径（**`SESSION_JOIN_ACK`**、分配 **`peer_id`**）**不重复**；见 [session-create-join.md](./session-create-join.md) 与 [peer-identity.md](./peer-identity.md)。

---

<!-- REQ: SESS-04 -->
