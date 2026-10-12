---
title: "权限模型与工作区：默认权限 vs 完全访问"
slug: "workbuddy-permissions-and-workspace"
section_name: "入门"
public_summary: "先把边界设对：工作区是内容边界，权限是行为边界。默认权限可控，完全访问高效但高风险。"
estimated_read_minutes: 8
status: "published"
---

这件事的本质：**你给 AI 的不是“能力”，而是“可操作的边界”**。边界没设好，你会得到两种“看起来效率很高但后患很大”的结果：一种是文件被误改、误删；另一种是内容被串用，比如把 A 客户的材料写进 B 客户的文档里。

![工作区与权限的关系](data:image/svg+xml;utf8,%3Csvg%20xmlns%3D%27http%3A//www.w3.org/2000/svg%27%20width%3D%27960%27%20height%3D%27420%27%20viewBox%3D%270%200%20960%20420%27%3E%3Cdefs%3E%3ClinearGradient%20id%3D%27g%27%20x1%3D%270%27%20y1%3D%270%27%20x2%3D%271%27%20y2%3D%271%27%3E%3Cstop%20offset%3D%270%27%20stop-color%3D%27%23ffffff%27/%3E%3Cstop%20offset%3D%271%27%20stop-color%3D%27%23f7f7f5%27/%3E%3C/linearGradient%3E%3C/defs%3E%3Crect%20x%3D%2716%27%20y%3D%2716%27%20width%3D%27928%27%20height%3D%27388%27%20rx%3D%2718%27%20fill%3D%27url(%23g)%27%20stroke%3D%27%23e9e9e7%27/%3E%3Ctext%20x%3D%2748%27%20y%3D%2768%27%20font-family%3D%27Arial%2C%20PingFang%20SC%2C%20Microsoft%20YaHei%27%20font-size%3D%2726%27%20font-weight%3D%27900%27%20fill%3D%27%23111111%27%3E%E8%BE%B9%E7%95%8C%E6%A8%A1%E5%9E%8B%3C/text%3E%3Ctext%20x%3D%2748%27%20y%3D%2798%27%20font-family%3D%27Arial%2C%20PingFang%20SC%2C%20Microsoft%20YaHei%27%20font-size%3D%2714%27%20fill%3D%27%235f5e58%27%3E%E5%B7%A5%E4%BD%9C%E5%8C%BA%E5%86%B3%E5%AE%9A%E2%80%9C%E8%83%BD%E7%9C%8B%E8%A7%81%E4%BB%80%E4%B9%88%E6%96%87%E4%BB%B6%E2%80%9D%EF%BC%8C%E6%9D%83%E9%99%90%E5%86%B3%E5%AE%9A%E2%80%9C%E5%8F%AF%E4%BB%A5%E4%B8%8D%E7%BB%8F%E7%A1%AE%E8%AE%A4%E5%B0%B1%E5%81%9A%E4%BB%80%E4%B9%88%E2%80%9D%3C/text%3E%3Crect%20x%3D%2748%27%20y%3D%27130%27%20width%3D%27404%27%20height%3D%27240%27%20rx%3D%2716%27%20fill%3D%27%23ffffff%27%20stroke%3D%27%23e9e9e7%27/%3E%3Crect%20x%3D%2748%27%20y%3D%27130%27%20width%3D%276%27%20height%3D%27240%27%20fill%3D%27%23c4f000%27/%3E%3Ctext%20x%3D%2780%27%20y%3D%27178%27%20font-family%3D%27Arial%2C%20PingFang%20SC%2C%20Microsoft%20YaHei%27%20font-size%3D%2720%27%20font-weight%3D%27850%27%20fill%3D%27%23222222%27%3E%E5%B7%A5%E4%BD%9C%E5%8C%BA%EF%BC%88Workspace%EF%BC%89%3C/text%3E%3Ctext%20x%3D%2780%27%20y%3D%27206%27%20font-family%3D%27Arial%2C%20PingFang%20SC%2C%20Microsoft%20YaHei%27%20font-size%3D%2714%27%20fill%3D%27%235f5e58%27%3E%E5%86%85%E5%AE%B9%E8%BE%B9%E7%95%8C%EF%BC%9A%E4%B8%80%E4%B8%AA%E4%BB%BB%E5%8A%A1%E8%83%BD%E8%AF%BB%E5%86%99%E5%93%AA%E4%B8%AA%E7%9B%AE%E5%BD%95%EF%BC%9F%3C/text%3E%3Ctext%20x%3D%2780%27%20y%3D%27240%27%20font-family%3D%27Arial%2C%20PingFang%20SC%2C%20Microsoft%20YaHei%27%20font-size%3D%2714%27%20fill%3D%27%235f5e58%27%3E%E9%9A%94%E7%A6%BB%E7%9B%AE%E5%BD%95%20%E2%86%92%20%E9%99%8D%E4%BD%8E%E8%AF%AF%E8%AF%BB%E3%80%81%E8%AF%AF%E6%94%B9%E3%80%81%E4%B8%B2%E7%94%A8%E9%A3%8E%E9%99%A9%3C/text%3E%3Crect%20x%3D%27508%27%20y%3D%27130%27%20width%3D%27404%27%20height%3D%27240%27%20rx%3D%2716%27%20fill%3D%27%23ffffff%27%20stroke%3D%27%23e9e9e7%27/%3E%3Crect%20x%3D%27508%27%20y%3D%27130%27%20width%3D%276%27%20height%3D%27240%27%20fill%3D%27%230f7b6c%27/%3E%3Ctext%20x%3D%27540%27%20y%3D%27178%27%20font-family%3D%27Arial%2C%20PingFang%20SC%2C%20Microsoft%20YaHei%27%20font-size%3D%2720%27%20font-weight%3D%27850%27%20fill%3D%27%23222222%27%3E%E6%9D%83%E9%99%90%EF%BC%88Permissions%EF%BC%89%3C/text%3E%3Ctext%20x%3D%27540%27%20y%3D%27206%27%20font-family%3D%27Arial%2C%20PingFang%20SC%2C%20Microsoft%20YaHei%27%20font-size%3D%2714%27%20fill%3D%27%235f5e58%27%3E%E8%A1%8C%E4%B8%BA%E8%BE%B9%E7%95%8C%EF%BC%9A%E6%93%8D%E4%BD%9C%E5%89%8D%E8%A6%81%E4%B8%8D%E8%A6%81%E7%A1%AE%E8%AE%A4%EF%BC%9F%3C/text%3E%3Ctext%20x%3D%27540%27%20y%3D%27240%27%20font-family%3D%27Arial%2C%20PingFang%20SC%2C%20Microsoft%20YaHei%27%20font-size%3D%2714%27%20fill%3D%27%235f5e58%27%3E%E9%BB%98%E8%AE%A4%E6%9D%83%E9%99%90%20%E2%86%92%20%E6%9C%89%E4%B8%8D%E7%A1%AE%E5%AE%9A%E5%B0%B1%E9%97%AE%E4%BD%A0%EF%BC%9B%E5%AE%8C%E5%85%A8%E8%AE%BF%E9%97%AE%20%E2%86%92%20%E6%9B%B4%E5%BF%AB%E4%BD%86%E9%9C%80%E8%87%AA%E6%8B%85%E9%A3%8E%E9%99%A9%3C/text%3E%3C/svg%3E)

## 工作区（Workspace）是内容边界

WorkBuddy 的任务通常不是“生成一段话”，而是要读写文件、生成产物、整理目录。官方在主界面与工作区章节里明确强调：工作目录既是效率设置，也是安全边界，建议按任务建目录空间，避免把发票、周报、客户材料混在同一个大目录里导致误读、误改和信息串用。[$TRAE_REF](https://workbuddy.homes/bluebook/%E7%AC%AC%E4%B8%80%E7%AF%87%20%E4%BD%BF%E7%94%A8%E6%89%8B%E5%86%8C%EF%BC%9A%E5%85%88%E6%8A%8A%20WorkBuddy%20%E7%94%A8%E8%B5%B7%E6%9D%A5/%E7%AC%AC%203%20%E7%AB%A0%20WorkBuddy%20%E7%9A%84%E4%B8%BB%E7%95%8C%E9%9D%A2%E3%80%81%E4%BB%BB%E5%8A%A1%E4%B8%8E%E5%B7%A5%E4%BD%9C%E5%8C%BA/)

你可以把工作区理解成“这个任务允许触达的文件集合”，它的价值不只在安全，也在效率：让 AI 不需要每次都猜你要读哪个文件夹。

建议把工作区当成“每个任务的工作台面”，而不是“电脑里最大的资料仓库”：

- 给每类任务一个固定目录：写文章、做表格、做口播稿不要混在一起
- 给每个客户或项目一个独立目录：避免跨项目串用材料
- 给每次交付留出一个明确的 `output/`：让产物落点可预测、可检查

一个最小可用的目录结构可以长这样：

```txt
WorkBuddy/
  00-inbox/            # 临时放入的原始材料（只进不出，避免被误改）
  10-projects/
    client-a/
      input/
      output/
      notes/
  90-scratch/          # 试错区：临时脚本、草稿、一次性产物
```

如果你只记一条规则：**重要资料不要直接在“总目录”里跑任务。先把要处理的材料复制到任务目录，再让 WorkBuddy 在这个目录内操作。**

## 权限（Permissions）是行为边界

权限不是“能不能做”，而是“做之前要不要确认”。它决定了 WorkBuddy 在执行过程中，是“遇到不确定就停下来问你”，还是“按它的理解直接继续做下去”。

为了把这件事说清楚，可以把权限简单拆成两档（你现在的标题也正是这个逻辑）：

| 维度 | 默认权限 | 完全访问 |
|---|---|---|
| 交互方式 | 关键改动前更倾向于询问确认 | 更倾向于自动继续执行 |
| 安全性 | 更稳、更适合新流程 | 更快，但更依赖你的边界设置 |
| 适合场景 | 新手上手、重要资料、跨目录任务 | 成熟 SOP、重复性任务、可容错场景 |
| 主要风险 | 频繁确认会打断节奏 | 误改误删、跨目录读写、内容串用 |

官方对“完全访问”有一个很关键的提醒：开启“允许完全访问”后，智能体可能读写授权目录外文件，因此需要谨慎使用，并优先按任务限定目录。[$TRAE_REF](https://workbuddy.homes/bluebook/%E7%AC%AC%E4%B8%80%E7%AF%87%20%E4%BD%BF%E7%94%A8%E6%89%8B%E5%86%8C%EF%BC%9A%E5%85%88%E6%8A%8A%20WorkBuddy%20%E7%94%A8%E8%B5%B7%E6%9D%A5/%E7%AC%AC%203%20%E7%AB%A0%20WorkBuddy%20%E7%9A%84%E4%B8%BB%E7%95%8C%E9%9D%A2%E3%80%81%E4%BB%BB%E5%8A%A1%E4%B8%8E%E5%B7%A5%E4%BD%9C%E5%8C%BA/)

这句话翻译成人话就是：**完全访问不是“更聪明”，只是“更少打断你”。如果边界没设好，它也会更快地把事情做错。**

![从目标到交付的任务链路示意](http://142.248.136.161/uploads/20260803081219-8d31621c-5589-4d99-b011-ed8fbac4e928.jpg)

## 怎么选：一个简单的决策

你可以按下面三条去选，不需要背复杂规则：

- 这次任务会不会碰重要文件：会就用默认权限
- 你是否能清楚指出输出位置和文件名：不清楚就用默认权限
- 你是否已经用同样方式跑通过 2~3 次：没跑通前不要开完全访问

当你非常确信“它接下来要做什么”时，完全访问可以让节奏更连贯。反过来，只要你自己都说不清它该怎么做，就不要指望它在完全访问下替你兜底。

## 最小建议

这里给的是“不会后悔”的最小策略，不追求极致效率，但能把风险压到可控范围：

- 新项目或新流程：先用默认权限跑通 3 次，确认不会误删、误改、误把产物写到奇怪的目录
- 重要资料：先备份，再自动化；最稳的备份方式是“副本 + 版本管理”，而不是只相信回收站
- 永远先设工作区，再谈权限：工作区没设好，权限越高越危险
- 给产物固定落点：比如统一输出到 `output/`，并要求它每次完成后列出“改动清单”

## 常见坑与排查

下面这些问题，基本都能用“工作区有没有设对、权限有没有开太高”来解释：

- 文件读不到或写不进去：先确认任务的工作区目录是不是你以为的那个目录
- 产物找不到：多数是写到默认目录或历史任务目录里了，让它在结果里列出完整路径再去找
- 内容串用：检查你是不是把多个客户/项目的材料放进同一工作区里
- 误删误改：第一时间先停用完全访问，把任务切回默认权限，再让它复盘“刚才改了哪些文件”
