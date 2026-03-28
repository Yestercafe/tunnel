<!-- gsd-project-start source:PROJECT.md -->
## Project

**Tunnel（传输隧道）**

面向公网部署的**中继隧道**：多个 **client** 通过同一服务在同一 **session** 内交换数据。创建者**开房**并获得 `session_id`（或邀请码），其它 **peer** 凭此加入。协议优先：先交付**帧格式、状态机、错误码**与**一致性测试**；实现语言为 **Go**，但第一步不强制完成端到端产品实现。

**面向谁：**需要在内网外、浏览器或脚本环境之间安全、可扩展地转发数据的开发者与小团队（**少量固定 peer**）。

**Core Value:** 在 **TLS 由边缘/服务器终止** 的前提下，用**同一套协议**同时支撑：**广播**、**私信（单播）**、**双向流**、**小消息与大块流**；并通过**可选应用信封**让 Web 前端、Copilot 管道等上层复用，而无需各自定义一套私有帧格式。

### Constraints

- **语言**：实现默认 **Go**（规范与测试可先于完整实现落地）。
- **交付顺序**：先**规范 + 一致性测试**，再展开实现与运维化。
- **传输**：须兼容 **WebSocket**；须能在**浏览器**中运行 client。
- **安全边界**：机密性与完整性主要依赖 **TLS（边缘/服务器）**；协议负责会话、成员与消息语义。
- **顺序模型**：**流内有序、流间可乱序**（在规范中用流 ID 或逻辑通道精确定义）。
<!-- gsd-project-end -->

<!-- gsd-stack-start source:research/STACK.md -->
## Technology Stack

## Recommended Stack
### Core Technologies
| Technology | Version | Purpose | Why Recommended |
|------------|---------|---------|-----------------|
| Go | 1.22+（建议跟踪当前稳定版） | 协议参考实现、一致性测试 runner | 静态编译、并发原语成熟、适合网络与二进制协议 |
| WebSocket (RFC 6455) | 协议标准 | 浏览器必选传输之一 | 浏览器原生 API；二进制帧可承载自定义协议 |
| `gobwas/ws` 或 `nhooyr/websocket` 等 | 以选型时最新为准 | Go 侧 WS 服务端/客户端 | 生态常用；需评估与「单连接多路复用」的契合度 |
| `github.com/stretchr/testify` | 1.9+ | 断言与测试套件组织 | Go 生态事实标准，适合一致性测试 |
### Supporting Libraries
| Library | Version | Purpose | When to Use |
|---------|---------|---------|-------------|
| `bufio` / `encoding/binary` | 标准库 | 定长与变长字段解析 | 帧解析与 golden test |
| `github.com/google/go-cmp` | 0.6+ | 深度比较结构体 | 复杂消息与状态机测试 |
### Development Tools
| Tool | Purpose | Notes |
|------|---------|-------|
| `go test ./...` | 运行一致性测试 | 可将向量放在 `testdata/` |
| `golangci-lint` | 静态检查 | 避免低级错误进入协议实现 |
## Installation
## Alternatives Considered
| Recommended | Alternative | When to Use Alternative |
|-------------|-------------|-------------------------|
| 先规范 + 测试，再选具体 WS 库 | 直接绑定某一框架 | 仅当团队已统一框架且愿意承担耦合 |
| 二进制帧自定义 | 仅用 JSON over WS | 大载荷与流式场景下带宽与解析成本更高 |
## What NOT to Use
| Avoid | Why | Use Instead |
|-------|-----|-------------|
| 在规范未定前锁死某一「多路复用库」 | 与自研帧/流 ID 模型可能冲突 | 规范中定义流抽象，实现再选库或自研 mux |
| 将 TLS 细节塞进 v1 核心帧 | 与「TLS 在边缘」决策冲突 | 在部署文档与能力位中扩展 |
## Stack Patterns by Variant
- 以 **WSS + 二进制帧** 为主路径；在规范中写清与 TCP 原生承载的差异（若有）。
- 共享同一「逻辑帧」编码；传输层仅负责把字节送进解析器。
## Version Compatibility
| 组件 | 兼容说明 |
|------|----------|
| Go 版本 | 建议在 `go.mod` 标明 `go 1.22` 或更新，便于使用新标准库特性 |
| WebSocket 子协议 | 若使用 Sec-WebSocket-Protocol，需在规范中登记名称与版本策略 |
## References
- RFC 6455 The WebSocket Protocol  
- Go 官方文档：Testing、Fuzzing（可用于生成随机帧向量）
<!-- gsd-stack-end -->

<!-- gsd-conventions-start source:CONVENTIONS.md -->
## Conventions

Conventions not yet established. Will populate as patterns emerge during development.
<!-- gsd-conventions-end -->

<!-- gsd-architecture-start source:ARCHITECTURE.md -->
## Architecture

Architecture not yet mapped. Follow existing patterns found in the codebase.
<!-- gsd-architecture-end -->

<!-- gsd-workflow-start source:GSD defaults -->
## GSD Workflow Enforcement

Before using Edit, Write, or other file-changing tools, start work through a GSD command so planning artifacts and execution context stay in sync.

Use these entry points:
- `/gsd-quick` for small fixes, doc updates, and ad-hoc tasks
- `/gsd-debug` for investigation and bug fixing
- `/gsd-execute-phase` for planned phase work

Do not make direct repo edits outside a GSD workflow unless the user explicitly asks to bypass it.
<!-- gsd-workflow-end -->



<!-- gsd-profile-start -->
## Developer Profile

> Profile not yet configured. Run `/gsd-profile-user` to generate your developer profile.
> This section is managed by `generate-claude-profile` -- do not edit manually.
<!-- gsd-profile-end -->
