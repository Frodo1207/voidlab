# Phase 3 设计文档

## 1. 设计目标

本阶段设计文档用于承接“运营效率与系统稳定性增强”的实现设计。

Phase 3 的目标不是继续扩展很多新业务，而是把 Phase 2 已经跑通的业务闭环，升级为一个能被多人长期、低风险、可审计运营的系统。

## 2. 设计结论

Phase 3 建议围绕五条主线推进：

1. 站点配置增强
2. 状态流规则化
3. 基础权限控制
4. 关键操作审计
5. 基础统计看板

这五条主线里，**状态流 + 权限 + 审计** 是 Phase 3 的核心骨架。

原因很简单：

- 只有配置化，没有权限和审计，系统仍然容易被误操作
- 只有权限，没有状态规则，团队运营动作仍然不标准
- 只有统计，没有可靠的动作记录，数据也很难被信任

## 3. 模块范围

- 站点配置增强
- 核心对象状态流规则
- 基础角色与权限
- 操作日志
- 基础统计页

## 4. 设计原则

### 4.1 先稳定运营动作，再扩展业务能力

Phase 3 不应优先增加更多页面或更多业务对象。

正确顺序是：

1. 先把已有动作做标准
2. 再让多人可协同
3. 再做更复杂的运营能力

### 4.2 权限必须围绕动作设计，而不是围绕页面设计

不要只做“某个页面能不能进”的权限。

更关键的是：

- 能不能创建
- 能不能编辑
- 能不能发布
- 能不能改状态
- 能不能看 Leads

因此 Phase 3 的权限设计应以“动作”作为判断核心。

### 4.3 状态必须是业务约束，不只是展示字段

当前系统中的 `status` 不能继续只当作一个普通字符串字段使用。

Phase 3 需要让状态承担真正的业务规则职责：

- 限制允许的流转方向
- 阻止非法变更
- 作为统计与审计的基础

### 4.4 审计必须优先服务运营追责与回溯

Phase 3 的日志不需要一开始就做成复杂审计平台。

但至少要满足两个问题：

1. 这是谁做的
2. 他什么时候做的

如果一个关键动作不能回答这两个问题，就不算真正可运营。

### 4.5 统计应以现有业务对象为中心

Phase 3 的统计不追求 BI 化，不追求复杂报表。

第一版只要能回答：

- 当前有多少内容
- 当前有多少活动
- 当前有多少 Builder
- 当前有多少 Leads
- Leads 分布在哪些状态

就已经足够支撑日常运营判断。

## 5. 当前系统状态判断

进入 Phase 3 之前，当前系统已经具备：

- 资讯、活动、Builder 的后台真实 CRUD
- 前台 `Insights / Events / Builders` API 化
- `Contact / Event / Builder` 三类入口统一进入 Leads
- Leads 后台列表、详情、状态更新、跟进日志
- 基础站点配置页

因此当前系统的主要问题已经不是“链路是否存在”，而是：

- 配置项还不够完整
- 状态规则还不严格
- 权限仍然非常粗
- 审计能力仍然较弱
- 统计还没有形成统一后台视图

这正是 Phase 3 的切入点。

## 6. 站点配置增强设计

### 6.1 设计目标

在 Phase 2 已有 `site_configs` 基础上，扩展成“基础运营配置中心”。

这里仍然不做通用低代码配置平台，只做明确、可维护、可验证的几类站点配置。

### 6.2 配置范围

Phase 3 第一版建议支持以下配置键：

- `home_banner`
- `home_featured`
- `contact_channels`
- `footer_config`
- `global_cta`
- `featured_content_slots`

### 6.3 数据结构建议

继续复用已有 `site_configs`：

- `id`
- `config_key`
- `config_value_json`
- `updated_by`
- `updated_at`

不新增独立配置表。

原因：

- 当前配置种类还不多
- `site_configs` 已经能承载需求
- 更利于快速扩展和统一管理

### 6.4 配置项建议

#### `footer_config`

建议至少包含：

- `copyright_text`
- `links`
- `contact_email`

#### `global_cta`

建议至少包含：

- `title`
- `description`
- `primary_label`
- `primary_path`
- `secondary_label`
- `secondary_path`

#### `featured_content_slots`

建议至少包含：

- 首页精选文章位
- 首页精选活动位
- 首页精选 Builder 位

说明：

Phase 2 的首页精选主要依赖“内容自身的 featured 标记”。

Phase 3 可以进一步增强为：

- 默认按 featured 读取
- 支持后台手动指定若干固定精选位

## 7. 状态流设计

### 7.1 设计目标

