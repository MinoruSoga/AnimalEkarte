const ENCODED_BACKSLASH = /%5c/i;

function hasAsciiControlCharacter(value: string): boolean {
  return [...value].some((character) => {
    const codePoint = character.codePointAt(0) ?? 0;
    return codePoint <= 0x1f || codePoint === 0x7f;
  });
}

/** Parses a same-origin browser path or rejects it without attempting navigation. */
export function parseInternalPath(candidate: unknown): string | null {
  if (typeof candidate !== "string" || candidate === "") return null;
  if (candidate.trim() !== candidate) return null;
  if (!candidate.startsWith("/") || candidate.startsWith("//")) return null;
  if (hasAsciiControlCharacter(candidate) || ENCODED_BACKSLASH.test(candidate)) {
    return null;
  }

  try {
    const parsed = new URL(candidate, window.location.origin);
    if (parsed.origin !== window.location.origin) return null;
    return `${parsed.pathname}${parsed.search}${parsed.hash}`;
  } catch {
    return null;
  }
}
