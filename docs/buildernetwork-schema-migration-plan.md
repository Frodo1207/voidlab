# BuilderNetwork Project / Opportunity 数据迁移方案

状态：Project/Opportunity API、前端聚合接口迁移与 V2 数据库约束强化已实现；生产一致性备份与副本迁移演练通过，尚未部署新版本
日期：2026-10-23
范围：SQLite 数据结构、旧数据映射、API/Token 兼容和迁移验收。

## 1. 当前结构与关键依赖

- `network_opportunities` 同时存项目标题/介绍、招募字段、`project_plan_json`、`team_intro` 和 `team_members_json`。
- `network_opportunities.profile_id` 当前兼任项目发起人和招募机会所有者。
- `network_contact_requests` 用 `target_type='opportunity'` 与 `target_id` 指向机会，但没有机会外键；因此旧 Opportunity ID 必须保留。
- 团队成员当前序列化在 JSON 中。写入时校验 Profile 已公开，但读取时成员身份来自保存的 JSON 副本。
- SQLite 启动通过 `CREATE TABLE IF NOT EXISTS`、兼容补列函数初始化，没有现存的版本化迁移表。
- `OpenSQLite` 已开启外键、WAL、5 秒 busy timeout，并把连接数限制为 1；迁移应沿用单连接写入特性。
- 旧的公开详情路由、工作台机会接口和 `network_opportunities:read/write` Agent Token scope 均依赖现有 Opportunity 形状。

## 2. 目标数据模型

### `network_projects`

建议字段：

| 字段 | 说明 |
| --- | --- |
| `id` | Project 主键 |
| `owner_profile_id` | 发起人 Profile |
| `slug` | Project 独立公开 URL 标识 |
| `title`、`summary` | 项目名称和介绍 |
| `status` | `draft`、`published`、`archived` |
| `project_plan_json` | 可选计划章节和灵活里程碑 |
| `team_intro` | 项目团队介绍 |
| `created_at`、`updated_at` | 时间戳；旧记录迁移时保留 |

建议唯一索引 `slug`，以及 `(status, created_at DESC)`、`(owner_profile_id, status)` 索引。

### `network_project_members`

用关系表保存项目成员，成员身份始终从当前公开 Profile 读取：

| 字段 | 说明 |
| --- | --- |
| `project_id` | 关联 Project |
| `profile_id` | 关联 Builder Profile |
| `role`、`contribution` | 项目角色及贡献说明 |
| `sort_order` | 展示顺序 |
| `created_at` | 加入时间 |

主键为 `(project_id, profile_id)`；增加按 `profile_id, project_id` 查询的索引。公开 Project 查询只返回当前已发布 Profile 的公开摘要，不复制姓名、slug、技能等易过期信息。Project Owner 单独返回，不要求其一定出现在成员清单中。

### `network_opportunities`

- 增加 `project_id`，作为所属 Project 的规范关联。
- 保留 `id`、`slug`、`profile_id`、`title`、`summary`、角色/技能/合作条件、状态、截止时间等现有字段。
- `profile_id` 在兼容期作为旧所有者字段保留；新写入必须保证它等于 Project 的 `owner_profile_id`。新查询以 Project Owner 为准，后续可单独评估移除冗余列。
- 目标约束为 `project_id NOT NULL REFERENCES network_projects(id)`；第一段迁移先使用可空列完成回填，数据校验通过后再通过 SQLite 表重建强化 `NOT NULL`。
- 旧 `project_plan_json`、`team_intro`、`team_members_json` 列先保留作兼容快照，不再作为新 API 的真源；确认客户端和回滚窗口结束后再另行移除。
- Project/Opportunity 使用关闭、归档状态代替硬删除，避免历史申请中的 `target_id` 失去对应对象。

新增索引建议：`(project_id, status, created_at DESC)`。旧机会状态索引在新发现 API 稳定前保留。

## 3. 旧数据映射规则

