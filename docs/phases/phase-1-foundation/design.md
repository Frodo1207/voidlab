# Phase 1 设计文档

## 1. 设计目标

Phase 1 的目标不是一次性做完整业务系统，而是把当前后台从“演示骨架”建设成“最小可用的运营后台”。

本阶段设计文档用于回答以下问题：

1. 后台第一阶段要有哪些页面
2. 后端第一阶段要提供哪些 API
3. 核心数据表应该如何设计
4. 后台、后端、媒体上传之间如何配合
5. 开发时应按什么顺序推进

## 2. 设计边界

### 2.1 本阶段包含

- 登录模块
- 资讯管理模块
- 活动管理模块
- Builder 管理模块
- 媒体上传模块
- 支撑上述模块的基础 API 和数据库结构

### 2.2 本阶段不包含

- Leads 模块
- Contact / Event / Builder 外部意向提交
- 站点配置模块
- 知识库模块
- AI Agent Skill
- 复杂权限系统
- 数据统计

## 3. 设计原则

Phase 1 设计遵循以下原则：

### 3.1 先把主数据录入系统搭好

本阶段主要承接三类主数据：

- Articles
- Events
- Builders

### 3.2 先支持后台录入，再考虑前台消费

Phase 1 可以先完成后台录入和基础 API，不要求前台一次性全部切换到 API。

### 3.3 保持轻量，但结构清晰

当前技术路线下，允许实现保持轻量，但必须保留清晰分层，避免把后续二期能力堵死。

## 4. 信息架构

### 4.1 后台页面结构

建议第一阶段后台页面结构如下：

1. 登录页
2. 后台首页
3. 资讯列表页
4. 资讯新建 / 编辑页
5. 活动列表页
6. 活动新建 / 编辑页
7. Builder 列表页
8. Builder 新建 / 编辑页
9. 媒体库页

### 4.2 后台导航结构

建议后台左侧主导航为：

- 仪表盘
- 资讯管理
- 活动管理
- Builder 管理
- 媒体资源

### 4.3 页面优先级

Phase 1 页面优先级建议如下：

#### P0

- 登录页
- 资讯列表页
- 资讯编辑页
- 活动列表页
- 活动编辑页
- Builder 列表页
- Builder 编辑页

#### P1

- 媒体库页
- 后台首页

## 5. 页面设计

### 5.1 登录页

#### 目标

提供后台基础登录能力。

#### 字段

- 用户名或邮箱
- 密码

#### 行为

- 提交登录
- 登录成功后跳转后台首页
- 登录失败展示错误信息

### 5.2 资讯列表页

#### 目标

查看和管理文章列表。

#### 列表字段

- 标题
- 分类
- 受众
- 状态
- 是否精选
- 更新时间
- 操作

#### 列表操作

- 新建文章
- 编辑
- 删除
- 切换状态

### 5.3 资讯编辑页

#### 目标

完成文章录入和编辑。

#### 表单字段

- 标题
- slug
- 摘要
- 分类
- 标签
- 受众
- 封面图
- 正文
- 来源名称
- 来源链接
- 是否精选
- 状态

#### 操作

- 保存草稿
- 发布
- 删除

### 5.4 活动列表页

#### 目标

查看和管理活动列表。

#### 列表字段

- 标题
- 时间
- 城市
- 类型
- 状态
- 更新时间
- 操作

### 5.5 活动编辑页

#### 表单字段

- 标题
- slug
- 时间
- 地点
- 城市
- 类型
- 状态
- 摘要
- 封面图
- 活动详情

#### 操作

- 保存草稿
- 发布
- 删除

### 5.6 Builder 列表页

#### 目标

查看和管理 Builder 列表。

#### 列表字段

- 姓名
- 角色
- 城市
- 是否 featured
- 可联系状态
- 状态
- 更新时间
- 操作

### 5.7 Builder 编辑页

#### 表单字段

- 姓名
- slug
- 头衔
- 城市
- 角色
- 简介
- 故事
- 能力标签
- 关注方向
- 合作方式
- 可联系状态
- 是否 featured
- 封面图
- 状态

#### 操作

- 保存草稿
- 发布
- 删除

### 5.8 媒体库页

#### 目标

提供基础资源上传与复用能力。

#### 页面能力

- 上传图片
- 查看已上传文件
- 复制文件 URL
- 返回文件元信息

## 6. 前端管理端结构设计

建议 `apps/admin` 第一阶段按页面与模块拆分：

```text
apps/admin/src/
  main.ts
  App.vue
  router/
  layouts/
  views/
    login/
    dashboard/
    articles/
    events/
    builders/
    media/
  components/
  services/
  stores/
  types/
```

