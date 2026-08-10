---
title: "Agent 开发 Cookbook（零基础）"
slug: "agent-cookbook-index"
section_name: "Agent Cookbook"
public_summary: "围绕 Codex、Trae 等 Agent 宿主，从 0 到 1 学会设计、实现并交付一个可验收的 Agent。"
status: "published"
---

这里是 `Agent 开发 Cookbook（零基础）` 的总索引。它不是正文，而是这套 Space 的总入口：你可以先在这里看清整套课的结构，再决定自己要从主干开始，还是从某个案例切进去。

## Space 结构

- `nav/`：导读、阅读路线、目录索引、资料来源
- `intro/`：从概念到第一个可验收 Agent 的主干章节
- `practice/`：3 个主案例，覆盖单 Agent、真实世界案例与多 Agent 系统

## 导读与导航（`nav/`）

- `00-space-intro.md`：这套课到底解决什么问题
- `01-reading-path.md`：怎么读这套课
- `02-directory-index.md`：一页导航
- `03-sources-and-citations.md`：资料来源与引用说明
- `99-changelog.md`：更新日志

## 入门（`intro/`）

- `00-what-is-agent.md`：什么是 Agent，不是什么
- `01-agent-mental-model.md`：Agent 的最小骨架
- `02-using-agent-hosts.md`：如何使用 Codex / Trae 这类 Agent 宿主
- `03-prompt-as-contract.md`：Prompt 不是咒语，是契约
- `04-tools-and-first-external-action.md`：工具调用与第一个外部动作
- `05-first-acceptable-agent.md`：第一个可验收 Agent

## 实战案例（`practice/`）

- `20-dialogue-agent.md`：案例一，对话 Agent
- `21-travel-assistant.md`：案例二，智能旅行助手
- `22-texas-holdem-multi-agent.md`：案例三，德州扑克多 Agent 系统

## 怎么开始最稳

如果你是第一次系统学 Agent 开发，建议先按 `intro/` 主干一路读到 `05-first-acceptable-agent.md`。这样你会先把“什么是 Agent、最小骨架是什么、怎么在宿主里工作、Prompt 怎么写成契约、工具怎么接、什么叫可验收”这几层搭起来，再进入案例。

如果你已经对主干概念不陌生，也可以直接从 `practice/20-dialogue-agent.md` 开始，先看最小单 Agent 是怎么成立的，再继续读旅行助手和多 Agent 系统。

## 当前状态

这套 Space 当前已经完成主干与 3 个主案例的正文补写。现在它不是“待填骨架”，而是一套可以直接阅读的零基础课程：前 6 篇负责建立共同骨架，后 3 篇负责把单 Agent、真实世界任务和多 Agent 系统一层层推开。

## 推荐入口

- 想先跑通一次最小闭环：从 `01-agent-mental-model.md` 开始
- 想先看怎么把任务写成系统：从 `03-prompt-as-contract.md` 开始
- 想先看真实案例：从 `20-dialogue-agent.md` 开始
