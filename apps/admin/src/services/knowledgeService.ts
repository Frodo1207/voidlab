import type {
  ContentStatus,
  KnowledgeAccessTokenRecord,
  KnowledgeAssetRecord,
  KnowledgeAssetUploadResult,
  KnowledgeEntryRecord,
  KnowledgeMarkdownImportResult,
  KnowledgeSpaceRecord,
  KnowledgeVisibilityMode
} from "../types";
import { apiRequest } from "./apiClient";

interface ApiKnowledgeSpaceRecord {
  id: number;
  title: string;
  slug: string;
  description: string;
  cover_label: string;
  icon: string;
  theme_tint: string;
  visibility_mode: KnowledgeVisibilityMode;
  directory_summary: string;
  intro_markdown: string;
  token_hint: string;
  cover_url: string;
  status: ContentStatus;
  entry_count: number;
  section_count: number;
  updated_at: string;
}

interface ApiKnowledgeEntryRecord {
  id: number;
  space_id: number;
  space_slug?: string;
  title: string;
  slug: string;
  section_name: string;
  sort_order: number;
  estimated_read_minutes: number;
  public_summary: string;
  content_markdown: string;
  cover_url: string;
  is_preview: boolean;
  status: ContentStatus;
  updated_at: string;
}

interface ApiKnowledgeMarkdownImportResult {
  title: string;
  slug: string;
  section_name: string;
  estimated_read_minutes: number;
  public_summary: string;
  content_markdown: string;
  cover_url: string;
  is_preview: boolean;
  status: ContentStatus;
}

interface ApiKnowledgeAssetRecord {
  id: number;
  space_id: number;
  media_asset_id: number;
  file_name: string;
  content_type: string;
  created_at: string;
}

interface ApiKnowledgeAssetUploadResult {
  asset: ApiKnowledgeAssetRecord;
  markdown_url: string;
  markdown_snippet: string;
  public_url: string;
}

interface ApiKnowledgeAccessTokenRecord {
  id: number;
  name: string;
  access_level: "basic" | "pro" | "vip";
  scope_type: "single_space" | "multi_space" | "all_published";
  space_ids: number[];
  is_active: boolean;
  expires_at: string;
  created_by: number;
  created_at: string;
  updated_at: string;
}

interface ApiKnowledgeAccessTokenCreateResponse {
  record: ApiKnowledgeAccessTokenRecord;
  token: string;
}

export interface KnowledgeSpacePayload {
  title: string;
  slug: string;
  description: string;
  cover_label: string;
  icon: string;
  theme_tint: string;
  visibility_mode: KnowledgeVisibilityMode;
  directory_summary: string;
  intro_markdown: string;
  token_hint: string;
  cover_url: string;
  status: ContentStatus;
}

export interface KnowledgeEntryPayload {
  space_id: number;
  title: string;
  slug: string;
  section_name: string;
  sort_order: number;
  estimated_read_minutes: number;
  public_summary: string;
  content_markdown: string;
  cover_url: string;
  is_preview: boolean;
  status: ContentStatus;
}

export interface KnowledgeAccessTokenCreatePayload {
  space_id?: number;
  space_ids?: number[];
  name: string;
  access_level: "basic" | "pro" | "vip";
  expires_at: string;
}

export async function listKnowledgeSpaces(): Promise<KnowledgeSpaceRecord[]> {
  const records = await apiRequest<ApiKnowledgeSpaceRecord[]>("/api/v1/knowledge/spaces");
  return records.map(mapKnowledgeSpace);
}

export async function getKnowledgeSpaceById(id: number): Promise<KnowledgeSpaceRecord | undefined> {
  const records = await listKnowledgeSpaces();
  return records.find((record) => record.id === id);
}

export async function createKnowledgeSpace(payload: KnowledgeSpacePayload): Promise<KnowledgeSpaceRecord> {
  const record = await apiRequest<ApiKnowledgeSpaceRecord>("/api/v1/knowledge/spaces", {
    method: "POST",
    body: JSON.stringify(payload)
  });
  return mapKnowledgeSpace(record);
}

export async function updateKnowledgeSpace(id: number, payload: KnowledgeSpacePayload): Promise<KnowledgeSpaceRecord> {
  const record = await apiRequest<ApiKnowledgeSpaceRecord>(`/api/v1/knowledge/spaces/${id}`, {
    method: "PUT",
    body: JSON.stringify(payload)
  });
  return mapKnowledgeSpace(record);
}

export async function deleteKnowledgeSpace(id: number): Promise<{ deleted: boolean; id: number }> {
  return apiRequest<{ deleted: boolean; id: number }>(`/api/v1/knowledge/spaces/${id}`, {
    method: "DELETE"
  });
}

export async function listKnowledgeEntries(spaceId?: number): Promise<KnowledgeEntryRecord[]> {
  const suffix = typeof spaceId === "number" ? `?space_id=${encodeURIComponent(String(spaceId))}` : "";
  const records = await apiRequest<ApiKnowledgeEntryRecord[]>(`/api/v1/knowledge/entries${suffix}`);
  return records.map(mapKnowledgeEntry);
}

export async function getKnowledgeEntryById(id: number): Promise<KnowledgeEntryRecord | undefined> {
  const records = await listKnowledgeEntries();
  return records.find((record) => record.id === id);
}

export async function createKnowledgeEntry(payload: KnowledgeEntryPayload): Promise<KnowledgeEntryRecord> {
  const record = await apiRequest<ApiKnowledgeEntryRecord>("/api/v1/knowledge/entries", {
    method: "POST",
    body: JSON.stringify(payload)
  });
  return mapKnowledgeEntry(record);
}

