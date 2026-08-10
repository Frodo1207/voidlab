# 活动运营

## 适用场景

当任务与这些动作有关时，读这一份：

- 新建活动草稿
- 修改活动时间、地点、类型、文案
- 补活动封面
- 发布 / 撤回 / 归档

## 字段约定

活动常用字段：

- `title`
- `slug`
- `summary`
- `city`
- `location`
- `event_type`
- `event_time`
- `cover_url`
- `content`
- `status`

## 默认策略

- 默认先建 `draft`
- 标题尽量自然，不写得像测试数据
- 城市、地点、类型、时间尽量彼此匹配
- 如果是虚构活动，也要写得像真实社区活动

## 虚构活动建议

为了让页面看起来像真实社区在运营，活动可以分成：

- `Builder Meetup`
- `Salon`
- `Workshop`
- `Open Day`
- `Hackathon Warmup`

活动文案里建议包含：

- 这场活动适合谁
- 会发生什么
- 为什么值得来

## 更新规则

如果只是状态变化，优先走：

- `PUT /api/v1/events/:id/status`

如果要改标题、时间、地点、封面等字段：

1. 先 `GET /api/v1/events/:id`
2. 合并变更
3. 再 `PUT /api/v1/events/:id`
