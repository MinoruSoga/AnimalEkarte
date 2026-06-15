import { normalizeKana } from "./normalize-kana";
import type { Pet } from "@/types";

/**
 * ペット行が検索語にヒットするか判定する。
 * カタカナ↔ひらがな区別なし (normalizeKana で統一) + 大文字小文字区別なし。
 * 検索対象: ownerName / ownerNameKana / ownerNumber / pet.name / petNameKana / species
 */
export function matchesPetSearch(pet: Pet, term: string): boolean {
  if (!term) return true;

  const normalize = (s: string) => normalizeKana(s).toLowerCase();
  const normalizedTerm = normalize(term);
  const ownerNumberStr = pet.ownerNumber?.toString() ?? "";

  return (
    normalize(pet.ownerName).includes(normalizedTerm) ||
    normalize(pet.ownerNameKana ?? "").includes(normalizedTerm) ||
    ownerNumberStr.includes(term) || // ownerNumber は数字列なのでカナ正規化不要 (raw term で比較)
    normalize(pet.name).includes(normalizedTerm) ||
    normalize(pet.petNameKana ?? "").includes(normalizedTerm) ||
    (pet.species != null && normalize(pet.species).includes(normalizedTerm))
  );
}
