# 文章运营

## 适用场景

当任务与这些动作有关时，读这一份：

- 新建资讯草稿
- 更新文章标题、摘要、正文
- 补封面
- 发布 / 撤回 / 归档

## 字段约定

文章常用字段：

- `title`
- `slug`
- `summary`
- `category`
- `audience`
- `tags`
- `cover_url`
- `content`
- `source_name`
- `source_url`
- `featured`
- `status`

## 标题规则

- 默认不要给标题加 `AI HOT｜`
- 标题直接使用事件本身
- 来源信息保留在：
  - `tags`
  - `source_name`
  - 正文 `来源` 段

## 正文规则

当前项目的资讯正文优先使用 `Markdown`。

推荐骨架：

```md
## 事件摘要

一句到两句讲清楚发生了什么。

## 关键信息

- 信息 1
- 信息 2
- 信息 3

## 为什么值得关注

1. 对开发者的影响
2. 对产品的影响
3. 对行业的影响

## 来源

- 来源 A
- 来源 B
```

## 更新规则

如果只是状态变化，优先走：

- `PUT /api/v1/articles/:id/status`

如果要改正文、标题、封面等字段：

1. 先 `GET /api/v1/articles/:id`
2. 合并变更
3. 再 `PUT /api/v1/articles/:id`

不要只发局部字段。
