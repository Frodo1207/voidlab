# 发布流

这份文件描述“一个运营任务从开始到结束”应该怎么走。

## 标准顺序

### 1. 握手

至少做：

1. `GET /healthz`
2. `GET /api/v1/auth/me`
3. 对目标资源做一次只读检查

### 2. 创建或读取对象

- 新对象：先建草稿
- 老对象：先读取当前详情

### 3. 补内容

根据资源类型补：

- 标题
- 摘要
- 正文
- 时间 / 地点
- 标签
- 封面

### 4. 状态变更

如果用户明确说要发，就走状态接口发布。

### 5. 汇报结果

只输出最必要的信息：

- 做了什么
- `id`
- `slug`
- `status`
- 关键变更

## 推荐输出

例如：

- `Created event draft`
- `id: 7`
- `slug: ai-builder-night-shanghai`
- `status: draft`
- `updated fields: title, summary, cover_url`
