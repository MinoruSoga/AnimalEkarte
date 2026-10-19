/**
 * chat-context.ts — ヘルプチャット用のマニュアルコンテキスト構築
 *
 * 自然文の質問（「レジ締めの手順は？」）は Fuse の閾値を超えて 0 件になりがちなため、
 * 質問全体での検索に加えて、助詞・文末表現・質問語を除いたキーワード単位でも検索する。
 *
 * UNIT-02 での強化:
 * - 語尾除去をループ適用し「確認したいです」→「確認」のように連なった語尾を剥がす
 * - 「〜方」「〜方法」などの形式名詞語尾を剥がして語幹を取り出す
 * - 同義語・表記ゆれ（ja-synonyms.ts）で検索語を展開する
 * - ヒット記事に「根拠ゲート」を適用する: 抽出した意味語グループのうち
 *   min(2, グループ数) 個以上が記事テキスト（または keywords）へ
 *   文字列として存在するときだけ採用する。弱い fuzzy ヒットだけでは
 *   コンテキストを返さないため、マニュアル外の質問は空（=フォールバック）になる
 * - Fuse のマッチスコアで候補を昇順ソートして topK を返す
 *
 * UNIT-05 での強化（弱ヒットの参照リンクフォールバック）:
 * - pre-gate の Fuse 候補が 1 件以上あるのに根拠ゲートを 1 件も
 *   通らなかった場合（弱ヒット）、上位 3 件を text 空のタイトル
 *   リンク候補として返す。バックエンドは context 項目を sources へ
 *   そのままエコーするため、ユーザーには辿れる記事リンクが残る。
 *   text は必ず空 — ゲートで落ちた抜粋を LLM へ渡すと根拠ゲートの
 *   意味が無くなる。空 <manual> しか見ない LLM は 記載なし
 *   フォールバックを返す（UNIT-04）。
 * - pre-gate 候補も無い真のノーマッチ（ストップワードのみの質問を
 *   含む）は従来どおり空配列を返す。
 *
 * UNIT-03 での強化（VN 対応・案1 決定論のみ。LLM/翻訳は呼ばない）:
 * - VN の機能語・疑問語（làm sao / cách / ở đâu 等）をストップワード化
 * - トークンを NFC 正規化し、根拠ゲートは大小文字・発音記号の違いを
 *   吸収して比較する（ラテン文字語のみ。"Tiêm"→"tiem" が keywords_vi の
 *   "tiêm" に一致する。JA 側の判定は従来どおり文字列そのまま）
 * - 記事側の VN 語彙 keywords_vi は Fuse の検索キー（manual-index.ts）と
 *   根拠ゲートの haystack の両方に供給する。片方だけだと Fuse ヒットが
 *   ゲートで落ちるため
 */
import type { ManualArticle, ScoredManualArticle } from "@/lib/manual-index";

import searchDict from "@/constants/manual-search-dict.json";

import type { SupportChatContextItem, SupportChatKeywordCandidate } from "../types";

import { expandJaTerm, segmentJaTerm } from "./ja-synonyms";

/** 助詞・句読点・空白など、キーワード間の区切りとして扱う文字群 */
const TOKEN_SPLIT = /[\s、。，．！？!?,・「」『』（）()【】.:;のをにがはでとへもや]+/u;

/**
 * 文末の活用形・丁寧語・形式名詞を剥がす（「登録したい」→「登録」、
 * 「作成方法」→「作成」、「入力しますか」→「入力」）。
 * 長い語尾ほど先に置く（してください より ください を先に食わせない）。
 * extractSearchTermGroups が繰り返し適用するので、連なった語尾は多段で剥がれる。
 */
const SUFFIX_STRIP = new RegExp(
  `(${searchDict.suffixes.map((sfx) => sfx.replace(/[.*+?^${}()|[\]\\]/g, "\\$&")).join("|")})$`,
  "u",
);

/** 検索語として意味を持たない一般語・質問表現 */
const STOPWORDS = new Set<string>(searchDict.stopwords);

/** SUFFIX_STRIP が収束するまで繰り返し剥がす */
function stripSuffixes(term: string): string {
  let prev = "";
  let cur = term;
  while (cur !== prev) {
    prev = cur;
    cur = cur.replace(SUFFIX_STRIP, "");
  }
  return cur;
}

