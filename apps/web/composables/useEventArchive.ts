import { computed, ref } from "vue";
import { resolvePublicApiPath, resolveUploadsUrl } from "../src/runtimeConfig";

export type EventStatus = "live" | "next" | "done";
export type EventType = string;

type ApiEnvelope<T> = {
  code: number;
  message: string;
  data: T;
};

type PublicEventRecord = {
  id: number;
  title: string;
  slug: string;
  summary: string;
  city: string;
  location: string;
  event_type: string;
  event_time: string;
  content: string;
  cover_url: string;
  status: string;
  updated_at: string;
};

export type EventItem = {
  id: number;
  slug: string;
  title: string;
  time: string;
  rawTime: string;
  location: string;
  city: string;
  status: EventStatus;
  type: EventType;
  summary: string;
  content: string;
  cover: string;
};

const fallbackCover = "/assets/event1.PNG";
const weekdayLabels = ["周日", "周一", "周二", "周三", "周四", "周五", "周六"];

const eventsState = ref<EventItem[]>([]);
const loadingState = ref(false);
const loadedState = ref(false);
const errorState = ref("");

let listPromise: Promise<EventItem[]> | null = null;

function parseEventDate(value: string) {
  const normalized = value.trim();
  if (!normalized) {
    return null;
  }

  const date = new Date(normalized.includes("T") ? normalized : normalized.replace(" ", "T"));
  return Number.isNaN(date.getTime()) ? null : date;
}

function isSameDay(left: Date, right: Date) {
  return (
    left.getFullYear() === right.getFullYear() &&
    left.getMonth() === right.getMonth() &&
    left.getDate() === right.getDate()
  );
}

function deriveEventStatus(rawTime: string): EventStatus {
  const eventDate = parseEventDate(rawTime);
  if (!eventDate) {
    return "next";
  }

  const now = new Date();
  if (isSameDay(eventDate, now)) {
    return "live";
  }

  return eventDate.getTime() > now.getTime() ? "next" : "done";
}

function formatEventTime(rawTime: string) {
  const eventDate = parseEventDate(rawTime);
  if (!eventDate) {
    return rawTime;
  }

  const month = String(eventDate.getMonth() + 1).padStart(2, "0");
  const day = String(eventDate.getDate()).padStart(2, "0");
  const hours = String(eventDate.getHours()).padStart(2, "0");
  const minutes = String(eventDate.getMinutes()).padStart(2, "0");

  return `${month}.${day} ${weekdayLabels[eventDate.getDay()]} ${hours}:${minutes}`;
}

function normalizeCoverUrl(value: string) {
  const cover = value.trim();
  if (!cover) {
    return fallbackCover;
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

function mapEvent(record: PublicEventRecord): EventItem {
  return {
    id: record.id,
    slug: record.slug,
    title: record.title,
    time: formatEventTime(record.event_time),
    rawTime: record.event_time,
    location: record.location || record.city || "待定",
    city: record.city || "待定",
    status: deriveEventStatus(record.event_time),
    type: record.event_type || "活动",
    summary: record.summary || "",
    content: record.content || "",
    cover: normalizeCoverUrl(record.cover_url)
  };
}

function upsertEvent(nextEvent: EventItem) {
  const nextEvents = [...eventsState.value];
  const index = nextEvents.findIndex((item) => item.slug === nextEvent.slug);

  if (index >= 0) {
    nextEvents[index] = nextEvent;
  } else {
    nextEvents.unshift(nextEvent);
  }

  eventsState.value = nextEvents;
}

async function requestPublicApi<T>(path: string) {
  const response = await fetch(resolvePublicApiPath(path));
  const envelope = (await response.json()) as ApiEnvelope<T>;

  if (!response.ok || envelope.code !== 0) {
    throw new Error(envelope.message || "请求失败");
  }

  return envelope.data;
}

export async function loadEvents(force = false) {
  if (loadedState.value && !force) {
    return eventsState.value;
  }

  if (listPromise && !force) {
    return listPromise;
  }

  loadingState.value = true;
  errorState.value = "";

  listPromise = requestPublicApi<PublicEventRecord[]>("/api/v1/public/events")
    .then((records) => {
      eventsState.value = records.map(mapEvent);
      loadedState.value = true;
      return eventsState.value;
    })
    .catch((error: unknown) => {
      errorState.value = error instanceof Error ? error.message : "加载活动失败";
      throw error;
    })
    .finally(() => {
      loadingState.value = false;
      listPromise = null;
    });

  return listPromise;
}

export async function loadEventBySlug(slug: string) {
  const normalizedSlug = slug.trim();
  if (!normalizedSlug) {
    return null;
  }

  const existing = eventsState.value.find((item) => item.slug === normalizedSlug);
  if (existing) {
    return existing;
  }

  const record = await requestPublicApi<PublicEventRecord>(`/api/v1/public/events/${normalizedSlug}`);
  const event = mapEvent(record);
  upsertEvent(event);
  return event;
}

export function useEventArchive() {
  if (!loadedState.value && !loadingState.value) {
    void loadEvents();
  }

  const featuredEvents = computed(() => eventsState.value.slice(0, 5));
  const liveCount = computed(() => eventsState.value.filter((event) => event.status === "live").length);
  const citiesCount = computed(() => new Set(eventsState.value.map((event) => event.city)).size);

  return {
    events: computed(() => eventsState.value),
    featuredEvents,
    liveCount,
    citiesCount,
    loading: computed(() => loadingState.value),
    loaded: computed(() => loadedState.value),
    error: computed(() => errorState.value),
    loadEvents,
    loadEventBySlug
  };
}

export function eventStatusLabel(status: EventStatus) {
  if (status === "live") return "报名中";
  if (status === "next") return "即将开始";
  return "已结束";
}

export function eventStatusBadgeClass(status: EventStatus) {
  if (status === "done") {
    return "border-white/30 text-white/50";
  }

  return "border-[var(--color-turquoise)] text-[var(--color-turquoise)]";
}

export function eventActionLabel(status: EventStatus) {
  if (status === "done") return "查看回顾";
  if (status === "next") return "预约席位";
  return "立即报名";
}
