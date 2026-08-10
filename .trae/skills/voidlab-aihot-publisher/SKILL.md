---
name: "voidlab-aihot-publisher"
description: "Publishes VOIDLAB AI news from AI HOT with cover generation and article push. Invoke when creating, covering, or publishing AI HOT based articles in VOIDLAB."
---

# VOIDLAB AIHOT Publisher

这个 skill 用来把 `AI HOT` 里的资讯快速变成 VOIDLAB 站内文章，并且补上封面图、回填封面地址、再发布到线上。

适用场景：

- 用户说“用 AI HOT 发一篇资讯”
- 用户说“抓一条 AI 新闻发到站里”
- 用户说“给这篇资讯补个封面并发布”
- 用户说“批量同步 AI HOT 精选”

不适用场景：

- 普通网页内容采编但不是来自 `AI HOT`
- 长篇深度报告
- 只做封面设计，不发文章
- 只修现有前端页面

## 目标

这个 skill 的目标不是只“新增一篇文章”，而是标准化整条资讯发布链路：

- 从 `AI HOT` 选择合适资讯
- 整理为站内文章草稿
- 生成或接入封面图
- 把封面回填到文章
- 将文章发布
- 在遇到媒体上传异常时自动走回退路径

## 固定依赖

执行时优先复用这些能力与文件：

- 内置 `aihot` skill
- 项目内 `voidlab-ops-agent` skill
- 项目内 `deploy` 与线上 VOIDLAB API
- 项目内 `assets/covers/`

如果当前项目空间存在这些 skill，应直接复用，不要另起一套同义流程。

## 输入要求

执行前优先收集这些信息：

- 目标资讯来源：默认 `AI HOT`
- 文章目标环境 URL
- Agent Token
- 是只建草稿，还是直接发布
- 封面图来源：用户上传 / 网图 / AI 生成

如果用户没有明确给 URL，按下面顺序推断：

1. 用户本轮明确给出的 URL
2. 环境变量中的稳定 URL
3. 当前项目默认线上地址：`http://142.248.136.161`

如果用户没有明确说要草稿还是发布，默认先建草稿。  
如果用户说“发布吧”“直接发”，则直接发布。

## 标准流程

执行时按下面顺序推进。

### 阶段 1：选题

1. 使用 `aihot` 获取最近 24 小时或最近一周的精选资讯
2. 优先选：
   - 模型发布
   - 平台能力升级
   - 对开发者或 Agent 工作流有价值的内容
3. 避免直接发：
   - 只有情绪没有信息密度的热帖
   - 无法核清核心事实的内容
   - 与 VOIDLAB 站点主题明显不匹配的内容

默认文章结构：

- 标题
- 摘要
- 事件摘要
- 关键信息
- 为什么值得关注
- 来源

现在这个项目的资讯正文默认应按 `Markdown` 文档来组织，而不是纯富文本段落拼接。

推荐正文骨架：

```md
## 事件摘要

一句到两句讲清这件事发生了什么。

## 关键信息

- 事实 1
- 事实 2
- 事实 3

## 为什么值得关注

1. 对开发者的影响
2. 对产品或 Agent 工作流的影响
3. 对行业格局的影响

## 来源

- AI HOT 精选
- 原始出处
```

### 阶段 2：连接 VOIDLAB API

写文章前必须做最小握手：

1. `GET /api/v1/auth/me`
2. `GET /api/v1/articles`

只有在这两步通过后，才允许继续写文章。

如果失败，按下面解释：

- `401`：token 无效或已停用
- `403`：token 权限不够
- `5xx`：服务本身异常，先不要继续写入

### 阶段 3：创建文章

默认创建文章草稿，除非用户明确要直接发布。

标题规则：

- 默认不要在标题前添加 `AI HOT｜`
- 标题直接使用整理后的事件标题本身
- `AI HOT` 作为来源信息保留在：
  - `tags`
  - `source_name`
  - 正文 `## 来源`

文章 payload 至少包含：

- `title`
- `slug`
- `summary`
- `category`
- `audience`
- `tags`
- `content`
- `source_name`
- `source_url`
- `status`

`content` 字段现在默认应写入 Markdown 文档正文。  
如果用户直接提供 `.md` 文本内容，优先原样保留其标题层级、列表和强调语法，再按现有 payload 写入 `content`。

推荐约定：

- `category`：优先按 AI HOT 原始主题，如 `ai-models`
- `audience`：默认 `builders`
- `tags`：包含 `AI HOT` 和主题词
- `featured`：默认 `false`

