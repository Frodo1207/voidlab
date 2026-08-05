import { computed, ref } from "vue";
import { resolvePublicApiPath, resolveUploadsUrl } from "../src/runtimeConfig";
import { extractMarkdownExcerpt } from "../src/markdown";

export type InsightCategory = string;
export type InsightAudience = string;

type ApiEnvelope<T> = {
  code: number;
  message: string;
  data: T;
};

type PublicArticleRecord = {
  id: number;
  title: string;
  slug: string;
  summary: string;
  category: string;
  audience: string;
  tags: string[];
  cover_url: string;
  content: string;
  source_name: string;
  source_url: string;
  featured: boolean;
  status: string;
  updated_at: string;
};

export interface Insight {
  id: number;
  slug: string;
  title: string;
  category: string;
  audience: string;
  publishedAt: string;
  rawPublishedAt: string;
  summary: string;
  whyItMatters: string;
  sourceName: string;
  sourceUrl: string;
  tags: string[];
  content: string[];
  rawContent: string;
  featured: boolean;
  cover: string;
  isNews: boolean;
}

const insightsState = ref<Insight[]>([]);
const loadingState = ref(false);
const loadedState = ref(false);
const errorState = ref("");

let listPromise: Promise<Insight[]> | null = null;

function normalizeCoverUrl(value: string) {
  const cover = value.trim();
  if (!cover) {
    return "";
  }

  if (
    cover.startsWith("http://") ||
    cover.startsWith("https://") ||
    cover.startsWith("data:")
  ) {
    return cover;
  }

  return resolveUploadsUrl(cover);
}

function splitContent(value: string) {
  return value
    .split(/\n{2,}|\r?\n/)
    .map((item) => item.trim())
    .filter(Boolean);
}

function formatPublishedAt(rawValue: string) {
  const normalized = rawValue.trim();
  if (!normalized) {
    return "";
  }

  const date = new Date(normalized.includes("T") ? normalized : normalized.replace(" ", "T"));
  if (Number.isNaN(date.getTime())) {
    return normalized;
  }

  const month = String(date.getMonth() + 1).padStart(2, "0");
  const day = String(date.getDate()).padStart(2, "0");
  return `${month}.${day}`;
}

function mapInsight(record: PublicArticleRecord): Insight {
  const rawContent = (record.content || "").trim();
  const content = splitContent(rawContent);
  const summary = record.summary || "";
  const excerpt = extractMarkdownExcerpt(rawContent, 220);

  return {
    id: record.id,
    slug: record.slug,
    title: record.title,
    category: record.category || "行业观察",
    audience: record.audience || "个人",
    publishedAt: formatPublishedAt(record.updated_at),
    rawPublishedAt: record.updated_at || "",
    summary: summary || excerpt,
    whyItMatters: summary || excerpt,
    sourceName: record.source_name || "VOIDLAB",
    sourceUrl: record.source_url || "#",
    tags: record.tags || [],
    content,
    rawContent,
    featured: Boolean(record.featured),
    cover: normalizeCoverUrl(record.cover_url),
    isNews: record.category === "快讯"
  };
}

function upsertInsight(nextInsight: Insight) {
  const nextInsights = [...insightsState.value];
  const index = nextInsights.findIndex((item) => item.slug === nextInsight.slug);

  if (index >= 0) {
    nextInsights[index] = nextInsight;
  } else {
    nextInsights.unshift(nextInsight);
  }

  insightsState.value = nextInsights;
}

async function requestPublicApi<T>(path: string) {
  const response = await fetch(resolvePublicApiPath(path));
  const envelope = (await response.json()) as ApiEnvelope<T>;

  if (!response.ok || envelope.code !== 0) {
    throw new Error(envelope.message || "请求失败");
  }

  return envelope.data;
}

export async function loadInsights(force = false) {
  if (loadedState.value && !force) {
    return insightsState.value;
  }

  if (listPromise && !force) {
    return listPromise;
  }

  loadingState.value = true;
  errorState.value = "";

  listPromise = requestPublicApi<PublicArticleRecord[]>("/api/v1/public/articles")
    .then((records) => {
      insightsState.value = records.map(mapInsight);
      loadedState.value = true;
      return insightsState.value;
    })
    .catch((error: unknown) => {
      errorState.value = error instanceof Error ? error.message : "加载资讯失败";
      throw error;
    })
    .finally(() => {
      loadingState.value = false;
      listPromise = null;
    });

  return listPromise;
}

export async function loadInsightBySlug(slug: string) {
  const normalizedSlug = slug.trim();
  if (!normalizedSlug) {
    return null;
  }

  const existing = insightsState.value.find((item) => item.slug === normalizedSlug);
  if (existing) {
    return existing;
  }

  const record = await requestPublicApi<PublicArticleRecord>(`/api/v1/public/articles/${normalizedSlug}`);
  const insight = mapInsight(record);
  upsertInsight(insight);
  return insight;
}

export function useInsightsFeed() {
  if (!loadedState.value && !loadingState.value) {
    void loadInsights();
  }

  const newsInsights = computed(() => insightsState.value.filter((insight) => insight.isNews));
  const mainInsights = computed(() => insightsState.value.filter((insight) => !insight.isNews));
  const featuredInsights = computed(() => {
    const preferred = mainInsights.value.filter((insight) => insight.featured);
    const source = preferred.length > 0 ? preferred : mainInsights.value;
    return source.slice(0, 5);
  });
  const categoryCount = computed(() => new Set(mainInsights.value.map((insight) => insight.category)).size);
  const audienceCount = computed(() => new Set(mainInsights.value.map((insight) => insight.audience)).size);

  return {
    insights: mainInsights,
    newsInsights,
    featuredInsights,
    categoryCount,
    audienceCount,
    loading: computed(() => loadingState.value),
    loaded: computed(() => loadedState.value),
    error: computed(() => errorState.value),
    loadInsights,
    loadInsightBySlug
  };
}