每条旧 Opportunity 生成一个 Project，保证现有项目详情可一对一迁移：

| 旧 `network_opportunities` | 新 Project | Opportunity |
| --- | --- | --- |
| `id` | 新生成 Project ID | 原 ID 不变 |
| `profile_id` | `owner_profile_id` | 保留为兼容字段，值不变 |
| `slug` | 初始 Project slug 复制旧 slug | 原 slug 不变 |
| `title`、`summary` | 初始项目标题/介绍 | 保留旧值，供旧机会详情兼容 |
| `project_plan_json` | 原样复制 | 旧列保留，不再作为真源 |
| `team_intro` | 原样复制 | 旧列保留，不再作为真源 |
| `team_members_json` | 解析后写入 `network_project_members` | 旧 JSON 保留 |
| 招募角色、技能、预算、方式、截止日期、状态 | 不迁移 | 原值不变 |
| `created_at`、`updated_at` | 原值复制 | 原值不变 |

Project 状态映射：旧 Opportunity 为 `draft` 时 Project 为 `draft`；其他旧状态映射为 `published`，Opportunity 自身状态保持原值。公开与否仍需同时满足 Project 已发布、Owner Profile 已发布；开放申请还要满足 Opportunity 为 `open` 且未过期。

`team_members_json` 迁移仅写入 `profile_id`、`role`、`contribution` 和 `sort_order`。显示身份改由 Profile 查询，不延用 JSON 中的姓名或 slug 快照。里程碑 `owner_member_ids` 仍使用 Profile ID，因此不需要改写。

`network_contact_requests` 无需改数据：Opportunity ID 不变，所有 `target_type='opportunity'` 的历史申请继续指向同一机会。不得把申请目标改成 Project ID。

## 4. 分阶段迁移策略

### 阶段 A：预检与备份

- 在数据库副本上运行迁移演练；生产执行前备份 SQLite 文件，并确认 WAL checkpoint/备份方式可还原一致状态。
- 本地开发库已通过 SQLite `.backup` 创建一致性副本，并以当前源码启动 API 完成 V1/V2 演练：迁移前 1 条 Opportunity、4 个 Profile、0 条申请；迁移后 1 条 Opportunity 关联 1 个 Project，Opportunity ID/slug/status 保持不变，`PRAGMA integrity_check` 与 `foreign_key_check` 通过，公开 Project 聚合 API 正常返回。该副本没有历史申请，申请目标保留由迁移自动化测试覆盖；未触碰生产数据库。
- 生产 SQLite 已使用当前 Go SQLite 驱动的 `VACUUM INTO` 在服务器侧生成一致性备份：`/opt/voidlab/backup/voidlab-pre-buildernetwork-v2-20261023T022621Z.sqlite`。在该备份副本上运行当前 API 的 V1/V2 启动迁移后，5 条 Opportunity 对应 5 个 Project；迁移前后 Opportunity ID/slug/status 摘要哈希相同，0 条 Opportunity 未关联，`integrity_check` 与 `foreign_key_check` 通过，Project 公开 API 返回 200。生产 API 健康且容器未替换；本轮未修改线上源库。
- 统计 Opportunity、申请、JSON 计划、团队成员行数，以及 status/profile 外键完整性。
- 校验 `team_members_json` 和 `project_plan_json` 可解析。遇到坏 JSON、重复成员或缺失 Profile 时，输出 Opportunity ID 并中止迁移；不静默丢弃数据。

### 阶段 B：版本化、可重复的回填

- 增加 BuilderNetwork 专属的 schema migration 版本表，不继续把多步迁移零散塞进通用补列函数。
- 一个事务中创建 Project/成员表、给 Opportunity 增加可空 `project_id`、为每条旧 Opportunity 创建 Project、回填成员关系、填写 `project_id` 并记录 migration 版本。
- migration 版本最后写入；任一校验或 SQL 操作失败时整笔事务回滚。
- 重复启动时按 migration 版本跳过已完成步骤；禁止重复生成 Project 或重复成员。
- 所有旧 Opportunity，包括草稿、暂停、已关闭和过期记录都要回填，不仅是公开列表中的记录。

