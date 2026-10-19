# BuilderNetwork PRD

## 1. 文档目的

本文档用于定义 VOIDLAB `BuilderNetwork` 模块的产品定位、业务对象、核心流程、信息架构、接口边界与一期实施范围，作为后续设计与开发的直接依据。

本模块不是对现有 `builders` 页面的简单改版，而是将其升级为一个面向 AI 时代协作的网络产品。

本文档重点回答：

1. BuilderNetwork 到底是什么产品
2. 它解决哪些真实业务问题
3. 核心业务对象和模块如何拆分
4. 为什么它需要 Agent 与 AI Native 能力
5. 第一阶段应该先做什么，不做什么

> 当前产品交互与 MVP 决策以 [BuilderNetwork 产品与交互设计](./buildernetwork-product-design.md) 为准。本 PRD 早期提出的会话系统、Agent 前台体验、自动推荐等属于候选能力，不代表一期范围；一期以邀请注册、作品/机会发布、申请审核和授权交换联系方式为核心闭环。

## 2. 产品定位

### 2.1 模块定义

BuilderNetwork 是一个以作品证明能力、以合作机会促成连接的协作市场。平台通过开放 API 和 Skill 支持用户自己的 Agent，但 Agent 不是一期的前台主交互。

它不是：

- 纯人物黄页
- 传统招聘网站
- 仅供展示的作品集页面
- 只靠手工筛选的成员目录

它是：

- 一个让用户发布作品、沉淀协作身份的网络
- 一个让用户发布当前状态与合作需求的机会市场
- 一个让用户主动寻找合作者、项目方、服务方的发现系统
- 一个先申请、经审核、再按双方授权交换联系方式的连接机制
- 一个通过开放 API 和 Skill 支持外部 Agent 操作的底层能力

产品中需明确区分：

- `Work`：用户展示的项目成果和能力证明。
- `Opportunity`：用户正在招募、寻找或提供的合作机会。

### 2.2 目标价值

对用户：

- 能把“我是谁”升级成“我做过什么、现在要做什么”
- 能展示作品及自己在作品中的真实贡献
- 能发布自己的合作状态，让别人主动来找
- 能主动搜索和连接潜在合作对象

对 VOIDLAB：

- 将社区成员、活动参与者、合作对象沉淀为结构化网络资产
- 从单点内容/活动运营，升级为长期关系与协作网络运营
- 通过开放 API 和 Skill 扩展社区运营与用户 Agent 的操作能力

## 3. 背景与问题

### 3.1 当前现状

当前仓库中的 `builders` 已具备如下基础：

- 前台人物列表页与详情页
- 后台 Builder 资料录入与发布
- 通过 Builder 详情页提交合作意向
- Leads 系统承接站外合作请求

但当前版本仍主要是静态人物资料卡，缺失以下关键能力：

- 用户无法上传和管理作品
- 一个作品无法关联多个协作者
- 用户无法发布“我现在在找什么”
- 用户无法主动高效查找人、作品、机会
- 用户之间没有结构化连接与消息流
- Agent 无法围绕网络行为执行高层动作

### 3.2 要解决的核心问题

BuilderNetwork 需要解决四类问题：

1. 身份问题：
   用户如何让别人快速理解“我是谁，我能做什么”。

2. 证据问题：
   用户如何通过作品证明自己的能力，而不是只靠自我介绍。

3. 机会问题：
   用户如何公开表达当前状态、需求、招募方向和合作意图。

4. 连接问题：
   用户如何主动找人、被别人找到，并通过申请审核建立可信连接。

## 4. 目标用户

一期建议覆盖三类核心用户，外部 Agent 作为 API / Skill 调用方支持：

### 4.1 Builder / 创作者 / 开发者

特征：

- 有作品、有技能、有项目经历
- 希望获得合作、兼职、共创、顾问、工作机会

核心动作：

- 建立个人档案
- 上传作品
- 发布当前合作状态
- 主动查找合作者或项目

### 4.2 项目发起者 / 需求方

特征：

- 需要找 AI 开发、设计、产品、运营、内容、行业专家

核心动作：

- 发布合作机会
- 搜索合适的人或作品
- 发起联系请求

