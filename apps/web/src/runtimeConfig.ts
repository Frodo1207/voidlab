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
    if (rawValue.startsWith("data:")) {
      return rawValue;
    }

    try {
      const url = new URL(rawValue);
      if (url.pathname.startsWith("/uploads/")) {
        // 兼容历史数据里写死的 IP / HTTP 资源地址。
        // 在当前 HTTPS / 域名环境下统一改写成同源 uploads 路径，避免 mixed-content。
        return `${resolvePublicApiPath(url.pathname)}${url.search}${url.hash}`;
      }
    } catch {
      return rawValue;
    }

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
