import type { ActorType, AuditLogRecord } from "../types";
import { apiRequest } from "./apiClient";

interface ApiAuditLogRecord {
  id: number;
  actor_type: ActorType;
  actor_id: number;
  actor_username: string;
  actor_role: string;
  agent_token_id?: number;
  action: string;
  entity_type: string;
  entity_id?: number;
  entity_label: string;
  detail: unknown;
  created_at: string;
}

export async function listAuditLogs(limit = 100): Promise<AuditLogRecord[]> {
  const records = await apiRequest<ApiAuditLogRecord[]>(`/api/v1/audit-logs?limit=${limit}`);
  return records.map(mapAuditLog);
}

function mapAuditLog(record: ApiAuditLogRecord): AuditLogRecord {
  return {
    id: record.id,
    actorType: record.actor_type,
    actorId: record.actor_id,
    actorUsername: record.actor_username,
    actorRole: record.actor_role,
    agentTokenId: record.agent_token_id,
    action: record.action,
    entityType: record.entity_type,
    entityId: record.entity_id,
    entityLabel: record.entity_label,
    detail: record.detail,
    createdAt: record.created_at
  };
}
