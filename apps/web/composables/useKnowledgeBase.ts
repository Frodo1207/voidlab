import { computed, ref } from "vue";
import { resolvePublicApiPath, resolveUploadsUrl } from "../src/runtimeConfig";

export type KnowledgeVisibilityMode = "directory_only" | "public" | "private_hidden";
export type KnowledgeEntryStatus = "published" | "draft" | "archived";
export type KnowledgeAccessState = "public" | "unlocked" | "locked";

type ApiEnvelope<T> = {
  code: number;
  message: string;
  data: T;
};

type PublicKnowledgeSpaceRecord = {
  id: number;
  title: string;
  slug: string;
  description: string;
  cover_label: string;
  icon: string;
  theme_tint: string;
  visibility_mode: KnowledgeVisibilityMode;
  entry_count: number;
  section_count: number;
  updated_at: string;
  token_hint: string;
  directory_summary: string;
  intro_markdown: string;
  cover_url: string;
  status: KnowledgeEntryStatus;
};

type PublicKnowledgeEntryRecord = {
  id: number;
  space_id: number;
  space_slug: string;
  title: string;
  slug: string;
  section_name: string;
  sort_order: number;
  estimated_read_minutes: number;
  public_summary: string;
  content_markdown?: string;
  cover_url: string;
  is_preview: boolean;
  status: KnowledgeEntryStatus;
  updated_at: string;
};

type PublicKnowledgeTOCResponse = {
  space: PublicKnowledgeSpaceRecord;
  entries: PublicKnowledgeEntryRecord[];
};

type PublicKnowledgeEntryResponse = {
  space: PublicKnowledgeSpaceRecord;
  entry: PublicKnowledgeEntryRecord;
};

type VerifyTokenResponse = {
  grant: string;
  space: {
    id: number;
    slug: string;
  };
  access: {
    access_level: "basic" | "pro" | "vip";
    scope_type: "single_space" | "multi_space" | "all_published";
    space_ids: number[];
  };
};

type KnowledgeGrantRecord = {
  grant: string;
  accessLevel: "basic" | "pro" | "vip" | "legacy";
  scopeType: "single_space" | "multi_space" | "all_published" | "legacy_single_space";
  spaceIds: number[];
  spaceSlugs: string[];
};

export interface KnowledgeEntry {
  id: number;
  spaceId: number;
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
  coverUrl: string;
  updatedAt: string;
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
  coverUrl: string;
  status: KnowledgeEntryStatus;
}

class KnowledgeApiError extends Error {
  status: number;
  code: number;

  constructor(message: string, status: number, code: number) {
    super(message);
    this.name = "KnowledgeApiError";
    this.status = status;
    this.code = code;
  }
}

const STORAGE_KEY = "voidlab-knowledge-grants";

const spacesState = ref<KnowledgeSpace[]>([]);
const tocState = ref<Record<string, KnowledgeEntry[]>>({});
const grantsState = ref<KnowledgeGrantRecord[]>([]);
const loadingSpacesState = ref(false);
const loadedSpacesState = ref(false);
const errorState = ref("");

let grantsLoaded = false;
let spacesPromise: Promise<KnowledgeSpace[]> | null = null;
const tocPromises = new Map<string, Promise<{ space: KnowledgeSpace; entries: KnowledgeEntry[] }>>();
const entryPromises = new Map<string, Promise<KnowledgeEntry | null>>();

function ensureGrantsLoaded() {
  if (grantsLoaded || typeof window === "undefined") {
    return;
  }

  grantsLoaded = true;

  try {
    const rawValue = window.localStorage.getItem(STORAGE_KEY) || window.sessionStorage.getItem(STORAGE_KEY);
    if (!rawValue) {
      grantsState.value = [];
      return;
    }

    const parsed = JSON.parse(rawValue) as unknown;
    if (Array.isArray(parsed)) {
      grantsState.value = parsed as KnowledgeGrantRecord[];
      return;
    }

    if (parsed && typeof parsed === "object") {
      grantsState.value = Object.entries(parsed as Record<string, string>).map(([spaceSlug, grant]) => ({
        grant,
        accessLevel: "legacy",
        scopeType: "legacy_single_space",
        spaceIds: [],
        spaceSlugs: [spaceSlug]
      }));
      return;
    }

    grantsState.value = [];
  } catch {
    grantsState.value = [];
  }
}

function persistGrants() {
  if (typeof window === "undefined") {
    return;
  }

  window.localStorage.setItem(STORAGE_KEY, JSON.stringify(grantsState.value));
  window.sessionStorage.setItem(STORAGE_KEY, JSON.stringify(grantsState.value));
}