如果用户没有给 slug，就自动转成 kebab-case。

### 阶段 4：封面图

如果用户说文章需要封面，则必须执行这一层。

封面来源优先级：

1. 用户提供图片
2. AI 生成
3. 用户明确同意后再去找图

如果用户没有特别要求，默认生成一张适合资讯封面的横版图。

封面图要求：

- 不带标题文字
- 能作为文章卡片封面
- 风格统一、不要低质 AI 海报感
- 优先深色科技感、编辑感、产品感

生成完成后，先把图片保存到项目目录：

- `assets/covers/<slug>-cover.jpg`

### 阶段 5：媒体上传与回退

优先尝试通过媒体接口上传：

- `POST /api/v1/media/upload`

如果上传成功：

- 使用返回的 `object_url` 或可访问路径作为文章 `cover_url`

如果上传失败，不要直接放弃。  
要先判断是否是已知的表结构问题。

当前项目已知问题：

- `media_assets` 可能缺少 `file_size` 字段
- 报错示例：`table media_assets has no column named file_size`

如果遇到这个问题，使用回退路径：

1. 直接把图片上传到服务器 `uploads` 目录
2. 确认公网地址可访问
3. 将文件名回填到文章 `cover_url`

当前项目回退写法示例：

- 文件名：`deepseek-v4-flash-0731-cover.jpg`
- 对外访问：`/uploads/deepseek-v4-flash-0731-cover.jpg`
- 文章 `cover_url`：直接写文件名或现有后端兼容的路径值

### 阶段 6：发布

如果当前文章还是草稿，且用户明确要发布，则调用：

- `PUT /api/v1/articles/:id/status`

状态改为：

- `published`

如果用户只说“先建草稿”，则停在 `draft`。

## Markdown 资讯规则

当前项目的官网前端已经支持 Markdown 资讯详情渲染，因此这个 skill 在写资讯时应默认遵循下面规则：

1. `content` 优先用 Markdown
2. 一级标题尽量不用，正文从 `##` 开始
3. 列表优先用 `-` 或 `1.`
4. 不要把整篇内容塞成单段纯文本
5. 如果来源本身就是 Markdown 文档，尽量少改结构，只做必要清洗

适合保留的 Markdown 元素：

- `##`、`###`
- 无序列表
- 有序列表
- 粗体强调
- 普通链接

当前不建议依赖的高级写法：

- HTML 片段混排
- 很复杂的表格
- 代码块作为资讯主体

## 输出要求

每次执行完后，输出尽量短，但要包含：

- 执行动作
- 文章 `id`
- `slug`
- `status`
- 是否已设置封面
- 如果走了回退路径，要明确说明

示例：

- `Created article from AI HOT`
- `id: 3`
- `slug: aihot-deepseek-v4-flash-0731-open-weights`
- `status: published`
- `cover: deepseek-v4-flash-0731-cover.jpg`
- `media upload fallback used: yes`

## 常见故障

### 1. token 可读但不可写

表现：

- `auth/me` 成功
- `GET /api/v1/articles` 成功
- `POST /api/v1/articles` 或状态变更返回 `403`

处理：

- 告诉用户 token 缺少 `articles:write`

### 2. 媒体上传失败

表现：

- `POST /api/v1/media/upload` 返回 `5302`
- 或数据库字段错误

处理：

- 直接切回 `uploads` 目录上传
- 继续回填文章封面，不阻断发稿

### 3. 页面能开但文章图片不显示

处理顺序：

1. 检查 `/uploads/<file>` 是否返回 `200`
2. 检查文章 `cover_url` 是否和后端前端约定一致
3. 再检查前端渲染逻辑

### 4. 已发布但内容未更新

处理：

- 先 `GET /api/v1/articles/:id`
- 确认当前状态和 `updated_at`
- 如果要改正文，必须走完整 `PUT /api/v1/articles/:id`
- 不要只传局部字段

## 推荐执行风格

- 先做连接检查，再写入
- 优先给出可发布结果，不在中途卡死
- 对已知后端问题直接走回退路径
- 不要把“媒体上传失败”等同于“文章无法发布”

## 当前项目的真实经验

这个项目当前已经验证过下面这条链路：

1. 用 `AI HOT` 选资讯
2. 用 Agent Token 创建文章
3. 用 AI 生成封面图
4. 媒体上传失败时走 `uploads` 回退
5. 回填 `cover_url`
6. 成功发布文章

所以这个 skill 应该默认优先沿用这条已验证链路，而不是在执行时重新摸索。
