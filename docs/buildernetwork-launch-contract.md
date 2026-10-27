# BuilderNetwork 上线契约基线

状态：第一步范围与接口基线，待进入实现前评审
日期：2026-10-23
适用范围：BuilderNetwork 首次正式上线

本文补充当前产品文档，用于锁定页面、业务对象、状态、权限和 API 的上线目标。既有产品设计继续作为产品原则参考；若与本文的当前上线决策冲突，以本文记录的上线决策为准。本文不表示接口或功能已经完成。

## 1. 上线目标

用户能够通过邀请注册并维护公开 Builder 主页和作品；发现 Builder、项目和招募机会；发布项目及招募机会；提交、处理合作申请；在双方同意后交换已授权的联系方式；并通过 Agent Token 在授权范围内读写资料。

页面不能依赖生产环境 Mock 数据。所有可见的主要操作必须有真实 API、明确的加载/错误/空状态和权限反馈。

## 2. 首发范围决策

| 能力 | 首发决策 |
| --- | --- |
| 注册 | 邀请码 + 邮箱 + 密码；不要求邮箱验证码 |
| Builder 主页 | 公开档案、技能、合作状态、档期、公开作品及作品贡献 |
| 作品 | 独立于项目和招募机会；支持创建、编辑、发布、下架和多成员贡献 |
| 项目 | 独立业务对象，承载项目介绍、计划、里程碑和团队 |
| 招募机会 | 关联项目，承载招募角色、技能、合作条件、截止日期和状态；一个项目可有多个机会 |
| 联系申请 | 申请 Builder 或具体招募机会；申请状态为待处理、已同意、已婉拒、已撤回 |
| 联系方式 | 私密存储；申请人按字段授权，接收方同意后才能按授权读取；可撤销后续展示 |
| Agent Token | 用户可创建、查看状态和停用 Token；按现有 scope 授权 |
| Discussions | 首发隐藏；不展示 Mock 主题或无效操作，后续单独建设 |
| 站内聊天、评分/交易/支付、自动撮合 | 不属于首发范围 |

## 3. 核心对象与生命周期

### Account 与 Profile

- Account 用于登录，首发沿用邀请制邮箱/密码注册，不包含邮件验证码流程。
- Profile 对外公开职业身份：姓名、角色、简介、城市、技能、合作方式、availability 和最早可合作时间。
- Profile 的公开状态与私密联系方式分离。公开 Profile 不代表公开邮箱、微信或手机号。
- Builder 合作状态使用 `open`、`limited`、`unavailable`，由用户维护，不由系统根据活跃度推断。

### Work

- Work 是已完成、进行中或可展示的作品，用来证明能力；它不是招募需求，也不是 Project。
- Work 字段包括标题、摘要、正文、封面、Demo/Repo 链接、标签、状态和成员贡献。
- Work 状态沿用 `draft`、`published`、`archived`。只有已发布且作者 Profile 公开的作品进入公开列表和个人主页。
- 成员关系包含 Builder、角色、贡献说明、顺序和是否为所有者。

### Project

- Project 表达项目本身，由发起方拥有。
- Project 承载 slug、标题、介绍、计划章节、里程碑、团队介绍和项目成员。
- 计划章节可选；里程碑不固定周期，时间可选，可使用固定日期、周期长度、相对描述、弹性安排或不指定。
- 团队成员关联公开 Profile，保存其在项目中的角色、贡献和排序；展示时返回当前公开身份摘要。

### Opportunity

- Opportunity 表达一个具体招募/合作需求，必须关联 Project。
- Opportunity 承载类型、招募角色、所需技能、合作方式、预算、周期、地点/远程、截止日期和状态。
- 状态沿用 `draft`、`open`、`paused`、`filled`、`closed`；过期由截止日期判定，不接受新的申请。
- 一个 Project 可关联多个 Opportunity；关闭一个机会不应自动关闭 Project 或其他机会。

### ContactRequest 与 ContactExchange

- ContactRequest 目标为 Builder 或具体 Opportunity；必须有申请理由，可选关联已发布作品和逐字段联系方式授权。
- 状态：`pending` → `accepted` 或 `declined`；申请人可在处理前将其设为 `withdrawn`。
- 首发不展示 `progress_status` 合作跟踪阶段。旧数据库字段可暂时保留，但新页面/API 不依赖它。
- 接受申请和建立联系方式授权快照必须处于同一事务。
- 联系方式只向申请双方返回双方各自明确授权的字段；接受前、拒绝/撤回后、撤销后均不可读取。
- 联系方式撤销阻止未来的平台读取，不能收回对方此前已经复制的信息。

## 4. 页面与 API 契约

所有列表使用统一分页结构：

```json
{
  "items": [],
  "total": 0,
  "page": 1,
  "page_size": 20,
  "facets": {}
}
```

现有发现接口已支持 `page`、`page_size`、`q`、`role`、`skill`、`city`、`availability`、`type`、`remote` 中的部分组合。上线目标补齐页面实际使用的筛选和计数。

- Builder 默认按最近更新排序；支持按姓名排序。
- Project 默认按最新发布排序；支持按标题排序。
- `facets` 提供全量匹配结果中的角色/机会类型计数，不能只统计当前页。
- 首页保留“热门项目”视觉区块，首发实际返回最新的 5 个有开放招募的 Project；不将其描述为真实热度排名。
- 项目列表以 Project 为单位，一个 Project 只出现一行，显示开放 Opportunity 数量/摘要；进入详情后选具体 Opportunity 再申请。