function grantAppliesToSpace(grant: KnowledgeGrantRecord, space: KnowledgeSpace | null, spaceSlug: string) {
  if (grant.scopeType === "all_published") {
    return true;
  }

  if (grant.spaceSlugs.includes(spaceSlug)) {
    return true;
  }

  if (!space) {
    return false;
  }

  return grant.spaceIds.includes(space.id);
}

function findGrantRecord(spaceSlug: string) {
  const space = getSpaceBySlug(spaceSlug);
  return grantsState.value.find((grant) => grantAppliesToSpace(grant, space, spaceSlug)) ?? null;
}

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

function mapSpace(record: PublicKnowledgeSpaceRecord): KnowledgeSpace {
  return {
    id: record.id,
    title: record.title,
    slug: record.slug,
    description: record.description || "",
    coverLabel: record.cover_label || "",
    icon: record.icon || "📘",
    themeTint: record.theme_tint || "from-[#f3f4f6] via-white to-white",
    visibilityMode: record.visibility_mode,
    entryCount: record.entry_count || 0,
    sectionCount: record.section_count || 0,
    lastUpdatedAt: record.updated_at || "",
    tokenHint: record.token_hint || "输入访问令牌即可解锁整个 Space",
    directorySummary: record.directory_summary || "",
    introMarkdown: record.intro_markdown || "",
    coverUrl: normalizeCoverUrl(record.cover_url || ""),
    status: record.status
  };
}

function mapEntry(record: PublicKnowledgeEntryRecord): KnowledgeEntry {
  return {
    id: record.id,
    spaceId: record.space_id,
    spaceSlug: record.space_slug,
    title: record.title,
    slug: record.slug,
    sectionName: record.section_name || "General",
    sortOrder: record.sort_order || 0,
    estimatedReadMinutes: record.estimated_read_minutes || 0,
    publicSummary: record.public_summary || "",
    contentMarkdown: record.content_markdown || "",
    status: record.status,
    isPreview: Boolean(record.is_preview),
    coverUrl: normalizeCoverUrl(record.cover_url || ""),
    updatedAt: record.updated_at || ""
  };
}

async function requestPublicApi<T>(path: string, init?: RequestInit) {
  const shouldRetry = (init?.method ?? "GET").toUpperCase() === "GET" && path.startsWith("/api/v1/public/knowledge/");
  const maxAttempts = shouldRetry ? 2 : 1;

  let lastError: KnowledgeApiError | Error | null = null;

  for (let attempt = 1; attempt <= maxAttempts; attempt += 1) {
    try {
      const response = await fetch(resolvePublicApiPath(path), init);
      const rawText = await response.text();

      let envelope: ApiEnvelope<T> | null = null;
      try {
        envelope = rawText ? (JSON.parse(rawText) as ApiEnvelope<T>) : null;
      } catch {
        envelope = null;
      }

      if (!response.ok || !envelope || envelope.code !== 0) {
        const error = new KnowledgeApiError(
          envelope?.message || (response.status >= 500 ? "知识库服务暂时繁忙" : "请求失败"),
          response.status,
          envelope?.code ?? -1
        );

        const retryable = shouldRetry && attempt < maxAttempts && (response.status >= 500 || response.status === 429);
        if (retryable) {
          await new Promise((resolve) => globalThis.setTimeout(resolve, 180));
          continue;
        }

        throw error;
      }

      return envelope.data;
    } catch (error) {
      lastError = error instanceof Error ? error : new Error("请求失败");
      const retryable =
        shouldRetry &&
        attempt < maxAttempts &&
        (!(lastError instanceof KnowledgeApiError) ||
          lastError.status >= 500 ||
          lastError.status === 429 ||
          lastError.status <= 0);
      if (retryable) {
        await new Promise((resolve) => globalThis.setTimeout(resolve, 180));
        continue;
      }
      throw lastError;
    }
  }

  throw lastError ?? new Error("请求失败");
}

function upsertSpace(nextSpace: KnowledgeSpace) {
  const spaces = [...spacesState.value];
  const index = spaces.findIndex((item) => item.slug === nextSpace.slug);

  if (index >= 0) {
    spaces[index] = { ...spaces[index], ...nextSpace };
  } else {
    spaces.push(nextSpace);
  }

  spacesState.value = spaces.sort((left, right) => {
    if (left.lastUpdatedAt === right.lastUpdatedAt) {
      return left.id - right.id;
    }
    return left.lastUpdatedAt < right.lastUpdatedAt ? 1 : -1;
  });
}

