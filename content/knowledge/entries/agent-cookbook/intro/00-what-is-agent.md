---
title: "什么是 Agent，不是什么"
slug: "agent-cookbook-what-is-agent"
section_name: "入门"
public_summary: "先搞清楚 Agent 的最小定义：目标、输入、工具、过程控制与可验收输出。"
estimated_read_minutes: 8
status: "published"
---

## 一句话结论

Agent 不是更会聊天的 AI，而是一种围绕目标组织动作、调用工具并交付结果的工作单元。你可以把它理解成一个“会工作的小系统”，而不只是一个“会回复你的对话框”。

![Agent 最小闭环示意图](data:image/svg+xml;utf8,%3Csvg%20xmlns%3D%27http%3A//www.w3.org/2000/svg%27%20width%3D%27960%27%20height%3D%27480%27%20viewBox%3D%270%200%20960%20480%27%3E%3Crect%20width%3D%27960%27%20height%3D%27480%27%20fill%3D%27%23f7f7f5%27/%3E%3Crect%20x%3D%2730%27%20y%3D%2730%27%20width%3D%27900%27%20height%3D%27420%27%20rx%3D%2724%27%20fill%3D%27%23ffffff%27%20stroke%3D%27%23e6e6e1%27/%3E%3Ctext%20x%3D%2760%27%20y%3D%2774%27%20font-family%3D%27Arial%2C%20PingFang%20SC%2C%20Microsoft%20YaHei%27%20font-size%3D%2730%27%20font-weight%3D%27800%27%20fill%3D%27%23111111%27%3EAgent%20%E6%9C%80%E5%B0%8F%E9%97%AD%E7%8E%AF%3C/text%3E%3Ctext%20x%3D%2760%27%20y%3D%27106%27%20font-family%3D%27Arial%2C%20PingFang%20SC%2C%20Microsoft%20YaHei%27%20font-size%3D%2715%27%20fill%3D%27%235f5e58%27%3E%E7%9B%AE%E6%A0%87%E3%80%81%E8%BE%93%E5%85%A5%E3%80%81%E5%B7%A5%E5%85%B7%E3%80%81%E8%BF%87%E7%A8%8B%E6%8E%A7%E5%88%B6%E4%B8%8E%E4%BA%A7%E7%89%A9%E5%BD%A2%E6%88%90%E4%B8%80%E4%B8%AA%E7%9C%9F%E6%AD%A3%E7%9A%84%E5%B7%A5%E4%BD%9C%E9%97%AD%E7%8E%AF%E3%80%82%3C/text%3E%3Crect%20x%3D%2760%27%20y%3D%27146%27%20width%3D%27156%27%20height%3D%2796%27%20rx%3D%2718%27%20fill%3D%27%23fbfbfa%27%20stroke%3D%27%23deded8%27/%3E%3Crect%20x%3D%27216%27%20y%3D%27168%27%20width%3D%2798%27%20height%3D%2752%27%20rx%3D%2726%27%20fill%3D%27%23f1f5f9%27%20stroke%3D%27%23cbd5e1%27/%3E%3Crect%20x%3D%27330%27%20y%3D%27146%27%20width%3D%27156%27%20height%3D%2796%27%20rx%3D%2718%27%20fill%3D%27%23fbfbfa%27%20stroke%3D%27%23deded8%27/%3E%3Crect%20x%3D%27486%27%20y%3D%27168%27%20width%3D%2798%27%20height%3D%2752%27%20rx%3D%2726%27%20fill%3D%27%23f1f5f9%27%20stroke%3D%27%23cbd5e1%27/%3E%3Crect%20x%3D%27600%27%20y%3D%27146%27%20width%3D%27156%27%20height%3D%2796%27%20rx%3D%2718%27%20fill%3D%27%23fbfbfa%27%20stroke%3D%27%23deded8%27/%3E%3Crect%20x%3D%27756%27%20y%3D%27168%27%20width%3D%2798%27%20height%3D%2752%27%20rx%3D%2726%27%20fill%3D%27%23f1f5f9%27%20stroke%3D%27%23cbd5e1%27/%3E%3Ctext%20x%3D%2792%27%20y%3D%27184%27%20font-family%3D%27Arial%2C%20PingFang%20SC%2C%20Microsoft%20YaHei%27%20font-size%3D%2721%27%20font-weight%3D%27750%27%20fill%3D%27%23111111%27%3E%E7%9B%AE%E6%A0%87%20%2B%20%E8%BE%93%E5%85%A5%3C/text%3E%3Ctext%20x%3D%2792%27%20y%3D%27214%27%20font-family%3D%27Arial%2C%20PingFang%20SC%2C%20Microsoft%20YaHei%27%20font-size%3D%2714%27%20fill%3D%27%235f5e58%27%3E%E8%A6%81%E5%AE%83%E5%B9%B2%E4%BB%80%E4%B9%88%EF%BC%8C%E5%85%81%E8%AE%B8%E5%AE%83%E7%9C%8B%E4%BB%80%E4%B9%88%E3%80%82%3C/text%3E%3Ctext%20x%3D%27360%27%20y%3D%27184%27%20font-family%3D%27Arial%2C%20PingFang%20SC%2C%20Microsoft%20YaHei%27%20font-size%3D%2721%27%20font-weight%3D%27750%27%20fill%3D%27%23111111%27%3E%E6%8E%A8%E7%90%86%20%2B%20%E5%B7%A5%E5%85%B7%3C/text%3E%3Ctext%20x%3D%27360%27%20y%3D%27214%27%20font-family%3D%27Arial%2C%20PingFang%20SC%2C%20Microsoft%20YaHei%27%20font-size%3D%2714%27%20fill%3D%27%235f5e58%27%3E%E6%83%B3%E6%80%8E%E4%B9%88%E5%81%9A%EF%BC%8C%E5%8A%A8%E4%BB%80%E4%B9%88%E5%B7%A5%E5%85%B7%E3%80%82%3C/text%3E%3Ctext%20x%3D%27634%27%20y%3D%27184%27%20font-family%3D%27Arial%2C%20PingFang%20SC%2C%20Microsoft%20YaHei%27%20font-size%3D%2721%27%20font-weight%3D%27750%27%20fill%3D%27%23111111%27%3E%E4%BA%A7%E7%89%A9%20%2B%20%E9%AA%8C%E6%94%B6%3C/text%3E%3Ctext%20x%3D%27634%27%20y%3D%27214%27%20font-family%3D%27Arial%2C%20PingFang%20SC%2C%20Microsoft%20YaHei%27%20font-size%3D%2714%27%20fill%3D%27%235f5e58%27%3E%E8%BE%93%E5%87%BA%E4%BB%80%E4%B9%88%EF%BC%8C%E6%80%8E%E4%B9%88%E7%AE%97%E5%81%9A%E5%AF%B9%E4%BA%86%E3%80%82%3C/text%3E%3Ctext%20x%3D%27245%27%20y%3D%27200%27%20font-family%3D%27Arial%2C%20PingFang%20SC%2C%20Microsoft%20YaHei%27%20font-size%3D%2716%27%20font-weight%3D%27700%27%20fill%3D%27%23334155%27%3E%E6%80%8E%E4%B9%88%E5%81%9A%EF%BC%9F%3C/text%3E%3Ctext%20x%3D%27517%27%20y%3D%27200%27%20font-family%3D%27Arial%2C%20PingFang%20SC%2C%20Microsoft%20YaHei%27%20font-size%3D%2716%27%20font-weight%3D%27700%27%20fill%3D%27%23334155%27%3E%E5%81%9A%E5%87%BA%E4%BA%86%E4%BB%80%E4%B9%88%EF%BC%9F%3C/text%3E%3Ctext%20x%3D%27782%27%20y%3D%27200%27%20font-family%3D%27Arial%2C%20PingFang%20SC%2C%20Microsoft%20YaHei%27%20font-size%3D%2716%27%20font-weight%3D%27700%27%20fill%3D%27%23334155%27%3E%E8%BF%99%E6%AC%A1%E8%BF%90%E8%A1%8C%E5%A5%BD%E4%B8%8D%E5%A5%BD%EF%BC%9F%3C/text%3E%3Crect%20x%3D%2760%27%20y%3D%27300%27%20width%3D%27796%27%20height%3D%27110%27%20rx%3D%2718%27%20fill%3D%27%23111111%27/%3E%3Ctext%20x%3D%2792%27%20y%3D%27342%27%20font-family%3D%27Arial%2C%20PingFang%20SC%2C%20Microsoft%20YaHei%27%20font-size%3D%2724%27%20font-weight%3D%27800%27%20fill%3D%27%23ffffff%27%3E%E5%8F%AA%E8%A6%81%E8%BF%99%E4%B8%AA%E9%97%AD%E7%8E%AF%E6%88%90%E7%AB%8B%EF%BC%8C%E6%89%8D%E5%BC%80%E5%A7%8B%E7%AE%97%E2%80%9CAgent%20%E5%BC%80%E5%8F%91%E2%80%9D%3C/text%3E%3Ctext%20x%3D%2792%27%20y%3D%27374%27%20font-family%3D%27Arial%2C%20PingFang%20SC%2C%20Microsoft%20YaHei%27%20font-size%3D%2715%27%20fill%3D%27%23d6d6d0%27%3E%E5%A6%82%E6%9E%9C%E5%8F%AA%E6%9C%89%E2%80%9C%E9%97%AE%E4%B8%80%E5%8F%A5%20%E2%86%92%20%E7%AD%94%E4%B8%80%E5%8F%A5%E2%80%9D%EF%BC%8C%E9%82%A3%E6%9B%B4%E5%83%8F%E6%98%AF%E8%81%8A%E5%A4%A9%EF%BC%8C%E8%80%8C%E4%B8%8D%E6%98%AF%E8%AE%BE%E8%AE%A1%E4%B8%80%E4%B8%AA%E5%B7%A5%E4%BD%9C%E7%B3%BB%E7%BB%9F%E3%80%82%3C/text%3E%3C/svg%3E)

## 为什么这一章特别重要

零基础最容易犯的第一个错，就是把 `Agent` 当成一个“更高级的聊天模式”。于是学习路径会变成：

- 学一点提示词
- 学一点工具调用
- 跑几个 demo
- 结果还是说不清楚自己到底做的是“一个 Agent”，还是“一个会多说几句的机器人”

`hello-agents` 把前面几章专门拿来讲“初识智能体”“智能体发展史”“大语言模型基础”，本质上也是先帮读者建立一个稳定的判断框架，而不是直接把人推进某个框架 API 里 [$TRAE_REF](https://raw.githubusercontent.com/datawhalechina/hello-agents/main/README.md)。我们这套课不想铺很长的历史，但这一章仍然必须完成同样的任务：先把“Agent 到底是什么”讲清楚。

## 这一章要讲什么

- Agent 和聊天机器人的区别
- Agent 和 workflow 的区别
- Agent 和普通脚本的区别
- 为什么“能交付结果”比“回答得像人”更重要

## 先记住一个最小定义

如果要用一句最短的话来定义 Agent，我会这样说：

> Agent 是一个围绕目标运行的工作单元。它会接收输入、根据上下文做决策、在边界内调用工具，并交付一个可以检查的结果。

这个定义里最关键的不是“智能”，而是“工作单元”这四个字。因为它意味着：

- 它有明确目标，不是漫无边际地对话
- 它有输入边界，不是什么都看
- 它有动作能力，不只是说建议
- 它有输出和验收，不是“看起来差不多就行”

## Agent 不是什么

### 它不等于聊天机器人

聊天机器人最核心的体验是“你问一句，它回一句”。重点是对话的连贯感、语气、知识面，甚至情绪价值。但 Agent 的重点不在“回得像不像人”，而在“它有没有把事情往结果推进”。

一个聊天机器人可以很聪明、很自然、很有风格，但如果它：

- 不知道目标是什么
- 不知道允许读什么
- 不会调用工具
- 不会生成可验收产物

那它更像一个高级对话界面，而不是一个 Agent。

### 它不等于 workflow

workflow 更像一条提前写死的流程：第一步做什么，第二步做什么，第三步做什么。它的优点是稳定，缺点是遇到变化时不够灵活。

Agent 不同的地方在于，它通常会在一定边界内自己判断：

- 先读哪份材料
- 先做哪一步
- 遇到缺失信息时要停下来，还是先给草稿
- 这次该调用哪个工具

所以你可以把 workflow 理解成“已经写死的路线”，把 Agent 理解成“带着地图和规则去完成任务的人”。

### 它不等于普通脚本

普通脚本最强的地方是确定性：输入固定、逻辑固定、输出固定。Agent 不是来替代脚本的，而是补脚本最不擅长的那一段：

- 面对模糊任务
- 需要理解自然语言输入
- 需要在多个动作之间做选择
- 需要先产出草稿再逐步修正

所以更准确的关系是：脚本像固定机械臂，Agent 像带判断能力的执行者。很多时候，一个好 Agent 反而会去调用脚本，而不是把所有事情都“自己想一遍”。

## 那什么情况下才算在做 Agent

最简单的判断标准，不是看你用了什么框架，也不是看你调了多少模型参数，而是看下面这个闭环有没有成立：

- 你是否给了它明确目标
- 你是否定义了输入边界
- 它是否真的调用了某种外部动作或工具
- 它是否交付了某种可以检查的结果
- 你是否有办法判断它这次做得好不好

只要这五件事成立，你基本就已经站在 Agent 开发的门槛里了。反过来，如果这里只有“问一句，回一句”，那更像是提示词工程，还没真正进入 Agent 开发。

## 为什么“能交付结果”比“回答得像人”更重要

这套课后面会反复讲一句话：**Agent 是拿来工作的，不是拿来表演聊天感的。**

这句话不是说对话体验不重要，而是说它在开发顺序上没那么重要。零基础最容易被“像不像人”“语气自然不自然”吸走注意力，但真实工作里更重要的其实是：

- 产物有没有落盘
- 输出结构稳不稳定
- 来源能不能追溯
- 出错时能不能回退

也就是说，一个 Agent 即使说话不够拟人，只要它能稳定地产出正确结果，它就是有价值的。反过来，一个“非常像人”的 Agent，如果每次都不给你真正可用的产物，那它在开发意义上并不成立。

## 这一章读完后，你至少要会判断两件事

第一，你能不能把一个东西区分成：

- 聊天机器人
- workflow
- 普通脚本
- Agent

第二，你能不能拿一个真实任务做快速判断：  
“这个任务如果要做成 Agent，最小闭环应该是什么？”

只要这两个判断建立起来，后面“最小骨架”“宿主工具”“Prompt 契约”“工具调用”这些章节才不会变成零散技巧，而会连成一条真正的开发路径。

## 继续阅读

如果这一章解决的是“Agent 到底是什么”，那下一章要解决的就是更具体的问题：一个 Agent 真正开始开发时，到底应该被拆成哪几个稳定部件。继续读 `01-agent-mental-model.md`，你会看到这套课后面反复复用的 6 个槽位。

## 交付物

- 一份“Agent 最小定义”笔记
- 一张最小闭环草图

## 验收标准

- 你能用自己的话说清楚 Agent 的最小闭环
- 你能解释为什么聊天机器人不一定算 Agent
- 你能举出一个真实任务，并判断它现在还只是聊天、workflow，还是已经具备 Agent 雏形
