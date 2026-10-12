---
title: "MCP 工具接入"
slug: "agent-cookbook-travel-planning-mcp-integration"
section_name: "案例二：旅游路线规划 Agent"
public_summary: "地图、图片、路线这些能力不要各接各的，放到统一工具层里，系统会顺很多。"
estimated_read_minutes: 8
status: "published"
---

## 为什么这里很适合讲 MCP

旅行助手这种项目，天然会用到一堆外部能力：地图、地理编码、路线、图片、天气、酒店资料。只要这些能力开始变多，就会遇到前面讲过的问题：每个工具的参数、返回格式、调用方式都不一样。

所以这个案例很适合用来理解 MCP 的位置。它不是来替代 Agent 的，也不是来替代工具本身的，而是把这些工具统一挂进系统里。

## 为什么不直接一个个调 API

直接调 API 当然可以，但项目一大，很快就会出现几个问题：

![MCP 在旅行助手里的位置](../../assets/agent-cookbook/travel-mcp-layer.svg)

- 每接一个服务都要自己写一套适配
- 参数名和结果格式很难统一
- 多个 Agent 想共用同一类工具时会重复接入
- 工具一多，排错会越来越乱

`voidagent` 这一章专门用了高德地图 MCP 和图片相关 API 集成来说明这件事。

## 这一页真正要做的决定

对这套课来说，这一页最重要的不是把所有工具都接完，而是先把工具层的原则放对：

- 哪些能力应该放进 MCP 层
- 哪些能力只是普通后端服务
- 哪些工具会被多个 Agent 共用
- 哪些输出要在工具层先做格式收口

比如地图查询、地点搜索、路线规划这类能力，就非常适合统一放在 MCP 工具层。这样无论是景点 Agent 还是行程规划 Agent，拿到的都是同一套入口。

## 继续阅读

下一篇：`practice/25-travel-planning-frontend-details.md`
