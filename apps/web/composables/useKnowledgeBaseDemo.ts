import { computed, ref } from "vue";
import { extractMarkdownExcerpt } from "../src/markdown";

export type KnowledgeVisibilityMode = "directory_only" | "public" | "private_hidden";
export type KnowledgeEntryStatus = "published" | "draft";

export interface KnowledgeEntry {
  id: number;
  spaceSlug: string;
  title: string;
  slug: string;
  sectionName: string;
  sortOrder: number;
  estimatedReadMinutes: number;
  publicSummary: string;
  contentMarkdown: string;
  status: KnowledgeEntryStatus;
  isPreview: boolean;
}

export interface KnowledgeSpace {
  id: number;
  title: string;
  slug: string;
  description: string;
  coverLabel: string;
  icon: string;
  themeTint: string;
  visibilityMode: KnowledgeVisibilityMode;
  entryCount: number;
  sectionCount: number;
  lastUpdatedAt: string;
  tokenHint: string;
  directorySummary: string;
  introMarkdown: string;
  demoToken: string;
}

const STORAGE_KEY = "voidlab-knowledge-unlocks";

const knowledgeEntries: KnowledgeEntry[] = [
  {
    id: 1,
    spaceSlug: "agent-builder-playbook",
    title: "01. 为什么先做 Agent 工作台，而不是先做大而全平台",
    slug: "why-agent-workbench-first",
    sectionName: "Part 1 / 基础判断",
    sortOrder: 1,
    estimatedReadMinutes: 8,
    publicSummary: "先确定 Agent 的实际工作边界，再反推后台结构，而不是一开始就做庞大的系统幻想。",
    status: "published",
    isPreview: true,
    contentMarkdown: `# 为什么先做 Agent 工作台

很多团队一开始谈 AI Agent，会立刻把问题描述成一个大平台：

- 要有自动内容生成
- 要有自动发布
- 要有工作流编排
- 要有权限体系

但在早期阶段，这种做法往往会带来一个问题：

> 你构建的是一套想象中的总系统，而不是一条已经跑通的工作链。

## 正确顺序

先确定一条真实高频动作：

1. 选题进入后台
2. 生成文章草稿
3. 人工复核
4. 发布

这条链路一旦跑通，你就拥有了真正的产品边界。

## 对 VOIDLAB 的意义

VOIDLAB 当前最适合从三类对象切入：

- 资讯
- 活动
- Builder 卡片

因为它们已经有明确的数据结构，也已经进入后台运营系统。`
  },
  {
    id: 2,
    spaceSlug: "agent-builder-playbook",
    title: "02. Agent Skill 不是功能集合，而是可审计的操作协议",
    slug: "skills-as-operating-protocol",
    sectionName: "Part 1 / 基础判断",
    sortOrder: 2,
    estimatedReadMinutes: 10,
    publicSummary: "Skill 的本质不是 API 文档，而是一套知道该如何判断、执行、回报的操作协议。",
    status: "published",
    isPreview: false,
    contentMarkdown: `# Skill 不是功能集合

很多人把 Skill 理解成：

\`发文章\`、\`更新内容\`、\`拉数据\`

但这些只是动作名，不是操作协议。

## 什么叫操作协议

一个真正能工作的 Skill，至少要回答：

1. 输入是什么
2. 缺什么信息时要问
3. 能默认什么
4. 失败时如何回报
5. 成功后如何写入审计

## 为什么这对知识库重要

知识库的前台访问不是“公开博客”，而是：

- 目录公开
- 内容受控
- 按 Space 解锁

所以后续 Agent 如果要运营知识库，就必须知道自己是在操作：

- Space
- Entry
- Token
- Access Log`
  },
  {
    id: 3,
    spaceSlug: "agent-builder-playbook",
    title: "03. 从文章、活动到知识库，后台模型怎么演进",
    slug: "content-model-evolution",
    sectionName: "Part 2 / 系统结构",
    sortOrder: 3,
    estimatedReadMinutes: 12,
    publicSummary: "知识库不是换个前端壳子的文章系统，它需要单独的 Space、Entry 和 Access Token 三层结构。",
    status: "published",
    isPreview: false,
    contentMarkdown: `# 内容模型演进

VOIDLAB 现有内容对象可以理解成三类：

| 对象 | 面向谁 | 是否公开 |
| --- | --- | --- |
| Article | 所有人 | 是 |
| Event | 社区成员 / 外部报名者 | 大部分公开 |
| Builder | 社区连接对象 | 大部分公开 |

知识库和这三者不同。

## 知识库的最小模型

### Space

一个专栏、专题或课程集合。

### Entry

Space 下的具体内容单元。

### Access Token

控制某个 Space 是否可读。

## 为什么不直接复用 Article

因为 Article 当前天然就是公开阅读模型。

而知识库的关键不是内容编辑本身，而是：

- 目录公开
- 内容受控
- 一次验证，整组解锁`
  },
  {
    id: 4,
    spaceSlug: "notion-style-founder-course",
    title: "01. 创始人为什么需要自己的知识操作系统",
    slug: "why-founders-need-knowledge-os",
    sectionName: "Module 1 / 方法底盘",
    sortOrder: 1,
    estimatedReadMinutes: 7,
    publicSummary: "不是为了记笔记，而是为了把决策、行动和复盘连成一个持续增长的系统。",
    status: "published",
    isPreview: true,
    contentMarkdown: `# 创始人为什么需要知识操作系统

创始人最容易掉进两个陷阱：

- 每天都在接触信息，但没有结构
- 每次都在做判断，但没有复用

## 真正的问题

问题不是你知道得不够多，而是：

> 你没有把已知的东西沉淀成可复用结构。

## 一套好的知识系统要做到

1. 记录来源
2. 记录判断
3. 记录行动
4. 记录结果

这样它才不是笔记，而是经营自己的操作系统。`
  },
  {
    id: 5,
    spaceSlug: "notion-style-founder-course",
    title: "02. 如何把一堆博客和教程重组为课程目录",
    slug: "restructure-blogs-into-course",
    sectionName: "Module 2 / 内容编排",
    sortOrder: 2,
    estimatedReadMinutes: 11,
    publicSummary: "从单篇内容思维切换到 Space 编排思维：入口、目录、章节、练习和升级路径。",
    status: "published",
    isPreview: false,
    contentMarkdown: `# 从博客到课程目录

一个博客列表，只是在堆内容。

一个知识 Space，才是在组织学习路径。

## 最小目录结构

1. 你先回答：这套内容想把人带到哪里
2. 再回答：用户进入时最先缺什么
3. 然后才决定每一篇放在哪里

## 一个简单的 Notion 式编排模板

### Start here

- 为什么学
- 学完能解决什么
- 如何使用这套内容

### Core modules

- 基础概念
- 方法框架
- 示例拆解

### Apply

- 作业
- 模板
- 常见错误`
  }
];

