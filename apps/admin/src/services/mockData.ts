import type {
  ArticleRecord,
  BuilderRecord,
  EventRecord,
  MediaAssetRecord
} from "../types";

export const articleMocks: ArticleRecord[] = [
  {
    id: 1,
    title: "AI Builder Network 正在形成新的协作方式",
    slug: "ai-builder-network-collaboration",
    summary: "用于后台骨架联调的文章示例。",
    category: "Research",
    audience: "Builders",
    tags: ["network", "ai"],
    coverUrl: "https://placehold.co/1200x630?text=Article",
    content: "Phase 1 文章编辑器示例内容。",
    sourceName: "VOIDLAB.AI",
    sourceUrl: "https://voidlab.ai",
    featured: true,
    status: "published",
    updatedAt: "2026-07-31 10:00"
  }
];

export const eventMocks: EventRecord[] = [
  {
    id: 1,
    title: "VOIDLAB Builders Night",
    slug: "voidlab-builders-night",
    summary: "用于后台骨架联调的活动示例。",
    city: "Shanghai",
    location: "Xuhui",
    eventType: "Meetup",
    eventTime: "2026-08-15 19:30",
    content: "Phase 1 活动编辑器示例内容。",
    coverUrl: "https://placehold.co/1200x630?text=Event",
    status: "draft",
    updatedAt: "2026-07-31 10:00"
  }
];

export const builderMocks: BuilderRecord[] = [
  {
    id: 1,
    name: "Alex Chen",
    slug: "alex-chen",
    title: "AI Product Builder",
    city: "Shanghai",
    role: "Founder",
    intro: "用于后台骨架联调的 Builder 示例。",
    story: "Phase 1 Builder 编辑器示例内容。",
    expertise: ["AI Product", "Growth"],
    focusAreas: ["Agent", "Workflow"],
    collaborationModes: ["Advisory", "Co-building"],
    contactable: true,
    featured: true,
    coverUrl: "https://placehold.co/1200x630?text=Builder",
    status: "published",
    updatedAt: "2026-07-31 10:00"
  }
];

export const mediaMocks: MediaAssetRecord[] = [
  {
    id: 1,
    fileName: "hero-cover.png",
    objectUrl: "https://placehold.co/1200x630?text=Media",
    contentType: "image/png",
    fileSizeLabel: "320 KB",
    createdAt: "2026-07-31 10:00"
  }
];
