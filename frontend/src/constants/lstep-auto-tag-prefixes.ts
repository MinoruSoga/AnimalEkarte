// Lステップ自動管理タグエントリ一覧
// BE-019 autoManagedPrefixes と完全同期（lstep_tag_service.go）。
// FE-REOPEN-002: BE の完全一致+前方一致ロジックに合わせて更新。
// この定数はクロス feature インポートを避けるため src/constants/ に配置する（features/ 配下禁止）。

const LSTEP_AUTO_TAG_ENTRIES = [
  // prefix entries (trailing _) — 新仕様タグ (FEAT-377/378: 日本語・接頭辞付き)
  "CPM_",
  "LTV_",
  "VISIT_",
  "PET_",
  "EXCL_",
  // prefix entries (trailing _) — 既存タグ (snake_case 英語)
  "cpm_",
  "last_visit_",
  "first_visit_",
  "ltv_",
  "visit_count_",
  "vaccine_",
  "checkup_",
  "next_checkup_",
  "refill_due_",
  "next_visit_",
  "reserved_",
  "no_show_",
  "breed_",
  "sex_",
  "pet_birthday_",
  "birth_year_",
  "chronic_",
  "dormant_",
  "cert_sent_",
  "post_surgery_",
  "post_discharge_",
  "pet_deceased_",
  // exact-match entries (no trailing _)
  "canceled_visit",
  "has_dog",
  "has_cat",
  "has_both",
  "spay_neutered",
  "intact",
] as const;

/** 指定タグ名が自動管理タグかどうか判定する（BE の isAutoManagedTag と同一ロジック） */
export function isAutoManagedTag(tagName: string): boolean {
  return LSTEP_AUTO_TAG_ENTRIES.some((entry) => tagName === entry || tagName.startsWith(entry));
}

/** 一致した禁止エントリを返す（バリデーションエラー表示用） */
export function getForbiddenPrefix(tagName: string): string | null {
  const matched = LSTEP_AUTO_TAG_ENTRIES.find(
    (entry) => tagName === entry || tagName.startsWith(entry),
  );
  return matched ?? null;
}
