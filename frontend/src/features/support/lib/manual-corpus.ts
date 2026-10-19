/**
 * manual-corpus.ts — 実効マニュアルコーパスの canonical JSON + SHA-256 revision。
 * backend/internal/manualarticle/corpus.go と byte 一致が必須:
 * フィールド順 category,slug,title,section,body,keywords,keywords_vi・空白なし・
 * HTML/Unicode エスケープなし（JSON.stringify と Go SetEscapeHTML(false) は一致）。
 * ソートは category→slug の UTF-16 コードユニット辞書順（localeCompare 不使用）。
 */
import type { ManualArticle } from "@/lib/manual-index";
import { MANUAL_CORPUS_BASE_REVISION } from "@/constants/manual-corpus.generated";

/** ベースコーパスの生成時 revision（マージ前）。EMR-410 の base hash 照合用 */
export const baseCorpusRevision = MANUAL_CORPUS_BASE_REVISION;

// Go encoding/json は U+2028/U+2029 を \uXXXX にエスケープする。JSON.stringify は
// 生の文字を出すため、Go 標準に合わせて実 Unicode 文字だけを置換する
// （文字列中のリテラル "\u2028" エスケープ表記は対象外）。
const U2028_RE = /\u2028/g;
const U2029_RE = /\u2029/g;

// EMR-441: backend sortCorpusArticles (Go string < = UTF-8 バイト順 = Unicode
// code point 順) と同一の順序で整列する比较器。JS の文字列 < は UTF-16 単位順のため、
// サロゲートペア slug (😀 等) と BMP slug (Ｚ 等) の混在で backend と異なる配列順に
// なり revision hash が不一致 (全 chat で 409) になるのを防ぐ。
function compareGoString(a: string, b: string): number {
  const ac = [...a];
  const bc = [...b];
  const n = Math.min(ac.length, bc.length);
  for (let i = 0; i < n; i++) {
    const ap = ac[i]!.codePointAt(0)!;
    const bp = bc[i]!.codePointAt(0)!;
    if (ap !== bp) return ap < bp ? -1 : 1;
  }
  if (ac.length === bc.length) return 0;
  return ac.length < bc.length ? -1 : 1;
}

export function canonicalCorpusJson(articles: ManualArticle[]): string {
  const sorted = [...articles].sort((a, b) =>
    a.category !== b.category
      ? compareGoString(a.category, b.category)
      : compareGoString(a.slug, b.slug),
  );
  return `[${sorted
    .map((a) =>
      JSON.stringify({
        category: a.category,
        slug: a.slug,
        title: a.title,
        section: a.section,
        body: a.content,
        keywords: a.keywords ?? [],
        keywords_vi: a.keywords_vi ?? [],
      })
        .replace(U2028_RE, "\\u2028")
        .replace(U2029_RE, "\\u2029"),
    )
    .join(",")}]`;
}

function sha256Hex(buf: ArrayBuffer): string {
  return Array.from(new Uint8Array(buf))
    .map((b) => b.toString(16).padStart(2, "0"))
    .join("");
}

/** 実効コーパス revision（canonical article array の SHA-256 hex）を非同期で計算する */
export async function computeManualRevision(articles: ManualArticle[]): Promise<string> {
  const json = canonicalCorpusJson(articles);
  const digest = await crypto.subtle.digest("SHA-256", new TextEncoder().encode(json));
  return sha256Hex(digest);
}

/** base コーパス revision と生成定数の一致を検証（初回送信前の整合性確認用） */
export async function verifyBaseCorpusRevision(articles: ManualArticle[]): Promise<boolean> {
  return (await computeManualRevision(articles)) === baseCorpusRevision;
}
