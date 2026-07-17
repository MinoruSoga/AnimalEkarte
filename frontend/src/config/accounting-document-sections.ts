// #190: 帳票レイアウト設定用定数 — feature 非依存の共有定数（R-F2-S9: accounting から抽出）
export const DOCUMENT_SECTION_KEYS = [
  "clinic_header",
  "owner_pet_info",
  "items_table",
  "payment_summary",
  "footer_note",
] as const;

export type DocumentSectionKey = (typeof DOCUMENT_SECTION_KEYS)[number];

export const DOCUMENT_SECTION_LABELS: Record<DocumentSectionKey, string> = {
  clinic_header: "病院情報ヘッダー",
  owner_pet_info: "飼主・ペット情報",
  items_table: "明細テーブル",
  payment_summary: "お会計サマリー",
  footer_note: "備考・フッター",
};
