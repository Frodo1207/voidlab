import { computed, ref } from "vue";
import { resolvePublicApiPath, resolveUploadsUrl } from "../src/runtimeConfig";

export type BuilderRole = string;

type ApiEnvelope<T> = {
  code: number;
  message: string;
  data: T;
};

type PublicBuilderRecord = {
  id: number;
  name: string;
  slug: string;
  title: string;
  city: string;
  role: string;
  intro: string;
  story: string;
  expertise: string[];
  focus_areas: string[];
  collaboration_modes: string[];
  contactable: boolean;
  featured: boolean;
  cover_url: string;
  status: string;
  updated_at: string;
};

export type BuilderItem = {
  id: number;
  slug: string;
  name: string;
  title: string;
  city: string;
  role: BuilderRole;
  intro: string;
  story: string;
  expertise: string[];
  openFor: string;
  focusAreas: string[];
  collaborationModes: string[];
  availabilityNote: string;
  featured: boolean;
  contactable: boolean;
  cover: string;
  updatedAt: string;
};

const buildersState = ref<BuilderItem[]>([]);
const loadingState = ref(false);
const loadedState = ref(false);
const errorState = ref("");

let listPromise: Promise<BuilderItem[]> | null = null;

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

function deriveOpenFor(record: PublicBuilderRecord) {
  if (record.collaboration_modes.length > 0) {
    return record.collaboration_modes.join(" / ");
  }

  if (record.contactable) {
    return "合作交流 / 项目沟通";
  }

  return "当前以内容展示为主";
}

function deriveAvailabilityNote(record: PublicBuilderRecord) {
  if (record.contactable) {
    return `当前可通过 VOIDLAB 发起合作，适合围绕 ${record.role || "Builder 方向"} 进一步沟通。`;
  }

  return "当前以资料展示为主，合作窗口会由 VOIDLAB 统一评估后开启。";
}

function mapBuilder(record: PublicBuilderRecord): BuilderItem {
  return {
    id: record.id,
    slug: record.slug,
    name: record.name,
    title: record.title,
    city: record.city || "线上",
    role: record.role || "Builder",
    intro: record.intro || "",
    story: record.story || "",
    expertise: record.expertise || [],
    openFor: deriveOpenFor(record),
    focusAreas: record.focus_areas || [],
    collaborationModes: record.collaboration_modes || [],
    availabilityNote: deriveAvailabilityNote(record),
    featured: Boolean(record.featured),
    contactable: Boolean(record.contactable),
    cover: normalizeCoverUrl(record.cover_url),
    updatedAt: record.updated_at || ""
  };
}

function upsertBuilder(nextBuilder: BuilderItem) {
  const nextBuilders = [...buildersState.value];
  const index = nextBuilders.findIndex((item) => item.slug === nextBuilder.slug);

  if (index >= 0) {
    nextBuilders[index] = nextBuilder;
  } else {
    nextBuilders.unshift(nextBuilder);
  }

  buildersState.value = nextBuilders;
}

async function requestPublicApi<T>(path: string) {
  const response = await fetch(resolvePublicApiPath(path));
  const envelope = (await response.json()) as ApiEnvelope<T>;

  if (!response.ok || envelope.code !== 0) {
    throw new Error(envelope.message || "请求失败");
  }

  return envelope.data;
}

export async function loadBuilders(force = false) {
  if (loadedState.value && !force) {
    return buildersState.value;
  }

  if (listPromise && !force) {
    return listPromise;
  }

  loadingState.value = true;
  errorState.value = "";

  listPromise = requestPublicApi<PublicBuilderRecord[]>("/api/v1/public/builders")
    .then((records) => {
      buildersState.value = records.map(mapBuilder);
      loadedState.value = true;
      return buildersState.value;
    })
    .catch((error: unknown) => {
      errorState.value = error instanceof Error ? error.message : "加载 Builder 失败";
      throw error;
    })
    .finally(() => {
      loadingState.value = false;
      listPromise = null;
    });

  return listPromise;
}

export async function loadBuilderBySlug(slug: string) {
  const normalizedSlug = slug.trim();
  if (!normalizedSlug) {
    return null;
  }

  const existing = buildersState.value.find((item) => item.slug === normalizedSlug);
  if (existing) {
    return existing;
  }

  const record = await requestPublicApi<PublicBuilderRecord>(`/api/v1/public/builders/${normalizedSlug}`);
  const builder = mapBuilder(record);
  upsertBuilder(builder);
  return builder;
}

export function useBuilderNetwork() {
  if (!loadedState.value && !loadingState.value) {
    void loadBuilders();
  }

  const featuredBuilders = computed(() => {
    const preferred = buildersState.value.filter((builder) => builder.featured);
    const source = preferred.length > 0 ? preferred : buildersState.value;
    return source.slice(0, 6);
  });

  const cityCount = computed(() => new Set(buildersState.value.map((builder) => builder.city)).size);
  const roleCount = computed(() => new Set(buildersState.value.map((builder) => builder.role)).size);

  return {
    builders: computed(() => buildersState.value),
    featuredBuilders,
    cityCount,
    roleCount,
    loading: computed(() => loadingState.value),
    loaded: computed(() => loadedState.value),
    error: computed(() => errorState.value),
    loadBuilders,
    loadBuilderBySlug
  };
}
