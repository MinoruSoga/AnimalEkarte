/**
 * Safe FormData readers — prefer these over `formData.get(...) as string`.
 * Missing keys and non-string File values become empty string (or null for optional).
 */

export function getFormString(formData: FormData, key: string): string {
  const value = formData.get(key);
  return typeof value === "string" ? value : "";
}

export function getFormOptionalString(
  formData: FormData,
  key: string,
): string | null {
  const value = formData.get(key);
  if (value === null) return null;
  return typeof value === "string" ? value : "";
}

export function getFormEnum<T extends string>(
  formData: FormData,
  key: string,
  isValid: (value: string) => value is T,
): T | null {
  const raw = getFormString(formData, key);
  return isValid(raw) ? raw : null;
}
