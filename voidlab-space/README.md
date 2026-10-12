# voidlab-space

这是一个独立的内容源空间。

它做两件事：

- 存放知识库 Markdown 源文件
- 本地实时预览这些内容在 VOIDLAB 风格下的最终样子

## 目录

```text
voidlab-space/
  content/
    entries/      可预览的正式内容
    drafts/       草稿
    sources/      原始素材
    assets/       本地图片等资源
    templates/    写作模板
  project-management/
    skills/       从 .trae 镜像出来的项目管理 / 运营 skill
  src/            本地预览壳
```

## 启动

```bash
cd voidlab-space
npm install
npm run dev
```

默认地址：

- `http://127.0.0.1:4174/`

## 写作约定

- 路径：`content/entries/<space-slug>/<section>/<file>.md`
- 例如：`content/entries/fde/intro/00-what-is-fde.md`
- 标题优先取 frontmatter 的 `title`
- slug 优先取 frontmatter 的 `slug`

## 下一步

- 先把你想迁出的内容慢慢移到 `content/entries`
- 确认预览样式满意后，再把“同步到正式知识库”的脚本接进来

## 项目管理 Skill 副本

为了让这套内容后面可以独立做项目管理，不继续直接依赖 `.trae`，当前已经把 `voidlab-ops-agent` 完整镜像到：

- `voidlab-space/project-management/skills/voidlab-ops-agent`

这个目录保留了：

- `SKILL.md`
- `references/*.md`
