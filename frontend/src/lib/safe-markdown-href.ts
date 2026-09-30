/**
 * safe-markdown-href — Markdown のリンク href を安全なスキームのみに限定する。
 * javascript:/data: 等の XSS ベクタを除外する（react-markdown の a レンダラー用）。
 */
export function getSafeMarkdownHref(href: unknown): string | undefined {
  if (typeof href !== "string") return undefined;
  const trimmed = href.trim();
  if (trimmed === "") return undefined;

  const decoded = (() => {
    try {
      return decodeURIComponent(trimmed);
    } catch {
      return trimmed;
    }
  })();

  // 制御文字（NULLバイト等）でスキーム判定を回避する攻撃を防ぐため意図的に制御文字を除去する
  // eslint-disable-next-line no-control-regex
  const compact = decoded.replace(/[\u0000-\u001F\u007F\s]+/g, "");
  const schemeMatch = /^([a-z][a-z0-9+.-]*):/i.exec(compact);
  if (schemeMatch) {
    const protocol = `${schemeMatch[1].toLowerCase()}:`;
    return protocol === "http:" || protocol === "https:" || protocol === "mailto:"
      ? trimmed
      : undefined;
  }

  if (trimmed.startsWith("//")) {
    return undefined;
  }

  return trimmed;
}
