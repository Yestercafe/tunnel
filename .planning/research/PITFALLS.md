# Pitfalls Research

**Domain:** 在已有协议/成帧/测试基线上，为隧道系统**首次**接入最小 Relay + Client（TCP+TLS、会话路由、并发与背压）  
**Researched:** 2026-04-04  
**Confidence:** **MEDIUM**（与 `docs/spec/v1/`、`pkg/framing` 及 Go 网络实践对齐；部分为行业模式，非本项目独有官方列表）

## Critical Pitfalls

### Pitfall 1：把「TLS 连接」当成「消息边界」

**What goes wrong:**  
在 `crypto/tls.Conn` 或 `net.Conn` 上调用 `Read`，假定一次调用返回**恰好一帧**或「读完一整条应用消息」。实际得到的是**任意长度分片**；若未与 `ParseFrame` / 半包逻辑配合，会出现随机截断、粘包漏解析、或把半包当致命错误关闭连接。

**Why it happens:**  
TLS 记录在底层被解密后仍经流抽象暴露；协议文档已定义长度前缀成帧，但实现者仍习惯「一次 Read 一条消息」的心智模型。

**How to avoid:**  
严格按 [connection-state.md](docs/spec/v1/connection-state.md)：**追加到缓冲区 → 至少 10 字节头 → `payload_len` → 收满再交付**；粘包时**循环**解析直到不足以构成下一帧。对 `ErrNeedMore`（或等价「半包」）**等待更多字节**，不当作协议错误。复用已有 `pkg/framing.ParseFrame` 语义。

**Warning signs:**  
单测用「整帧一次写入」全绿，但集成测试或真实网络下偶发 `PROTOCOL_ERROR`、首包解析失败、或仅在高延迟下复现的截断。

**Phase to address:**  
**RELAY-IMPL**（入站读循环与缓冲策略）、**CLIENT-IMPL**（出站/入站对称实现）

---

### Pitfall 2：成帧层状态与 session 层状态混在同一套「连接状态」里

**What goes wrong:**  
把「缓冲区还不够一帧」与「尚未 JOIN_ACK」混成单一枚举，或在未区分层次的情况下短路错误处理；导致半包被误判为鉴权失败，或 session 门禁与成帧致命错误恢复路径纠缠，难以测试与排障。

**Why it happens:**  
规范已拆成 [connection-state.md](docs/spec/v1/connection-state.md) 与 [session-state.md](docs/spec/v1/session-state.md)，实现时图省事合并状态机。

**How to avoid:**  
**成帧层**只产出完整帧 payload（及头字段）；**仅当**一帧完整交付后，再由上层解析 `msg_type`、路由。致命成帧错误（如 `ERR_FRAME_TOO_LARGE`、`ERR_PROTO_VERSION`）与「需更多字节」必须分支清晰（规范 STATE-01 已强调）。

**Warning signs:**  
同一函数既 `Read` 又解析 `SESSION_*` 又处理路由，且缺少明确的「先 ParseFrame 再分支」顺序。

**Phase to address:**  
**RELAY-IMPL**（分层模块边界）、**CLIENT-IMPL**（镜像结构便于对照测试）

---

### Pitfall 3：在 JOIN_ACK 之前路由数据面或信任 `peer_id`

**What goes wrong:**  
在 `SESSION_JOIN_ACK` 未确认前转发 `STREAM_*`、广播或单播，或未加入者被错误计入 session 成员表，造成跨会话串话或违反 [session-state.md](docs/spec/v1/session-state.md) 的门禁。

**Why it happens:**  
控制面与数据面消息类型多，实现时先写「 happy path 数据转发」，后补门禁。

**How to avoid:**  
Relay 维护：**连接 → 是否已绑定 session + 合法 peer_id**；仅在对该连接已发送（并记录）**有效 JOIN 成功语义**后，才允许该连接上的数据面路由。创建者路径亦须在 `SESSION_CREATE_ACK` / 等价约定明确后再进入可路由状态（与规范中创建者/加入者状态一致）。

**Warning signs:**  
「单测直接发数据帧」能通过，但缺少「未 JOIN 发数据须拒斥/错误码」的负例。

**Phase to address:**  
**RELAY-IMPL**（会话表 + 连接表不变量）、**E2E-DEMO**（负例路径）

---

### Pitfall 4：Session 路由表与并发：在 map 上无锁或锁顺序不一致

**What goes wrong:**  
`session_id → peers`、`peer_id → conn` 等在多 goroutine 下竞态：重复 `peer_id`、向已关闭连接写、或死锁（A 锁 session 再锁 peer，B 相反）。

**Why it happens:**  
每个连接一个读循环、一个写路径，快速实现时先共享全局 map。

**How to avoid:**  
明确**谁拥有**连接句柄（通常每连接一 goroutine 读，写路径经 channel 或每连接互斥）；**session 结构体**内保护成员列表的锁或与连接生命周期绑定的层次化锁。**文档化锁顺序**（例如：先锁 session 再锁 peer 条目，或仅用 channel 串行化 session 变更）。v1.1 规模小但仍应一次性做对，避免「以后再线程安全」。