### 阶段 C：强化关联约束与新写入路径

- 回填后验证所有 Opportunity 都有合法 Project，再重建 `network_opportunities` 表，将 `project_id` 强制为 `NOT NULL` 并加外键；保留 Opportunity ID、slug、旧列和现有数据。
- Project 创建/更新、成员维护独立写入 Project 表；Opportunity 写入只负责招募字段，并验证该 Project 归当前用户所有。
- “同时创建项目并发布机会”在同一事务内创建 Project、成员和 Opportunity，失败时不留半个对象。
- milestone owner 必须属于该 Project 的成员；Project 成员必须引用有效 Profile。未公开 Profile 不得在公开响应中泄露。
- 硬删除 Project/Opportunity 在首发不开放；历史申请需保留可读上下文。

## 5. API 与 Agent Token 兼容

- 新 Project API 使用 Project ID/slug；机会 API 使用 Opportunity ID/slug，并通过 `project_id` 关联 Project。
- 成员端提供 `GET/POST /api/v1/network/projects`、`GET/PUT /api/v1/network/projects/:id`；公开端提供 `GET /api/v1/public/network/projects` 与 `GET /api/v1/public/network/projects/:slug`。
- Project 管理路由要求 `network_projects:read/write`。既有 Opportunity scopes 不会自动获得 Project API 权限；Agent Token 创建页和 Trae/Codex/Claude Skill 包已增加新 scope 说明。
- 旧 `POST /api/v1/network/opportunities` 未传 `project_id` 时，在同一事务创建 Project、成员和 Opportunity；传入时只接受当前用户拥有的 Project。Opportunity 更新也在事务内同步兼容字段与 Project 真源。
- 旧 `/public/network/opportunities/:slug` 保留为兼容视图：旧客户端仍能读取 Project 介绍、计划和团队；新前端改用 Project 详情 API。
- 项目市场以 Project 为展示单位，从其开放 Opportunities 汇总筛选条件；详情页选择具体 Opportunity 后发起申请，申请目标继续使用 Opportunity ID。
- 首发项目详情隐藏 Discussions 入口，待讨论数据与权限接口就绪后再开放。
- 旧工作台 Opportunity DTO 可暂时包含 Project 摘要和旧详情字段，但后端从 Project/成员关系组装响应，不能继续把机会列当真源。
- 旧 Opportunity ID、slug 和申请目标不变，避免个人工作台、通知和已发出的申请失效。
- 新增 `network_projects:read` 与 `network_projects:write` scope，供新 Project 路由使用；保留既有 `network_opportunities:*` scope 供机会读写。
- 旧 Opportunity API 在兼容期继续接受旧的项目详情 payload，以保持已签发 Agent Token 的既有实际能力；新独立 Project API 必须要求 Project scope。API 文档要明确这一区别，不得让旧 Token 未经授权地调用全新 Project 管理能力。
- 结束兼容期前盘点 Agent Token 使用情况，并给 Skill/Bridge 和用户明确迁移说明；不静默扩大旧 token 的新路由权限。

## 6. 校验、回滚和上线门槛

迁移后必须通过以下校验：

1. 旧 Opportunity 总数等于迁移映射数；所有旧 ID、slug、owner 和招募字段保持一致。
2. 每条旧 Opportunity 恰好关联一个新 Project；不存在空 `project_id` 或 owner 不一致。
3. 计划 JSON、团队介绍、成员数和顺序与迁移前一致；Project 成员均关联有效 Profile。
4. `network_contact_requests` 的行数、ID、`target_type` 和 `target_id` 完全不变。
5. `PRAGMA foreign_key_check` 无结果；关键索引存在。
6. 新旧公开详情路由、工作台机会列表、已发申请详情、Agent Token 读写通过 API 回归测试。
7. 迁移重复执行不产生重复 Project/成员；中途失败后重启可安全继续或保持原状。

