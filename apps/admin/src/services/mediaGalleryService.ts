import type { MediaGalleryItemRecord, MediaGalleryRecord } from "../types";
import { apiRequest } from "./apiClient";

export interface MediaGalleryPayload {
  title: string;
  slug: string;
  description: string;
  isDefault: boolean;
}

export interface MediaGalleryItemPayload {
  altText: string;
  sortOrder: number;
  isActive: boolean;
}

export async function listMediaGalleries(): Promise<MediaGalleryRecord[]> {
  return apiRequest<MediaGalleryRecord[]>("/api/v1/media/galleries");
}

export async function createMediaGallery(payload: MediaGalleryPayload): Promise<MediaGalleryRecord> {
  return apiRequest<MediaGalleryRecord>("/api/v1/media/galleries", {
    method: "POST",
    body: JSON.stringify(payload)
  });
}

export async function updateMediaGallery(id: number, payload: MediaGalleryPayload): Promise<MediaGalleryRecord> {
  return apiRequest<MediaGalleryRecord>(`/api/v1/media/galleries/${id}`, {
    method: "PUT",
    body: JSON.stringify(payload)
  });
}

export async function deleteMediaGallery(id: number): Promise<{ deleted: boolean }> {
  return apiRequest<{ deleted: boolean }>(`/api/v1/media/galleries/${id}`, {
    method: "DELETE"
  });
}

export async function listMediaGalleryItems(id: number): Promise<MediaGalleryItemRecord[]> {
  return apiRequest<MediaGalleryItemRecord[]>(`/api/v1/media/galleries/${id}/items`);
}

export async function addMediaGalleryItems(id: number, mediaAssetIds: number[]): Promise<MediaGalleryItemRecord[]> {
  return apiRequest<MediaGalleryItemRecord[]>(`/api/v1/media/galleries/${id}/items`, {
    method: "POST",
    body: JSON.stringify({ media_asset_ids: mediaAssetIds })
  });
}

export async function updateMediaGalleryItem(
  galleryId: number,
  itemId: number,
  payload: MediaGalleryItemPayload
): Promise<MediaGalleryItemRecord[]> {
  return apiRequest<MediaGalleryItemRecord[]>(`/api/v1/media/galleries/${galleryId}/items/${itemId}`, {
    method: "PUT",
    body: JSON.stringify(payload)
  });
}

export async function deleteMediaGalleryItem(
  galleryId: number,
  itemId: number
): Promise<{ deleted: boolean }> {
  return apiRequest<{ deleted: boolean }>(`/api/v1/media/galleries/${galleryId}/items/${itemId}`, {
    method: "DELETE"
  });
}
