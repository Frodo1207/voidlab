---
title: "报名表单、状态与校验"
slug: "mini-program-ai-signup-form-and-state"
section_name: "主项目闭环"
public_summary: "专门解决活动报名里最容易出问题的一层：表单字段、状态变化和提交校验。"
estimated_read_minutes: 11
status: "published"
---

## 一句话结论

活动报名项目最像业务系统的地方，就在表单和状态是否真的站得住。

## 这一章要讲什么

- 报名表单字段怎么设计才够用
- 待提交、已报名、已通过、已拒绝这些状态怎么定义
- 名额已满、重复报名、信息不完整要怎么处理

## 交付物

- 一份状态图：`signup-state.md`
- 一份表单字段定义：`form-schema.md`

## 验收标准

- 你能描述报名提交前、中、后的状态变化和校验逻辑

## 常见坑

- 表单字段一开始就太多
- 状态命名混乱
- 只做前端校验，不处理后端重复提交

## 继续阅读

下一篇：`build/16-auth-and-admin.md`