### 6.1 推荐页面路由

- `/login`
- `/`
- `/articles`
- `/articles/new`
- `/articles/:id/edit`
- `/events`
- `/events/new`
- `/events/:id/edit`
- `/builders`
- `/builders/new`
- `/builders/:id/edit`
- `/media`

### 6.2 推荐前端模块划分

#### `services/`

封装 API 请求：

- `authService`
- `articleService`
- `eventService`
- `builderService`
- `mediaService`

#### `stores/`

第一阶段建议至少有：

- `authStore`

其余列表数据可以先采用页面内请求，不必过早做全局 store。

## 7. API 设计

### 7.1 API 基础规范

统一前缀：

- `/api/v1`

统一响应格式建议：

```json
{
  "code": 0,
  "message": "ok",
  "data": {}
}
```

错误响应建议：

```json
{
  "code": 4001,
  "message": "invalid request",
  "data": null
}
```

### 7.2 Auth API

#### POST `/api/v1/auth/login`

请求体：

```json
{
  "username": "admin",
  "password": "******"
}
```

返回：

- 用户信息
- token 或 session 信息

#### GET `/api/v1/auth/me`

用于获取当前登录用户。

#### POST `/api/v1/auth/logout`

用于退出登录。

### 7.3 Article API

#### GET `/api/v1/articles`

功能：

- 获取文章列表

支持参数：

- `page`
- `page_size`
- `status`
- `keyword`

#### GET `/api/v1/articles/:id`

功能：

- 获取文章详情

#### POST `/api/v1/articles`

功能：

- 创建文章

#### PUT `/api/v1/articles/:id`

功能：

- 更新文章

#### DELETE `/api/v1/articles/:id`

功能：

- 删除文章

### 7.4 Event API

#### GET `/api/v1/events`

获取活动列表。

#### GET `/api/v1/events/:id`

获取活动详情。

#### POST `/api/v1/events`

创建活动。

#### PUT `/api/v1/events/:id`

更新活动。

#### DELETE `/api/v1/events/:id`

删除活动。

### 7.5 Builder API

#### GET `/api/v1/builders`

获取 Builder 列表。

#### GET `/api/v1/builders/:id`

获取 Builder 详情。

#### POST `/api/v1/builders`

创建 Builder。

#### PUT `/api/v1/builders/:id`

更新 Builder。

#### DELETE `/api/v1/builders/:id`

删除 Builder。

### 7.6 Media API

#### POST `/api/v1/media/upload`

功能：

- 上传文件

返回：

- 文件 ID
- 文件 URL
- 文件名
- 内容类型

#### GET `/api/v1/media`

获取媒体列表。

## 8. 数据模型设计

Phase 1 第一阶段建议先落 5 张核心表：

- `users`
- `articles`
- `events`
- `builders`
- `media_assets`

### 8.1 users

用途：

- 后台用户登录

建议字段：

- `id`
- `username`
- `password_hash`
- `role`
- `created_at`
- `updated_at`

### 8.2 articles

用途：

- 资讯内容管理

建议字段：

- `id`
- `title`
- `slug`
- `summary`
- `category`
- `audience`
- `tags_json`
- `cover_media_id`
- `content`
- `source_name`
- `source_url`
- `featured`
- `status`
- `published_at`
- `created_by`
- `updated_by`
- `created_at`
- `updated_at`

说明：

- 第一阶段标签可先用 `tags_json`
- 不急于单独拆 tag 表

### 8.3 events

用途：

- 活动管理

建议字段：

- `id`
- `title`
- `slug`
- `event_time`
- `location`
- `city`
- `event_type`
- `status`
- `summary`
- `cover_media_id`
- `content`
- `created_by`
- `updated_by`
- `created_at`
- `updated_at`

### 8.4 builders

用途：

- Builder 档案管理

建议字段：

- `id`
- `name`
- `slug`
- `title`
- `city`
- `role`
- `intro`
- `story`
- `expertise_json`
- `focus_areas_json`
- `collaboration_modes_json`
- `availability_note`
- `open_for`
- `contactable`
- `featured`
- `cover_media_id`
- `status`
- `created_by`
- `updated_by`
- `created_at`
- `updated_at`

### 8.5 media_assets

用途：

- 管理上传后的文件资源

建议字段：

- `id`
- `object_key`
- `object_url`
- `file_name`
- `content_type`
- `file_size`
- `uploaded_by`
- `created_at`

## 9. 状态设计

### 9.1 Article 状态

第一阶段建议支持：

- `draft`
- `published`

### 9.2 Event 状态

第一阶段建议支持：

- `draft`
- `published`

### 9.3 Builder 状态

第一阶段建议支持：

- `draft`
- `published`

