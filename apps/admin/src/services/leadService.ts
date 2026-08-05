import type { LeadLogRecord, LeadRecord, LeadSourceType, LeadStatus } from "../types";
import { apiRequest } from "./apiClient";

interface ApiLeadLogRecord {
  id: number;
  lead_id: number;
  action: string;
  content: string;
  created_by: number;
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
  logs?: ApiLeadLogRecord[];
}

export interface LeadPayload {
  source_type: LeadSourceType;
  source_id?: number;
  name: string;
  contact: string;
  message: string;
  status?: LeadStatus;
  notes?: string;
  owner_id?: number;
}

export async function listLeads(): Promise<LeadRecord[]> {
  const records = await apiRequest<ApiLeadRecord[]>("/api/v1/leads");
  return records.map(mapLead);
}

export async function getLeadById(id: number): Promise<LeadRecord | undefined> {
  const record = await apiRequest<ApiLeadRecord>(`/api/v1/leads/${id}`);
  return mapLead(record);
}

export async function createLead(payload: LeadPayload): Promise<LeadRecord> {
  const record = await apiRequest<ApiLeadRecord>("/api/v1/leads", {
    method: "POST",
    body: JSON.stringify(payload)
  });
  return mapLead(record);
}

export async function updateLeadStatus(id: number, status: LeadStatus): Promise<LeadRecord> {
  const record = await apiRequest<ApiLeadRecord>(`/api/v1/leads/${id}/status`, {
    method: "PUT",
    body: JSON.stringify({ status })
  });
  return mapLead(record);
}

export async function addLeadLog(id: number, content: string, action = "note"): Promise<LeadRecord> {
  const record = await apiRequest<ApiLeadRecord>(`/api/v1/leads/${id}/logs`, {
    method: "POST",
    body: JSON.stringify({ action, content })
  });
  return mapLead(record);
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
    logs: (record.logs ?? []).map(mapLeadLog)
  };
}

function mapLeadLog(record: ApiLeadLogRecord): LeadLogRecord {
  return {
    id: record.id,
    leadId: record.lead_id,
    action: record.action,
    content: record.content,
    createdBy: record.created_by,
    createdAt: record.created_at
  };
}
