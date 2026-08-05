import type { ManagedUserRecord, UserRole } from "../types";
import { apiRequest } from "./apiClient";

interface ApiManagedUserRecord {
  id: number;
  username: string;
  role: UserRole;
  display_name: string;
  is_active: boolean;
  created_at: string;
  updated_at: string;
}

export interface UserCreatePayload {
  username: string;
  password: string;
  role: UserRole;
}

export async function listUsers(): Promise<ManagedUserRecord[]> {
  const records = await apiRequest<ApiManagedUserRecord[]>("/api/v1/users");
  return records.map(mapManagedUser);
}

export async function createUser(payload: UserCreatePayload): Promise<ManagedUserRecord> {
  const record = await apiRequest<ApiManagedUserRecord>("/api/v1/users", {
    method: "POST",
    body: JSON.stringify(payload)
  });
  return mapManagedUser(record);
}

export async function updateUserRole(id: number, role: UserRole): Promise<ManagedUserRecord> {
  const record = await apiRequest<ApiManagedUserRecord>(`/api/v1/users/${id}/role`, {
    method: "PUT",
    body: JSON.stringify({ role })
  });
  return mapManagedUser(record);
}

export async function updateUserStatus(id: number, isActive: boolean): Promise<ManagedUserRecord> {
  const record = await apiRequest<ApiManagedUserRecord>(`/api/v1/users/${id}/status`, {
    method: "PUT",
    body: JSON.stringify({ is_active: isActive })
  });
  return mapManagedUser(record);
}

export async function resetUserPassword(id: number, password: string): Promise<void> {
  await apiRequest(`/api/v1/users/${id}/password`, {
    method: "PUT",
    body: JSON.stringify({ password })
  });
}

function mapManagedUser(record: ApiManagedUserRecord): ManagedUserRecord {
  return {
    id: record.id,
    username: record.username,
    role: record.role,
    displayName: record.display_name,
    isActive: record.is_active,
    createdAt: record.created_at,
    updatedAt: record.updated_at
  };
}
