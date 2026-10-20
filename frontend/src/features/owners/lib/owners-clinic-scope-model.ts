/**
 * EMR-463: 飼主・ペット一覧の有効な医院選択を決める純関数群。
 * 契約: docs/design/domains/owners-pets/technical-spec.md#owners-list-scope-contract-emr463
 *
 * 決定順: URL 適用済み選択 ∩ 現在許可 → 保存選択 ∩ 現在許可 → 現在の全許可。
 * - 許可集合が `[]` なら一覧 API を呼ばない（空選択を clinic_ids 省略へ変換すると現在医院へ戻るため）。
 * - 選択が GET /pets `clinic_ids` の上限（50）を超えたら超過として停止する。
 *   先頭 50 件への切捨て・分割取得はしない。
 * - URL・保存値に含まれる許可外 ID は送らない。ID は文字列のまま扱う。
 */

/** GET /pets `clinic_ids` のカンマ区切り最大件数（既存 API 契約）。 */
export const OWNERS_CLINIC_SELECTION_LIMIT = 50;

export type OwnersClinicSelectionSource = "url" | "saved" | "all";

export type OwnersClinicSelectionDecision =
  | { kind: "no-permitted-clinics" }
  | {
      kind: "over-limit";
      selectedClinicIds: string[];
      source: OwnersClinicSelectionSource;
      limit: number;
    }
  | { kind: "ready"; selectedClinicIds: string[]; source: OwnersClinicSelectionSource };

/** URL `?clinics=` を ID 配列にする。未指定・空は null（未適用）。 */
export function parseClinicsParam(param: string | null): string[] | null {
  if (param === null) return null;
  const ids = param
    .split(",")
    .map((id) => id.trim())
    .filter((id, index, all) => id !== "" && all.indexOf(id) === index);
  return ids.length > 0 ? ids : null;
}

/** 候補の順序を保ったまま、現在許可集合に含まれる ID だけを残す。 */
export function intersectWithPermitted(
  candidates: readonly string[],
  permittedClinicIds: readonly string[],
): string[] {
  const permitted = new Set(permittedClinicIds);
  const result: string[] = [];
  for (const id of candidates) {
    if (permitted.has(id) && !result.includes(id)) result.push(id);
  }
  return result;
}

export function decideOwnersClinicSelection({
  urlClinicIds,
  savedClinicIds,
  permittedClinicIds,
  limit = OWNERS_CLINIC_SELECTION_LIMIT,
}: {
  urlClinicIds: readonly string[] | null;
  savedClinicIds: readonly string[] | null;
  permittedClinicIds: readonly string[];
  limit?: number;
}): OwnersClinicSelectionDecision {
  if (permittedClinicIds.length === 0) return { kind: "no-permitted-clinics" };

  let selectedClinicIds: string[] = [];
  let source: OwnersClinicSelectionSource = "all";
  const fromUrl = urlClinicIds ? intersectWithPermitted(urlClinicIds, permittedClinicIds) : [];
  if (fromUrl.length > 0) {
    selectedClinicIds = fromUrl;
    source = "url";
  } else {
    const fromSaved = savedClinicIds
      ? intersectWithPermitted(savedClinicIds, permittedClinicIds)
      : [];
    if (fromSaved.length > 0) {
      selectedClinicIds = fromSaved;
      source = "saved";
    } else {
      selectedClinicIds = [...permittedClinicIds];
    }
  }

  if (selectedClinicIds.length > limit) {
    return { kind: "over-limit", selectedClinicIds, source, limit };
  }
  return { kind: "ready", selectedClinicIds, source };
}

/** 2 つの選択が同じ医院集合か（順序は問わない）。 */
export function isSameClinicSet(a: readonly string[], b: readonly string[]): boolean {
  if (a.length !== b.length) return false;
  const set = new Set(a);
  return b.every((id) => set.has(id));
}

/**
 * 明示トグル後の選択。許可外 ID はトグルできず、結果が空になる操作は無効（null）。
 * 既存の現在選択に含まれる許可外 ID も落とす。
 */
export function toggleClinicSelection(
  currentClinicIds: readonly string[],
  clinicId: string,
  permittedClinicIds: readonly string[],
): string[] | null {
  if (!permittedClinicIds.includes(clinicId)) return null;
  const current = intersectWithPermitted(currentClinicIds, permittedClinicIds);
  const next = current.includes(clinicId)
    ? current.filter((id) => id !== clinicId)
    : [...current, clinicId];
  return next.length > 0 ? next : null;
}

/**
 * 現在許可の取得に成功したときだけ使う、保存選択の失権除外。
 * 通信障害・契約未提供は失権の証明ではないため、この関数を呼ばない。
 * - unchanged: 保存なし、または全件が現在許可内
 * - replace: 許可外を除いた残りで上書き
 * - remove: 全件が許可外（次回は現在の全許可が初期値になる）
 */
export type SavedSelectionPruneAction =
  | { kind: "unchanged" }
  | { kind: "replace"; clinicIds: string[] }
  | { kind: "remove" };

export function pruneSavedSelection(
  savedClinicIds: readonly string[] | null,
  permittedClinicIds: readonly string[],
): SavedSelectionPruneAction {
  if (savedClinicIds === null) return { kind: "unchanged" };
  const kept = intersectWithPermitted(savedClinicIds, permittedClinicIds);
  if (kept.length === savedClinicIds.length) return { kind: "unchanged" };
  if (kept.length === 0) return { kind: "remove" };
  return { kind: "replace", clinicIds: kept };
}
