# Phase 1 — Technical Research

**Phase:** 1 — 协议基础（帧与 TLS 字节流承载）  
**Researched:** 2026-03-29  
**Requirement IDs:** FRAME-01, FRAME-02, FRAME-03, TRANS-01

## Objective

回答：**如何在 TLS 连续字节流上定义可解析、可演进、可测试的二进制帧？** 产出供 `PLAN.md` 引用的技术选项与推荐默认值。

## Findings

### 1. 字节序与定宽字段

- **网络字节序（大端）** 是二进制协议常见默认，与 `encoding/binary` 的 `BigEndian` 一致，便于文档与代码对照。  
- 魔数（magic）可选：用于快速识别流是否为隧道协议；若加魔数，须固定 4 字节并写入规范，避免与 TLS 应用数据混淆（TLS 之上已是明文负载，魔数在**解密后**的第一个字节开始）。

### 2. 长度前缀 vs 定长帧

- **变长 payload** 需要 **长度字段**（通常为 `uint32` 或 `varint`）。  
- **uint32 大端**：实现简单，最大 ~4GiB/帧，与 `binary.BigEndian` 一致。  
- **最大帧长** 必须在规范中写死（例如 4–16 MiB），防止 DoS；超过则 **协议错误** 并定义是否断开。

### 3. 版本字段（FRAME-02）

- 建议 **主版本 8 bit 或 16 bit** + 可选 **次版本**，或单一 `uint16` 协议版本号。  
- **策略**：未知主版本 → **必须**关闭连接并带错误码；同主版本未知次版本 → 可 **忽略扩展字段**（若帧头有 ext length）。

### 4. Capability（FRAME-03）

- **位图（uint32 或 uint64）** + **未知位必须忽略**（forward compatibility），除非该位标记为「必须理解」。  
- 可选：**capability 扩展 TLV** 放在固定头之后，由 `capability_ext_len` 指定。

### 5. TLS 字节流成帧（TRANS-01）

- **半包**：`Read` 返回少于一帧；需 **缓冲**（`bufio.Reader` 或自维护 `[]byte`）。  
- **粘包**：一次 `Read` 含多帧；循环 **解析直到缓冲区不足**。  
- **错误恢复**：帧头损坏、长度非法 → **丢弃连接**或进入 **显式错误状态**（在 Phase 5 细化）；v1 规范至少定义 **错误码枚举** 占位。

### 6. 与后续阶段边界

- Phase 1 **不**定义 session、路由、流 ID 的语义字段占位即可（若需预留，用 **reserved** 零填充并说明「Phase 2 定义」）。

## Recommendations (for planner)

| 主题 | 推荐 |
|------|------|
| 字节序 | 大端（big-endian） |
| 长度字段 | `uint32` 表示 payload 或「头+payload」总长（须在规范二选一并图示） |
| 版本 | 独立 `uint16` 或 `uint8` major + `uint8` minor |
| Capability | `uint32` 位图，未知位忽略 |
| 参考实现语言 | Go：`encoding/binary`, `io.Reader`/`bufio` |

## Validation Architecture

Phase 1 交付以 **规范文档 + 静态检查** 为主；可执行测试在 **Phase 6** 集中。本阶段仍建议：

- **go test**：对 **帧编解码辅助函数**（若计划中包含最小 `internal/framing` 或 `pkg/framing`）做 **round-trip** 与 **golden bytes**。  
- **文档校验**：Makefile 或脚本检查规范中 **术语表、字段表、示例十六进制** 三节齐全。  
- **采样**：每完成一个 PLAN，运行 `go test ./...`（若模块已初始化）与 `grep` 检查 REQ-ID 在 `docs/spec` 中出现。

## RESEARCH COMPLETE

本研究足以支撑三份可执行 PLAN（帧布局、版本/capability、传输绑定），并与 REQUIREMENTS.md 中 FRAME-* / TRANS-01 对齐。
