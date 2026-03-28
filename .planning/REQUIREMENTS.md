# Requirements: Tunnel（传输隧道）

**Defined:** 2026-03-29  
**Core Value:** 在 TLS 由边缘/服务器终止的前提下，用同一套协议支撑广播、私信、双向流与可选应用信封，并可在浏览器通过 WebSocket 接入。

## v1 Requirements

### 协议基础（帧与传输）

- [ ] **FRAME-01**: 规范定义**二进制帧格式**（字段顺序、字节序、最大/最小长度规则）
- [ ] **FRAME-02**: 帧头包含**协议版本**字段，并说明**兼容策略**（拒绝/忽略未知版本）
- [ ] **FRAME-03**: 帧头包含 **capability**（或等价能力位），并说明**未知能力位的处理规则**
- [ ] **TRANS-01**: 规范定义 **WebSocket（WSS 二进制）** 上映射：WS message 与逻辑帧的对应关系（含分片/重组若需要）

### 会话与成员

- [ ] **SESS-01**: **创建者**可创建 session，并获得 **session_id** 与/或**邀请码**的格式说明
- [ ] **SESS-02**: peer 可凭 **session_id / 邀请码** 加入同一 session
- [ ] **SESS-03**: 规范描述 **成员身份**（peer id）如何分配与在帧中引用
- [ ] **SESS-04**: 规范说明可选 **短 token / join 凭证** 的携带位置与校验失败时的行为

### 路由与流语义

- [ ] **ROUTE-01**: 支持 **广播**：同一 session 内，**除发送者外**所有成员可收到
- [ ] **ROUTE-02**: 支持 **私信（单播）**：可指定目标 peer，仅目标接收
- [ ] **STREAM-01**: 支持 **双向流**（在单 peer–relay 连接上可持续收发）
- [ ] **STREAM-02**: 明确 **流 ID（或逻辑通道）** 语义：**单流内有序**；**不同流之间允许乱序**

### 应用信封

- [ ] **APP-01**: 定义**可选应用信封**（至少包含字段占位）：如 **content-type**、**请求 id**、**关联 id**；并说明与帧 payload 的边界

### 状态机、错误与文档化假设

- [ ] **STATE-01**: 文档化连接级与 session 级 **状态机**（状态、事件、转移）
- [ ] **ERR-01**: 定义 **错误码目录** 与 **关闭/断开原因**（若适用 WebSocket close code 映射则一并说明）
- [ ] **SEC-01**: 文档化安全假设：**TLS 在边缘/服务器终止**；协议内**不负责 E2E 加密**（v1）

### 一致性测试

- [ ] **TEST-01**: 提供可 **`go test`（或等价）运行的一致性测试**，覆盖关键解析与状态转移；包含 **golden 向量**（`testdata/` 或等价）

## v2 Requirements（延后）

### 传输

- **TRANS-02**: 原生 TCP（或 QUIC）承载与帧映射（与 WS 共享逻辑帧）

### 安全

- **SEC-02**: 端到端加密或密钥协商（若产品需要）

## Out of Scope

| Feature | Reason |
|---------|--------|
| v1 完整生产级 Relay 运维（监控、配额、多区域） | 先协议与一致性测试 |
| 大规模水平扩展与集群分片 | 当前假设少量固定 peer |
| 完整身份联邦（SSO/OAuth） | v1 以 token/邀请码为主 |

## Traceability

| Requirement | Phase | Status |
|-------------|-------|--------|
| FRAME-01 | Phase 1 | Pending |
| FRAME-02 | Phase 1 | Pending |
| FRAME-03 | Phase 1 | Pending |
| TRANS-01 | Phase 1 | Pending |
| SESS-01 | Phase 2 | Pending |
| SESS-02 | Phase 2 | Pending |
| SESS-03 | Phase 2 | Pending |
| SESS-04 | Phase 2 | Pending |
| ROUTE-01 | Phase 3 | Pending |
| ROUTE-02 | Phase 3 | Pending |
| STREAM-01 | Phase 3 | Pending |
| STREAM-02 | Phase 3 | Pending |
| APP-01 | Phase 4 | Pending |
| STATE-01 | Phase 5 | Pending |
| ERR-01 | Phase 5 | Pending |
| SEC-01 | Phase 5 | Pending |
| TEST-01 | Phase 6 | Pending |

**Coverage:**

- v1 requirements: 17 total  
- Mapped to phases: 17  
- Unmapped: 0 ✓  

---

*Requirements defined: 2026-03-29*  
*Last updated: 2026-03-29 after roadmap creation*
