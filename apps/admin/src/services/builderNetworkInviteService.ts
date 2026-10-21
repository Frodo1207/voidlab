import type { NetworkInviteRecord } from "../types";
import { apiRequest } from "./apiClient";

interface ApiNetworkInviteRecord {
  id: number;
  max_uses: number;
  used_count: number;
  expires_at: string;
  status: "active" | "disabled";
  created_by: number;
  created_at: string;
}

interface ApiNetworkInviteCreateResponse {
  code: string;
  max_uses: number;
  expires_at: string;
}

export interface NetworkInviteCreatePayload {
  maxUses: number;
  expiresAt: string;
}

export async function listNetworkInvites(): Promise<NetworkInviteRecord[]> {
  const records = await apiRequest<ApiNetworkInviteRecord[]>("/api/v1/network/invitations");
  return records.map(mapNetworkInvite);
}

export async function createNetworkInvite(
  payload: NetworkInviteCreatePayload
): Promise<ApiNetworkInviteCreateResponse> {
  return apiRequest<ApiNetworkInviteCreateResponse>("/api/v1/network/invitations", {
    method: "POST",
    body: JSON.stringify({
      max_uses: payload.maxUses,
      expires_at: payload.expiresAt
    })
  });
}

export async function updateNetworkInviteStatus(
  id: number,
  status: NetworkInviteRecord["status"]
): Promise<void> {
  await apiRequest<{ updated: boolean }>(`/api/v1/network/invitations/${id}/status`, {
    method: "PUT",
    body: JSON.stringify({ status })
  });
}

function mapNetworkInvite(record: ApiNetworkInviteRecord): NetworkInviteRecord {
  return {
    id: record.id,
    maxUses: record.max_uses,
    usedCount: record.used_count,
    expiresAt: record.expires_at,
    status: record.status,
    createdBy: record.created_by,
    createdAt: record.created_at
  };
}