const knowledgeSpaces: KnowledgeSpace[] = [
  {
    id: 1,
    title: "Agent Builder Playbook",
    slug: "agent-builder-playbook",
    description: "把 Agent 工作台、后台运营和内容系统连接起来的一套实战知识空间。",
    coverLabel: "OPERATING PROTOCOL",
    icon: "◫",
    themeTint: "from-[#dbeafe] via-[#eff6ff] to-white",
    visibilityMode: "directory_only",
    entryCount: 3,
    sectionCount: 2,
    lastUpdatedAt: "2026-08-02",
    tokenHint: "输入访问令牌即可解锁整个 Space",
    directorySummary: "目录公开、正文受控，适合做课程、专栏、内训资料或会员知识资产。",
    demoToken: "VOIDLAB-AGENT-2026",
    introMarkdown: `# Agent Builder Playbook

这个 Space 聚焦一个核心问题：

**如何把 Agent 从一个演示概念，收成能进入日常运营系统的真实执行体。**

你会看到的不是零散博客，而是一条从：

- 工作台判断
- Skill 协议设计
- 后台结构演进

逐层推进的知识路径。`
  },
  {
    id: 2,
    title: "Founder Knowledge OS",
    slug: "notion-style-founder-course",
    description: "面向创始人和独立 builder 的知识操作系统课程原型。",
    coverLabel: "KNOWLEDGE OS",
    icon: "✦",
    themeTint: "from-[#fef3c7] via-[#fff8eb] to-white",
    visibilityMode: "directory_only",
    entryCount: 2,
    sectionCount: 2,
    lastUpdatedAt: "2026-08-01",
    tokenHint: "目录公开，输入令牌后可阅读全部章节",
    directorySummary: "更偏课程和方法论，适合会员区、训练营、顾问交付和体系化内容产品。",
    demoToken: "VOIDLAB-FOUNDER-2026",
    introMarkdown: `# Founder Knowledge OS

这不是普通博客合集，而是一个“知识如何服务经营动作”的课程型空间。

它更像你在 Notion 里整理好的：

- 首页说明
- 章节目录
- 分模块内容
- 统一阅读体验`
  }
];

