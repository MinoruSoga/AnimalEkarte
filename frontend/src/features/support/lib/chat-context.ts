/**
 * chat-context.ts — ヘルプチャット用のマニュアルコンテキスト構築
 *
 * 自然文の質問（「レジ締めの手順は？」）は Fuse の閾値を超えて 0 件になりがちなため、
 * 質問全体での検索に加えて、助詞・文末表現・質問語を除いたキーワード単位でも検索し、
 * ヒットを統合して上位 K 件をコンテキストとして返す。
 */
import type { ManualArticle } from "@/lib/manual-index";

import type { SupportChatContextItem } from "../types";

/** 助詞・句読点・空白など、キーワード間の区切りとして扱う文字群 */
const TOKEN_SPLIT = /[\s、。，．！？!?,・「」『』（）()【】のをにがはでとへもや]+/u;

/** 文末の活用形・丁寧語を剥がす（「登録したい」→「登録」） */
const SUFFIX_STRIP =
  /(してください|してほしい|しています|できました|できます|されました|されます|しまいました|しました|します|ません|したい|ませ|です|ます|たい|って|とは|かな)$/u;

/** 検索語として意味を持たない一般語 */
const STOPWORDS = new Set([
  "方法",
  "手順",
  "使い方",
  "つかいかた",
  "やり方",
  "やりかた",
  "操作",
  "設定",
  "どう",
  "どうやって",
  "どこ",
  "何",
  "なに",
  "なん",
  "教えて",
  "おしえて",
  "ください",
  "です",
  "ます",
  "できますか",
  "可能",
]);

/**
 * 自然文の質問から検索キーワードを抽出する。
 * 助詞・文末表現・質問語を取り除き、2 文字以上の語を重複なく返す。
 */
export function extractSearchTerms(message: string): string[] {
  const seen = new Set<string>();
  for (const raw of message.split(TOKEN_SPLIT)) {
    const term = raw.replace(SUFFIX_STRIP, "");
    if (term.length >= 2 && !STOPWORDS.has(term) && !seen.has(term)) {
      seen.add(term);
    }
  }
  return [...seen];
}

/**
 * 質問全体 + 抽出キーワードで検索し、重複を除いた上位記事を
 * コンテキスト（タイトル・リンク用識別子・本文抜粋）として返す。
 */
export function buildChatContext(
  search: (query: string) => ManualArticle[],
  message: string,
  topK: number,
  excerptLength: number,
): SupportChatContextItem[] {
  const picked = new Map<string, ManualArticle>();
  const add = (results: ManualArticle[]) => {
    for (const article of results) {
      const key = `${article.category}/${article.slug}`;
      if (!picked.has(key)) {
        picked.set(key, article);
      }
    }
  };

  add(search(message));
  for (const term of extractSearchTerms(message)) {
    if (picked.size >= topK) break;
    add(search(term));
  }

  return [...picked.values()].slice(0, topK).map((a) => ({
    title: a.title,
    category: a.category,
    slug: a.slug,
    text: a.content.slice(0, excerptLength),
  }));
}