/**
 * 自然文の質問から検索キーワードの「同義語グループ」を抽出する。
 * 助詞・文末表現・質問語を取り除き、2 文字以上の語を同義語展開つきで
 * 重複なく返す。各グループの先頭は質問文中の表層形。
 */
function extractSearchTermGroups(message: string): string[][] {
  const seen = new Set<string>();
  const groups: string[][] = [];
  for (const raw of message.split(TOKEN_SPLIT)) {
    // NFC へ正規化（NFD 入力と NFC のキーワード/stopword が一致しない事故を防ぐ）
    const token = raw.normalize("NFC");
    // 助詞で区切られない結合語（「新しい飼い主」）は辞書最長一致で分割する
    for (const seg of segmentJaTerm(stripSuffixes(token))) {
      // 小文字化: VN 疑問語の先頭大文字（"Cách"→"cách"）を stopword に
      // 一致させ、グループ変種も小文字に揃える（JA は無変換。根拠ゲートは
      // 両側を小文字比較するため挙動不変）
      const term = seg.toLowerCase();
      if (term.length < 2 || STOPWORDS.has(term)) continue;
      const variants = expandJaTerm(term);
      // 同じグループに属する別表層（「新しい」と「新規」）は同一キーで重複排除
      const key = [...variants].sort().join("|");
      if (seen.has(key)) continue;
      seen.add(key);
      groups.push([...variants]);
    }
  }
  return groups;
}

/** createManualSearcher が付与する matchScore を読む（無ければスコア不明=最後尾） */
function scoreOf(article: ManualArticle): number {
  const s = (article as Partial<ScoredManualArticle>).matchScore;
  return typeof s === "number" ? s : Number.POSITIVE_INFINITY;
}

/**
 * ベトナム語の比較用正規化: 結合文字（声調・母音記号）を剥がし、
 * đ/Đ を d へ畳み込む。発音記号なしの入力（"tiem"）でも
 * keywords_vi の "tiêm" に一致するようにするため。
 * JA 文字には適用しない（濁点が落ちて意味が変わるため）。
 */
function foldVn(s: string): string {
  return s.normalize("NFD").replace(/[đĐ]/g, "d").replace(/\p{M}/gu, "");
}

/**
 * 記事の根拠テキスト（タイトル・セクション・本文・keywords・keywords_vi）に
 * 変種のどれかが文字列として存在するか。
 * 比較は NFC + 小文字で行う（"Nhập"→"nhập"）。さらにラテン文字を含む
 * 変種は発音記号を畳んだ haystack とも比較する（"tiem"→"tiêm"）。
 */
function groupHasEvidence(article: ManualArticle, variants: readonly string[]): boolean {
  const base = `${article.searchText} ${(article.keywords ?? []).join(" ")} ${(article.keywords_vi ?? []).join(" ")}`;
  const normHaystack = base.normalize("NFC").toLowerCase();
  const foldHaystack = foldVn(normHaystack);
  return variants.some((v) => {
    const nv = v.normalize("NFC").toLowerCase();
    if (normHaystack.includes(nv)) return true;
    // ラテン文字を含む語のみ発音記号畳み込みで再比較（JA 判定は従来どおり）
    if (!/\p{Script=Latin}/u.test(nv)) return false;
    return foldHaystack.includes(foldVn(nv));
  });
}

/** 弱ヒット時に参照リンク候補として返す最大件数（バックエンドの context 上限 6 内） */
const WEAK_HIT_CANDIDATE_LIMIT = 3;

/**
 * 質問全体 + 抽出キーワード（同義語展開つき）で検索し、根拠ゲートを通った
 * 重複なし上位記事をコンテキスト（タイトル・リンク用識別子・本文抜粋）として返す。
 * 根拠ゲートを通る記事が無い場合、pre-gate 候補があれば text 空の
 * タイトルリンク候補を返す（弱ヒット）。pre-gate 候補も無ければ
 * 空配列（= マニュアル未収載としてフォールバック）。
 */
