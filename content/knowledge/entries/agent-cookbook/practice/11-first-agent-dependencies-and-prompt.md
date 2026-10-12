---
title: "依赖与系统提示词"
slug: "agent-cookbook-first-agent-dependencies-and-prompt"
section_name: "案例一：5 分钟实现第一个智能体"
public_summary: "先把这个最小案例要解决什么任务说清楚，再装依赖、写系统提示词。"
estimated_read_minutes: 9
status: "published"
---

## 先看任务

先别急着做复杂系统，先跑一个最小例子。这个案例要做的是一个能处理分步任务的智能旅行助手。用户会说一句话：

```text
你好，请帮我查询一下今天北京的天气，然后根据天气推荐一个合适的旅游景点。
```

这个任务不能一上来就直接编一个答案，而是要先查天气，再根据天气结果去推荐景点。也就是说，它至少要完成两轮动作：

- 先调用天气工具
- 拿到天气结果之后，再调用景点推荐工具

这就已经是一个最小但完整的 Agent 闭环了。

## 先装依赖

先装 3 个 Python 包：

- `requests`：负责发 HTTP 请求
- `tavily-python`：负责做实时搜索
- `openai`：负责调用兼容 OpenAI 接口的大语言模型服务

安装命令：

```bash
pip install requests tavily-python openai
```

## 写系统提示词

驱动这个 Agent 的关键，不只是模型本身，还有给模型的系统说明。这里需要明确告诉它：

- 它是谁
- 它有哪些工具
- 每次回复要长什么样
- 什么时候该结束任务

一个最小版本可以这样写：

```python
AGENT_SYSTEM_PROMPT = """
你是一个智能旅行助手。你的任务是分析用户的请求，并使用可用工具一步步地解决问题。

# 可用工具:
- `get_weather(city: str)`: 查询指定城市的实时天气。
- `get_attraction(city: str, weather: str)`: 根据城市和天气搜索推荐的旅游景点。

# 输出格式要求:
你的每次回复必须严格遵循以下格式，包含一对Thought和Action：

Thought: [你的思考过程和下一步计划]
Action: [你要执行的具体行动]

Action的格式必须是以下之一：
1. 调用工具：function_name(arg_name="arg_value")
2. 结束任务：Finish[最终答案]

# 重要提示:
- 每次只输出一对Thought-Action
- Action必须在同一行，不要换行
- 当收集到足够信息可以回答用户问题时，必须使用 Action: Finish[最终答案] 格式结束
"""
```

这段提示词不用写得多花，关键是规则要清楚。要让模型知道：每次只能做一步，动作必须结构化，信息够了就结束。

## 继续阅读

下一篇：`practice/12-first-agent-tools.md`
