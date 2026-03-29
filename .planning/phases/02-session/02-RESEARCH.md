# Phase 2 — Technical Research

**Phase:** 2 — 会话生命周期与成员  
**Researched:** 2026-03-29  
**Requirement IDs:** SESS-01, SESS-02, SESS-03, SESS-04

## Objective

确定 **session_id**、**邀请码**、**peer 标识** 与 **join 凭证** 在 **v1 二进制协议**中的承载方式；控制面消息置于 Phase 1 已定义的 **帧 payload** 之内。

## Findings

### 1. session_id 与邀请码

- **session_id**：建议 **UUID** 字符串（**36** 字符 ASCII，`8-4-4-4-12`）或 **16 字节原始 UUID**；文档采用 **UTF-8 字符串** 便于调试与日志。  
- **invite_code**：短链式 **Base32 无填充**、长度 **8～12** 字符，与 session 一对一映射，由 **Relay 在创建 session 时生成**。

### 2. 控制消息与数据消息

- **控制面**（创建/加入/错误应答）与 **数据面**（后续路由）共用同一 **帧格式**；**payload** 内第一字节或固定头为 **msg_type**（`uint8`），后续为消息体。  
- v1 规范用 **opcode 表** + **字段表** 描述，避免与 Phase 3 **路由头**重复设计。

### 3. peer_id

- Relay 在 peer **成功加入 session** 后分配 **uint64**，会话内唯一；**0** 保留表示「未分配/无效」。  
- 在协议中的引用方式：后续 Phase 在帧或子头中带 **peer_id (uint64, BE)**；本阶段只定义 **语义与分配规则**。

### 4. Join token

- **可选**：在 **JOIN** 类消息体中带 **token**（UTF-8 或固定长度 secret）；校验失败 → **`ERR_JOIN_DENIED`** 占位，关闭或拒绝（与 Phase 5 ERR 对齐）。

## Validation Architecture

- **文档**：`rg SESS-` 于 `docs/spec/v1/`；README 索引齐全。  
- **代码**：Phase 2 以规范为主；若增加 `pkg/session` 解析示例，则 `go test ./...`。  
- **Nyquist**：与 Phase 1 相同，以可 grep 的 REQ 标记收束。

## RESEARCH COMPLETE
