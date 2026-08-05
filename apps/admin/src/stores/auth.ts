import { computed, ref } from "vue";
import { defineStore } from "pinia";
import type { AuthSession, UserProfile, UserRole } from "../types";
import {
  clearSession,
  login as loginRequest,
  persistSession,
  restoreSession
} from "../services/authService";

export const useAuthStore = defineStore("auth", () => {
  const initialSession = restoreSession();
  const token = ref<string | null>(initialSession?.token ?? null);
  const user = ref<UserProfile | null>(initialSession?.user ?? null);

  const isAuthenticated = computed(() => Boolean(token.value));
  const role = computed<UserRole | null>(() => user.value?.role ?? null);

  async function login(username: string, password: string): Promise<AuthSession> {
    const session = await loginRequest(username, password);
    token.value = session.token;
    user.value = session.user;
    persistSession(session);
    return session;
  }

  function logout(): void {
    token.value = null;
    user.value = null;
    clearSession();
  }

  function hasAnyRole(roles: readonly UserRole[]): boolean {
    if (!roles.length) {
      return true;
    }

    if (!role.value) {
      return false;
    }

    return roles.includes(role.value);
  }

  return {
    token,
    user,
    role,
    isAuthenticated,
    login,
    hasAnyRole,
    logout
  };
});