const unlockState = ref<Record<string, boolean>>({});
let unlocksLoaded = false;

function ensureUnlockStateLoaded() {
  if (unlocksLoaded || typeof window === "undefined") {
    return;
  }

  unlocksLoaded = true;

  try {
    const rawValue = window.sessionStorage.getItem(STORAGE_KEY);
    unlockState.value = rawValue ? JSON.parse(rawValue) as Record<string, boolean> : {};
  } catch {
    unlockState.value = {};
  }
}

function persistUnlockState() {
  if (typeof window === "undefined") {
    return;
  }

  window.sessionStorage.setItem(STORAGE_KEY, JSON.stringify(unlockState.value));
}

function normalizeToken(token: string) {
  return token.trim().toUpperCase();
}

export function useKnowledgeBaseDemo() {
  ensureUnlockStateLoaded();

  const spaces = computed(() => knowledgeSpaces);
  const entries = computed(() =>
    knowledgeEntries
      .filter((item) => item.status === "published")
      .map((item) => ({
        ...item,
        previewText: item.publicSummary || extractMarkdownExcerpt(item.contentMarkdown, 180)
      }))
  );

  function getSpaceBySlug(spaceSlug: string) {
    return knowledgeSpaces.find((space) => space.slug === spaceSlug) ?? null;
  }

  function getEntriesBySpace(spaceSlug: string) {
    return entries.value
      .filter((entry) => entry.spaceSlug === spaceSlug)
      .sort((left, right) => left.sortOrder - right.sortOrder);
  }

  function getEntryBySlug(spaceSlug: string, entrySlug: string) {
    return entries.value.find((entry) => entry.spaceSlug === spaceSlug && entry.slug === entrySlug) ?? null;
  }

  function isSpaceUnlocked(spaceSlug: string) {
    return Boolean(unlockState.value[spaceSlug]);
  }

  function canReadEntry(spaceSlug: string, entrySlug: string) {
    const entry = getEntryBySlug(spaceSlug, entrySlug);
    if (!entry) {
      return false;
    }

    return entry.isPreview || isSpaceUnlocked(spaceSlug);
  }

  function unlockSpace(spaceSlug: string, token: string) {
    const space = getSpaceBySlug(spaceSlug);
    if (!space) {
      return { success: false, message: "知识空间不存在" };
    }

    if (normalizeToken(token) !== normalizeToken(space.demoToken)) {
      return { success: false, message: "令牌不正确，请重新输入" };
    }

    unlockState.value = {
      ...unlockState.value,
      [spaceSlug]: true
    };
    persistUnlockState();

    return { success: true, message: "知识空间已解锁" };
  }

  function lockSpace(spaceSlug: string) {
    if (!unlockState.value[spaceSlug]) {
      return;
    }

    const nextUnlocks = { ...unlockState.value };
    delete nextUnlocks[spaceSlug];
    unlockState.value = nextUnlocks;
    persistUnlockState();
  }

  function getSpaceStats(spaceSlug: string) {
    const spaceEntries = getEntriesBySpace(spaceSlug);
    const previewCount = spaceEntries.filter((entry) => entry.isPreview).length;
    const totalMinutes = spaceEntries.reduce((sum, entry) => sum + entry.estimatedReadMinutes, 0);

    return {
      entryCount: spaceEntries.length,
      previewCount,
      totalMinutes
    };
  }

  return {
    spaces,
    entries,
    getSpaceBySlug,
    getEntriesBySpace,
    getEntryBySlug,
    isSpaceUnlocked,
    canReadEntry,
    unlockSpace,
    lockSpace,
    getSpaceStats
  };
}
