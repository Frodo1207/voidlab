import { restoreSession } from "./sessionStorage";
import { resolveApiPath } from "./runtimeConfig";

interface ApiEnvelope<T> {
  code: number;
  message: string;
  data: T;
}

export async function apiRequest<T>(path: string, init?: RequestInit): Promise<T> {
  const session = restoreSession();
  const headers = new Headers(init?.headers ?? {});
  const isFormDataBody = typeof FormData !== "undefined" && init?.body instanceof FormData;

  if (!headers.has("Content-Type") && init?.body && !isFormDataBody) {
    headers.set("Content-Type", "application/json");
  }

  if (session?.token) {
    headers.set("Authorization", `Bearer ${session.token}`);
  }

  const response = await fetch(resolveApiPath(path), {
    ...init,
    headers
  });

  const envelope = (await response.json()) as ApiEnvelope<T>;

  if (!response.ok || envelope.code !== 0) {
    throw new Error(envelope.message || "请求失败");
  }

  return envelope.data;
}