| 页面/动作 | 目标 API 行为 | 当前差距 |
| --- | --- | --- |
| 首页发现 | 返回公开作品、Builder 和最新 5 个有开放招募的 Project | 当前页面请求全量资料，再在前端选择展示内容 |
| Builder 列表 | 分页搜索、角色/技能/城市/档期/远程筛选 | 页面仍请求全量 Profile 后在浏览器过滤 |
| Builder 个人页 | Profile 详情 + 该 Builder 发起或参与的公开作品 | 页面请求全量作品后在浏览器筛选；新增按 Profile 获取作品的接口 |
| 项目列表 | 分页搜索和筛选独立 Project；机会信息作为项目的招募摘要 | 当前返回的是 Opportunity 列表，页面名和业务对象不一致 |
| 项目详情 | 一次返回 Project、团队成员公开身份、计划和相关开放机会 | 当前详情是 Opportunity；团队成员身份还要再请求全量 Profile |
| 发布/编辑 | Project 和 Opportunity 分别维护；可提供组合式“创建项目并发布机会”流程 | 当前一个 Opportunity 表单同时写入项目计划和团队 |
| 工作台 | 读写当前用户 Profile、Work、Project、Opportunity、申请、联系方式授权和 Token | 基本模块已有接口；申请更新后仍全量刷新，已接受申请逐条读取联系方式 |
| Discussions | 首发无页面入口、无 API 要求 | 当前组件是 Mock 数据，发帖按钮未接通；上线前需隐藏 |

### 目标路由建议

以下是目标契约命名建议，可在实现时保留旧路由作为兼容别名：

- `GET /api/v1/public/network/discovery/profiles`
- `GET /api/v1/public/network/profiles/:slug`
- `GET /api/v1/public/network/profiles/:slug/works`
- `GET /api/v1/public/network/discovery/projects`
- `GET /api/v1/public/network/projects/:slug`
- `GET /api/v1/network/projects`、`POST /api/v1/network/projects`、`PUT /api/v1/network/projects/:id`
- 现有 Opportunity 读写接口继续使用，但创建/更新时要求关联 Project。
- 申请列表返回申请双方公开摘要和方向相关的目标信息；已接受时只附带当前用户有权读取的联系方式。
- 接受/拒绝/撤回接口返回更新后的申请视图，前端不需要重载整个工作台。

## 5. 身份、权限与可见性

- 游客可读取已发布且未过期的公开内容，不得读取草稿、私密联系方式或未发布成员数据。
- 登录用户可读写自己的 Profile、Work、Project、Opportunity 和申请；服务端按所有权校验，不能只依赖前端隐藏按钮。
- 申请接收方可处理发给自己的申请；申请人可撤回自己的待处理申请。
- 接受/拒绝/撤回操作继续要求用户会话；联系方式交换按参与方身份、申请状态、撤销状态和授权快照校验。
- Agent Token 与用户会话权限分开；Agent Token 仅可使用签发时获批的 scopes，不得通过新接口扩大权限。
- 邀请码、Token、申请状态变化和联系方式访问保留审计记录；生产响应不得包含开发 Mock 字段或虚构联系方式。

## 6. 兼容与迁移原则

- 当前 `network_opportunities` 内的计划和团队字段需要迁移至 Project；原有每条 Opportunity 至少映射为一个 Project，避免丢失计划、团队、详情 slug 和申请关系。
- `ContactRequest.target_type=opportunity` 继续指向原 Opportunity ID；迁移不得改写已有申请的目标含义。
- 旧的 `/public/network/opportunities/:slug` 路由在前端完成切换前保留，必要时返回兼容视图。
- SQLite 迁移可重复执行，有迁移前备份、迁移后校验和回滚方案；不直接删除旧列或旧接口。
- `progress_status` 暂不从旧 SQLite 表删除，只停止作为首发业务状态读写。

## 7. 首发验收基线

1. 新用户必须有有效邀请码才能注册；注册使用邮箱和密码，不要求邮箱验证码。
2. 用户可维护公开 Profile 和作品；Builder 个人页只展示该 Builder 创建或参与的公开作品及其贡献。
3. 用户可创建一个 Project 并关联一个或多个 Opportunity；项目成员可跳转到对应 Builder 主页。
4. Builder/Project/Opportunity 列表分页筛选正确，结果数量和筛选计数有明确服务端定义。
5. 申请人可提交申请和作品；接收方可接受或拒绝；申请人可撤回待处理申请。
6. 未接受申请、非参与方或已撤销交换均无法读取联系方式；双方只能看到各自授权的字段。
7. 用户只能修改自己拥有的对象；越权、无效状态和重复操作会得到稳定错误响应。
8. Agent Token scope 对应的读取/写入边界经测试覆盖，敏感操作有审计记录。
9. 首发页面没有生产 Mock、失效按钮或 Discussions 入口。
10. 旧 SQLite 数据迁移后，Profile、作品、项目/机会、团队和申请关系数量及关联通过校验。

## 8. 后续实现前需评审

- 项目列表卡片以 Project 为单位；当一个 Project 有多个开放 Opportunity 时，列表显示机会数量/摘要，详情页选择具体 Opportunity 后申请。
- 首发首页按最新 5 个有开放招募的 Project 展示；若未来要运营精选或按实际热度排序，再增加精选/热度规则。
- discovery `facets` 至少返回当前筛选下的角色、机会类型及对应总数；Builder 支持最近更新/姓名排序，Project 支持最新/标题排序。
- Discussions 已定为首发隐藏；若未来纳入首发，需重新估算模型、API、权限和测试范围。
