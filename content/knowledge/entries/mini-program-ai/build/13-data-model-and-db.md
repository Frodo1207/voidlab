---
title: "数据模型与数据库设计"
slug: "mini-program-ai-data-model-and-db"
section_name: "主项目闭环"
public_summary: "围绕活动、用户、报名三类核心对象，把项目数据结构定下来。"
estimated_read_minutes: 12
status: "published"
---

## 一句话结论

数据库设计不清楚，后面的接口、状态和审核流一定会反复返工。

## 这一章要讲什么

- 活动报名小程序最少需要哪些表和字段
- 活动、用户、报名三者怎么关联
- 哪些字段是展示字段，哪些字段是状态字段

## 交付物

- 一份表结构草案：`schema.md`

## 验收标准

- 你能解释报名记录为什么要单独成表，以及活动容量如何控制

## 常见坑

- 把表设计成页面长什么样就长什么样
- 状态字段随手命名
- 一开始不考虑审核和名额上限

## 继续阅读

下一篇：`build/14-api-and-business-flow.md`
