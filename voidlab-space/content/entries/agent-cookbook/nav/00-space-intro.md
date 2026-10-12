---
title: "初识 Agent 开发 Cookbook：从会用 AI 到会做 Agent"
slug: "agent-cookbook-space-intro"
section_name: "导读与导航"
public_summary: "这套课关注的不是某个框架，而是如何在 Codex、Trae 等宿主中，从 0 到 1 做出一个可验收的 Agent。"
estimated_read_minutes: 7
status: "published"
---

## 一句话结论

这套 Cookbook 解决的不是“怎么把 AI 用得更顺”，而是“怎么从会聊天，走到会定义任务、接工具、控边界、交付结果”。如果你以前主要是在和 AI 对话，这套课会把你往前推一步：开始把它当成一个可以设计、限制、验收的工作系统。

![Agent 开发 Cookbook 学习路线图](data:image/svg+xml;utf8,%3Csvg%20xmlns%3D%27http%3A//www.w3.org/2000/svg%27%20width%3D%27960%27%20height%3D%27540%27%20viewBox%3D%270%200%20960%20540%27%3E%3Crect%20width%3D%27960%27%20height%3D%27540%27%20fill%3D%27%23f7f7f5%27/%3E%3Crect%20x%3D%2736%27%20y%3D%2736%27%20width%3D%27888%27%20height%3D%27468%27%20rx%3D%2724%27%20fill%3D%27%23ffffff%27%20stroke%3D%27%23e5e5e0%27/%3E%3Ctext%20x%3D%2764%27%20y%3D%2782%27%20font-family%3D%27Arial%2C%20PingFang%20SC%2C%20Microsoft%20YaHei%27%20font-size%3D%2730%27%20font-weight%3D%27800%27%20fill%3D%27%23111111%27%3EAgent%20%E5%BC%80%E5%8F%91%20Cookbook%20%E8%B7%AF%E7%BA%BF%3C/text%3E%3Ctext%20x%3D%2764%27%20y%3D%27114%27%20font-family%3D%27Arial%2C%20PingFang%20SC%2C%20Microsoft%20YaHei%27%20font-size%3D%2715%27%20fill%3D%27%235f5e58%27%3E%E5%85%88%E8%B7%91%E9%80%9A%E6%9C%80%E5%B0%8F%E9%97%AD%E7%8E%AF%EF%BC%8C%E5%86%8D%E8%BF%9B%E5%85%A5%E7%9C%9F%E5%AE%9E%E4%B8%96%E7%95%8C%E6%A1%88%E4%BE%8B%E3%80%82%3C/text%3E%3Crect%20x%3D%2764%27%20y%3D%27150%27%20width%3D%27220%27%20height%3D%27140%27%20rx%3D%2718%27%20fill%3D%27%23fbfbfa%27%20stroke%3D%27%23deded8%27/%3E%3Crect%20x%3D%2764%27%20y%3D%27150%27%20width%3D%278%27%20height%3D%27140%27%20fill%3D%27%23c4f000%27/%3E%3Ctext%20x%3D%2790%27%20y%3D%27188%27%20font-family%3D%27Arial%2C%20PingFang%20SC%2C%20Microsoft%20YaHei%27%20font-size%3D%2722%27%20font-weight%3D%27750%27%20fill%3D%27%23111111%27%3E%E5%85%88%E5%BB%BA%E5%BF%83%E6%99%BA%E6%A8%A1%E5%9E%8B%3C/text%3E%3Ctext%20x%3D%2790%27%20y%3D%27222%27%20font-family%3D%27Arial%2C%20PingFang%20SC%2C%20Microsoft%20YaHei%27%20font-size%3D%2714%27%20fill%3D%27%235f5e58%27%3E%E4%BB%80%E4%B9%88%E6%98%AF%20Agent%EF%BC%8C%E6%9C%80%E5%B0%8F%E9%AA%A8%E6%9E%B6%E6%98%AF%E4%BB%80%E4%B9%88%EF%BC%9F%3C/text%3E%3Ctext%20x%3D%2790%27%20y%3D%27248%27%20font-family%3D%27Arial%2C%20PingFang%20SC%2C%20Microsoft%20YaHei%27%20font-size%3D%2714%27%20fill%3D%27%235f5e58%27%3E%E6%90%9E%E6%B8%85%E7%9B%AE%E6%A0%87%E3%80%81%E8%BE%93%E5%85%A5%E3%80%81%E5%B7%A5%E5%85%B7%E3%80%81%E7%8A%B6%E6%80%81%E3%80%81%E8%BE%93%E5%87%BA%E3%80%82%3C/text%3E%3Crect%20x%3D%27370%27%20y%3D%27150%27%20width%3D%27220%27%20height%3D%27140%27%20rx%3D%2718%27%20fill%3D%27%23fbfbfa%27%20stroke%3D%27%23deded8%27/%3E%3Crect%20x%3D%27370%27%20y%3D%27150%27%20width%3D%278%27%20height%3D%27140%27%20fill%3D%27%230f7b6c%27/%3E%3Ctext%20x%3D%27396%27%20y%3D%27188%27%20font-family%3D%27Arial%2C%20PingFang%20SC%2C%20Microsoft%20YaHei%27%20font-size%3D%2722%27%20font-weight%3D%27750%27%20fill%3D%27%23111111%27%3E%E5%85%88%E5%AD%A6%E4%BC%9A%E5%9C%A8%E5%AE%BF%E4%B8%BB%E9%87%8C%E5%B7%A5%E4%BD%9C%3C/text%3E%3Ctext%20x%3D%27396%27%20y%3D%27222%27%20font-family%3D%27Arial%2C%20PingFang%20SC%2C%20Microsoft%20YaHei%27%20font-size%3D%2714%27%20fill%3D%27%235f5e58%27%3ECodex%20%E3%80%81Trae%20%E8%BF%99%E7%B1%BB%20Agent%20%E5%AE%BF%E4%B8%BB%E6%80%8E%E4%B9%88%E7%9C%8B%E4%BB%BB%E5%8A%A1%E3%80%81%E7%9C%8B%E6%96%87%E4%BB%B6%E3%80%81%E7%9C%8B%E4%BA%A7%E7%89%A9%E3%80%82%3C/text%3E%3Crect%20x%3D%27676%27%20y%3D%27150%27%20width%3D%27220%27%20height%3D%27140%27%20rx%3D%2718%27%20fill%3D%27%23fbfbfa%27%20stroke%3D%27%23deded8%27/%3E%3Crect%20x%3D%27676%27%20y%3D%27150%27%20width%3D%278%27%20height%3D%27140%27%20fill%3D%27%234f46e5%27/%3E%3Ctext%20x%3D%27702%27%20y%3D%27188%27%20font-family%3D%27Arial%2C%20PingFang%20SC%2C%20Microsoft%20YaHei%27%20font-size%3D%2722%27%20font-weight%3D%27750%27%20fill%3D%27%23111111%27%3E%E5%81%9A%E5%87%BA%E7%AC%AC%E4%B8%80%E4%B8%AA%20Agent%3C/text%3E%3Ctext%20x%3D%27702%27%20y%3D%27222%27%20font-family%3D%27Arial%2C%20PingFang%20SC%2C%20Microsoft%20YaHei%27%20font-size%3D%2714%27%20fill%3D%27%235f5e58%27%3E%E8%B7%91%E9%80%9A%E4%B8%80%E6%AC%A1%E5%8F%AF%E9%AA%8C%E6%94%B6%E9%97%AD%E7%8E%AF%EF%BC%9A%E7%9B%AE%E6%A0%87%20%E2%86%92%20%E5%B7%A5%E5%85%B7%20%E2%86%92%20%E4%BA%A7%E7%89%A9%E3%80%82%3C/text%3E%3Cpath%20d%3D%27M284%20220%20L354%20220%27%20stroke%3D%27%23999%27%20stroke-width%3D%273%27/%3E%3Cpolygon%20points%3D%27354%20220%2C340%20212%2C340%20228%27%20fill%3D%27%23999%27/%3E%3Cpath%20d%3D%27M590%20220%20L660%20220%27%20stroke%3D%27%23999%27%20stroke-width%3D%273%27/%3E%3Cpolygon%20points%3D%27660%20220%2C646%20212%2C646%20228%27%20fill%3D%27%23999%27/%3E%3Crect%20x%3D%2764%27%20y%3D%27336%27%20width%3D%27832%27%20height%3D%27136%27%20rx%3D%2718%27%20fill%3D%27%23111111%27/%3E%3Ctext%20x%3D%2790%27%20y%3D%27380%27%20font-family%3D%27Arial%2C%20PingFang%20SC%2C%20Microsoft%20YaHei%27%20font-size%3D%2724%27%20font-weight%3D%27800%27%20fill%3D%27%23ffffff%27%3E3%20%E4%B8%AA%E6%A1%88%E4%BE%8B%E5%8F%AA%E8%A7%A3%E5%86%B3%203%20%E4%BB%B6%E4%BA%8B%3C/text%3E%3Ctext%20x%3D%2790%27%20y%3D%27414%27%20font-family%3D%27Arial%2C%20PingFang%20SC%2C%20Microsoft%20YaHei%27%20font-size%3D%2715%27%20fill%3D%27%23d6d6d0%27%3E%E5%AF%B9%E8%AF%9D%20Agent%20%E6%95%99%E4%BD%A0%E5%81%9A%E6%9C%80%E5%B0%8F%E5%8D%95%20Agent%EF%BC%9B%E6%97%85%E8%A1%8C%E5%8A%A9%E6%89%8B%E6%95%99%E4%BD%A0%E7%9C%9F%E5%AE%9E%E4%B8%96%E7%95%8C%E4%BB%BB%E5%8A%A1%E9%97%AD%E7%8E%AF%EF%BC%9B%E5%BE%B7%E5%B7%9E%E6%89%91%E5%85%8B%E6%95%99%E4%BD%A0%E5%A4%9A%20Agent%20%E7%8A%B6%E6%80%81%E6%B5%81%E4%B8%8E%E5%8D%8F%E4%BD%9C%E3%80%82%3C/text%3E%3C/svg%3E)

## 为什么还要单独做这套课

现在已经有不少 Agent 教程了，但大部分会落到两种极端：一种是偏概念，历史、范式、框架一大堆，看完还是不会做；另一种是偏平台操作，照着点完一个 demo，却不知道背后的共同骨架是什么。`hello-agents` 本身其实已经是很完整的一套系统性教程，从智能体基础、经典范式、低代码平台、主流框架、自研框架，一直延伸到记忆、上下文工程、协议、评估和综合案例 [$TRAE_REF](https://raw.githubusercontent.com/datawhalechina/hello-agents/main/README.md)。它很完整，但也正因为完整，对零基础读者来说，第一次进入时容易觉得内容面很大。[$TRAE_REF](https://raw.githubusercontent.com/datawhalechina/hello-agents/main/README.md)

这套 Cookbook 的选择是反过来：先不追求覆盖全部知识树，而是先把真正会反复碰到的那条最短路径讲透。也就是：

- 先理解 Agent 的最小心智模型
- 先学会在 Codex、Trae 这类宿主里工作
- 先做出一个真正可验收的 Agent
- 再进入两个更像真实项目的综合案例

## 这套课不追求什么

- 不追求一上来讲完所有框架
- 不追求把理论历史铺得很长
- 不追求做一堆横向小案例

这不是为了“偷内容”，而是因为零基础最容易死在前两周：概念太多、名词太多、案例太散，最后什么都摸过一点，但没有一件事真正做出来。所以我们只保留 3 个主案例，不做案例大杂烩。

## 这套课的默认读者

这套课默认你满足下面两条即可：

- 你已经会把 AI 当作工作工具，但还不会把它做成一个稳定的 Agent
- 你愿意在 Codex、Trae 这类宿主里，跟着一条主线把东西真正做出来

你不需要一开始就会某个 Agent 框架，也不需要先懂协议、评测、训练这些更后面的东西。那些都重要，但不是第一次入门最该优先解决的问题。

## 这套课的最小方法论

后面所有章节，其实都围绕同一个方法论反复展开：

1. 先把任务写清楚：目标、输入、约束、输出、验收  
2. 再把 Agent 的共同骨架看清楚：工具、状态、边界、产物  
3. 最后再把它放进真实案例里，看它如何变成一个真正能工作的系统

如果你能吃透这 3 步，后面换框架、换平台、换案例，迁移都不会太难。

## 课程最后会把你带到哪里

你最终会沿着这条路线往前走：

- 先理解 Agent 是什么
- 再做出一个单 Agent
- 再做一个真实世界旅行助手
- 最后做一个多 Agent 的德州扑克系统

这 3 个层级分别解决 3 种问题：

- `对话 Agent`：让你第一次真正做出“不是普通聊天”的最小 Agent
- `智能旅行助手`：让你理解工具、信息源、结构化交付怎么一起工作
- `德州扑克多 Agent 系统`：让你理解角色边界、状态流、共享上下文和协作/博弈

## 先怎么读最稳

如果你是第一次系统学 Agent 开发，最稳的读法仍然是：

- 先看这篇，知道整套课想把你带到哪里
- 再看“什么是 Agent”“Agent 的最小骨架”“如何使用 Agent 宿主”
- 然后进入“Prompt 不是契约”“工具调用”“第一个可验收 Agent”
- 最后再进入 3 个主案例

原因很简单：后面的案例虽然更有意思，但它们本质上都是前面主干能力的放大版。先把主干站稳，案例才不会变成“照着抄一遍，然后不知道为什么这样设计”。