**Warning signs:**  
`go test -race` 报警、偶发 `nil` 连接、或仅在高并发压测下出现的重复投递。

**Phase to address:**  
**RELAY-IMPL**（核心）；**E2E-DEMO** 中带 `-race` 的集成测试

---

### Pitfall 5：背压缺失 — 无界队列或阻塞写拖垮整个 Relay

**What goes wrong:**  
向慢速 peer 转发时：无界 channel/切片堆积导致内存暴涨；或在**持锁**情况下对 TLS `Write` 阻塞，阻塞其他 session 的进度；或 goroutine 无上限地为每消息 spawn。

**Why it happens:**  
最小实现优先「功能正确」，忽略「一对多广播时最慢消费者」的耦合。

**How to avoid:**  
为每连接出站路径设**有界**缓冲或显式丢弃/关闭策略（v1.1 可文档化「最小实现行为」）；避免在全局锁内做网络写；考虑每连接独立写 goroutine + 有界队列。与规范中**流内有序**一致的前提下，明确「慢 peer 是否反压发送方」的语义（至少实现上不可无限缓冲）。

**Warning signs:**  
两 peer 正常、三 peer 或一慢一快时延迟飙升或 OOM；CPU 不高但内存持续上涨。

**Phase to address:**  
**RELAY-IMPL**；**E2E-DEMO** 可加入简单慢消费者场景（若范围允许）

---

### Pitfall 6：TLS 配置与安全假设偏离规范

**What goes wrong:**  
生产代码误留 `InsecureSkipVerify`；或未校验服务端名与证书 SAN；或假设「TLS 上了」等于应用层无需关心会话凭证。与 [security-assumptions.md](docs/spec/v1/security-assumptions.md) 不一致。

**Why it happens:**  
本地自签证书调试方便，未及时收回。

**How to avoid:**  
开发/测试用构建标签或显式 `DEV` 环境变量区分；CI 使用固定测试证书或 `httptest`/`tls` 测试助手。文档保留：**机密性在 TLS 边缘终止**，协议 token 仍须按 [join-credentials.md](docs/spec/v1/join-credentials.md) 实现。

**Warning signs:**  
仓库内硬编码 `true` 的跳过校验；无测试覆盖证书失败路径。

**Phase to address:**  
**CLIENT-IMPL**（Dial TLS）、**RELAY-IMPL**（Load 证书）；**E2E-DEMO** 用可信测试夹具

---

### Pitfall 7：测试策略与协议先行代码库脱节

**What goes wrong:**  
仅重复 `pkg/framing` 级单测，无「半包 + 粘包」读循环测试；无两 peer 真实 TCP+TLS 管道；或测试依赖 wall-clock sleep 导致 CI 不稳定。未利用现有 `testdata/` 负例做端到端对齐。

**Why it happens:**  
v1.0 已交付一致性测试，实现阶段误以为「再测一遍 framing 即可」。

**How to avoid:**  
**集成测试**：`net.Listener` + `tls` + 控制 `Conn` 写入分片（强制半包/粘包）。**场景**：创建 session、加入、广播、单播、未 JOIN 发数据、错误码。与 **TEST-01** 精神一致：可重复、无手动步骤依赖。对 Relay 路由测**确定性顺序**（若规范允许流间乱序，断言集合而非严格时序）。

**Warning signs:**  
覆盖率主要落在 `pkg/framing`，`cmd/` 或 `internal/relay` 几乎无测试；问题只在手动 CLI 复现。

**Phase to address:**  
**E2E-DEMO**（主责）；**RELAY-IMPL / CLIENT-IMPL** 交付时即带 table-driven 集成测试

---

## Technical Debt Patterns

| Shortcut | Immediate Benefit | Long-term Cost | When Acceptable |
|----------|-------------------|----------------|-----------------|
| 全局无界 `[]byte` 接收缓冲 | 实现简单 | 恶意/错误对端大 `payload_len` 前已耗尽内存 | **Never** — 须在成帧层 enforce `MaxPayloadLen`（已与 `pkg/framing` 一致） |
| 单锁保护整个 Relay | 代码少 | 延迟与死锁风险随功能增长 | 仅 v1.1 极小原型且通过 `-race` 与简单压测 |
| 日志代替可观测错误码 | 排障快 | 客户端无法程序化重试 | v1.1 可最小日志，但规范错误码路径必须可达 |
| 仅 happy-path E2E | 先演示 | 门禁与负例未覆盖 | 演示前至少补 1～2 个负例 |

## Integration Gotchas

