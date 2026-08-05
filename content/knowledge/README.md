## Knowledge 内容仓库（Markdown）

这个目录用于存放「准备导入 VOIDLAB Knowledge Base 的源内容」：

- 每篇条目建议用一个独立的 `*.md` 文件保存（便于版本管理、协作编辑、审阅与回滚）。
- 文件可以带 YAML frontmatter（`--- ... ---`），后端的 `ParseMarkdownImport` 会读取并自动解析出标题、slug、摘要等字段。

### 推荐结构

- `entries/`：可直接导入的条目 Markdown（最终稿 / 待导入）
- `drafts/`：草稿与拆解（不一定能导入）
- `sources/`：原始材料（链接、摘录、截图说明等）
- `templates/`：写作与导入模板

### Frontmatter 字段（常用）

示例：

```yaml
---
title: "WorkBuddy 蓝皮书：目录与阅读路线"
slug: "workbuddy-bluebook-reading-guide"
section_name: "WorkBuddy"
public_summary: "用真实任务为主线的 27 章蓝皮书，总览 + 推荐阅读路线。"
cover_url: ""
is_preview: false
status: "published"
---
```

可用字段（大小写不敏感，会转小写读取）：

- `title`
- `slug`
- `section_name`（或 `section`）
- `public_summary`（或 `summary`）
- `estimated_read_minutes`（或 `read_minutes`）
- `cover_url`（或 `cover`）
- `is_preview`
- `status`（会被规范化）

### 文件命名建议

尽量使用：`YYYYMMDD-topic-slug.md` 或 `topic-slug.md`，避免中文空格与特殊符号。