说明：

更复杂的状态流留到 Phase 3。

## 10. 后端结构设计

建议 `apps/api` 第一阶段按业务域拆分：

```text
apps/api/internal/
  http/
    router.go
    auth_handler.go
    article_handler.go
    event_handler.go
    builder_handler.go
    media_handler.go
  service/
    auth_service.go
    article_service.go
    event_service.go
    builder_service.go
    media_service.go
  repository/
    user_repository.go
    article_repository.go
    event_repository.go
    builder_repository.go
    media_repository.go
  domain/
    user.go
    article.go
    event.go
    builder.go
    media.go
```

### 10.1 Handler 层职责

- 参数解析
- 响应处理
- 调用 service

### 10.2 Service 层职责

- 校验业务输入
- 处理 slug 规则
- 处理状态变更
- 组织 repository 调用

### 10.3 Repository 层职责

- SQLite 查询和写入
- 屏蔽 SQL 细节

## 11. 鉴权设计

Phase 1 推荐采用轻量方案。

### 11.1 推荐方案

- 后台登录成功后获取 token
- 后续请求带 token
- 后端中间件校验 token

### 11.2 当前角色

Phase 1 只需要支持：

- `admin`
- `editor`

### 11.3 权限控制范围

Phase 1 不要求做复杂 RBAC，只需要：

- 未登录不能访问后台受保护路由
- 已登录用户可访问 Phase 1 页面

## 12. 媒体上传设计

### 12.1 存储方式

文件存储到 MinIO。

数据库只存：

- 文件标识
- 文件 URL
- 文件名
- 类型
- 上传人

### 12.2 上传流程

1. 前端上传文件
2. 后端接收文件
3. 后端写入 MinIO
4. 后端写入 `media_assets`
5. 返回文件 URL 和文件 ID

### 12.3 文件使用方式

文章、活动、Builder 都通过 `cover_media_id` 关联媒体资源。

## 13. 前后台交互设计

### 13.1 新建文章

1. 进入文章编辑页
2. 填写内容
3. 上传封面
4. 保存
5. 返回列表

### 13.2 新建活动

1. 进入活动编辑页
2. 填写活动信息
3. 上传封面
4. 保存
5. 返回列表

### 13.3 新建 Builder

1. 进入 Builder 编辑页
2. 填写成员资料
3. 上传封面
4. 保存
5. 返回列表

## 14. 开发顺序建议

为了降低实现难度，建议严格按以下顺序开发：

### Step 1

完成后台路由和登录页面骨架。

### Step 2

完成后端基础鉴权和用户表。

### Step 3

完成媒体上传能力。

### Step 4

完成文章模块：

- 表
- API
- 后台列表页
- 后台编辑页

### Step 5

完成活动模块：

- 表
- API
- 后台列表页
- 后台编辑页

### Step 6

完成 Builder 模块：

- 表
- API
- 后台列表页
- 后台编辑页

### Step 7

补充后台首页和基础联调。

## 15. 风险与约束

### 15.1 当前前台仍为静态数据

这是已知状态，不是问题。Phase 1 不要求前台完全 API 化。

### 15.2 后台不要提前做复杂组件系统

第一阶段重点是把流程跑通，不要在 UI 抽象上投入过多时间。

### 15.3 数据结构先保守

像标签、能力项、合作方式这些多值字段，第一阶段可以先用 JSON 存储，减少表拆分成本。

## 16. 验收用例

### 16.1 登录用例

- 输入正确账号密码可登录
- 输入错误账号密码提示失败
- 未登录访问后台页面会被拦截

### 16.2 文章用例

- 可创建文章
- 可编辑文章
- 可删除文章
- 可切换草稿 / 发布状态

### 16.3 活动用例

- 可创建活动
- 可编辑活动
- 可删除活动
- 可设置时间、城市、类型和状态

### 16.4 Builder 用例

- 可创建 Builder
- 可编辑 Builder
- 可删除 Builder
- 可设置 featured 和 contactable

### 16.5 媒体用例

- 可上传图片
- 可返回文件 URL
- 可在文章、活动、Builder 中选择或引用

## 17. 本阶段完成标志

当以下条件同时满足时，可视为 Phase 1 完成：

1. 后台已具备可用登录能力
2. 后台已具备文章、活动、Builder 三个模块的基础 CRUD
3. 后台已具备媒体上传能力
4. 后端已具备对应 API 与核心表结构
5. 非开发人员已经可以通过后台完成主数据录入

## 18. 一句话总结

Phase 1 的设计核心，不是把所有业务都做完，而是先搭好一个：

**可以录内容、可以录活动、可以录 Builder、可以上传资源的最小可用后台。**
