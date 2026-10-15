# 每日资讯 MVP

## 已实现

- 后台 `/daily-news`：采集、查看来源、编辑关联文章、审核、发布、复制日报、管理员发送。
- 数据保存在现有 SQLite；官网继续读取现有 `articles`，无需更换资讯系统。
- AI HOT 匿名公开 API 采集滚动 24 小时的最多 8 条精选，保留来源、canonical 和版权信息。
- 最多创建 5 篇未入库资讯的 Markdown 摘要草稿。不是全文抓取，也不自动编造深度分析。
- 日期唯一、原文去重、事务建稿；重复采集同一天返回已有结果，不覆盖人工修改。
- 草稿允许 0–5 篇，无新资讯也允许生成空日报。
- 状态：`pending -> approved -> published`，发送状态另计为 `pending / sending / failed / sent`。
- 发布时冻结群消息正文。已发送任务不会再次发送；失败手动重试复用发送 UUID。

## 日常操作

1. 在每日资讯页点击“采集资讯”，或等待定时采集。
2. 查看来源，通过关联入口编辑站内文章草稿。
3. 审核通过后点击“发布”，将关联草稿发布到官网并生成最终站内链接。
4. 复制链接列表到其他社群，或等待定时发飞书。管理员可手动发送。

## 定时配置

环境变量见 `deploy/.env.example`，Compose 已透传：

```dotenv
DAILY_NEWS_ENABLED=false
DAILY_NEWS_SITE_URL=https://voidlab.club
DAILY_NEWS_COLLECT_TIME=09:00
DAILY_NEWS_SEND_TIME=10:30
FEISHU_APP_ID=
FEISHU_APP_SECRET=
FEISHU_CHAT_ID=oc_79d4c3853b7a26bb4856e05de33a7642
```

当前已确认目标群：

- 群名：`VoidLab灵感中`
- `chat_id`：`oc_79d4c3853b7a26bb4856e05de33a7642`

时间为北京时间。默认关闭调度，启用后 API 进程每分钟检查：

- 到采集时间生成当天草稿，重启后当天可补采。
- 到发送时间，仅发送已人工审核并发布的当天任务；未审核不会发送。
- 首次自动发送失败不会持续轰炸群，后台查看错误后手动重试。
- 到时间后才发布的当天任务，可在下一次分钟检查时发送。
- 不自动补发历史日报。

飞书使用服务端应用机器人，独立于本地 TRAE 插件的授权。需要机器人有消息发送权限并已加入目标群。凭据仅通过服务器环境配置，不放进源码或日志；不要把密钥发送到聊天中。本轮不配置真实凭据、不调用真实群、不部署。

如果你现在还没有机器人凭据，先做这两步即可：

1. 保留上面的 `FEISHU_CHAT_ID`
2. 后续补上 `FEISHU_APP_ID` 和 `FEISHU_APP_SECRET`

这样链路就只差飞书机器人密钥，不用再重新查群。

## 接口

以下接口沿用现有登录及角色控制：

- `GET /api/v1/daily-news`：最近 30 天任务。
- `POST /api/v1/daily-news/collect`：`{"date":"YYYY-MM-DD"}`，仅允许北京时间今天。
- `POST /api/v1/daily-news/:date/approve`：审核通过。
- `POST /api/v1/daily-news/:date/publish`：发布草稿并冻结群消息。
- `POST /api/v1/daily-news/:date/send`：管理员发送或重试。

## 主动触发

你可以随时主动触发，不用等定时：

- 后台：进入“每日资讯”页面，点击“采集资讯”
- API：调用 `POST /api/v1/daily-news/collect`

示例：

```bash
curl -X POST http://localhost:18081/api/v1/daily-news/collect \
  -H "Authorization: Bearer <后台登录token>" \
  -H "Content-Type: application/json" \
  -d '{"date":"2026-10-11"}'
```

注意：

- 现在只允许采集“北京时间今天”的滚动 24 小时资讯
- 重复触发同一天会直接返回已有任务，不会覆盖你已经手工改过的草稿

## 首版边界

- 运行调度需要 API 服务持续在线；目前按单实例设计。
- 首版按 AI HOT 返回顺序取精选，不支持人工调整日报条目、全自动审核、封面生成或多群分发。
- 群发依赖真实飞书应用权限和成员资格，尚需配置后的实际联调。
- 飞书 UUID 的去重有效期受平台限制；超过有效期的失败重试，先核对群消息，不能视为永久恰好一次保证。
- 定时执行错误写服务日志，人工操作记录审计；暂未增加独立告警通知渠道。