回滚边界：

- 迁移 DDL 和回填应在事务内完成；失败时回到迁移前结构。
- 新版本开始写入 Project 后，旧二进制无法理解多 Opportunity 共享 Project 的语义。因此不承诺“只回滚应用镜像、继续使用新数据库”安全。
- 发布前保存数据库备份；若新版本需要完全回退，恢复数据库备份并同时回退应用。若新版本已接收有效用户写入，优先修复并向前迁移，避免用旧备份覆盖新数据。
- 旧详情列在兼容期保留，但不将它们当作数据库回滚机制。

## 7. 上线前需要确认的产品规则

- 是否接受 SQLite 使用两次迁移：第一版回填并保留旧列，后续再强化 `project_id NOT NULL`/清理旧列。建议接受，以降低一次性表重建风险。
- Project 无开放 Opportunity 时，是否仍出现在项目市场。推荐：已发布 Project 可浏览，但没有开放机会时不显示申请按钮；首页区块仍只取有开放机会的最新 5 个。
- 用户要求删除账号/Profile 时，若拥有 Project 或历史申请，首发采取阻止硬删除并先归档，还是实现匿名化/转移 Owner。建议首发阻止硬删除并给出明确提示，后续再设计数据删除流程。

## 8. 当前实现进度

已实现：

- 新增 `builder_network_schema_migrations` 版本表，并以单个事务执行版本 1。
- 创建 `network_projects`、`network_project_members` 及查询索引。
- 为 `network_opportunities` 增加可空 `project_id` 外键列；旧字段、机会 ID 和 slug 保持不变。
- 为所有未关联 Project 的旧 Opportunity 创建一对一 Project，复制标题、介绍、计划、团队介绍和时间戳；草稿仍映射为草稿，其他旧机会映射为已发布 Project。
- 把团队 JSON 里的 Profile ID、角色和贡献回填到成员关联表；姓名、slug 等身份快照不迁移。
- 对非法 JSON、缺失 Owner/成员 Profile、重复成员和外键异常报错并回滚，不记录已应用版本。
- 补充测试覆盖旧表补列、数据回填、成员顺序、历史申请目标不变、重复运行及异常数据回滚。
- 增加 Project Service/Repository/HTTP 读写与公开详情接口；公开查询仅返回已发布 Project、公开 Profile 成员和开放机会。
- 新增独立 Project Agent Token scopes，旧 token 不会自动获得新权限。
- Opportunity 创建支持显式关联自有 Project；旧式请求未传 `project_id` 时原子创建 Project 与 Opportunity。旧详情/工作台响应从 Project 与成员关联表组装计划和团队。
- 补充 Project 所有权、可见性、Agent scope 隔离和 Opportunity 创建事务回滚测试；API 全量测试通过。

本轮新增完成：

- 已实现 migration version 2：在单个事务中重建 `network_opportunities`，要求 `project_id NOT NULL REFERENCES network_projects(id) ON DELETE RESTRICT`，保留 Opportunity ID、slug、所有字段、两个原有索引以及逻辑申请目标。
- `OpenSQLite` 按顺序运行 V1 回填和 V2 约束强化；V2 遇到未关联 Opportunity 时中止，不会静默丢弃或改写数据。
- 前端项目列表和详情已切换至 `/api/v1/public/network/projects` 聚合接口；DEV 预览 mock 仍可用，真实申请保持指向选定 Opportunity。
- 补充 V2 测试覆盖旧表约束升级、字段及历史申请保留、外键和 NOT NULL、幂等执行及未关联数据回滚。

当前本地代码、后端测试、前端构建、本地演练及生产数据副本迁移演练均已完成。服务器保留上述迁移前备份；生产 API 仍运行原容器，本轮未执行部署或切换线上数据库。
