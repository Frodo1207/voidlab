# Phase 2 验收文档

## 1. 文档目的

本文档用于确认 Phase 2 是否已经达到“前台与后台开始形成真实业务闭环”的完成标准。

Phase 2 的目标不是把系统一次做成完整 CRM 或完整运营平台，而是让网站从“后台可维护”推进到“前台能真实承接内容与意向，并统一沉淀到后台运营系统”。

## 2. 验收结论

截至当前版本，Phase 2 已基本完成。

结论依据：

- 前台 `Insights / Events / Builders` 已切换到真实公开 API
- `Contact / Event / Builder` 三类外部入口已真实入池 Leads
- Leads 后台列表、详情、状态流转、跟进日志已可用
- 基础站点配置已配置化，并可从后台修改后同步影响前台关键区域

因此可以判断：

**Phase 2 已达到“业务闭环可运行”的阶段目标，可以进入下一阶段。**

## 3. 验收范围回顾

根据 Phase 2 需求与设计文档，本阶段验收范围包括：

1. 前台 API 化
2. Leads 中台
3. 外部意向统一入池
4. 基础站点配置

## 4. 验收清单

### 4.1 前台 API 化

验收项：

- 资讯列表页读取公开文章接口
- 资讯详情页按 slug 读取公开文章详情
- 活动列表页读取公开活动接口
- 活动详情页按 slug 读取公开活动详情
- Builder 列表页读取公开 Builder 接口
- Builder 详情页按 slug 读取公开 Builder 详情
- 首页核心内容区读取真实内容源

当前状态：

- 已完成

说明：

- 已提供 `public/articles`、`public/events`、`public/builders` 三组公开读接口
- 前台 `Insights / Events / Builders` 及首页核心内容区均已切到真实 API
- 前台保留了必要的展示层兜底逻辑，以兼容后台尚未结构化的展示字段

### 4.2 Leads 中台

验收项：

- 可查看 Leads 列表
- 可查看 Lead 详情
- 可更新 Lead 状态
- 可新增跟进备注
- 可保留跟进日志

当前状态：

- 已完成

说明：

- 已新增 `leads` 与 `lead_logs` 数据表
- 后台已具备 Leads 列表页与详情页
- 已实现状态流转与日志记录能力

### 4.3 外部意向统一入池

验收项：

- Contact 页面提交后可创建 Lead
- 活动详情页报名 / 预约后可创建 Lead
- Builder 详情页合作发起后可创建 Lead
- 线索来源可区分 `contact / event / builder`

当前状态：

- 已完成

说明：

- 已新增：
  - `POST /api/v1/contact/submit`
  - `POST /api/v1/events/:id/rsvp`
  - `POST /api/v1/builders/inquiry`
- 三类入口已统一写入 Leads
- `event` 与 `builder` 来源已支持绑定真实 `source_id`

### 4.4 基础站点配置

验收项：

- 后台可维护首页 Banner 文案
- 后台可维护首页指标与区块说明
- 后台可维护 Contact 页联系方式卡片
- 前台可读取公开站点配置

当前状态：

- 已完成

说明：

- 已新增 `site_configs` 数据表
- 后台已新增站点配置页面
- 已提供 `GET /api/v1/public/site-configs`
- 首页和 Contact 页已开始使用配置源

## 5. 验证记录

当前阶段已完成以下验证：

### 5.1 编译验证

- `apps/api` 执行 `go build ./...` 通过
- `apps/admin` 执行 `npm run lint` 通过
- `apps/admin` 执行 `npm run build` 通过
- `apps/web` 执行 `npm run lint` 通过
- `apps/web` 执行 `npm run build` 通过

### 5.2 运行时验证

已分别验证：

- `public/articles` 列表与详情
- `public/events` 列表与详情
- `public/builders` 列表与详情
- `contact / event / builder` 三类入口提交入库
- Leads 列表、详情、状态更新、日志追加
- `public/site-configs` 与后台 `site-configs` 读写链路

## 6. 当前实现边界

虽然 Phase 2 已完成，但仍有一些明确边界：

- 站点配置仍是“基础配置”，不是完整配置平台
- Leads 仍是轻量运营中台，不是完整 CRM
- 前台仍存在部分展示文案兜底，不代表后台模型已经完全产品化
- 统计、权限、自动分配、通知等高级运营能力尚未进入本阶段

这些都属于后续阶段的范围，不影响 Phase 2 完成判断。

## 7. Phase 2 输出物确认

本阶段已经产出：

- 前台公开内容 API
- Leads 数据表与后台页面
- 外部入口统一入池能力
- 基础站点配置表与后台页面
- Phase 2 需求文档
- Phase 2 设计文档
- Phase 2 验收文档

## 8. 进入下一阶段的前提是否满足

进入下一阶段需要的前提包括：

1. 前台主内容区已不依赖静态数据
2. 外部意向已能统一进入后台
3. 后台已具备基础运营跟进能力
4. 站点关键文案已开始配置化

当前判断：

- 已满足

## 9. 下一阶段建议

Phase 2 完成后，建议进入 Phase 3，优先推进：

1. 数据清理与运营稳定性优化
2. Leads 来源对象与运营动作体验增强
3. 站点配置范围扩展
4. 基础统计与运营看板
5. 更稳定的内容发布与运营流程

## 10. 一句话结论

Phase 2 已从“后台可维护”推进为“前台展示、外部转化、后台跟进都能真实跑起来的业务闭环阶段”，可以正式进入下一阶段。