### 4.3 社区运营 / VOIDLAB 运营人员

特征：

- 负责筛选、推荐、组织、撮合、维护秩序

核心动作：

- 审核资料与作品
- 维护精选对象
- 运营机会广场
- 监控消息与异常行为

### 4.4 用户自有 Agent / Skill 调用方

特征：

- 用户在平台之外使用的 Agent，通过用户授权的 API Token 调用平台能力
- 不作为 BuilderNetwork 一期独立前台角色或聊天入口

核心动作：

- 读取和维护用户授权范围内的 Profile、Work、Opportunity
- 搜索公开 Builder、作品和合作机会
- 查询申请状态；需要用户明确授权才能接受或婉拒申请

## 5. 产品原则

BuilderNetwork 建议遵循以下原则：

1. 作品优先于自述
2. 机会优先于空泛社交
3. 关系建立先申请、再审核、后交换联系方式
4. 联系方式默认私密，只有在双方授权且申请被接受后才展示
5. Agent 通过开放 API 和 Skill 接入，关键动作必须可审计、可控、可撤销

## 6. 核心业务对象

一期核心对象以 Profile、Work、WorkMember、Opportunity、ContactRequest 和 ContactExchange 为主。

产品与交互设计中还补充了 `Account`、`InviteCode`、`PrivateContact`、`ContactExchange` 和 `WorkFeedback` 等对象，详见 [BuilderNetwork 产品与交互设计](./buildernetwork-product-design.md)。

### 6.1 Profile

用于表达“人是谁”。

建议字段：

- 基础身份：`id` `slug` `display_name` `headline`
- 所在信息：`city` `timezone`
- 角色信息：`primary_role` `secondary_roles`
- 介绍信息：`bio_short` `bio_long`
- 技能信息：`skills` `focus_areas`
- 合作信息：`service_modes` `contact_policy`
- 可见性信息：`status` `visibility`
- 运营信息：`featured` `verified`

### 6.2 Work

用于表达“做过什么”。

建议字段：

- 基础信息：`id` `slug` `title` `summary`
- 内容信息：`body_markdown`
- 作品类型：`work_type`
- 外部链接：`demo_url` `repo_url` `video_url`
- 媒体资源：`cover_url` `gallery`
- 标签信息：`tags`
- 状态信息：`status`
- 归属信息：`created_by_profile_id`

### 6.3 WorkMember

用于表达“一个作品关联哪些人、各自贡献什么”。

建议字段：

- `work_id`
- `profile_id`
- `role_in_work`
- `contribution_summary`
- `sort_order`
- `is_owner`

这是本模块最关键的一张关系表，它决定了网络不是单点人物页，而是通过作品自然长出关系。

### 6.4 Opportunity

用于表达“我现在在找什么”。

建议字段：

- 基础信息：`id` `profile_id` `title` `summary`
- 机会类型：`opportunity_type`
- 需求内容：`looking_for_roles` `required_skills`
- 合作边界：`engagement_mode` `budget_hint` `duration_hint`
- 地域信息：`city_scope` `remote_allowed`
- 状态信息：`status` `expires_at`

### 6.5 SearchIndex / Discovery View

用于承接“我去找别人”。

不一定作为独立表实现，但产品层需要明确支持：

- 按人搜索
- 按作品搜索
- 按机会搜索
- 语义搜索
- 标签筛选
- 地域筛选
- 当前开放状态筛选

### 6.6 ContactRequest

用于表达针对 Builder 或合作机会提交的申请，并承载申请理由与联系方式交换授权。

建议字段：

- `id`
- `from_profile_id`
- `to_profile_id`
- `context_type` (`profile` / `opportunity`)
- `context_id`
- `message`（申请理由）
- `related_work_ids`
- `contact_consent`
- `status` (`pending` / `accepted` / `declined` / `withdrawn` / `expired`)
- `created_by_actor_type`
- `created_by_actor_id`

### 6.7 ContactExchange

用于表达“申请被接受后，双方按授权交换联系方式”。

建议字段：

- `request_id`
- 双方授权交换的联系方式字段
- `revealed_at`
- `revoked_at`

