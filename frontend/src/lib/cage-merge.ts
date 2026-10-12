/**
 * EMR-306 ケージ一時連結: 代表/構成メンバー判定の shared predicate。
 * feature 間 import 禁止（frontend/CLAUDE.md）のため src/lib に置く
 * （resolveCageNameById の cage-name.ts と同じ昇格経緯）。
 * MasterItem / master Cage の両方が満たす最小構造で受ける。
 */
export interface CageMergeItem {
  id: string | number;
  /** 構成ケージが指す代表ケージ ID。wire の欠落・null は呼び出し側で正規化済み前提。 */
  mergedIntoCageId?: string | null;
}

/**
 * cageId が連結グループの代表ケージか（いずれかのケージの mergedIntoCageId が
 * この id を指す）。id は string/number union を string 正規化して比較する。
 */
export function isCageRepresentative(
  cages: readonly CageMergeItem[],
  cageId: string | number,
): boolean {
  const id = String(cageId);
  return cages.some((c) => c.mergedIntoCageId === id);
}

/**
 * 代表 cageId → 構成メンバー一覧（引数の順序を保持）。メンバーを持たない
 * 代表 id のキーは含まれない。ジェネリックで呼び出し側の要素型を保持する。
 */
export function cageMembersByRepresentative<T extends CageMergeItem>(
  cages: readonly T[],
): Map<string, T[]> {
  const membersByRep = new Map<string, T[]>();
  for (const cage of cages) {
    if (cage.mergedIntoCageId == null) continue;
    const members = membersByRep.get(cage.mergedIntoCageId);
    if (members) {
      members.push(cage);
    } else {
      membersByRep.set(cage.mergedIntoCageId, [cage]);
    }
  }
  return membersByRep;
}

export interface CageMergeSelectableItem extends CageMergeItem {
  isActive: boolean;
}

export function isCageMergeEligible(
  cage: CageMergeSelectableItem,
  occupiedCageIds: ReadonlySet<string> | undefined,
  membersByRepresentativeId: ReadonlyMap<string, readonly unknown[]>,
): boolean {
  const id = String(cage.id);
  return (
    occupiedCageIds !== undefined &&
    cage.isActive === true &&
    cage.mergedIntoCageId == null &&
    !membersByRepresentativeId.has(id) &&
    !occupiedCageIds.has(id)
  );
}
