export function normalizeMainlandPhone(value: string) {
  const digits = value.replace(/\D/g, "");
  const normalized = digits.length === 13 && digits.startsWith("86") ? digits.slice(2) : digits;
  if (/^1[3-9]\d{9}$/.test(normalized)) {
    return normalized;
  }
  return "";
}

function looksLikePhoneValue(value: string) {
  const trimmed = value.trim();
  if (!trimmed || trimmed.includes("@")) {
    return false;
  }

  if (/[A-Za-z_]/.test(trimmed)) {
    return false;
  }

  const digits = trimmed.replace(/\D/g, "");
  return digits.length >= 7;
}

export function normalizeFlexibleContact(value: string) {
  const trimmed = value.trim();
  const normalizedPhone = normalizeMainlandPhone(trimmed);
  return normalizedPhone || trimmed;
}

export function validateFlexibleContact(value: string) {
  const trimmed = value.trim();
  if (!trimmed) {
    return "请输入联系方式。";
  }

  if (looksLikePhoneValue(trimmed) && !normalizeMainlandPhone(trimmed)) {
    return "如果填写的是手机号，请输入有效的 11 位手机号。";
  }

  return "";
}
