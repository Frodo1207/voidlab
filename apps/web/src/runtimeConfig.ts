const publicApiBaseUrl = (import.meta.env.VITE_PUBLIC_API_BASE_URL || "").trim().replace(/\/+$/, "");

function normalizePath(path: string) {
  if (!path) {
    return "/";
  }

  return path.startsWith("/") ? path : `/${path}`;
}

export function resolvePublicApiPath(path: string) {
  const normalizedPath = normalizePath(path);
  return publicApiBaseUrl ? `${publicApiBaseUrl}${normalizedPath}` : normalizedPath;
}

export function resolveUploadsUrl(value: string) {
  const rawValue = value.trim();
  if (!rawValue) {
    return "";
  }

  if (rawValue.startsWith("http://") || rawValue.startsWith("https://") || rawValue.startsWith("data:")) {
    return rawValue;
  }

  if (rawValue.startsWith("/")) {
    if (rawValue.startsWith("/uploads/")) {
      return resolvePublicApiPath(rawValue);
    }
    return rawValue;
  }

  return resolvePublicApiPath(`/uploads/${rawValue}`);
}
