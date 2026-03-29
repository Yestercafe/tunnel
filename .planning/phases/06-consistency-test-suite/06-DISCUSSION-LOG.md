# Phase 6: 一致性测试套件 - Discussion Log

> **Audit trail only.** Do not use as input to planning, research, or execution agents.  
> Decisions are captured in **06-CONTEXT.md**。

**Date:** 2026-03-29  
**Phase:** 6 — 一致性测试套件  
**Areas discussed:** Golden 向量载体、覆盖范围、CI、负例与错误路径  
**Mode:** 用户授权助手 **自由选择**（等价于按推荐默认全选灰色区域）

---

## 1. Golden 向量载体

| Option | Description | Selected |
|--------|-------------|----------|
| A | 以 `testdata/` 文件为主，子目录分 framing / appenvelope | ✓ |
| B | 仅表驱动在 `_test.go` 内 |  |
| C | 默认 `//go:embed` |  |

**User's choice:** A（由助手代选）  
**Notes:** 与 ROADMAP「testdata 或等价」一致；hex 或 bin 均可，命名表达场景。

---

## 2. 覆盖范围与分层

| Option | Description | Selected |
|--------|-------------|----------|
| A | framing → appenvelope → 少量路由/流合成帧 | ✓ |
| B | 仅 framing |  |
| C | 第一轮即全协议矩阵 |  |

**User's choice:** A（由助手代选）  
**Notes:** 第三层仅引用规范示例，不展开完整状态机。

---

## 3. CI

| Option | Description | Selected |
|--------|-------------|----------|
| A | 添加最小 GitHub Actions 跑 `go test ./...` | ✓ |
| B | 仅文档约定本地执行 |  |

**User's choice:** A（由助手代选）  
**Notes:** 满足「可在 CI 运行」字面要求。

---

## 4. 负例与错误路径

| Option | Description | Selected |
|--------|-------------|----------|
| A | 解析层 + ErrCode 表；PROTOCOL_ERROR 编码随实现同步 | ✓ |
| B | 本阶段必须实现完整 PROTOCOL_ERROR 编解码 |  |

**User's choice:** A（由助手代选）  
**Notes:** 避免无实现先写测试；若计划引入小编码函数则同事务提交。

---

## Claude's Discretion

- 用户 **「你可以自由选择」** —— 全部灰色区域按 **CONTEXT 中推荐默认** 锁定；testdata 命名与 06-01/06-02 任务切分由规划阶段细化。

## Deferred Ideas

- Fuzz、testify、非 Go golden — 未纳入 Phase 6（见 CONTEXT `<deferred>`）。
