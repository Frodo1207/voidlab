import { resolvePublicApiPath } from "../src/runtimeConfig";

const TOKEN_KEY = "voidlab.network.token";

export type NetworkProfile = {
  id: number;
  slug: string;
  display_name: string;
  primary_role: string;
  headline: string;
  bio: string;
  city: string;
  skills: string[];
  collaboration_modes: string[];
  availability: string;
  available_from: string;
};

export type NetworkWork = {
  id: number;
  profile_id: number;
  slug: string;
  title: string;
  summary: string;
  body: string;
  cover_url: string;
  demo_url: string;
  repo_url: string;
  tags: string[];
  status: string;
  author_name: string;
  members?: Array<{ profile_id: number; display_name?: string; role: string; contribution: string; sort_order: number; is_owner: boolean }>;
};

export type NetworkOpportunity = {
  id: number;
  profile_id: number;
  slug: string;
  title: string;
  summary: string;
  opportunity_type: string;
  looking_for_roles: string[];
  required_skills: string[];
  engagement_mode: string;
  budget_hint: string;
  duration_hint: string;
  city_scope: string;
  remote_allowed: boolean;
  status: string;
  expires_at: string;
  owner_name: string;
};

export type NetworkContactRequest = {
  id: number;
  from_profile_id: number;
  target_type: string;
  target_id: number;
  message: string;
  status: string;
  progress_status: string;
  exchange_revoked: boolean;
  applicant_name: string;
  applicant_role: string;
  share_email: boolean;
  share_wechat: boolean;
  share_phone: boolean;
  related_work_ids: number[];
};

export type NetworkSession = {
  token: string;
  account: { id: number; email: string };
  profile: NetworkProfile;
};

export type NetworkAgentToken = {
  id: number;
  profile_id: number;
  name: string;
  scopes: string[];
  is_active: boolean;
  last_used_at: string;
  created_at: string;
  updated_at: string;
};

export type NetworkAgentTokenCreateResult = {
  token: string;
  record: NetworkAgentToken;
};

type Envelope<T> = {
  code: number;
  message: string;
  data: T;
};

export class NetworkRequestError extends Error {
  constructor(
    message: string,
    readonly status: number,
    readonly code: number,
    readonly path: string
  ) {
    super(message);
    this.name = "NetworkRequestError";
  }
}

export function getNetworkToken() {
  return localStorage.getItem(TOKEN_KEY) || "";
}

export function clearNetworkToken() {
  localStorage.removeItem(TOKEN_KEY);
}

export function saveNetworkSession(session: NetworkSession) {
  localStorage.setItem(TOKEN_KEY, session.token);
}

export async function networkRequest<T>(path: string, init: RequestInit = {}, authenticated = false): Promise<T> {
  const headers = new Headers(init.headers);
  if (init.body && !headers.has("Content-Type")) {
    headers.set("Content-Type", "application/json");
  }
  if (authenticated) {
    const token = getNetworkToken();
    if (token) {
      headers.set("Authorization", `Bearer ${token}`);
    }
  }

  let response: Response;
  try {
    response = await fetch(resolvePublicApiPath(path), { ...init, headers });
  } catch {
    throw new Error("无法连接市场服务，请确认 API 已启动。");
  }
  let envelope: Envelope<T>;
  try {
    envelope = (await response.json()) as Envelope<T>;
  } catch {
    throw new NetworkRequestError("市场服务返回了无效响应，请稍后重试。", response.status, -1, path);
  }
  if (!response.ok || envelope.code !== 0) {
    throw new NetworkRequestError(envelope.message || "请求失败", response.status, envelope.code, path);
  }
  return envelope.data;
}

export function postJSON<T>(path: string, data: unknown, authenticated = false) {
  return networkRequest<T>(path, { method: "POST", body: JSON.stringify(data) }, authenticated);
}
