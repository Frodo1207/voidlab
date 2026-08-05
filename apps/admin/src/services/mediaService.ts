import type { MediaAssetRecord } from "../types";
import { apiRequest } from "./apiClient";

export async function listMediaAssets(): Promise<MediaAssetRecord[]> {
  return apiRequest<MediaAssetRecord[]>("/api/v1/media");
}

export async function uploadMediaAsset(file: File): Promise<MediaAssetRecord> {
  const formData = new FormData();
  formData.append("file", file);

  return apiRequest<MediaAssetRecord>("/api/v1/media/upload", {
    method: "POST",
    body: formData
  });
}