服务端必须校验申请状态与双方授权，不得仅靠前端隐藏联系方式。站内会话/即时消息不属于一期必做范围。

## 7. 开放 API 与 Skill 能力定义

BuilderNetwork 不以 Agent 聊天框或自动匹配作为一期前台体验。平台以稳定 API 承载能力，再通过 Skill 让用户自己的 Agent 按用户授权完成操作。

### 7.1 可开放的结构化能力

可逐步开放：

- 读取和编辑个人档案
- 创建、编辑和发布作品
- 创建、编辑和发布合作机会
- 搜索 Builder、作品和合作机会
- 查询、接受或婉拒联系申请（接受动作需要明确授权）

### 7.2 Skill 使用原则

Skill 只封装 API 能力，不绕过用户权限；不得自动接受申请，不得在申请接受且双方授权前读取私密联系方式。自动推荐、语义搜索和代操作可在后续阶段按需建设。

## 8. 主要业务流程

### 8.1 用户上传作品

1. 用户进入我的工作台
2. 填写作品标题、摘要、链接、封面、标签
3. 可选关联参与成员并说明各自角色/贡献
4. 预览后发布，必要时由运营审核

结果：

- 作品进入发现广场或作品列表
- 相关成员个人主页同步展示该作品
- 作品可作为申请时的能力证明

### 8.2 用户发布当前状态 / 机会

1. 用户进入“发布需求”
2. 填写目标、角色/技能、合作方式、预算和预计周期
3. 预览并发布，进入“找项目”列表
4. 用户可编辑、暂停招募、标记已找到合作者或关闭

结果：

- Builder 可以查看详情并提交申请理由
- 需求方在工作台查看和处理申请

### 8.3 用户主动找人

1. 用户进入“找 Builder”或“找项目”
2. 按角色、技能、档期、合作方式、项目类型等筛选
3. 查看 Profile、Work 或 Opportunity 详情
4. 提交申请理由，可关联作品并授权交换选定的联系方式
5. 接收方接受后，双方可查看彼此已授权的联系方式
6. 双方在站外推进沟通；工作台记录申请与合作进度

### 8.4 申请处理与合作推进

1. 接收方在工作台查看申请者公开资料、申请理由和关联作品。
2. 接收方接受或婉拒申请；接受只代表同意建立联系，不代表达成交易。
3. 接受后，服务端按双方授权展示联系方式。
4. 双方可将进度更新为已交换联系方式、沟通中、合作中、已完成或已结束。
5. 申请关闭或机会关闭后保留必要历史记录，供双方查阅及运营审计。

### 8.5 Agent / Skill 后续接入

Skill 基于与前台相同的 API，按授权搜索、创建内容和管理申请；一期不建设 Agent 聊天框、自动接受申请或绕过用户确认的外联能力。

## 9. 模块拆解

一期建议拆为 8 个功能模块。

### 9.1 Profile 模块

职责：

- 维护个人档案
- 维护合作状态
- 维护可见性与联系策略

后台能力：

- 创建/编辑/审核 Profile
- 精选/隐藏/认证

### 9.2 Work 模块

职责：

- 发布作品
- 上传作品媒体
- 维护作品成员关系

后台能力：

- 审核作品
- 管理标签
- 管理违规内容

### 9.3 Opportunity 模块

职责：

- 发布当前状态和合作机会
- 控制有效期和开放程度

后台能力：

- 置顶
- 下架
- 审核

### 9.4 Discovery 模块

职责：

- 支持人、作品、机会的统一发现
- 一期支持关键词、标签和结构化筛选；语义搜索后续评估

### 9.5 ContactRequest 模块

职责：

- 控制连接请求
- 防止直接打扰
- 承载申请理由、关联作品和联系方式交换授权

### 9.6 ContactExchange 模块

职责：

- 在申请接受后按双方授权展示联系方式
- 记录联系方式交换及访问审计
- 站内会话与即时消息不属于一期范围

### 9.7 Agent Action 模块

职责：

- 暴露可控的 API，供用户 Skill / Agent 调用
- 对 Token 和操作范围进行 scope 控制
- 对关键动作写入审计日志

