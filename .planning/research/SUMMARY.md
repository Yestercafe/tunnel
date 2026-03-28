# Research Summary

**Project:** Tunnel（传输隧道）  
**Synthesized:** 2026-03-29

## Key Findings

### Stack

- **Go** 适合作为参考实现与一致性测试宿主；v1 主路径为 **TCP+TLS 字节流**。  
- 帧解析优先 **标准库 + 测试向量**；可选 **v2 WebSocket 适配** 与逻辑帧解耦。  
- 测试侧推荐 **testify**、**go-cmp**、**testdata 向量**。

### Table Stakes

- Session 创建/加入、**广播 + 私信**、**双向流**、**流内有序/流间乱序**、**可选应用信封**、**版本与能力位**、**TLS 字节流成帧**。

### Architecture

- 清晰分层：**传输 → 帧 → 会话/路由 → 可选应用层**。  
- Relay 负责成员与转发；顺序与关联主要由 **流 ID + 信封关联 id** 解决。

### Watch Out For

- **粘包/半包与帧边界**、**广播/私信语义**、**版本/能力协商**、**缺少机器可回归向量**、**低估 join 凭证必要性**。

## Files

| File | Role |
|------|------|
| `STACK.md` | 实现栈与工具 |
| `FEATURES.md` | 能力分层与依赖 |
| `ARCHITECTURE.md` | 组件与构建顺序 |
| `PITFALLS.md` | 风险与对应阶段 |

## Next Steps

- 将上表能力落实为 **REQ-*** 与 **Phase 映射**（见 `REQUIREMENTS.md` / `ROADMAP.md`）。  
- Phase 1 起优先冻结 **帧头与 TLS 字节流成帧**，减少后期互操作成本。
