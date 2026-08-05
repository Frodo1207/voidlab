---
title: "多 Agent 系统设计：从单次任务到协作系统"
slug: "workbuddy-multi-agent-design"
section_name: "系统沉淀"
public_summary: "多 Agent 设计不是让角色变多，而是让任务分工、状态流和交接边界变清楚。"
status: "published"
---

## 什么时候需要多 Agent

当一个任务同时包含研究、生成、审核、发布等不同角色时。

## 设计重点

- 谁负责哪一段
- 上下文怎么传递
- 谁做最终确认

## 最忌讳

角色堆得很多，但没有清晰边界和状态流。
