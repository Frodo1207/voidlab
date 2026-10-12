---
title: "运行结果与分析"
slug: "agent-cookbook-first-agent-result-and-analysis"
section_name: "案例一：5 分钟实现第一个智能体"
public_summary: "看这个最小 Agent 跑起来之后会发生什么，再把它真正学到的东西拆开看清楚。"
estimated_read_minutes: 7
status: "published"
---

## 看运行结果

一个正常的执行过程，大概会像这样：

```text
Thought: 先获取北京今天的天气，再根据天气推荐景点。
Action: get_weather(city="北京")

Observation: 北京当前天气：Sunny，气温26摄氏度

Thought: 现在已经知道天气，可以推荐合适的景点。
Action: get_attraction(city="北京", weather="Sunny")

Observation: 北京在晴天最值得去的旅游景点是颐和园，因其湖景和古建筑非常适合晴天游览。另一个推荐是长城。

Thought: 信息已经足够，可以整理成最终回复。
Action: Finish[今天北京天气晴朗，气温26摄氏度，适合外出游玩。推荐去颐和园或者长城。]
```

这段输出最值得看的，不是最后那句答案，而是中间真的走了两步。它没有跳过查天气，也没有直接硬编景点，而是按顺序把任务拆开了。

## 看懂这个例子

这个例子虽然小，但一个最小 Agent 该有的几层东西都已经跑出来了：

- 它能拆任务，不是一次性瞎答
- 它能调用工具，不只是生成文本
- 它能记住上一轮结果，再继续下一轮
- 它能在信息足够时结束任务，给出最终答案

到这里，第一个智能体的核心体验就已经有了。后面再往上加页面、后端、数据库，本质上都是在这个闭环之上继续搭系统。

## 继续阅读

下一组案例：`案例二：旅游路线规划 Agent`
