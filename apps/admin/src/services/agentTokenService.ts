import type { AgentScope, AgentTokenRecord } from "../types";
import { apiRequest } from "./apiClient";

interface ApiAgentTokenRecord {
  id: number;
  name: string;
  scopes: AgentScope[];
  is_active: boolean;
  last_used_at: string;
  created_by: number;
  created_at: string;
  updated_at: string;
}

interface ApiAgentTokenCreateResponse {
  record: ApiAgentTokenRecord;
  token: string;
}

export interface AgentTokenCreatePayload {
  name: string;
  scopes: AgentScope[];
}

export async function listAgentTokens(): Promise<AgentTokenRecord[]> {
  const records = await apiRequest<ApiAgentTokenRecord[]>("/api/v1/agent-tokens");
  return records.map(mapAgentToken);
}

export async function createAgentToken(payload: AgentTokenCreatePayload): Promise<{ record: AgentTokenRecord; token: string }> {
  const response = await apiRequest<ApiAgentTokenCreateResponse>("/api/v1/agent-tokens", {
    method: "POST",
    body: JSON.stringify(payload)
  });

  return {
    record: mapAgentToken(response.record),
    token: response.token
  };
}

export async function updateAgentTokenStatus(id: number, isActive: boolean): Promise<AgentTokenRecord> {
  const record = await apiRequest<ApiAgentTokenRecord>(`/api/v1/agent-tokens/${id}/status`, {
    method: "PUT",
    body: JSON.stringify({ is_active: isActive })
  });
  return mapAgentToken(record);
}

function mapAgentToken(record: ApiAgentTokenRecord): AgentTokenRecord {
  return {
    id: record.id,
    name: record.name,
    scopes: record.scopes,
    isActive: record.is_active,
    lastUsedAt: record.last_used_at,
    createdBy: record.created_by,
    createdAt: record.created_at,
    updatedAt: record.updated_at
  };
}
