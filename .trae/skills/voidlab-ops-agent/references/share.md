# 分享说明

如果你要把这个 skill 发给别人，不要只发 `SKILL.md`。

应该把整个文件夹一起发：

- `voidlab-ops-agent/`

里面应至少包含：

- `SKILL.md`
- `references/overview.md`
- `references/articles.md`
- `references/events.md`
- `references/builders.md`
- `references/media.md`
- `references/publishing-flow.md`
- `references/error-handling.md`
- `references/share.md`

## 对方怎么安装

对方把整个文件夹放到他项目里的：

```text
.trae/skills/voidlab-ops-agent/
```

就可以作为一个完整 skill 被识别。

## 为什么这样更好

相比只发一个超长 `SKILL.md`，这种结构更适合：

- 长期维护
- 分工协作
- 单独补某一块运营规则
- 直接作为一个完整 skill 包分享
