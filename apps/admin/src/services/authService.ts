import type { AuthSession, UserProfile } from "../types";
import { apiRequest } from "./apiClient";
export { clearSession, persistSession, restoreSession } from "./sessionStorage";

interface ApiUserProfile {
  id: number;
  username: string;
  role: "admin" | "editor" | "ops";
  display_name: string;
  is_active: boolean;
}

interface ApiAuthSession {
  token: string;
  user: ApiUserProfile;
}

export async function login(username: string, password: string): Promise<AuthSession> {
  if (!username.trim() || !password.trim()) {
    throw new Error("请输入账号和密码");
  }

  const session = await apiRequest<ApiAuthSession>("/api/v1/auth/login", {
    method: "POST",
    body: JSON.stringify({
      username,
      password
    })
  });

  return {
    token: session.token,
    user: mapUser(session.user)
  };
}

export async function getCurrentUser(): Promise<UserProfile> {
  const user = await apiRequest<ApiUserProfile>("/api/v1/auth/me");
  return mapUser(user);
}

function mapUser(user: ApiUserProfile): UserProfile {
  return {
    id: user.id,
    username: user.username,
    role: user.role,
    displayName: user.display_name,
    isActive: user.is_active
  };
}