function setEntries(spaceSlug: string, entries: KnowledgeEntry[]) {
  // 注意：TOC 接口不返回 `content_markdown`，而正文接口会返回。
  // 如果这里直接覆盖，会导致「刚渲染出正文 -> 随后又被 TOC 覆盖成空」的闪烁/白屏问题。
  // 因此在写入 TOC 时，需要保留已加载过的正文内容。
  const existingEntries = tocState.value[spaceSlug] ?? [];
  const existingMap = new Map(existingEntries.map((item) => [item.slug, item]));

  const merged = entries.map((next) => {
    const existing = existingMap.get(next.slug);
    if (!existing) {
      return next;
    }

    return {
      ...next,
      contentMarkdown: next.contentMarkdown || existing.contentMarkdown || ""
    };
  });

  tocState.value = {
    ...tocState.value,
    [spaceSlug]: [...merged].sort((left, right) => {
      if (left.sortOrder === right.sortOrder) {
        return left.id - right.id;
      }
      return left.sortOrder - right.sortOrder;
    })
  };
}

function upsertEntry(spaceSlug: string, nextEntry: KnowledgeEntry) {
  const currentEntries = tocState.value[spaceSlug] ?? [];
  const entries = [...currentEntries];
  const index = entries.findIndex((item) => item.slug === nextEntry.slug);

  if (index >= 0) {
    entries[index] = { ...entries[index], ...nextEntry };
  } else {
    entries.push(nextEntry);
  }

  setEntries(spaceSlug, entries);
}

export async function loadKnowledgeSpaces(force = false) {
  if (loadedSpacesState.value && !force) {
    return spacesState.value;
  }

  if (spacesPromise && !force) {
    return spacesPromise;
  }

  loadingSpacesState.value = true;
  errorState.value = "";

  spacesPromise = requestPublicApi<PublicKnowledgeSpaceRecord[]>("/api/v1/public/knowledge/spaces")
    .then((records) => {
      spacesState.value = records.map(mapSpace);
      loadedSpacesState.value = true;
      return spacesState.value;
    })
    .catch((error: unknown) => {
      errorState.value = error instanceof Error ? error.message : "加载知识库失败";
      throw error;
    })
    .finally(() => {
      loadingSpacesState.value = false;
      spacesPromise = null;
    });

  return spacesPromise;
}

export async function loadKnowledgeSpaceBySlug(spaceSlug: string, force = false) {
  const normalizedSlug = spaceSlug.trim();
  if (!normalizedSlug) {
    return null;
  }

  const activePromise = tocPromises.get(normalizedSlug);
  if (activePromise && !force) {
    return activePromise;
  }

  const promise = requestPublicApi<PublicKnowledgeTOCResponse>(`/api/v1/public/knowledge/spaces/${normalizedSlug}/toc`)
    .then((payload) => {
      const nextSpace = mapSpace(payload.space);
      const nextEntries = payload.entries.map(mapEntry);
      upsertSpace(nextSpace);
      setEntries(normalizedSlug, nextEntries);
      return { space: nextSpace, entries: nextEntries };
    })
    .finally(() => {
      tocPromises.delete(normalizedSlug);
    });

  tocPromises.set(normalizedSlug, promise);
  return promise;
}

export async function loadKnowledgeEntryBySlug(spaceSlug: string, entrySlug: string, force = false) {
  ensureGrantsLoaded();

  const normalizedSpaceSlug = spaceSlug.trim();
  const normalizedEntrySlug = entrySlug.trim();
  if (!normalizedSpaceSlug || !normalizedEntrySlug) {
    return null;
  }

  const cacheKey = `${normalizedSpaceSlug}:${normalizedEntrySlug}`;
  const existing = getEntryBySlug(normalizedSpaceSlug, normalizedEntrySlug);
  if (existing?.contentMarkdown && !force) {
    return existing;
  }

  const activePromise = entryPromises.get(cacheKey);
  if (activePromise && !force) {
    return activePromise;
  }

  const headers = new Headers();
  const grant = getSpaceGrant(normalizedSpaceSlug);
  if (grant) {
    headers.set("X-Knowledge-Grant", grant);
  }

  const promise = requestPublicApi<PublicKnowledgeEntryResponse>(
    `/api/v1/public/knowledge/spaces/${normalizedSpaceSlug}/entries/${normalizedEntrySlug}`,
    { headers }
  )
    .then((payload) => {
      const nextSpace = mapSpace(payload.space);
      const nextEntry = mapEntry(payload.entry);
      upsertSpace(nextSpace);
      upsertEntry(normalizedSpaceSlug, nextEntry);
      return nextEntry;
    })
    .finally(() => {
      entryPromises.delete(cacheKey);
    });

  entryPromises.set(cacheKey, promise);
  return promise;
}

export function getSpaceBySlug(spaceSlug: string) {
  return spacesState.value.find((space) => space.slug === spaceSlug) ?? null;
}

export function getEntriesBySpace(spaceSlug: string) {
  return tocState.value[spaceSlug] ?? [];
}

export function getEntryBySlug(spaceSlug: string, entrySlug: string) {
  return getEntriesBySpace(spaceSlug).find((entry) => entry.slug === entrySlug) ?? null;
}