### 9.8 Moderation / Operation 模块

职责：

- 审核资料和作品
- 处理异常连接行为
- 运营精选对象和机会广场

## 10. 前台信息架构

一期前台建议新增如下页面。

### 10.1 发现广场

市场独立壳层首页，区块包括：

- 精选主视觉
- 社区作品展示
- 热门 Builder
- 精选合作机会
- 分别跳转到“找 Builder”和“找项目”的入口

### 10.2 找 Builder 与个人主页

列表卡片和个人主页建议区块：

- 职业身份、技能和个人介绍
- 合作方式、开放状态与最早可合作时间
- 代表作品
- 当前机会
- 发起合作申请

### 10.3 Work 列表页 / 详情页

建议区块：

- 作品封面
- 项目摘要
- 参与成员
- 使用技术
- 作品链接
- 协作角色分工

### 10.4 找项目

列表与详情建议展示：

- 项目目标、招募角色、技能要求
- 合作方式、预算、周期、地点/远程
- 招募状态、申请数量与申请截止时间
- 提交申请理由和关联作品的入口

### 10.5 发布需求与我的工作台

发布需求表单支持项目目标、角色/技能、合作方式、预算、周期与截止时间。工作台包含：

- 个人资料、作品和档期管理
- 已发布合作机会及其状态
- 收到的申请与我发出的申请
- 申请处理结果和合作进度
- 仅对已接受且双方授权的申请展示联系方式

### 10.6 邀请制注册

注册前验证邀请码；随后完成账户身份验证、姓名/常用称呼、核心职业身份和一句话介绍。注册成功后引导用户补充资料与作品。

## 11. 后台信息架构

后台建议从当前 `Builder 管理` 升级为以下结构：

- `Profile 管理`
- `作品管理`
- `机会管理`
- `联系请求`
- `邀请码管理`
- `联系方式与授权审计`
- `标签与分类`
- `API / Skill 操作日志`

### 11.1 Profile 管理

支持：

- 新建与编辑
- 精选
- 认证
- 隐藏

### 11.2 作品管理

支持：

- 作品审核
- 成员关系维护
- 封面与附件管理

### 11.3 机会管理

支持：

- 发布与下架
- 置顶与精选
- 状态维护

### 11.4 联系申请与隐私管理

支持：

- 查看异常请求
- 处理举报、拉黑与违规行为
- 管理联系方式授权访问审计

## 12. API 设计原则

BuilderNetwork 必须遵循现有 VOIDLAB 架构原则：

- 前台不直接碰数据库
- 后台不直接碰数据库
- Skill / Agent 不直接碰数据库
- 所有动作都通过 API 承接

建议 API 分组如下：

- `/api/v1/auth/invitations`
- `/api/v1/auth/register`
- `/api/v1/network/profiles`
- `/api/v1/network/works`
- `/api/v1/network/opportunities`
- `/api/v1/network/search`
- `/api/v1/network/contact-requests`
- `/api/v1/network/contact-exchanges`

### 12.1 典型人类用户接口

- `GET /api/v1/public/network/profiles`
- `GET /api/v1/public/network/profiles/:slug`
- `GET /api/v1/public/network/works`
- `GET /api/v1/public/network/works/:slug`
- `GET /api/v1/public/network/opportunities`
- `POST /api/v1/auth/invitations/verify`
- `POST /api/v1/auth/register`
- `POST /api/v1/network/contact-requests`
- `POST /api/v1/network/contact-requests/:id/accept`
- `POST /api/v1/network/contact-requests/:id/decline`
- `GET /api/v1/network/contact-exchanges/:request_id`

### 12.2 典型 Agent 接口

建议以 scope 方式暴露，覆盖人类用户和 Skill 的统一权限边界：

- `network_profiles:read`
- `network_profiles:write`
- `network_works:read`
- `network_works:write`
- `network_opportunities:read`
- `network_opportunities:write`
- `network_contact_requests:read`
- `network_contact_requests:write`
- `network_contact_exchange:read`
- `network_contact_exchange:write`

典型动作：

- 发布作品
- 发布机会
- 搜索候选人
- 创建联系申请
- 接受或婉拒申请（需用户明确授权）
- 读取联系方式交换结果（仅限已接受且双方授权）

