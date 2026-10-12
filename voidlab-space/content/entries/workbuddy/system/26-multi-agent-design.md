---
title: "多 Agent 系统设计：从单次任务到协作系统"
slug: "workbuddy-multi-agent-design"
section_name: "系统沉淀"
public_summary: "多 Agent 设计不是让角色变多，而是让任务分工、状态流和交接边界变清楚。"
estimated_read_minutes: 15
status: "published"
---

多 Agent 最容易被误解成“多开几个聪明角色”。但真正的多 Agent 设计，重点从来不是角色数量，而是**角色边界、状态流、交接协议、汇总责任**。如果这些没定义清楚，多 Agent 只会把原本一个人的混乱，放大成一群角色的混乱。

这篇讨论的是：什么时候真的需要多 Agent，怎么拆角色不会互相踩，怎么让系统在可控范围内协作，而不是彼此复读、重复劳动、互相覆盖。

## 一句话结论

多 Agent 不是“让更多 Agent 同时干活”，而是**把一个复杂任务拆成多个单职责节点，并设计好它们之间传什么、什么时候传、谁最终负责汇总与验收**。

## 什么时候需要多 Agent

当一个任务同时包含研究、生成、审核、发布等不同角色时。

更具体一点，只有满足下面两个条件，才值得上多 Agent：

- 单 Agent 已经难以稳定处理，因为任务跨越了多种能力边界
- 这些边界可以清楚拆成“先后工序”或“并行工序”

典型适合多 Agent 的任务：

- 研究 → 写作 → 审核 → 发布
- 线索整理 → 数据分析 → 汇报生成
- 内容策划 → 脚本 → 镜头拆解 → 发布包

不适合多 Agent 的任务：

- 单一步骤、低复杂度、低风险的小任务
- 你自己都说不清楚目标与验收标准的任务
- 输入材料严重不稳定、输出形状还没跑通的任务

## 设计重点

- 谁负责哪一段
- 上下文怎么传递
- 谁做最终确认

我建议再补一个关键问题：**谁拥有“系统真相”**。如果没有一个统一的“真相源”，多 Agent 一定会出现版本漂移。

## 最忌讳

角色堆得很多，但没有清晰边界和状态流。

再展开一点，多 Agent 最常见的四种失败模式是：

- 角色重叠：两个 Agent 都在做“分析”，谁也说不清差别
- 上下文过载：每个 Agent 都拿到全部材料，结果谁都不专注
- 无人汇总：每个 Agent 都完成了子任务，但没有人对最终交付负责
- 状态断裂：中间产物没有固定格式，交接时只能重新解释

## 多 Agent 的最小架构

一个最小可用的多 Agent 系统，其实只需要 4 类角色：

| 角色 | 负责什么 | 典型输出 |
|---|---|---|
| 协调者 Coordinator | 拆任务、分派、收集结果、做最终汇总 | plan.md / handoff.md / final.md |
| 执行者 Worker | 完成某个单职责步骤 | research.md / script.md / report.md |
| 审核者 Reviewer | 检查质量、对照验收标准挑问题 | review.md / checklist.md |
| 发布者 Publisher | 把最终结果推到目标系统 | publish_log.md / delivery.md |

注意：不是每个系统都要 4 个角色，但任何一个复杂系统都应该有人承担这 4 类职责。

## 角色怎么拆，才不会互相踩

### 原则 1：按“单一职责”拆，不按“身份想象”拆

错误拆法：

- 内容专家
- 增长专家
- 运营专家

这类角色听起来厉害，但边界严重重叠。

更稳的拆法：

- 资料研究
- 结构整理
- 成稿输出
- 质量审核

也就是说，优先按“任务节点”拆，而不是按“专家称号”拆。

### 原则 2：每个 Agent 只拿必要上下文

研究 Agent 不需要看到发布规则，发布 Agent 也不需要读完整原始材料。

上下文传递最稳的方式是：中间产物文件化，而不是让后一个 Agent 继承前一个 Agent 的整段对话。

推荐目录结构：

```txt
multi-agent/
  input/
  work/
    01-plan.md
    02-research.md
    03-draft.md
    04-review.md
  output/
    final.md
  handoff.md
```

### 原则 3：只有一个角色拥有“最终汇总权”

多 Agent 系统里必须有一个角色拥有最终汇总权，通常就是 Coordinator。它负责：

- 汇总所有子结果
- 对照验收标准判断是否通过
- 决定是否返工
- 输出最终交付

如果没有这个角色，多 Agent 的最后一步通常会变成“大家都做完了，但没人敢拍板”。

## 状态流：多 Agent 真正的核心

多 Agent 设计最重要的其实不是角色，而是状态流。你可以把一个复杂任务抽象成下面这条链：

`planned → in_progress → awaiting_review → needs_revision → approved → delivered`

每个状态转移都应该有对应产物：

- `planned`：`plan.md`
- `in_progress`：阶段产物
- `awaiting_review`：提交审核说明
- `needs_revision`：返工意见
- `approved`：通过说明
- `delivered`：交付记录

只要状态流清楚，Agent 数量反而没那么重要。

## 多 Agent 的交接协议

交接协议至少要包含 5 件事：

- 当前状态：做到哪一步了
- 输入范围：本 Agent 使用了哪些材料
- 输出位置：结果落在哪个文件
- 未解决问题：哪些不确定项留给下一个节点
- 验收标准：下一个节点拿什么标准继续

你可以把它写成 `handoff.md`，并要求每个 Agent 完成后都更新。

## 可复用模板

### 模板 A：多 Agent 拆分卡

```md
# multi-agent design card

## 任务目标
- 

## 统一验收标准
- 

## 角色设计
1. Coordinator：
   - 输入：
   - 输出：
2. Worker A：
   - 输入：
   - 输出：
3. Reviewer：
   - 输入：
   - 输出：

## 状态流
- planned →
- in_progress →
- awaiting_review →
- approved →
- delivered
```

### 模板 B：handoff.md

```md
# handoff

## 当前角色
- 

## 当前状态
- 

## 已完成
- 

## 输出位置
- 

## 未解决问题
- 

## 下一个角色要做什么
- 
```

### 模板 C：WorkBuddy 任务说明（可直接复制）

```txt
目标：把一个复杂任务拆成多 Agent 协作系统，要求角色边界清晰、交接文件化、最终交付可验收。

输入：
- 原始任务目标
- 输入材料目录
- 验收标准

动作：
1) 先设计角色：Coordinator / Worker / Reviewer / Publisher（按需要裁剪）
2) 明确每个角色的输入、输出、边界与不负责项
3) 设计状态流（planned → in_progress → awaiting_review → approved → delivered）
4) 生成 handoff.md 模板与中间产物目录结构
5) 输出 multi-agent design card

约束：
- 不要按“听起来厉害”的身份拆角色；按任务节点拆
- 每个角色只拿必要上下文
- 必须指定一个最终汇总责任人

输出：
- output/multi-agent-design-card.md
- output/handoff-template.md
- output/workflow-state-machine.md

验收：
- 我能一眼看出每个角色负责什么、不负责什么
- 中间产物路径明确，交接不靠口头解释
- 最终交付责任明确，不会出现“谁都做了但没人拍板”
```

## 常见坑与排查

- 角色太多：先砍到 3~4 个职责节点；多 Agent 不是多越好
- 角色边界重叠：把“这个角色不负责什么”也写出来
- 所有人都拿全部上下文：强制文件化交接，只传必要中间结果
- 返工没有入口：补 `review.md` 与 `needs_revision` 状态，否则只能整体重来

## 最忌讳

角色堆得很多，但没有清晰边界和状态流。
