import type { ActorType, AuditLogRecord, DashboardStatsRecord, LeadRecord, LeadStatusStatsRecord, LeadSourceType, LeadStatus } from "../types";
import { apiRequest } from "./apiClient";

interface ApiLeadStatusStatsRecord {
  new: number;
  contacted: number;
  following: number;
  converted: number;
  invalid: number;
}

interface ApiDashboardStatsRecord {
  article_count: number;
  published_article_count: number;
  event_count: number;
  published_event_count: number;
  builder_count: number;
  published_builder_count: number;
  lead_count: number;
  actionable_lead_count: number;
  lead_status_distribution: ApiLeadStatusStatsRecord;
  recent_activities: ApiAuditLogRecord[];
  recent_actionable_leads: ApiLeadRecord[];
}

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
  detail_json: string;
  created_at: string;
}

interface ApiLeadRecord {
  id: number;
  source_type: LeadSourceType;
  source_id?: number;
  name: string;
  contact: string;
  message: string;
  status: LeadStatus;
  notes: string;
  owner_id?: number;
  created_at: string;
  updated_at: string;
}

export async function getDashboardStats(): Promise<DashboardStatsRecord> {
  const record = await apiRequest<ApiDashboardStatsRecord>("/api/v1/dashboard/stats");
  return mapDashboardStats(record);
}

function mapDashboardStats(record: ApiDashboardStatsRecord): DashboardStatsRecord {
  return {
    articleCount: record.article_count,
    publishedArticleCount: record.published_article_count,
    eventCount: record.event_count,
    publishedEventCount: record.published_event_count,
    builderCount: record.builder_count,
    publishedBuilderCount: record.published_builder_count,
    leadCount: record.lead_count,
    actionableLeadCount: record.actionable_lead_count,
    leadStatusDistribution: mapLeadStatusStats(record.lead_status_distribution),
    recentActivities: record.recent_activities.map(mapAuditLog),
    recentActionableLeads: record.recent_actionable_leads.map(mapLead)
  };
}

function mapLeadStatusStats(record: ApiLeadStatusStatsRecord): LeadStatusStatsRecord {
  return {
    new: record.new,
    contacted: record.contacted,
    following: record.following,
    converted: record.converted,
    invalid: record.invalid
  };
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
    detail: parseDetail(record.detail_json),
    createdAt: record.created_at
  };
}

function mapLead(record: ApiLeadRecord): LeadRecord {
  return {
    id: record.id,
    sourceType: record.source_type,
    sourceId: record.source_id,
    name: record.name,
    contact: record.contact,
    message: record.message,
    status: record.status,
    notes: record.notes,
    ownerId: record.owner_id,
    createdAt: record.created_at,
    updatedAt: record.updated_at,
    logs: []
  };
}

function parseDetail(detailJSON: string): unknown {
  try {
    return JSON.parse(detailJSON);
  } catch {
    return {};
  }
}
