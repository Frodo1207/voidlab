import type { ArticleRecord, ContentStatus } from "../types";
import { apiRequest } from "./apiClient";

interface ApiArticleRecord {
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
  status: ContentStatus;
  updated_at: string;
}

export async function listArticles(): Promise<ArticleRecord[]> {
  const records = await apiRequest<ApiArticleRecord[]>("/api/v1/articles");
  return records.map(mapArticle);
}

export async function getArticleById(id: number): Promise<ArticleRecord | undefined> {
  const record = await apiRequest<ApiArticleRecord>(`/api/v1/articles/${id}`);
  return mapArticle(record);
}

export interface ArticlePayload {
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
  status: ContentStatus;
}

export async function createArticle(payload: ArticlePayload): Promise<ArticleRecord> {
  const record = await apiRequest<ApiArticleRecord>("/api/v1/articles", {
    method: "POST",
    body: JSON.stringify(payload)
  });
  return mapArticle(record);
}

export async function updateArticle(id: number, payload: ArticlePayload): Promise<ArticleRecord> {
  const record = await apiRequest<ApiArticleRecord>(`/api/v1/articles/${id}`, {
    method: "PUT",
    body: JSON.stringify(payload)
  });
  return mapArticle(record);
}

export async function deleteArticle(id: number): Promise<{ deleted: boolean; id: number }> {
  return apiRequest<{ deleted: boolean; id: number }>(`/api/v1/articles/${id}`, {
    method: "DELETE"
  });
}

function mapArticle(record: ApiArticleRecord): ArticleRecord {
  return {
    id: record.id,
    title: record.title,
    slug: record.slug,
    summary: record.summary,
    category: record.category,
    audience: record.audience,
    tags: record.tags,
    coverUrl: record.cover_url,
    content: record.content,
    sourceName: record.source_name,
    sourceUrl: record.source_url,
    featured: record.featured,
    status: record.status,
    updatedAt: record.updated_at
  };
}
