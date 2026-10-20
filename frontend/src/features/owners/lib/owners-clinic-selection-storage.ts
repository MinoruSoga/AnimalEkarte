import { CANONICAL_CLINIC_ID_PATTERN } from "../api/get-owners-view-scope";

/**
 * EMR-463: 飼主・ペット一覧の医院選択の保存先（localStorage）。
 *
 * - key は `owners-list.clinics.v1.<user.id>`。key へ医院 ID を含めない。別ユーザーの key は読まない。
 * - 保存内容は選択した医院 ID（正規形 10 進文字列）の配列だけ。許可集合・患者／飼主情報は保存しない。
 * - 保存値は権限ではない。読み出し後に必ず現在の許可集合と交差して使う。
 * - storage が使えない環境（private mode・blocked 等）では例外を投げず「未保存」として扱う。
 */

const STORAGE_KEY_PREFIX = "owners-list.clinics.v1.";

export function ownersClinicSelectionStorageKey(userId: string): string {
  return `${STORAGE_KEY_PREFIX}${userId}`;
}

function resolveStorage(storage: Storage | null | undefined): Storage | null {
  if (storage !== undefined) return storage;
  try {
    return typeof window === "undefined" ? null : window.localStorage;
  } catch {
    return null;
  }
}

/** 正規形 ID の重複なし配列だけを受理する。それ以外の保存値は未保存（null）として扱う。 */
function parseStoredSelection(raw: string): string[] | null {
  let parsed: unknown;
  try {
    parsed = JSON.parse(raw);
  } catch {
    return null;
  }
  if (!Array.isArray(parsed)) return null;
  const ids: string[] = [];
  for (const item of parsed) {
    if (typeof item !== "string" || !CANONICAL_CLINIC_ID_PATTERN.test(item)) return null;
    if (!ids.includes(item)) ids.push(item);
  }
  return ids;
}

/** 保存済み選択を読む。未保存・読取不能・不正形は null。 */
export function readOwnersClinicSelection(
  userId: string,
  storage?: Storage | null,
): string[] | null {
  if (userId === "") return null;
  const target = resolveStorage(storage);
  if (!target) return null;
  try {
    const raw = target.getItem(ownersClinicSelectionStorageKey(userId));
    return raw === null ? null : parseStoredSelection(raw);
  } catch {
    return null;
  }
}

/** 選択を保存する。正規形でない ID は保存しない。成功時 true。 */
export function writeOwnersClinicSelection(
  userId: string,
  clinicIds: readonly string[],
  storage?: Storage | null,
): boolean {
  if (userId === "") return false;
  const target = resolveStorage(storage);
  if (!target) return false;
  const ids = clinicIds.filter(
    (id, index) => CANONICAL_CLINIC_ID_PATTERN.test(id) && clinicIds.indexOf(id) === index,
  );
  try {
    target.setItem(ownersClinicSelectionStorageKey(userId), JSON.stringify(ids));
    return true;
  } catch {
    return false;
  }
}

/** 保存済み選択を削除する（全件が現在許可外になったときだけ使う）。成功時 true。 */
export function removeOwnersClinicSelection(userId: string, storage?: Storage | null): boolean {
  if (userId === "") return false;
  const target = resolveStorage(storage);
  if (!target) return false;
  try {
    target.removeItem(ownersClinicSelectionStorageKey(userId));
    return true;
  } catch {
    return false;
  }
}