## 13. 数据模型建议

一期建议在现有 `builders` 的基础上新增以下核心表：

- `network_accounts`
- `network_invite_codes`
- `network_profiles`
- `network_profile_links`
- `network_private_contacts`
- `network_works`
- `network_work_members`
- `network_work_feedback`
- `network_opportunities`
- `network_contact_requests`
- `network_contact_exchanges`
- `network_saved_searches`
- `network_agent_actions`

建议策略：

- 现有 `builders` 不直接强改为废弃
- 可先迁移为 `network_profiles` 的前身数据源
- 一期允许双轨并存，逐步完成前台升级

## 14. 风控与治理

该模块涉及用户身份、申请和私密联系方式，必须在一期就考虑风控。

### 14.1 联系策略

默认连接流程：

1. 申请者提交申请理由、可选作品及联系方式交换授权。
2. 接收方查看公开资料与申请内容，再接受或婉拒。
3. 接受后，服务端按双方授权展示联系方式。
4. 私密联系方式在此前不可读取；接受申请不代表达成合作。

### 14.2 Agent 约束

Skill / Agent 操作必须受到用户授权和平台权限约束。

建议加入：

- API scope 控制
- 请求频率限制
- 关键操作的发起者记录
- 审计日志
- 用户撤回授权
- 禁止在授权前读取联系方式或自动接受申请

### 14.3 内容治理

需支持：

- 作品下架
- Profile 隐藏
- 机会关闭
- 申请限流、撤回与举报
- 联系方式访问审计

## 15. 一期范围（MVP）

一期优先跑通从邀请加入到建立合作联系的闭环。

### 15.1 必做

1. 邀请码验证、账户注册和基础 Profile 创建
2. Builder Profile、技能、合作方式与档期维护
3. Work 发布、展示及关联协作者
4. Opportunity 发布、筛选、编辑与状态管理
5. 申请理由、申请处理、状态追踪与授权交换联系方式
6. 发现广场、找 Builder、找项目、发布需求和我的工作台
7. API 鉴权、操作审计及最小 Skill 接入能力

### 15.2 可延后

- 点赞评论
- 关注关系
- 公开时间线
- 复杂推荐排序
- 站内即时聊天
- 星级评分、支付与交易担保
- 自动匹配与 Agent 自动外联
- 多租户组织空间

## 16. 实施阶段建议

### Phase 1：邀请与内容起盘

目标：

- 建立数据模型
- 跑通邀请码注册、Profile、Work、Opportunity 基础 CRUD
- 完成发现广场、找人、找项目、发布需求和工作台最小页面

### Phase 2：申请与联系方式交换

目标：

- 跑通申请提交、理由查看、接受/婉拒和状态追踪
- 按双方授权安全展示联系方式
- 加入申请限流、举报与审计

### Phase 3：API 与 Skill 扩展

目标：

- 稳定开放 Profile、Work、Opportunity、ContactRequest API
- 提供用户可复用的 Skill
- 根据使用数据评估语义搜索和推荐能力

## 17. 本轮实现建议

结合当前仓库，建议按以下顺序推进：

1. 确认注册认证方式及联系方式字段范围
2. 建立邀请注册、账户和 Profile 权限边界
3. 新增 Work、WorkMember 和作品互动
4. 新增 Opportunity 及状态管理
5. 新增 ContactRequest、ContactExchange 和服务端授权校验
6. 落地市场页面与工作台，再提供 API / Skill

这样可以最大化复用现有：

- 后台登录与权限
- 审计日志
- Leads
- 媒体管理
- API 分层
- Agent token 机制

## 18. 成功标准

一期完成后，应满足以下标准：

1. 用户可以建立结构化个人档案
2. 用户可以上传作品并关联多个协作者
3. 用户可以发布当前机会或合作状态
4. 用户可以搜索人、作品、机会
5. 用户可以提交申请、查看处理状态并维护合作进度
6. 申请接受且双方授权后，才能查看交换的联系方式
7. 用户可通过 API / Skill 在授权范围内管理资料和内容
8. 全部关键动作可审计、可控
