# 应用信封（APP-01）

## 范围

本规范定义 **`STREAM_DATA`**（见 [streams-lifecycle.md](./streams-lifecycle.md)）payload 中 **`application_data`** 在可选 **应用信封** 时的字节布局与 **v1 互操作** 元数据编码。与 [routing-modes.md](./routing-modes.md) 中 **Relay** 行为一致：**Relay 不解析** 信封内的结构化内容，仅按帧转发；与 **ROUTE** 不透明转发哲学一致。

## 编码决策（v1 互操作）

- **v1 规范必选、互操作格式**：**`envelope` 字段所覆盖的字节**（见下文分解）为 **UTF-8 编码的 JSON 对象**（一个 JSON object，键值均为规范所允许的标量形式）。
- **TLV**：**不**作为 v1 规范层面的必选互操作格式。若某实现私下在 `envelope` 或 `body` 中使用 TLV，属于**实现私有约定**，**不得**标为 v1 互操作；互通双方须在实现文档中另行约定。

## `flags` 与 `application_data`（相对 payload 起点）

与 [streams-lifecycle.md](./streams-lifecycle.md) **`STREAM_DATA`** 表一致：

| 字段 | 偏移 | 长度 |
|------|------|------|
| `flags` | 22 | 1 |
| `payload_len` | 23 | 2 |
| `application_data` | 25 | `payload_len` |

**`flags` 位定义（本规范发布后）：**

- **bit 0**：**`FIN`**（与既有 STREAM 语义一致）。
- **bit 1**：**`HAS_APP_ENVELOPE`** — 为 **`1`** 时表示 **`application_data`** 按本节「有信封」分解；为 **`0`** 时表示 **无信封** 前缀，语义与 Phase 3 **opaque** 一致。
- **bit 2–7**：保留；发送方 **MUST** 置 **`0`**；接收方 **MUST** 忽略未知位。

### `HAS_APP_ENVELOPE = 0`

**`@25` 起共 `payload_len` 字节**全部为 **应用 `body`**（不透明字节），**无** `envelope_len` 前缀。

### `HAS_APP_ENVELOPE = 1`

**`application_data`**（总长仍为 **`payload_len`**）分解为：

1. **`envelope_len`**：**uint16 BE**，占 **`application_data` 内偏移 0–1**（即相对 **payload** 起点为 **@25–@26**）。
2. **`envelope`**：紧随其后的 **`envelope_len` 字节**，为 UTF-8 JSON 文本（见「JSON schema 子集」）。
3. **`body`**：剩余字节，直至 **`application_data` 末尾**。

**约束：** **`2 + envelope_len ≤ payload_len`**。若违反则解码失败，错误码占位（**ERR_***）见 Phase 5 错误枚举；实现 **MUST** 拒绝该帧或按实现策略报错，**不得**静默截断互操作。

## JSON schema 子集（互操作）

`envelope` 解码为 JSON object 后，v1 互操作 **至少** 支持下列键（**UTF-8 字符串**值，允许空字符串 `""`）：

| 键 | 含义（示意） |
|----|----------------|
| **`content_type`** | 类比 MIME，例如 `application/json`。 |
| **`request_id`** | 请求/事务分支标识。 |
| **`correlation_id`** | 跨帧或往返关联标识（如 Copilot 管道）。 |

**限制：** 值为 **标量字符串** 的 **扁平** 键值对；**禁止** 在 v1 互操作信封中出现 **嵌套 object / array**（若需扩展，由后续版本规范另行定义）。

## `envelope_len` 上限

- **最大 `envelope_len`（v1）**：**4096** 字节。
- 须 **≤** 帧级 **`payload_len`** 与 **`2 + envelope_len ≤ payload_len`** 一致。
- 若 **`envelope_len` > 4096** 或 JSON 无效/非 UTF-8，视为协议或解码错误（**ERR_***，Phase 5）。

## Relay 不透明

**Relay** 对 **`envelope`** 内 JSON **不解析**、不校验键名；仅按 **`STREAM_DATA`** 帧边界与路由前缀转发，与数据面路由规范一致。

---

<!-- REQ: APP-01 -->
