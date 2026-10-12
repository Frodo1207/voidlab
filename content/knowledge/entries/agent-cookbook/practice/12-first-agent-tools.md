---
title: "准备工具"
slug: "agent-cookbook-first-agent-tools"
section_name: "案例一：5 分钟实现第一个智能体"
public_summary: "先给这个最小 Agent 准备两只手：一只查天气，一只查景点。"
estimated_read_minutes: 8
status: "published"
---

这一节会准备两件工具：

- 一个查天气
- 一个查景点

## 工具一：查询真实天气

第一个工具负责查天气。这里直接用免费的 `wttr.in` 服务：

```python
import requests

def get_weather(city: str) -> str:
    url = f"https://wttr.in/{city}?format=j1"

    try:
        response = requests.get(url, timeout=10)
        response.raise_for_status()
        data = response.json()

        current_condition = data["current_condition"][0]
        weather_desc = current_condition["weatherDesc"][0]["value"]
        temp_c = current_condition["temp_C"]

        return f"{city}当前天气：{weather_desc}，气温{temp_c}摄氏度"
    except requests.exceptions.RequestException as e:
        return f"错误：查询天气时遇到网络问题 - {e}"
    except (KeyError, IndexError) as e:
        return f"错误：解析天气数据失败 - {e}"
```

这个工具做的事很直接：请求天气接口，取出当前天气和温度，再整理成一句清楚的话返回。

## 工具二：根据天气推荐景点

第二个工具用来查资料并给出推荐。这里用 `Tavily` 做搜索：

```python
import os
from tavily import TavilyClient

def get_attraction(city: str, weather: str) -> str:
    api_key = os.environ.get("TAVILY_API_KEY")
    if not api_key:
        return "错误：未配置 TAVILY_API_KEY 环境变量。"

    tavily = TavilyClient(api_key=api_key)
    query = f"{city} 在 {weather} 天气下适合去的旅游景点推荐"

    try:
        response = tavily.search(
            query=query,
            search_depth="basic",
            include_answer=True
        )

        if response.get("answer"):
            return response["answer"]

        results = []
        for result in response.get("results", []):
            results.append(f"- {result['title']}: {result['content']}")

        if not results:
            return "抱歉，没有找到相关景点推荐。"

        return "根据搜索结果，找到以下信息：\n" + "\n".join(results)
    except Exception as e:
        return f"错误：执行搜索时出现问题 - {e}"
```

到这里，Agent 已经有了两只“手”：

- 一只手查天气
- 一只手查景点

## 把工具收进一个字典

为了让主循环能统一调用工具，可以先把工具放到一个字典里：

```python
available_tools = {
    "get_weather": get_weather,
    "get_attraction": get_attraction,
}
```

## 继续阅读

下一篇：`practice/13-first-agent-model-and-loop.md`