export function isSpaceUnlocked(spaceSlug: string) {
  ensureGrantsLoaded();
  return Boolean(findGrantRecord(spaceSlug));
}

export function isSpacePublic(spaceSlug: string) {
  const space = getSpaceBySlug(spaceSlug);
  return space?.visibilityMode === "public";
}

export function getSpaceAccessState(spaceSlug: string): KnowledgeAccessState {
  if (isSpacePublic(spaceSlug)) {
    return "public";
  }

  return isSpaceUnlocked(spaceSlug) ? "unlocked" : "locked";
}

export function getSpaceGrant(spaceSlug: string) {
  ensureGrantsLoaded();
  return findGrantRecord(spaceSlug)?.grant || "";
}

export function canReadEntry(spaceSlug: string, entrySlug: string) {
  const entry = getEntryBySlug(spaceSlug, entrySlug);
  if (!entry) {
    return false;
  }

  return entry.isPreview || isSpacePublic(spaceSlug) || isSpaceUnlocked(spaceSlug);
}

export async function unlockKnowledgeSpace(spaceSlug: string, token: string) {
  ensureGrantsLoaded();

  const normalizedSpaceSlug = spaceSlug.trim();
  if (!normalizedSpaceSlug) {
    return { success: false, message: "知识空间不存在" };
  }

  const payload = await requestPublicApi<VerifyTokenResponse>(
    `/api/v1/public/knowledge/spaces/${normalizedSpaceSlug}/verify-token`,
    {
      method: "POST",
      headers: {
        "Content-Type": "application/json"
      },
      body: JSON.stringify({
        token: token.trim()
      })
    }
  );

  const nextGrant: KnowledgeGrantRecord = {
    grant: payload.grant,
    accessLevel: payload.access.access_level,
    scopeType: payload.access.scope_type,
    spaceIds:
      payload.access.scope_type === "single_space" && (payload.access.space_ids?.length ?? 0) === 0
        ? [payload.space.id]
        : payload.access.space_ids ?? [],
    spaceSlugs: [normalizedSpaceSlug]
  };
  grantsState.value = [
    nextGrant,
    ...grantsState.value.filter((item) => item.grant !== nextGrant.grant)
  ];
  persistGrants();

  return { success: true, message: "知识空间已解锁" };
}

export function lockKnowledgeSpace(spaceSlug: string) {
  ensureGrantsLoaded();

  const target = findGrantRecord(spaceSlug);
  if (!target) {
    return;
  }

  grantsState.value = grantsState.value.filter((item) => item.grant !== target.grant);
  persistGrants();
}

export function getSpaceStats(spaceSlug: string) {
  const spaceEntries = getEntriesBySpace(spaceSlug);
  const previewCount = spaceEntries.filter((entry) => entry.isPreview).length;
  const totalMinutes = spaceEntries.reduce((sum, entry) => sum + entry.estimatedReadMinutes, 0);

  return {
    entryCount: spaceEntries.length,
    previewCount,
    totalMinutes
  };
}

export function resolveKnowledgeAssetUrl(spaceSlug: string, source: string) {
  const normalizedSource = source.trim();
  if (!normalizedSource) {
    return "";
  }

  if (
    normalizedSource.startsWith("http://") ||
    normalizedSource.startsWith("https://") ||
    normalizedSource.startsWith("data:")
  ) {
    return normalizedSource;
  }

  const match = normalizedSource.match(/^knowledge-asset:\/\/(\d+)$/i);
  if (!match) {
    return resolveUploadsUrl(normalizedSource);
  }

  const assetID = match[1];
  const grant = getSpaceGrant(spaceSlug);
  const path = resolvePublicApiPath(`/api/v1/public/knowledge/spaces/${spaceSlug}/assets/${assetID}`);
  return grant ? `${path}?grant=${encodeURIComponent(grant)}` : path;
}

export function useKnowledgeBase() {
  ensureGrantsLoaded();

  if (!loadedSpacesState.value && !loadingSpacesState.value) {
    void loadKnowledgeSpaces();
  }

  return {
    spaces: computed(() => spacesState.value),
    loading: computed(() => loadingSpacesState.value),
    loaded: computed(() => loadedSpacesState.value),
    error: computed(() => errorState.value),
    loadKnowledgeSpaces,
    loadKnowledgeSpaceBySlug,
    loadKnowledgeEntryBySlug,
    getSpaceBySlug,
    getEntriesBySpace,
    getEntryBySlug,
    isSpacePublic,
    getSpaceAccessState,
    isSpaceUnlocked,
    getSpaceGrant,
    canReadEntry,
    unlockSpace: unlockKnowledgeSpace,
    lockSpace: lockKnowledgeSpace,
    getSpaceStats,
    resolveKnowledgeAssetUrl
  };
}
