---
title: "数据库基础：你必须补上的一课"
slug: "fde-0to1-database-basics"
section_name: "第二部分：基础补全"
public_summary: "面向转行：用最短路径掌握数据库与数据建模（SQL、索引、事务、迁移）。"
estimated_read_minutes: 16
status: "published"
---

## 一句话结论

数据库不是后端专属技能，它决定了你的产品如何存活、如何变更、如何扩展。

## 这一章要讲什么

- 表、字段、主键、外键：最小数据建模
- 常见查询：筛选、排序、分页、聚合
- 索引与性能：为什么慢
- 事务与一致性：为什么会“写进去但读不到”
- 迁移：为什么团队必须有 schema 版本

## 交付物

- 一份“数据模型草图”：`schema-sketch.md`
- 一份 SQL 练习：`sql-exercises.md`

## 验收标准

- 你能解释：一个列表页的数据是如何被查询、分页并返回的

## 常见坑

- 只会“能查出来”，不会考虑分页、索引和慢查询
- 没有迁移（migration）习惯：schema 改了但没有版本记录
- 把业务口径写在代码里不写在文档里，久了就对不齐

## 继续阅读

下一篇：`foundation/14-api-and-auth-minimum.md`
