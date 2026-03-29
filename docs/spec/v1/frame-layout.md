# 帧布局（FRAME-01）

## 范围与术语

- **逻辑帧（frame）**：TLS 解密后字节流上可解析的最小单元，由 **固定长度帧头** 与 **变长 payload** 组成。
- **固定帧头（fixed header）**：本规范定义的 **10 字节**，含长度、版本、capability；语义细节见 [version-capability.md](./version-capability.md)。
- **payload**：紧跟帧头之后的 **0～N** 字节；N 由帧头中 **payload_len** 给出。

## 帧结构总览

```
字节偏移   0      1      2      3  |  4      5  |  6      7      8      9  |  10 …
         +-------------------------+------------+---------------------------+--------…
         |    payload_len (uint32)   |  version   |    capability (uint32)    | payload
         +-------------------------+------------+---------------------------+--------…
           big-endian                uint16 BE      big-endian
```

- **总长度**：`10 + payload_len` 字节。

## 字段定义表

| 字段 | 偏移（字节） | 长度（字节） | 类型 | 说明 |
|------|-------------|-------------|------|------|
| `payload_len` | 0 | 4 | `uint32` | **payload** 字节数，不含帧头；大端。 |
| `version` | 4 | 2 | `uint16` | 协议版本；大端。 |
| `capability` | 6 | 4 | `uint32` | 能力位图；大端。 |
| `payload` | 10 | `payload_len` | 字节序列 | 原始负载；可为空。 |

## 字节序

帧头内所有多字节字段均为 **big-endian（大端）**。

## 长度语义

- **`payload_len` 表示「固定帧头（10 字节）之后」的 payload 长度**，不包含帧头本身。
- 这样实现可用 `10 + payload_len` 得到整帧长度，避免「总长」与「payload 长」混用。

## 最小 / 最大帧长

| 量 | 值 |
|----|-----|
| 固定帧头 | **10** 字节 |
| 最小整帧长度 | **10** 字节（`payload_len = 0`） |
| 最大 `payload_len` | **16 MiB** = **16 777 216** 字节 |
| 最大整帧长度 | **16 777 226** 字节（10 + 16 777 216） |

超过最大 `payload_len` 的帧为 **协议错误**；发送方 MUST NOT 发送；接收方 MUST 拒绝并进入错误处理（错误码命名见 [transport-binding.md](./transport-binding.md)，与 Phase 5 对齐）。

## Reserved

本阶段帧头无额外 **reserved** 零填充字段；未使用的 **capability** 位定义见 version-capability 文档。后续阶段若扩展帧头，将通过 **新版本号** 或 **capability** 协商。

## 示例（完整帧）

下列为一帧：**payload_len = 0**，**version = 0x0001**，**capability = 0x00000000**（仅帧头，无 payload）。

十六进制（每行 16 字节）：

```
00 00 00 00  00 01  00 00 00 00
```

| 片段 | 含义 |
|------|------|
| `00 00 00 00` | payload_len = 0 |
| `00 01` | version = 0x0001 |
| `00 00 00 00` | capability = 0 |

带 3 字节 payload `48 65 6c`（ASCII `"Hel"`）的示例：

```
00 00 00 03  00 01  00 00 00 00  48 65 6c
```

---

<!-- REQ: FRAME-01 -->