| Integration | Common Mistake | Correct Approach |
|---------------|----------------|------------------|
| `crypto/tls` 服务端 | 证书链/ALPN 配置与客户端不一致导致握手失败 | 最小化可重复：测试套件内自包含证书或文档化 `mkcert` 步骤 |
| `net.Conn` 与 `ParseFrame` | 每次 `Read` 直接 `ParseFrame` 而不保留半包 | 持久化读缓冲区，循环解析 |
| Session 广播 | 向发送者本人再投递一份（若产品未定义） | 与 [routing-modes.md](docs/spec/v1/routing-modes.md) 产品语义一致；通常排除自环或明确文档 |
| 邀请码 / `session_id` | 两字段索引不同步 | Relay 侧单一事实来源（创建时同时登记），原子更新 |

## Performance Traps

| Trap | Symptoms | Prevention | When It Breaks |
|------|----------|------------|----------------|
| 每帧分配大 slice | GC 压力、延迟尖刺 | 复用 buffer、`sync.Pool`（按需） | 高频小消息或大块流并存时 |
| 广播时串行同步写 | 尾延迟高 | 每 peer 异步写 + 有界队列 | peer 数 > 2 或跨地域 RTT |
| TLS 小读缓冲导致 syscall 多 | CPU 在系统调用上 | 较大用户态缓冲、一次 Read 多读 | 吞吐压测时（参见 Go TLS 缓冲相关讨论） |

## Security Mistakes

| Mistake | Risk | Prevention |
|---------|------|------------|
| 跳过 TLS 校验连接公网 Relay | MITM、会话劫持 | 生产路径强制校验；测试隔离 |
| 日志打印 `join` token / 完整 payload | 凭证泄露 | 脱敏或 debug 级别门控 |
| 未限制单 IP 连接数（最小实现可接受但需知） | 简易 DoS | v1.1 文档列入已知限制；后续阶段加固 |

## UX Pitfalls（面向演示与 CLI）

| Pitfall | User Impact | Better Approach |
|---------|-------------|-----------------|
| 错误信息仅 `PROTOCOL_ERROR` 无附加上下文 | 集成方无法区分重试与致命失败 | 记录规范错误码 + 简要原因（不泄露敏感信息） |
| CLI 参数 session 与 invite 混用 | 用户误加入失败 | 与规范一致命名，`--session-id` / `--invite-code` 显式区分 |

## "Looks Done But Isn't" Checklist

- [ ] **成帧：** 半包与粘包均在测试中出现（不仅是「整帧一次读」）
- [ ] **门禁：** 未 `JOIN_ACK` 的数据面帧被拒绝且有对应错误/行为
- [ ] **路由：** 同 session 两 peer 广播与单播双向可达；错误 session 不可达
- [ ] **TLS：** 非跳过校验路径在 CI 中执行至少一次
- [ ] **并发：** `go test -race` 覆盖 Relay 核心路径
- [ ] **资源：** 慢 peer 或断连不会导致无界内存增长（有界或关闭连接）

## Recovery Strategies

| Pitfall | Recovery Cost | Recovery Steps |
|---------|---------------|----------------|
| 成帧与 session 状态耦合 | HIGH | 抽出读缓冲+ParseFrame 层；上层仅收完整 payload |
| 无背压导致 OOM | MEDIUM | 加有界队列与降级策略；必要时断开最慢 peer |
| TLS 误配置上线 | HIGH | 配置审计；分环境证书；自动化握手测试 |

## Pitfall-to-Phase Mapping

| Pitfall | Prevention Phase | Verification |
|---------|------------------|--------------|
| TLS 无消息边界 / 半包粘包 | RELAY-IMPL + CLIENT-IMPL | 分片读集成测试；与 `connection-state.md` 行为一致 |
| 成帧 vs session 状态混淆 | RELAY-IMPL | 代码审查检查分层；单元测试只测一层 |
| JOIN 前路由 | RELAY-IMPL | 负例测试未 JOIN 发 `STREAM_*` |
| 并发 map 竞态 | RELAY-IMPL | `-race`；并发加入/广播测试 |
| 背压 | RELAY-IMPL | 慢消费者或阻塞写模拟 |
| TLS 安全捷径 | CLIENT-IMPL + RELAY-IMPL | 无 InsecureSkipVerify 的生产构建；证书失败测试 |
| 测试不足 | E2E-DEMO | TCP+TLS 双 peer；可重复脚本/`go test` |

## Sources

- 项目规范：`docs/spec/v1/connection-state.md`、`session-state.md`、`transport-binding.md`、`routing-modes.md`、`security-assumptions.md`
- 实现参考：`pkg/framing`（`ErrNeedMore`、`MaxPayloadLen`）
- Go 网络实践：TLS/流式 `Read` 返回分片为常态；需自建成帧循环（与 Stack Overflow / Go issue 讨论一致 — **MEDIUM** 置信度）
- 背压与死锁：通用服务模式（有界队列、避免锁内 IO）— **MEDIUM** 置信度

---
*Pitfalls research for: Tunnel v1.1 最小 Relay + Client（协议先行代码库上的首次运行时集成）*  
*Researched: 2026-04-04*
