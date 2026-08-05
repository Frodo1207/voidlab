---
title: "模型与成本：Auto / Max / 积分机制怎么控"
slug: "workbuddy-model-and-cost"
section_name: "入门"
public_summary: "模型选择影响质量、上下文和消耗。先建立成本感，再决定什么时候该开 Max。"
status: "published"
---

## Auto 模式

适合绝大多数日常任务，优点是省心，缺点是消耗不够透明。

## Max 模式

适合长任务或上下文很长的场景。它的价值是保细节，不是默认更高级。

## 积分机制的现实问题

作者提到 WorkBuddy 不是按 token 计费，而是按积分消耗，这会让任务成本预估不够直观。

## 使用建议

- 默认先用 Auto
- 遇到长任务再开 Max
- 重要任务先小样测试，再放大执行
