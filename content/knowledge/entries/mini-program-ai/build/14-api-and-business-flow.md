---
title: "接口设计与业务流转"
slug: "mini-program-ai-api-and-business-flow"
section_name: "主项目闭环"
public_summary: "从接口角度把活动展示、报名提交、审核处理和状态查询串起来。"
estimated_read_minutes: 12
status: "published"
---

## 一句话结论

接口不是页面的附属物，而是把业务流真正跑起来的中轴。

## 这一章要讲什么

- 活动列表、详情、报名、审核分别需要什么接口
- 用户和管理员各自会触发哪些业务动作
- 哪些状态转换必须在接口层保证一致性

## 交付物

- 一份接口清单：`api-contract.md`

## 验收标准

- 你能把报名流程讲成一条清晰的业务状态流

## 常见坑

- 只有接口名，没有业务语义
- 前端自己猜状态变化
- 审核和报名写成同一条模糊接口

## 继续阅读

下一篇：`build/15-signup-form-and-state.md`