Phase 3 需要把当前核心对象的状态从“可随便写的字符串”升级为“受控业务流”。

### 7.2 核心对象范围

- Article
- Event
- Builder
- Lead

### 7.3 Article 状态流

建议状态：

- `draft`
- `published`
- `archived`

建议流转：

- `draft -> published`
- `published -> draft`
- `published -> archived`
- `archived -> draft`

约束：

- 未完成必要字段时，禁止发布
- 已归档内容默认不出现在前台公开接口中

### 7.4 Event 状态流

当前 Event 已有内容状态 `draft / published`，Phase 3 建议继续保持“内容状态”和“展示状态”分离。

建议内容状态：

- `draft`
- `published`
- `archived`

前台展示状态仍根据 `event_time` 推导：

- `next`
- `live`
- `done`

约束：

- 只有 `published` 的活动可进入公开接口
- 只有 `draft` 或 `published` 的活动可编辑
- `archived` 活动默认不在首页推荐

### 7.5 Builder 状态流

建议状态：

- `draft`
- `published`
- `archived`

建议流转：

- `draft -> published`
- `published -> draft`
- `published -> archived`
- `archived -> draft`

约束：

- 只有 `published` 的 Builder 可进入前台公开网络

### 7.6 Lead 状态流

Phase 2 已有：

- `new`
- `contacted`
- `following`
- `converted`
- `invalid`

Phase 3 建议正式定义允许流转方向：

- `new -> contacted`
- `new -> invalid`
- `contacted -> following`
- `contacted -> invalid`
- `following -> converted`
- `following -> invalid`
- `following -> contacted`

不建议允许：

- `converted -> new`
- `invalid -> following`

说明：

如果后续确实需要重新激活，应使用明确的“重新打开”动作，而不是任意改状态。

### 7.7 实现建议

状态流逻辑应放在 `service` 层，而不是前端限制。

前端可以做交互约束，但最终判断必须由后端执行。

## 8. 权限设计

### 8.1 设计目标

当前系统虽然已有 `admin / editor` 角色字段，但权限仍基本未生效。

Phase 3 应把角色从“数据字段”升级为“真实访问控制”。

### 8.2 角色定义

建议保留并扩展为：

- `admin`
- `editor`
- `ops`

如果当前实现更偏销售语义，也可以命名为 `sales`，但从当前项目语境看，`ops` 更贴近实际运营角色。

### 8.3 动作权限矩阵建议

#### `admin`

可执行：

- 全部内容管理
- 全部状态流转
- Leads 查看与跟进
- 站点配置修改
- 统计查看
- 用户与角色管理（若本阶段不做用户管理页，则先保留能力设计）

#### `editor`

可执行：

- 文章管理
- 活动管理
- Builder 管理
- 媒体管理
- 查看部分站点配置

限制：

- 不可管理 Leads
- 不可修改关键站点配置
- 不可进行高风险发布配置操作

#### `ops`

可执行：

- Leads 查看
- Leads 状态更新
- Leads 跟进日志
- 基础统计查看

限制：

- 不可编辑文章、活动、Builder
- 不可修改站点配置

### 8.4 技术实现建议

建议在 API 层新增基于角色的中间件，例如：

- `RequireRole("admin")`
- `RequireAnyRole("admin", "editor")`
- `RequireAnyRole("admin", "ops")`

后台前端的菜单和按钮也应同步按角色隐藏，但真正的权限校验必须由 API 负责。

## 9. 操作日志设计

### 9.1 设计目标

Phase 3 需要让关键操作可追踪。

这里的日志与 Leads 跟进日志不是一回事。

- Leads 跟进日志：偏业务沟通记录
- 操作日志：偏系统动作审计

### 9.2 审计对象范围

建议第一版覆盖：

- Article 创建 / 更新 / 发布 / 归档
- Event 创建 / 更新 / 发布 / 归档
- Builder 创建 / 更新 / 发布 / 归档
- Lead 状态变更
- Site Config 更新

### 9.3 数据模型建议

建议新增表：

- `operation_logs`

字段建议：

- `id`
- `object_type`
- `object_id`
- `action`
- `before_json`
- `after_json`
- `operator_id`
- `operator_role`
- `created_at`

### 9.4 记录策略

不是所有动作都必须记录完整前后快照。

第一版建议：

- 创建：记录 `after_json`
- 更新：记录 `before_json + after_json`
- 状态变更：记录旧状态和新状态
- 配置更新：记录配置 key 与前后值

### 9.5 实现建议

日志记录应尽量在 `service` 层完成，避免：

- 前端漏记
- repository 层不了解业务上下文

