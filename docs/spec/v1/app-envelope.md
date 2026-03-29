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

## 端到端示例 A：JSON 请求/响应

下列为 **一帧** **`STREAM_DATA`**：**`HAS_APP_ENVELOPE = 1`**，**`FIN = 0`**，故 **`flags = 0x02`**（仅 bit1）。**`body`** 为 UTF-8 JSON 负载 **`{"ok":true}`**（演示最小应答体）。**`envelope`** UTF-8 文本为：

`{"content_type":"application/json","request_id":"req-7a2f"}`（长度 **59** 字节；**无** `correlation_id` 键，互操作上允许）。

**配对语义（v1 示例约定）：** **应答帧**亦为 **`STREAM_DATA`** 且 **`HAS_APP_ENVELOPE = 1`** 时，在 **`envelope`** 中使用 **相同** **`request_id`**（本例 **`req-7a2f`**）以与请求配对；**`body`** 承载应答 JSON（具体形状由应用定义，本规范不约束）。

### 请求帧：`application_data` 分解（相对 **payload** 起点）

| 相对 payload 偏移 | 长度（字节） | 含义 |
|-------------------|-------------|------|
| @25–26 | 2 | **`envelope_len` = `0x003B`（59，uint16 BE）** |
| @27–85 | 59 | **`envelope`**（上列 JSON UTF-8） |
| @86–96 | 11 | **`body`**：`{"ok":true}` |

本例 **`payload_len` = 72**（`0x0048`），即 **`application_data`** 总长 72 字节。

**`application_data` 起始十六进制（连续 16 字节，可对照上表）：**

`00 3B 7B 22 63 6F 6E 74 65 6E 74 5F 74 79 70 65`

（`00 3B` = **59** 大端；随后为 JSON 开头 `{"content_type`…）

---

## 端到端示例 B：Copilot 往返

**场景：** **Peer A** 与 **Peer B** 在 **同一条** 双向逻辑流上交换 **`STREAM_DATA`**（**同一** **`stream_id = 7`**，大端 **`0x00000007`**）。路由为 **`UNICAST`**（**`routing_mode = 2`**），见 [routing-modes.md](./routing-modes.md) **ROUTE-02**。**`msg_type = 0x11`（`STREAM_DATA`）**。

两帧信封 JSON **共享** 同一非空 **`correlation_id`**：**`corr-copilot-01`**；**`request_id`** 分别为 **`req-a1`**（A→B）与 **`req-b1`**（B→A），用于区分调用分支。

**Relay 不解释信封** — 仅按帧转发；与上文「Relay 不透明」及 **04-01** 正文一致。

### 帧 1：Peer A → Peer B

| 相对 payload 偏移 | 示例值 | 含义 |
|-------------------|--------|------|
| @0 | `0x11` | **`STREAM_DATA`** |
| @1 | `0x02` | **`UNICAST`** |
| @2–9 | （示例） | **`src_peer_id`** = A |
| @10–17 | （示例） | **`dst_peer_id`** = B |
| @18–21 | `00 00 00 07` | **`stream_id`** = 7 |
| @22 | `0x02` | **`flags`**：`HAS_APP_ENVELOPE`，无 **FIN** |
| @23–24 | `0x00 6A` | **`payload_len`** = 106（`application_data` 总长） |
| @25+ | 见下 | **`application_data`**：`envelope_len` + `envelope` + `body` |

**`application_data` 前缀（**`envelope_len` = 92 = `0x005C`**，**`body`** = `prompt-bytes`）：**

`00 5c 7b 22 63 6f 6e 74 65 6e 74 5f 74 79 70 65 22 3a 22 61 70 70 6c 69 63 61 74 69 6f 6e 2f 6a 73 6f 6e 22 2c 22 72 65 71 75 65 73 74 5f 69 64 22 3a 22 72 65 71 2d 61 31 22 2c 22 63 6f 72 72 65 6c 61 74 69 6f 6e 5f 69 64 22 3a 22 63 6f 72 72 2d 63 6f 70 69 6c 6f 74 2d 30 31 22 7d 70 72 6f 6d 70 74 2d 62 79 74 65 73`

（前 **8** 个十六进制字节：`00 5c 7b 22 63 6f 6e 74 65 6e 74` — `00 5c` = 92；其后为 JSON 起始。）

### 帧 2：Peer B → Peer A

| 相对 payload 偏移 | 示例值 | 含义 |
|-------------------|--------|------|
| @0 | `0x11` | **`STREAM_DATA`** |
| @1 | `0x02` | **`UNICAST`** |
| @2–9 | （示例） | **`src_peer_id`** = B |
| @10–17 | （示例） | **`dst_peer_id`** = A |
| @18–21 | `00 00 00 07` | **同一** **`stream_id`** = 7 |
| @22 | `0x02` | **`flags`** |
| @23–24 | `0x00 69` | **`payload_len`** = 105 |
| @25+ | 见下 | **`application_data`** |

**`envelope`** 中 **`request_id`** 为 **`req-b1`**，**`correlation_id`** 仍为 **`corr-copilot-01`**；**`body`** = `reply-bytes`。

**`application_data` 起始十六进制（前 16 字节）：**

`00 5c 7b 22 63 6f 6e 74 65 6e 74 5f 74 79 70 65 22`

（与帧 1 相同 **`envelope_len`**；**`envelope`** 内 **`request_id`** 字段值不同，完整字节序列由 UTF-8 编码唯一确定。）

---

<!-- REQ: APP-01 -->
