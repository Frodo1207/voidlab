---
title: "阅读路线：先跑通，再进案例"
slug: "agent-cookbook-reading-path"
section_name: "导读与导航"
public_summary: "给零基础读者两条阅读路线：最短闭环路线，以及完整案例路线。"
estimated_read_minutes: 6
status: "published"
---

## 路线 A：最短闭环

如果你想先做出第一个能用的 Agent，优先按这个顺序读：

- `00-what-is-agent.md`
- `01-agent-mental-model.md`
- `02-using-agent-hosts.md`
- `03-prompt-as-contract.md`
- `04-tools-and-first-external-action.md`
- `05-first-acceptable-agent.md`

这条路线的目标不是“学全”，而是先跑通一次真正的交付闭环。

它适合第一次系统接触 Agent 开发的人。因为这 6 篇会先把共同骨架搭起来：什么是 Agent、最小骨架是什么、宿主怎么用、Prompt 怎么写成契约、工具怎么接、什么叫可验收交付。等这条线走完，你再看案例时，会更容易看懂“为什么这里要加这一层”。

## 路线 B：完整案例

如果你已经接受“先用做出来的东西理解概念”，那读完入门主干后，可以直接进入：

- `20-dialogue-agent.md`
- `21-travel-assistant.md`
- `22-texas-holdem-multi-agent.md`

这条路线的作用不是替代主干，而是把主干能力逐层放大。`20-dialogue-agent.md` 会先把最小任务型对话系统立起来，`21-travel-assistant.md` 会把它推进到真实世界任务，`22-texas-holdem-multi-agent.md` 则会把单 Agent 扩展成多角色系统。

## 路线 C：按问题进入

- 想先搞清楚“Agent 到底是什么”：读 `00-what-is-agent.md`
- 想先搞清楚“共同骨架是什么”：读 `01-agent-mental-model.md`
- 想先搞清楚“宿主怎么用”：读 `02-using-agent-hosts.md`
- 想先搞清楚“Prompt 怎么写”：读 `03-prompt-as-contract.md`
- 想先搞清楚“工具怎么接”：读 `04-tools-and-first-external-action.md`
- 想先搞清楚“第一个能交付的 Agent 怎么成立”：读 `05-first-acceptable-agent.md`
- 想先看案例怎么落地：从 `20-dialogue-agent.md` 开始

如果你是带着明确问题来的，这条读法会更省时间。但最稳的方式仍然是：按问题切进去之后，至少把前后相邻的一篇也补上。这样你看到的就不只是单点答案，而是一段完整链路。

## 推荐读法

零基础最稳的读法仍然是：先走路线 A，再走路线 B。因为后面的旅行助手和德州扑克多 Agent，本质上都是前面主干能力的放大版。

如果你只打算先花一个晚上建立整体感觉，先读 `01-agent-mental-model.md`、`03-prompt-as-contract.md` 和 `20-dialogue-agent.md`，通常最容易抓住这套课的主心骨。