## 10. 基础统计设计

### 10.1 设计目标

Phase 3 的统计页主要服务：

- 每日运营巡检
- 内容与线索总量判断
- Leads 跟进状态分布判断

### 10.2 看板内容建议

建议统计页第一版包含：

- 文章总数
- 已发布文章数
- 活动总数
- 已发布活动数
- Builder 总数
- 已发布 Builder 数
- Leads 总数
- 新 Leads 数
- 跟进中 Leads 数
- 已转化 Leads 数
- 无效 Leads 数

### 10.3 数据来源

直接基于现有表统计：

- `articles`
- `events`
- `builders`
- `leads`

第一版不引入缓存层，不引入独立统计仓库。

### 10.4 页面建议

建议新增后台页面：

- `Dashboard / 运营看板`

当前已有仪表盘页，可以考虑直接升级现有 Dashboard，而不是新建独立页面。

## 11. API 设计

### 11.1 状态流 API

建议为核心对象增加明确状态更新接口，而不是继续依赖通用更新接口隐式改状态。

例如：

- `PUT /api/v1/articles/:id/status`
- `PUT /api/v1/events/:id/status`
- `PUT /api/v1/builders/:id/status`
- `PUT /api/v1/leads/:id/status`

说明：

- Leads 状态接口已存在
- 文章 / 活动 / Builder 建议在 Phase 3 增加独立状态接口

### 11.2 Site Config API

在 Phase 2 现有基础上扩展：

- `GET /api/v1/site-configs`
- `PUT /api/v1/site-configs/:key`
- `GET /api/v1/public/site-configs`

不建议做：

- 任意 key 的批量动态写入
- 无约束 schema 的配置提交

### 11.3 Statistics API

建议新增：

- `GET /api/v1/dashboard/summary`

返回示例结构可包含：

- `articles.total`
- `articles.published`
- `events.total`
- `events.published`
- `builders.total`
- `builders.published`
- `leads.total`
- `leads.by_status`

### 11.4 Operation Logs API

建议新增：

- `GET /api/v1/operation-logs`
- `GET /api/v1/operation-logs/:id`

第一版只读，不在后台支持手工写入。

## 12. 后台页面设计

### 12.1 站点配置页增强

在现有 `site-configs` 页面基础上继续扩展：

- 页脚配置
- 全局 CTA
- 精选位配置

### 12.2 仪表盘升级

当前 Dashboard 应升级为基础运营看板，显示：

- 内容总量
- 活动总量
- Builder 总量
- Leads 总量
- Leads 状态分布

### 12.3 操作日志页

建议新增：

- 操作日志列表页

建议字段：

- 操作时间
- 操作人
- 对象类型
- 对象 ID
- 动作
- 简要说明

### 12.4 权限驱动菜单显示

后台菜单应根据角色裁剪：

- `editor` 不显示 Leads 管理
- `ops` 不显示内容管理和站点配置

## 13. 后端结构建议

建议新增或增强以下结构：

- `internal/service/permission_service.go`
- `internal/service/dashboard_service.go`
- `internal/service/operation_log_service.go`
- `internal/repository/operation_log_repository.go`
- `internal/http/dashboard_handler.go`
- `internal/http/operation_log_handler.go`

同时逐步把状态流校验收敛到各模块 service 中。

## 14. 推荐开发顺序

Phase 3 不建议五块同时开做。

推荐顺序如下：

1. 状态流规则化
2. 基础权限控制
3. 仪表盘统计
4. 操作日志
5. 站点配置增强

原因：

- 状态流和权限是“防错骨架”
- 统计是“运营反馈面”
- 日志是“追责与回溯”
- 配置增强虽然重要，但优先级略低于系统稳定性骨架

## 15. 验收用例建议

### 15.1 状态流

- 非法状态流转会被拒绝
- 合法状态流转可成功执行

### 15.2 权限

- `editor` 无法访问 Leads API
- `ops` 无法编辑文章 / 活动 / Builder
- `admin` 可访问全部模块

### 15.3 审计

- 发布文章后可看到操作日志
- 更新站点配置后可看到操作日志
- 修改 Leads 状态后可看到操作日志

### 15.4 统计

- Dashboard 能显示各核心对象总量
- Leads 状态统计与数据库结果一致

### 15.5 配置

- 后台修改页脚配置后，前台对应区域同步生效
- 后台修改精选位配置后，首页展示发生变化

## 16. 一句话结论

Phase 3 的核心不是“新增能力”，而是把已有业务闭环升级为一个**有规则、有权限、有审计、可看板化的稳定运营系统**。
