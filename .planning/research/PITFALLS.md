# Pitfalls Research

**Domain:** TLS 字节流中继、二进制协议  
**Researched:** 2026-03-29  
**Confidence:** MEDIUM

## Pitfall: 字节流粘包/半包与逻辑帧边界不一致

**Warning signs:** 解析随机失败、升级 TLS 后偶发截断、大包永远无法读完。  
**Prevention:** 在规范中明确：**长度前缀或定界规则**、**最大帧长**、**半包缓冲策略**；为大块载荷定义 **分片与重组状态机**（若在帧层拆分）。  
**Suggested phase:** 帧格式与传输绑定阶段（Phase 1）。

## Pitfall: 广播与私信语义混用导致重复投递或遗漏

**Warning signs:** 接收方收到重复、或私信被广播。  
**Prevention:** 帧内 **delivery 类型**（broadcast / unicast）与 **目标 peer 字段** 必填规则写死；Relay 做单元测试。  
**Suggested phase:** 路由与流阶段（Phase 3）。

## Pitfall: 顺序假设错误（全局序 vs 流内序）

**Warning signs:** 多流并发时业务乱序崩溃。  
**Prevention:** 规范写明 **顺序仅保证在单流内**；信封层 **关联 id** 用于跨流关联。  
**Suggested phase:** 路由与流 + 应用信封（Phase 3–4）。

## Pitfall: 能力协商与版本号被忽略导致互操作失败

**Warning signs:** 新旧 client 连上即断或静默丢字段。  
**Prevention:** 握手或首帧 **能力位**；未知能力 **必须可忽略或明确拒绝**。  
**Suggested phase:** 协议基础阶段（Phase 1）。

## Pitfall: 仅有人工测试无向量回归

**Warning signs:** 改一字段破坏多实现。  
**Prevention:** **一致性测试**目录化、CI 运行；关键帧 **golden bytes**。  
**Suggested phase:** 一致性测试阶段（Phase 6）。

## Pitfall: 将「TLS 在边缘」误解为协议内无需认证

**Warning signs:** 任意人猜 session_id 即可窃听或注入。  
**Prevention:** v1 仍建议 **短 token / join 凭证**；文档标注威胁模型。  
**Suggested phase:** 会话生命周期（Phase 2）与安全说明（Phase 5）。
