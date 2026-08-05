import type { ContentStatus, EventRecord } from "../types";
import { apiRequest } from "./apiClient";

interface ApiEventRecord {
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
  status: ContentStatus;
  updated_at: string;
}

export interface EventPayload {
  title: string;
  slug: string;
  summary: string;
  city: string;
  location: string;
  event_type: string;
  event_time: string;
  cover_url: string;
  content: string;
  status: ContentStatus;
}

export async function listEvents(): Promise<EventRecord[]> {
  const records = await apiRequest<ApiEventRecord[]>("/api/v1/events");
  return records.map(mapEvent);
}

export async function getEventById(id: number): Promise<EventRecord | undefined> {
  const record = await apiRequest<ApiEventRecord>(`/api/v1/events/${id}`);
  return mapEvent(record);
}

export async function createEvent(payload: EventPayload): Promise<EventRecord> {
  const record = await apiRequest<ApiEventRecord>("/api/v1/events", {
    method: "POST",
    body: JSON.stringify(payload)
  });
  return mapEvent(record);
}

export async function updateEvent(id: number, payload: EventPayload): Promise<EventRecord> {
  const record = await apiRequest<ApiEventRecord>(`/api/v1/events/${id}`, {
    method: "PUT",
    body: JSON.stringify(payload)
  });
  return mapEvent(record);
}

export async function deleteEvent(id: number): Promise<{ deleted: boolean; id: number }> {
  return apiRequest<{ deleted: boolean; id: number }>(`/api/v1/events/${id}`, {
    method: "DELETE"
  });
}

function mapEvent(record: ApiEventRecord): EventRecord {
  return {
    id: record.id,
    title: record.title,
    slug: record.slug,
    summary: record.summary,
    city: record.city,
    location: record.location,
    eventType: record.event_type,
    eventTime: record.event_time,
    content: record.content,
    coverUrl: record.cover_url,
    status: record.status,
    updatedAt: record.updated_at
  };
}
