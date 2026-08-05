import type { BuilderRecord, ContentStatus } from "../types";
import { apiRequest } from "./apiClient";

interface ApiBuilderRecord {
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
  status: ContentStatus;
  updated_at: string;
}

export interface BuilderPayload {
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
  status: ContentStatus;
}

export async function listBuilders(): Promise<BuilderRecord[]> {
  const records = await apiRequest<ApiBuilderRecord[]>("/api/v1/builders");
  return records.map(mapBuilder);
}

export async function getBuilderById(id: number): Promise<BuilderRecord | undefined> {
  const record = await apiRequest<ApiBuilderRecord>(`/api/v1/builders/${id}`);
  return mapBuilder(record);
}

export async function createBuilder(payload: BuilderPayload): Promise<BuilderRecord> {
  const record = await apiRequest<ApiBuilderRecord>("/api/v1/builders", {
    method: "POST",
    body: JSON.stringify(payload)
  });
  return mapBuilder(record);
}

export async function updateBuilder(id: number, payload: BuilderPayload): Promise<BuilderRecord> {
  const record = await apiRequest<ApiBuilderRecord>(`/api/v1/builders/${id}`, {
    method: "PUT",
    body: JSON.stringify(payload)
  });
  return mapBuilder(record);
}

export async function deleteBuilder(id: number): Promise<{ deleted: boolean; id: number }> {
  return apiRequest<{ deleted: boolean; id: number }>(`/api/v1/builders/${id}`, {
    method: "DELETE"
  });
}

function mapBuilder(record: ApiBuilderRecord): BuilderRecord {
  return {
    id: record.id,
    name: record.name,
    slug: record.slug,
    title: record.title,
    city: record.city,
    role: record.role,
    intro: record.intro,
    story: record.story,
    expertise: record.expertise,
    focusAreas: record.focus_areas,
    collaborationModes: record.collaboration_modes,
    contactable: record.contactable,
    featured: record.featured,
    coverUrl: record.cover_url,
    status: record.status,
    updatedAt: record.updated_at
  };
}
