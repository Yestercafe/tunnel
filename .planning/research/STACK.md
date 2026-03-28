# Stack Research

**Domain:** 公网中继隧道协议（Go 实现、浏览器 WebSocket client）  
**Researched:** 2026-03-29  
**Confidence:** MEDIUM（协议 v1 未定稿，以下为生态与实现侧建议）

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

```bash
go mod init example.com/tunnel  # 若尚未初始化
go get github.com/stretchr/testify@latest
```

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

**若浏览器是唯一 client：**
- 以 **WSS + 二进制帧** 为主路径；在规范中写清与 TCP 原生承载的差异（若有）。

**若同时需要原生 TCP：**
- 共享同一「逻辑帧」编码；传输层仅负责把字节送进解析器。

## Version Compatibility

| 组件 | 兼容说明 |
|------|----------|
| Go 版本 | 建议在 `go.mod` 标明 `go 1.22` 或更新，便于使用新标准库特性 |
| WebSocket 子协议 | 若使用 Sec-WebSocket-Protocol，需在规范中登记名称与版本策略 |

## References

- RFC 6455 The WebSocket Protocol  
- Go 官方文档：Testing、Fuzzing（可用于生成随机帧向量）
