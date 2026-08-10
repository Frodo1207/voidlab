---
name: "voidlab-kb-ops"
description: "Syncs local Markdown in content/knowledge to VOIDLAB Knowledge Base (Spaces/Entries/Assets) via admin API. Invoke when publishing/updating KB content or running a KB content sync."
---

# VOIDLAB Knowledge Ops

这个 Skill 用来“运营”VOIDLAB 的知识库内容：把仓库里的 Markdown 内容（`content/knowledge/**`）按规则同步到线上/本地的 Knowledge Base（Space / Entry / Asset），并确保目录分组（`section_name`）与排序（`sort_order`）稳定可控。

如果用户说类似这些话，就应当调用本 Skill：

- “把这些 Markdown 上传到知识库 / 同步到知识库 / 批量导入知识库”
- “给 WorkBuddy Space 建目录并先塞一些内容看看效果”
- “按 section 分组、按顺序排好、发布/撤回”

## 依赖与复用

本 Skill 与 `voidlab-ops-agent` 同属于“通过 API 做后台运营”的路线。优先复用其：

- URL 解析规则
- Agent Token 鉴权方式与 scope 检查
- 写操作前的握手与失败处理

当你需要更广泛的后台操作（文章/活动/Builder/媒体），先调用 `voidlab-ops-agent`；当重点是知识库内容的同步与编排，调用本 Skill。

## 必要环境变量

优先使用 Agent Token（可审计、可限权）：

- `VOIDLAB_API_BASE_URL`：例如 `http://127.0.0.1:18081`
- `VOIDLAB_AGENT_TOKEN`：从后台创建的 agent token（需包含 `knowledge:read` / `knowledge:write`，图片还需要 `media:write`）

如果用户没有配置环境变量，允许在当次对话里让用户直接提供：

- API Base URL
- Agent Token

除非用户明确要操作线上，否则本地默认用 `http://127.0.0.1:18081`。

## 内容仓库约定（Source of Truth）

默认内容源目录：

- `content/knowledge/entries/<space-slug>/**/*.md`

推荐约定：

1. 目录结构表达 section（分组），文件名表达排序
   - `entries/workbuddy/intro/01-xxx.md` → `section_name=入门`，`sort_order=101`
   - `entries/workbuddy/practice/11-xxx.md` → `section_name=实战案例`，`sort_order=211`
   - `entries/workbuddy/system/22-xxx.md` → `section_name=系统沉淀`，`sort_order=322`
   - `entries/workbuddy/nav/00-xxx.md` → `section_name=导读与导航`，`sort_order=0xx`

2. `slug` 必须稳定
   - 优先读取 frontmatter 的 `slug`
   - 若缺失：用文件名推导（去掉 `NN-` 前缀）生成 slug

3. `section_name` 的确定优先级
   - frontmatter `section_name`
   - 否则按目录映射：`nav/intro/practice/system`
   - 否则默认 `General`

> 注意：系统层面 Space → Entry 只有两层，但前端目录会按 Entry 的 `section_name` 分组显示。

## 同步流程（Upsert）

### 0. Preflight（写操作前必做）

1. `GET /healthz`
2. `GET /api/v1/auth/me`
3. `GET /api/v1/knowledge/spaces`（验证最小读权限）

任何一步失败都先停下来解决鉴权/URL/服务状态问题。

### 1. 确保 Space 存在

- `GET /api/v1/knowledge/spaces` 查 slug
- 不存在则 `POST /api/v1/knowledge/spaces` 创建（title/slug/status/visibility 等按用户要求填）

### 2. 获取 Space 下现有 Entry 映射

- `GET /api/v1/knowledge/entries?space_id=<id>`
- 构建 `slug -> entry` 映射，作为幂等 upsert 的依据

### 3. 将 Markdown 解析成可导入 payload

对每个 `*.md`：

1. 调用 `POST /api/v1/knowledge/entries/import-markdown`（multipart file）获得解析结果（title/slug/public_summary/content_markdown/…）
2. 合并 frontmatter 覆盖策略（以文件为真源），并按目录/文件名规则补齐：
   - `space_id`
   - `section_name`
   - `sort_order`
   - `status`（默认 `draft`，除非用户明确要直接发布）
   - `is_preview`（按 frontmatter 或默认 false）

### 4. Upsert 到系统

- 若 slug 已存在：`PUT /api/v1/knowledge/entries/:id`
- 若 slug 不存在：`POST /api/v1/knowledge/entries`

### 5. 图片资产（可选增强）

若 Markdown 中包含本地图片引用（例如 `![](./assets/x.png)`）：

1. 逐个上传到 `POST /api/v1/knowledge/spaces/:id/assets`
2. 用返回的 `knowledge-asset://<id>` 替换 Markdown 链接
3. 再次 `PUT /api/v1/knowledge/entries/:id` 更新正文

默认策略：先不自动做图片替换，除非用户明确说“把图片也一起上传并替换引用”。

## 安全策略

- 默认 `dry-run`（只生成计划与差异，不写入）除非用户说“直接同步/直接上传/直接发布”
- 发布（`status=published`）需要二次确认（或显式的 “全部发布” 指令）
- 不自动删除线上 Entry。删除必须显式指定（避免误删）

## 输出要求（每次同步的交付）

同步结束后输出：

- 目标 Space（slug / id）
- 新建了哪些 Entry
- 更新了哪些 Entry（标题/摘要/正文/排序/分组/状态）
- 失败项与下一步修复建议

