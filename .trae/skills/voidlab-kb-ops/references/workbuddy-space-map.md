## WorkBuddy Space 目录映射（建议）

最终在 Knowledge Base 里建议用一个 Space（`workbuddy`），靠 `section_name` 分组来形成“4 个 section”的目录感：

- `导读与导航`
- `入门`
- `实战案例`
- `系统沉淀`

推荐把本地内容源目录这样组织（便于自动推导）：

- `content/knowledge/entries/workbuddy/nav/`
- `content/knowledge/entries/workbuddy/intro/`
- `content/knowledge/entries/workbuddy/practice/`
- `content/knowledge/entries/workbuddy/system/`

排序建议（`sort_order`）：

- 导读与导航：`0~99`
- 入门：`100~199`
- 实战案例：`200~299`
- 系统沉淀：`300~399`

文件名建议带两位数字前缀：`NN-title.md`，用于稳定排序与复盘时的“位置感”。

