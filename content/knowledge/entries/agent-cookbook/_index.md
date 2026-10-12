---
title: "Agent 开发 Cookbook（零基础）"
slug: "agent-cookbook-index"
section_name: "Agent Cookbook"
public_summary: "围绕 Codex、Trae 等 Agent 宿主，从 0 到 1 学会设计、实现并交付一个可验收的 Agent。"
status: "published"
---

这里是 `Agent 开发 Cookbook（零基础）` 的总索引。它不是正文，而是这套 Space 的总入口：你可以先在这里看清整套课的结构，再决定自己要从主干开始，还是从某个案例切进去。

## Space 结构

- `nav/`：课程导读
- `intro/`：从大语言模型、Agent、MCP、Skill 到最小 Agent 开发前的主干章节
- `practice/`：主案例，覆盖单 Agent、真实世界案例与多 Agent 系统

## 课程导读（`nav/`）

- `00-space-intro.md`：课程背景、前置要求与最终目标

## 入门（`intro/`）

- `00-what-is-agent.md`：什么是大语言模型
- `01-agent-mental-model.md`：什么是 Agent
- `02-using-agent-hosts.md`：什么是 MCP
- `03-prompt-as-contract.md`：什么是 Skill

## 实战（`practice/`）

### 案例一：5 分钟实现第一个智能体

- `11-first-agent-dependencies-and-prompt.md`：依赖与系统提示词
- `12-first-agent-tools.md`：准备工具
- `13-first-agent-model-and-loop.md`：接入模型与跑主循环
- `14-first-agent-result-and-analysis.md`：运行结果与分析

### 案例二：旅游路线规划 Agent

- `20-travel-planning-agent.md`：项目概述与架构设计
- `21-travel-planning-prerequisites.md`：快速体验：先把项目跑起来
- `22-travel-planning-frontend-page.md`：数据模型设计
- `23-travel-planning-backend-and-db.md`：多 Agent 协作设计
- `24-travel-planning-mcp-integration.md`：MCP 工具接入
- `25-travel-planning-frontend-details.md`：前端页面设计
- `26-travel-planning-features.md`：功能实现与交互细节

## 怎么开始最稳

如果你是第一次系统学 Agent 开发，建议先看 `nav/00-space-intro.md`，再按 `intro/` 主干一路读到 `03-prompt-as-contract.md`。这样你会先把“大语言模型是什么、Agent 是什么、MCP 和 Skill 在系统里各自扮演什么角色”这几层搭起来，再进入案例。

如果已经对主干概念不陌生，也可以直接从 `practice/11-first-agent-dependencies-and-prompt.md` 开始，先跑通一个最小 Agent，再进入 `案例二：旅游路线规划 Agent`，把完整项目的背景、前置条件和技术搭建路径看清楚。

## 当前状态

这套 Space 现在先把最前面的结构收成了一个短导读：只讲背景、前置要求和最终目标。后面会继续按这个更紧的结构补正文。

## 推荐入口

- 想先建立概念入口：从 `00-what-is-agent.md` 开始
- 想先看第一个最小示例：从 `11-first-agent-dependencies-and-prompt.md` 开始
- 想先看完整项目案例：从 `20-travel-planning-agent.md` 开始