export async function updateKnowledgeEntry(id: number, payload: KnowledgeEntryPayload): Promise<KnowledgeEntryRecord> {
  const record = await apiRequest<ApiKnowledgeEntryRecord>(`/api/v1/knowledge/entries/${id}`, {
    method: "PUT",
    body: JSON.stringify(payload)
  });
  return mapKnowledgeEntry(record);
}

export async function importKnowledgeEntryMarkdown(file: File): Promise<KnowledgeMarkdownImportResult> {
  const formData = new FormData();
  formData.append("file", file);

  const result = await apiRequest<ApiKnowledgeMarkdownImportResult>("/api/v1/knowledge/entries/import-markdown", {
    method: "POST",
    body: formData
  });

  return {
    title: result.title,
    slug: result.slug,
    sectionName: result.section_name,
    estimatedReadMinutes: result.estimated_read_minutes,
    publicSummary: result.public_summary,
    contentMarkdown: result.content_markdown,
    coverUrl: result.cover_url,
    isPreview: result.is_preview,
    status: result.status
  };
}

export async function deleteKnowledgeEntry(id: number): Promise<{ deleted: boolean; id: number }> {
  return apiRequest<{ deleted: boolean; id: number }>(`/api/v1/knowledge/entries/${id}`, {
    method: "DELETE"
  });
}

export async function listKnowledgeAssets(spaceId: number): Promise<KnowledgeAssetRecord[]> {
  const records = await apiRequest<ApiKnowledgeAssetRecord[]>(`/api/v1/knowledge/spaces/${spaceId}/assets`);
  return records.map(mapKnowledgeAsset);
}

export async function uploadKnowledgeAsset(spaceId: number, file: File): Promise<KnowledgeAssetUploadResult> {
  const formData = new FormData();
  formData.append("file", file);

  const result = await apiRequest<ApiKnowledgeAssetUploadResult>(`/api/v1/knowledge/spaces/${spaceId}/assets`, {
    method: "POST",
    body: formData
  });

  return {
    asset: mapKnowledgeAsset(result.asset),
    markdownUrl: result.markdown_url,
    markdownSnippet: result.markdown_snippet,
    publicUrl: result.public_url
  };
}

export async function listKnowledgeAccessTokens(spaceId?: number): Promise<KnowledgeAccessTokenRecord[]> {
  const suffix = typeof spaceId === "number" ? `?space_id=${encodeURIComponent(String(spaceId))}` : "";
  const records = await apiRequest<ApiKnowledgeAccessTokenRecord[]>(`/api/v1/knowledge/access-tokens${suffix}`);
  return records.map(mapKnowledgeAccessToken);
}

export async function createKnowledgeAccessToken(
  payload: KnowledgeAccessTokenCreatePayload
): Promise<{ record: KnowledgeAccessTokenRecord; token: string }> {
  const response = await apiRequest<ApiKnowledgeAccessTokenCreateResponse>("/api/v1/knowledge/access-tokens", {
    method: "POST",
    body: JSON.stringify(payload)
  });

  return {
    record: mapKnowledgeAccessToken(response.record),
    token: response.token
  };
}

export async function updateKnowledgeAccessTokenStatus(
  id: number,
  isActive: boolean
): Promise<KnowledgeAccessTokenRecord> {
  const record = await apiRequest<ApiKnowledgeAccessTokenRecord>(`/api/v1/knowledge/access-tokens/${id}/status`, {
    method: "PUT",
    body: JSON.stringify({ is_active: isActive })
  });
  return mapKnowledgeAccessToken(record);
}

function mapKnowledgeSpace(record: ApiKnowledgeSpaceRecord): KnowledgeSpaceRecord {
  return {
    id: record.id,
    title: record.title,
    slug: record.slug,
    description: record.description,
    coverLabel: record.cover_label,
    icon: record.icon,
    themeTint: record.theme_tint,
    visibilityMode: record.visibility_mode,
    directorySummary: record.directory_summary,
    introMarkdown: record.intro_markdown,
    tokenHint: record.token_hint,
    coverUrl: record.cover_url,
    status: record.status,
    entryCount: record.entry_count,
    sectionCount: record.section_count,
    updatedAt: record.updated_at
  };
}

function mapKnowledgeEntry(record: ApiKnowledgeEntryRecord): KnowledgeEntryRecord {
  return {
    id: record.id,
    spaceId: record.space_id,
    spaceSlug: record.space_slug ?? "",
    title: record.title,
    slug: record.slug,
    sectionName: record.section_name,
    sortOrder: record.sort_order,
    estimatedReadMinutes: record.estimated_read_minutes,
    publicSummary: record.public_summary,
    contentMarkdown: record.content_markdown,
    coverUrl: record.cover_url,
    isPreview: record.is_preview,
    status: record.status,
    updatedAt: record.updated_at
  };
}

function mapKnowledgeAsset(record: ApiKnowledgeAssetRecord): KnowledgeAssetRecord {
  return {
    id: record.id,
    spaceId: record.space_id,
    mediaAssetId: record.media_asset_id,
    fileName: record.file_name,
    contentType: record.content_type,
    createdAt: record.created_at
  };
}

function mapKnowledgeAccessToken(record: ApiKnowledgeAccessTokenRecord): KnowledgeAccessTokenRecord {
  return {
    id: record.id,
    name: record.name,
    accessLevel: record.access_level,
    scopeType: record.scope_type,
    spaceIds: record.space_ids ?? [],
    isActive: record.is_active,
    expiresAt: record.expires_at,
    createdBy: record.created_by,
    createdAt: record.created_at,
    updatedAt: record.updated_at
  };
}
