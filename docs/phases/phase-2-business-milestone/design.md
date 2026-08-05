# Phase 2 设计文档

## 1. 设计目标

本阶段设计文档用于承接“当前业务闭环跑通”的实现设计。

Phase 2 的目标不是新增很多后台页面，而是把当前网站从“后台可维护”推进到“前台能真实承接业务、并把外部意向统一沉淀进 Leads”。

## 2. 设计结论

Phase 2 建议围绕三条主线推进：

1. 前台 API 化
2. Leads 中台化
3. 外部意向统一入池

这三条主线里，**Leads 是 Phase 2 的核心枢纽**。

## 3. 模块范围

- 前台 API 化改造
- Contact 提交
- Event 报名 / 预约
- Builder 合作发起
- Leads 模块
- 基础站点配置

## 4. 设计原则

### 4.1 先统一 Leads，再扩入口

不要先把 Contact、Event、Builder 三个表单各自做一套后台处理逻辑。

正确顺序是：

1. 先定义统一 Lead 模型
2. 再让三个入口统一写入 Lead

### 4.2 先做最小可运营闭环

Phase 2 的重点不是复杂 CRM，而是先形成：

- 外部意向被接住
- 后台有人能跟进
- 状态能更新

### 4.3 前台切 API 要围绕业务闭环推进

前台 API 化不需要一次性改全站，而应优先改：

- Insights
- Events
- Builders

这三类页面与后台主数据直接对应，改造收益最高。

## 5. Leads 设计

### 5.1 Lead 的定位

Lead 是当前系统所有对外意向的统一承接对象。

它不是单一表单记录，而是一个运营跟进对象。

### 5.2 Lead 来源

Phase 2 第一版建议只支持三类来源：

- `contact`
- `event`
- `builder`

### 5.3 Lead 核心字段

建议第一版至少包含：

- `id`
- `source_type`
- `source_id`
- `name`
- `contact`
- `message`
- `status`
- `notes`
- `created_at`
- `updated_at`

### 5.4 Lead 状态设计

建议第一版状态如下：

- `new`
- `contacted`
- `following`
- `converted`
- `invalid`

这套状态已经足够支撑当前运营动作。

### 5.5 Lead 日志

Phase 2 建议为 Leads 预留跟进日志能力。

第一版可以使用独立表：

- `lead_logs`

字段建议：

- `id`
- `lead_id`
- `action`
- `content`
- `created_by`
- `created_at`

说明：

第一版不需要复杂审计系统，但至少要能留下跟进备注。

## 6. 数据模型设计

Phase 2 建议新增以下核心表：

- `leads`
- `lead_logs`
- `site_configs`

### 6.1 leads

建议字段：

- `id`
- `source_type`
- `source_id`
- `name`
- `contact`
- `message`
- `status`
- `owner_id`
- `created_at`
- `updated_at`

### 6.2 lead_logs

建议字段：

- `id`
- `lead_id`
- `action`
- `content`
- `created_by`
- `created_at`

### 6.3 site_configs

建议字段：

- `id`
- `config_key`
- `config_value_json`
- `updated_by`
- `updated_at`

说明：

Phase 2 的配置化能力只需要先服务于基础站点配置，不需要做通用配置平台。

## 7. API 设计

### 7.1 Leads API

建议新增：

- `GET /api/v1/leads`
- `GET /api/v1/leads/:id`
- `POST /api/v1/leads`
- `PUT /api/v1/leads/:id/status`
- `POST /api/v1/leads/:id/logs`

说明：

- 后台直接使用 `GET /api/v1/leads`
- 三类前台入口原则上也可以统一落到 `POST /api/v1/leads`

### 7.2 Contact API

建议新增：

- `POST /api/v1/contact/submit`

功能：

- 接收 Contact 页提交
- 自动创建 Lead
- `source_type = contact`

### 7.3 Event RSVP API

建议新增：

- `POST /api/v1/events/:id/rsvp`

功能：

- 接收活动报名 / 预约
- 自动创建 Lead
- `source_type = event`
- `source_id = event_id`

### 7.4 Builder Inquiry API

建议新增：

