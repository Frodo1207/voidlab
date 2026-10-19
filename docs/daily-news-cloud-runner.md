# VOIDLAB 每日资讯云端 Runner

这个 runner 用于在云端常驻执行每日资讯任务，不依赖本地电脑开机。

## 运行方式

- `api` 容器继续提供正式接口、文章发布、媒体上传和飞书发送能力。
- `dailynews-runner` 容器使用 `VOIDLAB_AGENT_TOKEN` 调用正式 API，按北京时间每天执行一次：
  - `collect`
  - `top-up`
  - `approve`
  - `publish`
  - `generate covers`
  - `upload assets`
  - `bind covers`
  - `build digest`
  - `send Feishu`
- 调用协议与项目中的 `voidlab-daily-news-ops` 工作流保持一致，只是把它固化成云端常驻 runner。

## 环境变量

- `VOIDLAB_API_BASE_URL`
  - runner 访问 API 的地址，容器内默认用 `http://api:8080`
- `VOIDLAB_AGENT_TOKEN`
  - 必填，建议至少包含：
    - `daily_news:read`
    - `daily_news:write`
    - `daily_news:send`
    - `articles:read`
    - `articles:write`
    - `media:read`
    - `media:write`
- `DAILY_NEWS_RUNNER_ENABLED`
  - `true` 时启用 runner
- `DAILY_NEWS_RUNNER_TIME`
  - 北京时间，格式 `HH:MM`，默认 `10:30`
- `DAILY_NEWS_TARGET_COUNT`
  - 默认 `5`
- `DAILY_NEWS_COVER_COMMAND`
  - 可选。若配置，runner 会执行该命令生成封面，并读取 `VOIDLAB_DAILY_NEWS_OUTPUT` 指向的 JPG 文件再上传。

## 封面策略

- 优先执行 `DAILY_NEWS_COVER_COMMAND` 生成无文字 AI 视觉图。
- 若未配置生成命令，runner 会从 `base-gallery` 中挑选已有图片兜底。
- 无论封面来自生图还是图库，最终都通过正式文章接口 `PUT /api/v1/articles/:id` 回写 `cover_url`，不会直接改数据库。
- 站内与飞书分发均只使用 VOIDLAB `insights` 链接，不输出或分发 `AI HOT` 字样。

## 幂等与续跑

- 如果当天任务已经存在，runner 会读取任务状态并继续未完成步骤。
- 如果当天已经 `sent`，runner 只记录 `already-sent`，不会重复制造第二条群消息。
- 运行摘要会拆成 `generate_covers`、`upload_assets`、`bind_covers`、`build_digest`、`send_feishu` 等细阶段；失败或降级原因会分别写入 `failures` 或 `warnings`。
- 只有当站内文章、封面绑定、摘要构建和飞书发送都达到完成态时，runner 才会把当天任务视为真正完成。

## 部署建议

- 保持 API 内建旧定时器 `DAILY_NEWS_ENABLED=false`，避免与 runner 重复调度。
- 生产上启用 `dailynews-runner` 容器，并为其配置有效的 `VOIDLAB_AGENT_TOKEN`。
- 如需 AI 生图，建议把封面生成脚本打进镜像，或在部署环境中提供可执行命令。
