---
title: "模板库：把输出结构固定下来"
slug: "workbuddy-template-library"
section_name: "系统沉淀"
public_summary: "模板库的意义不是省事，而是让同类任务在结构、质量和验收上保持稳定。"
estimated_read_minutes: 14
status: "published"
---

模板库不是“提示词大全”，也不是“收藏一堆格式”。它的核心价值是：把你曾经跑通的交付形状固化下来，让同类任务的结构、质量与验收口径保持稳定。

你可以把模板库理解成一种工程化资产：它不是为了让你少写几句话，而是为了让你在团队协作、批量生产、自动化时不会因为“每次都重新定义输出”而崩盘。

## 一句话结论

模板库里只放三类东西：**可复用输入、可复用输出、可复用验收**。缺任意一个，都不算模板，只算灵感。

## 什么值得进模板库

- 重复率高
- 输出结构稳定
- 质量标准明确

换成人话就是：你愿意下次继续按同一套结构交付，并且你能清楚判断“通过/不通过”。

## 模板至少包括

## 模板至少包括

- 输入说明

我建议再加两项（否则很难长期维护）：

- 版本与变更记录：模板为什么改了、改了什么
- 适用边界：什么情况不要用这个模板

## 最小模板规范（推荐）

你可以把每个模板写成一个独立文件（`.md` 或 `.txt`），遵循固定字段。这样你后面无论封装成 Skill、还是用于自动化，都不会乱。

```txt
name: {模板名}
type: {doc|sheet|slides|workflow}
use_for: {适用场景}
dont_use_for: {不适用场景}

inputs:
- {输入材料清单与口径}

outputs:
- {文件名/格式/落点}

constraints:
- {硬约束：不能改什么、必须遵守什么规范}

acceptance:
- {验收规则：通过/不通过}

version:
- v1: {日期} {变更说明}
```

## 一个模板从哪里来

模板不是设计出来的，是“从一次成功里抽出来的”。最稳的抽取方式：

1) 先让任务跑通 2~3 次（同结构、同口径）  
2) 把当时写过的任务说明与验收规则复制出来  
3) 把“你每次都要重复强调的部分”写进 constraints  
4) 把“你每次都要检查的部分”写进 acceptance  

当你做到这一步，你其实已经在做 Skill 设计，只是还没封装。

## 模板库怎么组织（不要像文件夹地狱）

模板库最容易变成“分类越分越细，最后谁也找不到”。我建议只按两层组织：

- 第一层：按交付物类型（doc/sheet/slides/workflow）
- 第二层：按业务主题或场景（sales/content/research/ops）

示例结构：

```txt
templates/
  doc/
    sales-weekly-report.md
    meeting-summary.md
  sheet/
    leads-pipeline.md
    cost-model.md
  slides/
    weekly-brief.md
    product-demo.md
  workflow/
    file-organization.md
    info-digest.md
```

## 模板与 Skill 的关系

模板库解决的是“可复用交付形状”，Skill 解决的是“可复用执行链路”。你可以把它们理解成：

- 模板：把结果长成什么样（输出结构 + 验收）
- Skill：怎么稳定得到这个结果（流程 + 工具 + 资源）

所以模板往往是 Skill 的核心组成部分，但模板本身不等于 Skill。

## 判断是否成熟（何时进入模板库）
- 输出格式
- 验收标准

更具体的成熟标准（建议同时满足）：

- 连续 3 次用同一结构交付，质量差异不大
- 你能在 1 分钟内说清验收标准（并且能拒绝不合格产物）
- 模板的输入范围清晰（不会每次都临场加材料）

## 可复用模板（示例）

下面给你两个“我们这套 Space 里已经跑通过”的模板示例，你可以直接拿去建模板文件。

### 示例 1：销售邮件 → 销售周报 PPT（slides + sheet + provenance）

```txt
name: sales_email_to_weekly_brief
type: workflow
use_for: 将一周客户邮件整理成可投屏汇报的销售周报
dont_use_for: 涉及敏感报价明细或需要强合规审批的客户材料

inputs:
- input/emails.md（只读；允许去重但不改原文）

outputs:
- output/sales-weekly-brief_v1.pptx
- output/leads.xlsx
- output/provenance.md

constraints:
- 不要杜撅事实；无法确认写“待确认”
- PPT <= 10 页；每页一个观点
- 所有关键结论必须在 provenance 里可追溯

acceptance:
- 文件落点正确（output/）
- 关键结论可追溯
- 线索清单去重后数量一致

version:
- v1: 2026-08-05 初版
```

### 示例 2：每日资讯摘要（doc/workflow）

```txt
name: daily_info_digest
type: workflow
use_for: 将当天资讯输入整理成 3 分钟可读摘要
dont_use_for: 需要深度长文研究的专题报告

inputs:
- rules/topics.md
- rules/rubric.md
- 今日原始材料（链接/要点/文件）

outputs:
- output/YYYY-MM-DD.md

constraints:
- 今日重点 <= 3
- 每条重点必须带来源

acceptance:
- 3 分钟内可读，且能回答“关注/忽略/下一步行动”
- 每条重点可追溯到来源

version:
- v1: 2026-08-05 初版
```

## 常见坑与排查

- 模板越来越多但没人用：说明你在收集“好看格式”而不是“可复用交付形状”；强制每个模板都带 acceptance。
- 模板写得太抽象：把输入与输出写成具体文件名与目录落点；否则无法验收。
- 模板没有边界：补 `dont_use_for`，否则会被滥用导致结果翻车。
