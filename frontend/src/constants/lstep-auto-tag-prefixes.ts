// Lステップ自動管理タグのプレフィックス定数
// BE-019 autoManagedPrefixes と同期。FE-002 / FE-005 / FE-007 が共通インポート。
// この定数はクロス feature インポートを避けるため src/constants/ に配置する（features/ 配下禁止）。

export const LSTEP_AUTO_TAG_PREFIXES = [
  "cpm_",
  "vaccine_",
  "last_visit_",
  "ltv_",
  "checkup_",
  "pet_birthday_",
  "has_",
  "dormant_",
  "visit_count_",
  "pet_species_",
] as const;

export type LstepAutoTagPrefix = (typeof LSTEP_AUTO_TAG_PREFIXES)[number];

/** 指定タグ名が自動管理タグかどうか判定する */
export function isAutoManagedTag(tagName: string): boolean {
  return LSTEP_AUTO_TAG_PREFIXES.some((prefix) => tagName.startsWith(prefix));
}

/** 禁止プレフィックスを返す（バリデーションエラー表示用） */
export function getForbiddenPrefix(tagName: string): string | null {
  const matched = LSTEP_AUTO_TAG_PREFIXES.find((prefix) =>
    tagName.startsWith(prefix)
  );
  return matched ?? null;
}