- `POST /api/v1/builders/:id/inquiry`

功能：

- 接收 Builder 合作发起
- 自动创建 Lead
- `source_type = builder`
- `source_id = builder_id`

### 7.5 前台公开数据 API

为支撑前台 API 化，建议补充公开读接口：

- `GET /api/v1/public/articles`
- `GET /api/v1/public/articles/:slug`
- `GET /api/v1/public/events`
- `GET /api/v1/public/events/:slug`
- `GET /api/v1/public/builders`
- `GET /api/v1/public/builders/:slug`

说明：

后台管理接口与前台公开接口建议分开，避免把后台接口直接暴露给前台。

## 8. 前台交互设计

### 8.1 Insights API 化

改造目标：

- 列表页读取公开文章接口
- 详情页按 slug 读取公开文章详情

### 8.2 Events API 化

改造目标：

- 活动列表页读取公开活动接口
- 活动详情页按 slug 读取公开活动详情
- 详情页报名表单提交到 RSVP 接口

### 8.3 Builders API 化

改造目标：

- Builder 列表页读取公开 Builder 接口
- Builder 详情页按 slug 读取公开 Builder 详情
- 详情页合作发起表单提交到 inquiry 接口

### 8.4 Contact 表单

Contact 页面增加真实表单提交能力，字段建议至少包含：

- 姓名
- 联系方式
- 留言内容

提交后：

- 创建 Lead
- 返回成功提示

## 9. 后台页面设计

Phase 2 新增后台页面建议如下：

1. Leads 列表页
2. Lead 详情页
3. 基础站点配置页

### 9.1 Leads 列表页

建议展示字段：

- 姓名
- 联系方式
- 来源
- 当前状态
- 创建时间
- 操作

### 9.2 Lead 详情页

建议展示：

- Lead 基础信息
- 来源对象
- 留言内容
- 当前状态
- 跟进日志

并支持：

- 更新状态
- 新增跟进备注

### 9.3 站点配置页

Phase 2 第一版只建议支持：

- 联系方式配置
- 首页精选位配置
- Banner 文案配置

## 10. 后端结构建议

建议 Phase 2 在 `apps/api` 中新增：

- `lead_repository.go`
- `lead_service.go`
- `lead_handler.go`
- `public_handler.go`
- `contact_handler.go`

并在 SQLite 初始化中新增：

- `leads`
- `lead_logs`
- `site_configs`

## 11. 开发顺序建议

建议严格按以下顺序推进：

### Step 1

先做 Leads 数据模型和后台 Leads API。

### Step 2

再做 Contact 提交入库。

### Step 3

再做 Event 报名 / 预约入库。

### Step 4

再做 Builder 合作发起入库。

### Step 5

再做前台 `Insights / Events / Builders` API 化。

### Step 6

最后补基础站点配置。

这样推进的好处是：

- Leads 中台先稳定
- 外部入口统一写入模型
- 前台 API 化有明确目标对象

## 12. 验收用例

### 12.1 Contact 用例

- Contact 表单可成功提交
- 提交后自动创建 Lead

### 12.2 Event 用例

- 活动页可提交报名 / 预约
- 提交后自动创建 Lead

### 12.3 Builder 用例

- Builder 详情页可提交合作发起
- 提交后自动创建 Lead

### 12.4 Leads 用例

- 后台可查看 Leads 列表
- 后台可查看 Lead 详情
- 后台可更新 Lead 状态
- 后台可记录跟进备注

### 12.5 前台 API 化用例

- Insights、Events、Builders 前台页面均不再依赖静态数据
- 后台更新后前台可见

## 13. 当前阶段完成标志

当以下条件同时满足时，可视为 Phase 2 完成：

1. 前台 Insights / Events / Builders 已切换为 API 数据
2. Contact / Event / Builder 三类外部意向都能统一进入 Leads
3. 后台可查看和跟进 Leads
4. 团队已无需改代码即可完成当前主业务运营

## 14. 一句话总结

Phase 2 的本质不是“再做几个页面”，而是：

**把前台展示、外部意向和后台跟进这三段链路真正接起来，让网站从“可维护后台”进入“可运营业务系统”。**
