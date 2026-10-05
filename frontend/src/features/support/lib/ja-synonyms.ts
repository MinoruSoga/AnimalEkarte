/**
 * ja-synonyms.ts — JA 検索クエリの表記ゆれ・同義語グループ
 *
 * チャット質問側の語とマニュアル記事側の語がズレるときの橋渡し。
 * 「同じ意味として扱ってよい」表記ゆれ・同義語のみをまとめる
 * （関連語の羅列＝キーワードスタッフィングはしない）。
 *
 * グループ内の語は相互に言い換え可能として、検索語・根拠判定の両方で展開する。
 */

export const JA_SYNONYM_GROUPS: readonly (readonly string[])[] = [
  // 飼い主さん → 飼い主 は suffix strip で処理するため不要
  ["飼い主", "飼主", "オーナー"],
  ["作り", "作る", "作成"],
  ["新しい", "新規"],
  ["ワクチン", "ワクチン接種", "予防接種"],
  ["締め方", "締め", "レジ締め"],
  ["会計処理", "会計"],
  ["手続き", "手続", "手順"],
];

/**
 * 1 語を同義語グループへ展開する。どのグループにも属さなければその語だけを返す。
 * 戻り値は base term を先頭に含む。
 */
export function expandJaTerm(term: string): readonly string[] {
  for (const group of JA_SYNONYM_GROUPS) {
    if (group.includes(term)) {
      return [term, ...group.filter((w) => w !== term)];
    }
  }
  return [term];
}

/** 分かち書き辞書: 同義語グループに登録された語だけを知っている語とする */
const SEGMENT_DICT = new Set(JA_SYNONYM_GROUPS.flat());

/**
 * 形態素解析なしでの簡易分かち書き。
 * 助詞で区切られず結合したトークン（例:「新しい飼い主」）を、辞書内の語との
 * 最長一致で先頭から切り出す。辞書語を先頭に含まないトークンはそのまま返す。
 * 例: "新しい飼い主" → ["新しい","飼い主"], "ワクチン接種" → ["ワクチン接種"]
 */
export function segmentJaTerm(term: string): string[] {
  if (term.length < 2 || SEGMENT_DICT.has(term)) return [term];
  const parts: string[] = [];
  let rest = term;
  while (rest.length >= 2) {
    let matched: string | undefined;
    for (const w of SEGMENT_DICT) {
      if (rest.startsWith(w) && (matched === undefined || w.length > matched.length)) {
        matched = w;
      }
    }
    if (matched === undefined) break;
    parts.push(matched);
    rest = rest.slice(matched.length);
  }
  // 未消費部分は 1 つの語として残す（先頭辞書語が無かった場合は全体）
  if (rest.length > 0) parts.push(rest);
  // 末尾が 1 字だけ余った場合は前の語へ戻す（「新しい飼い主X」対策）
  if (parts.length > 1 && parts[parts.length - 1].length === 1) {
    const tail = parts.pop() ?? "";
    parts[parts.length - 1] += tail;
  }
  return parts;
}
