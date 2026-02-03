# 架构决策记录 (Architecture Decision Records)

## 概述

架构决策记录 (ADR) 是记录重要架构决策的文档。每个 ADR 记录一个特定的架构决策，包括决策的背景、决策内容、决策原因和后果。

## ADR 目录结构

```
docs/adr/
├── README.md              # 本文件
├── 0001-use-go-for-cli.md # 使用 Go 语言开发 CLI 工具
├── 0002-modular-architecture.md # 模块化架构设计
├── 0003-yaml-config.md    # 使用 YAML 配置文件
├── 0004-plugin-checkers.md # 插件化检查器架构
├── 0005-multi-format-reports.md # 多格式报告输出
├── template.md            # ADR 模板
└── index.md              # ADR 索引
```

## ADR 编号规则

- ADR 使用四位数字编号，从 0001 开始
- 编号按时间顺序递增
- 每个 ADR 都有唯一的编号

## ADR 状态

每个 ADR 有以下状态之一：

- **提议** (Proposed): 已提议但未决定
- **接受** (Accepted): 已接受并实施
- **弃用** (Deprecated): 已弃用但可能仍在使用
- **替代** (Superseded): 已被新的 ADR 替代

## 如何创建新的 ADR

1. 复制 `template.md` 文件
2. 使用下一个可用的编号命名文件
3. 填写 ADR 内容
4. 提交 PR 进行审查
5. 合并后更新 `index.md`

## ADR 审查流程

1. **提议阶段**: 创建 ADR 草案
2. **技术审查**: 技术团队审查
3. **架构审查**: 架构师审查
4. **决策**: 正式接受或拒绝
5. **实施**: 根据决策实施
6. **文档更新**: 更新相关文档

## 重要原则

- **透明性**: 所有重要架构决策都应该有记录
- **可追溯性**: 能够追溯决策的历史和原因
- **可维护性**: ADR 应该易于理解和维护
- **一致性**: 遵循相同的格式和结构

## 相关资源

- [ADR 规范](https://adr.github.io/)
- [架构决策记录最佳实践](https://docs.microsoft.com/en-us/azure/architecture/framework/decision-guide/adr/)
- [软件架构决策记录](https://thinkrelevance.com/blog/2011/11/15/documenting-architecture-decisions)

---

*最后更新: 2026-02-03*
