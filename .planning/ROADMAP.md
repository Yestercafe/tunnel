# Roadmap: Tunnel（传输隧道）

## Overview

从**可互操作的二进制协议**出发：先冻结**帧、版本、能力与 TLS 字节流上的成帧**，再定义**会话与成员**、**广播/私信与流语义**、**可选应用信封**，最后补齐**状态机、错误与安全假设**，并以**一致性测试与 golden 向量**收束。实现语言为 **Go**，但 v1 交付以**规范与测试**为主。**v1 不采用 WebSocket**。

## Phases

- [x] **Phase 1: 协议基础 — 帧与 TLS 字节流承载** — 定义帧格式、版本、能力位与 TCP+TLS 上的成帧/解析   (completed 2026-03-29)
- [x] **Phase 2: 会话生命周期与成员** — 创建/加入、peer 标识、可选 join token   (completed 2026-03-29)
- [ ] **Phase 3: 路由与多路流** — 广播、私信、双向流、流内序与流间乱序  
- [ ] **Phase 4: 可选应用信封** — content-type、请求/关联 id 与 payload 边界  
- [ ] **Phase 5: 状态机、错误与安全假设** — 状态转移、错误码、关闭语义、TLS 假设文档化  
- [ ] **Phase 6: 一致性测试套件** — go test、testdata 向量、关键回归场景  

## Phase Details

### Phase 1: 协议基础 — 帧与 TLS 字节流承载

**Goal**: 交付可实现的**帧布局**与在 **TLS（典型 TCP）字节流**上的**成帧与解析**规则，含版本与能力字段。  
**Depends on**: 无（首阶段）  
**Requirements**: FRAME-01, FRAME-02, FRAME-03, TRANS-01  
**Success Criteria**（必须为真）:

1. 两个独立实现（或同一实现的 client/server 模拟）能在 **TLS 保护的连续字节流**上按规范交换**合法帧**并解析成功（含粘包/半包场景）  
2. **版本升级/拒绝策略**在文档中有可执行规则  
3. **未知 capability** 的处理规则明确且无歧义  

**Plans**: 3 plans  

Plans:

- [x] 01-01: 帧字段表、字节序、长度与边界条件成文  
- [x] 01-02: 版本号与 capability 位图语义 + 示例十六进制帧  
- [x] 01-03: 字节流上帧边界（长度前缀等）、最大帧长与错误恢复策略  

**UI hint**: no  

---

### Phase 2: 会话生命周期与成员

**Goal**: 定义 **session 创建/加入**、**peer 标识**与 **可选 join 凭证** 的协议语义。  
**Depends on**: Phase 1  
**Requirements**: SESS-01, SESS-02, SESS-03, SESS-04  
**Success Criteria**:

1. 文档可推导出 Relay 所需的 **最小成员表** 字段  
2. **创建者开房 / peer 加入** 的消息序列可画图且与状态机一致  
3. token 无效、session 不存在等路径有**明确错误/关闭行为**（引用 ERR 占位）  

**Plans**: 3 plans  

Plans:

- [x] 02-01: session_id / 邀请码格式与创建/加入消息  
- [x] 02-02: peer id 分配与在帧中的引用方式  
- [x] 02-03: 可选短 token 的携带位置与失败语义  

**UI hint**: no  

---

### Phase 3: 路由与多路流

**Goal**: 在单连接上支持 **广播**、**私信**、**双向流**，并写清 **流内有序、流间乱序**。  
**Depends on**: Phase 2  
**Requirements**: ROUTE-01, ROUTE-02, STREAM-01, STREAM-02  
**Success Criteria**:

1. 每种投递模式（广播/私信）均有**帧级路由字段**说明与示例  
2. **流 ID** 定义与顺序保证范围无歧义  
3. 至少一组**并发多流**场景的文字用例（非实现）  

**Plans**: 3 plans  

Plans:

- [x] 03-01: 广播语义（排除发送者）与帧字段  
- [x] 03-02: 私信语义与目标 peer 字段  
- [x] 03-03: 流创建/数据/结束与顺序规则  

**UI hint**: no  

---

### Phase 4: 可选应用信封

**Goal**: 定义**可选**应用层元数据，使 HTTP 式前端与 Copilot 管道可共享隧道。  
**Depends on**: Phase 3  
**Requirements**: APP-01  
**Success Criteria**:

1. 信封字段列表、出现条件、与二进制 payload 的边界清晰  
2. 至少 **2 个**示例：JSON 请求/响应式；**Copilot 往返**式（关联 id）  
3. 无信封时行为与有信封时行为一致且可测  

**Plans**: 2 plans  

Plans:

- [ ] 04-01: 信封编码（如长度前缀 TLV 或 JSON 头 — 由规范选定）  
- [ ] 04-02: 示例场景与正/反向关联 id 用法  

**UI hint**: no  

---

### Phase 5: 状态机、错误与安全假设

**Goal**: 收敛 **状态机**、**错误码**、**关闭原因**，并文档化 **TLS 在边缘** 的威胁模型。  
**Depends on**: Phase 4  
**Requirements**: STATE-01, ERR-01, SEC-01  
**Success Criteria**:

1. 状态机图或表覆盖连接与会话主路径  
2. 错误码表可被测试引用；与连接/TLS 中止的对应关系有说明（不要求 WebSocket）  
3. **SEC-01** 明确 v1 不提供的保证与推荐部署方式  

**Plans**: 3 plans  

Plans:

- [ ] 05-01: 连接/session 状态机终稿  
- [ ] 05-02: 错误码与关闭码目录  
- [ ] 05-03: 安全假设与威胁模型（短文档）  

**UI hint**: no  

---

### Phase 6: 一致性测试套件

**Goal**: 以 **Go** 驱动可重复执行的**一致性测试**与 **golden 向量**，锁定规范。  
**Depends on**: Phase 5  
**Requirements**: TEST-01  
**Success Criteria**:

1. `go test ./...` 可在 CI 运行并通过  
2. `testdata/`（或等价）含**多组**代表性帧与负例  
3. 覆盖 Phase 1–5 中**已冻结**的关键行为（解析、路由模式、错误路径）  

**Plans**: 2 plans  

Plans:

- [ ] 06-01: 测试工程布局与向量格式约定  
- [ ] 06-02: 实现解析器/夹具最小闭环与回归集  

**UI hint**: no  

---

## Progress

**Execution Order:** 1 → 2 → 3 → 4 → 5 → 6  

| Phase | Plans Complete | Status | Completed |
|-------|----------------|--------|-----------|
| 1. 帧与 TLS 字节流 | 3/3 | Complete    | 2026-03-29 |
| 2. 会话与成员 | 3/3 | Complete    | 2026-03-29 |
| 3. 路由与流 | 0/3 | Not started | - |
| 4. 应用信封 | 0/2 | Not started | - |
| 5. 状态机与错误 | 0/3 | Not started | - |
| 6. 一致性测试 | 0/2 | Not started | - |
