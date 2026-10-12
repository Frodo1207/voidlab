---
title: "接入模型与跑主循环"
slug: "agent-cookbook-first-agent-model-and-loop"
section_name: "案例一：5 分钟实现第一个智能体"
public_summary: "把模型接进来，再把 Thought、Action、Observation 的循环真正跑起来。"
estimated_read_minutes: 9
status: "published"
---

## 接入模型

这里用一个兼容 OpenAI 风格接口的简单客户端：

```python
from openai import OpenAI

class OpenAICompatibleClient:
    def __init__(self, model: str, api_key: str, base_url: str):
        self.model = model
        self.client = OpenAI(api_key=api_key, base_url=base_url)

    def generate(self, prompt: str, system_prompt: str) -> str:
        messages = [
            {"role": "system", "content": system_prompt},
            {"role": "user", "content": prompt},
        ]
        response = self.client.chat.completions.create(
            model=self.model,
            messages=messages,
            stream=False,
        )
        return response.choices[0].message.content
```

这里不绑定具体服务商。只要提供 `API_KEY`、`BASE_URL` 和 `MODEL_ID`，不管接 OpenAI、Azure，还是本地兼容接口，整体用法都差不多。

## 跑主循环

下面这段代码，就是把前面讲过的 `Thought → Action → Observation` 循环真的跑起来：

```python
import os
import re

API_KEY = "YOUR_API_KEY"
BASE_URL = "YOUR_BASE_URL"
MODEL_ID = "YOUR_MODEL_ID"
os.environ["TAVILY_API_KEY"] = "YOUR_TAVILY_API_KEY"

llm = OpenAICompatibleClient(
    model=MODEL_ID,
    api_key=API_KEY,
    base_url=BASE_URL,
)

user_prompt = "你好，请帮我查询一下今天北京的天气，然后根据天气推荐一个合适的旅游景点。"
prompt_history = [f"用户请求: {user_prompt}"]

for i in range(5):
    full_prompt = "\n".join(prompt_history)
    llm_output = llm.generate(full_prompt, system_prompt=AGENT_SYSTEM_PROMPT)

    match = re.search(
        r"(Thought:.*?Action:.*?)(?=\n\s*(?:Thought:|Action:|Observation:)|\Z)",
        llm_output,
        re.DOTALL,
    )
    if match:
        llm_output = match.group(1).strip()

    prompt_history.append(llm_output)

    action_match = re.search(r"Action: (.*)", llm_output, re.DOTALL)
    if not action_match:
        prompt_history.append("Observation: 错误：未能解析到 Action 字段。")
        continue

    action_str = action_match.group(1).strip()

    if action_str.startswith("Finish"):
        final_answer = re.match(r"Finish\[(.*)\]", action_str).group(1)
        print(final_answer)
        break

    tool_name = re.search(r"(\w+)\(", action_str).group(1)
    args_str = re.search(r"\((.*)\)", action_str).group(1)
    kwargs = dict(re.findall(r'(\w+)="([^"]*)"', args_str))

    if tool_name in available_tools:
        observation = available_tools[tool_name](**kwargs)
    else:
        observation = f"错误：未定义的工具 '{tool_name}'"

    prompt_history.append(f"Observation: {observation}")
```

这段代码不长，但一个最小 Agent 该有的东西已经都串起来了：

- 模型负责思考
- 工具负责执行
- 观察结果会回到下一轮
- 信息够了就结束

## 继续阅读

下一篇：`practice/14-first-agent-result-and-analysis.md`