export function buildChatContext(
  search: (query: string) => ManualArticle[],
  message: string,
  topK: number,
  excerptLength: number,
): SupportChatContextItem[] {
  const groups = extractSearchTermGroups(message);
  if (groups.length === 0) return [];

  const candidates = new Map<string, { article: ManualArticle; best: number }>();
  const collect = (results: ManualArticle[]) => {
    for (const article of results) {
      const key = `${article.category}/${article.slug}`;
      const score = scoreOf(article);
      const existing = candidates.get(key);
      if (!existing || score < existing.best) {
        candidates.set(key, { article, best: score });
      }
    }
  };

  collect(search(message));
  const queried = new Set<string>([message]);
  for (const group of groups) {
    for (const variant of group) {
      if (queried.has(variant)) continue;
      queried.add(variant);
      collect(search(variant));
    }
  }

  // 根拠ゲート: 異なる意味語グループのうち min(2, グループ数) 個以上が
  // 記事へ文字列として現れる場合のみ採用。単一グループの質問はその語の
  // 出現を必須とし、fuzzy 一致だけの弱いヒットは捨てる。
  const required = Math.min(2, groups.length);
  const admitted = [...candidates.values()].filter(
    ({ article }) => groups.filter((g) => groupHasEvidence(article, g)).length >= required,
  );
  admitted.sort((a, b) => a.best - b.best);

  // 弱ヒット: pre-gate 候補はあるが根拠ゲートを 1 件も通らなかった。
  // 抜粋は出せないがタイトルリンク候補を返せば、バックエンドが sources
  // としてエコーしフォールバック回答に辿れる手がかりが残る。
  // text は必ず空 — ゲート落ちした本文を LLM へ渡してはいけない。
  if (admitted.length === 0) {
    if (candidates.size === 0) return []; // 真のノーマッチ
    const weak = [...candidates.values()].sort((a, b) => a.best - b.best);
    return weak.slice(0, Math.min(topK, WEAK_HIT_CANDIDATE_LIMIT)).map(({ article }) => ({
      title: article.title,
      category: article.category,
      slug: article.slug,
      text: "",
    }));
  }

  return admitted.slice(0, topK).map(({ article }) => ({
    title: article.title,
    category: article.category,
    slug: article.slug,
    text: article.content.slice(0, excerptLength),
  }));
}

// EMR-438: 候補 slug の charset フィルタは撤去した。upsert は空でない slug 全般を
// 受け付けるため、ASCII/桁数で絞ると日本語 slug 等の正規記事が keyword-only 検索の
// 候補から漏れていた（Codex P2）。長さは backend 契約 (100バイト) に合わせて制御。

/**
 * EMR-410: ハイブリッド検索用の keyword 候補 ID 一覧（上位20・順位順・重複なし）。
 * pre-evidence（根拠ゲート前）の Fuse 候補をそのまま ID 化して送り、
 * 採否は BE が正本コーパス上で決定する。本文は送らない。
 */
export function buildKeywordCandidates(
  search: (query: string) => ManualArticle[],
  message: string,
): SupportChatKeywordCandidate[] {
  const groups = extractSearchTermGroups(message);
  if (groups.length === 0) return [];

  const candidates = new Map<string, { article: ManualArticle; best: number }>();
  const collect = (results: ManualArticle[]) => {
    for (const article of results) {
      const key = `${article.category}/${article.slug}`;
      const score = scoreOf(article);
      const existing = candidates.get(key);
      if (!existing || score < existing.best) {
        candidates.set(key, { article, best: score });
      }
    }
  };
  collect(search(message));
  const queried = new Set<string>([message]);
  for (const group of groups) {
    for (const variant of group) {
      if (queried.has(variant)) continue;
      queried.add(variant);
      collect(search(variant));
    }
  }
  return (
    [...candidates.values()]
      // EMR-438: 候補の slug charset フィルタは撤去した。upsert は空でない slug 全般を
      // 受け付けるため、ASCII/桁数で絞ると日本語 slug 等の正規記事が keyword-only 検索の
      // 候補から漏れる（Codex P2）。backend 契約の slug 100バイト上限のみこちらで担保し、
      // 400 "invalid keyword candidate" を防ぐ（charset 制限はない）。
      .filter(({ article }) => new TextEncoder().encode(article.slug).length <= 100)
      .sort((a, b) => a.best - b.best)
      .slice(0, 20)
      .map(({ article }) => ({ category: article.category, slug: article.slug }))
  );
}
