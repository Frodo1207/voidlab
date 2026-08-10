# Builder 运营

## 适用场景

当任务与这些动作有关时，读这一份：

- 新建 Builder 卡片
- 更新 Builder 介绍、标签、合作方式
- 发布 / 撤回 / 归档 Builder

## 字段约定

Builder 常用字段：

- `name`
- `slug`
- `title`
- `city`
- `role`
- `intro`
- `story`
- `expertise`
- `focus_areas`
- `collaboration_modes`
- `cover_url`
- `contactable`
- `featured`
- `status`

## 默认策略

- 默认先建 `draft`
- `intro` 用一句话讲清楚他是谁
- `story` 讲成长路径或擅长方向
- 数组字段更新时，默认是整组写回，不是局部 patch

## 更新规则

如果只是状态变化，优先走：

- `PUT /api/v1/builders/:id/status`

如果要改介绍、标签、合作方式等字段：

1. 先 `GET /api/v1/builders/:id`
2. 合并变更
3. 再 `PUT /api/v1/builders/:id`
