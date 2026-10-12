---
title: "Schema + CRUD：把数据跑通"
slug: "fde-0to1-schema-crud"
section_name: "第三部分：最小项目闭环"
public_summary: "从数据模型到接口到页面：把一条 CRUD 链路跑通。"
estimated_read_minutes: 14
status: "published"
---

## 一句话结论

只要 CRUD 跑通，你就完成了 60% 的转行入门门槛。

## 这一章要讲什么

- 从 schema 设计开始，跑通：迁移 → CRUD API → 页面展示
- 列表页为什么必须做分页/筛选（这是最常见的真实需求）
- “写进去但读不到/读不一致”通常是哪里出了问题

## 交付物

- 数据表设计与迁移
- CRUD API
- 前端页面（列表/详情/编辑）

## 验收标准

- 列表页能分页、筛选
- 任何一次写入，都能在刷新后稳定读到

## 常见坑

- 没有唯一键/索引，数据一多就慢到不可用
- 迁移靠手改数据库，导致本地/线上 schema 不一致
- 只验证“能新增”，不验证“编辑/删除/权限/错误处理”

## 继续阅读

下一篇：`loop/22-deploy-and-observability.md`
