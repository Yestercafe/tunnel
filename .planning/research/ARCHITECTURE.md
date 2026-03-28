# Architecture Research

**Domain:** 中继隧道 — 会话、路由、传输  
**Researched:** 2026-03-29  
**Confidence:** MEDIUM

## Major Components

| 组件 | 职责 | 边界 |
|------|------|------|
| **Relay（服务器）** | 接受连接、验证 token（若启用）、维护 session 与成员、转发帧 | 不解析 E2E 明文（当前假设） |
| **Client** | 与 relay 建立 **TLS 字节流**（典型 TCP）、加入 session、发送/接收帧 | v1 以 Go/进程为主；浏览器为后续适配 |
| **Session 抽象** | session_id、成员集合、创建/加入规则 | 与传输层无关 |
| **流抽象** | 流 ID、方向、顺序域；承载大块与小消息 | 多路复用的核心 |
| **应用信封（可选）** | 与业务相关的元数据 | 位于帧 payload 内或紧接帧头之后（由规范定） |

## Data Flow

1. Client A 与 Relay 建立 **TLS（边缘终止）** 下的连接（典型为 **TCP + TLS**）。  
2. A 执行 **创建 session** 或 **加入 session**（携带 session_id / 邀请码 / token）。  
3. A 发送 **数据帧**：默认 **广播** 至同 session 其它成员；或 **私信** 至指定 peer。  
4. Relay 根据帧内路由信息复制或单播至目标连接。  
5. 接收方 client 解析帧 → 可选解析应用信封 → 交付上层（前端、Copilot 管道等）。

## Component Boundaries

- **传输层**：v1 为 **TLS 字节流**；未来可增 WebSocket/WebTransport **适配层**；只保证字节双向可达。  
- **帧层**：版本、能力、流 ID、路由模式（广播/单播）、长度与校验（若有）。  
- **会话层**：谁在哪个 session、成员变更事件。  
- **应用层**：完全可选；通过信封字段关联请求/响应。

## Suggested Build Order（与路线图一致）

1. **帧格式 + TLS 字节流成帧**（无业务也可测互通）。  
2. **会话状态机**（创建/加入/离开/心跳若有）。  
3. **路由与流**（广播、私信、流内顺序）。  
4. **应用信封**。  
5. **错误码与关闭语义**。  
6. **一致性测试套件**。

## Build Order Dependencies

- 路由依赖会话成员存在。  
- 大块流依赖流 ID 与分片规则先于实现。  
- 一致性测试依赖错误码与状态机稳定后再冻结向量。
