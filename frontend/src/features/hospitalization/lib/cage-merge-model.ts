import type { MasterItem } from "@/types";
import { resolveCageNameById } from "@/lib/cage-name";

/**
 * EMR-306 ケージ一時連結: 入院ボード merge mode の pure model。
 * merge-mode drag source id は `merge-cage-<id>` — patient drag の id
 * （hospitalization id）や drop target の `cage-<id>` と衝突しない。
 */
export const MERGE_CAGE_DRAG_PREFIX = "merge-cage-";

export function mergeCageDragId(cageId: string): string {
  return `${MERGE_CAGE_DRAG_PREFIX}${cageId}`;
}

/** 連結メンバーケージか（mergedIntoCageId が代表ケージを指す）。 */
export function isMergedMember(cage: MasterItem): boolean {
  return cage.mergedIntoCageId != null;
}

/**
 * 代表ケージか: いずれかのケージの mergedIntoCageId がこの cageId を指す場合。
 * 両辺とも string で比較する（transform 後の mergedIntoCageId は string、cage.id も string）。
 */
export function isCageRepresentative(cages: readonly MasterItem[], cageId: string): boolean {
  return cages.some((c) => c.mergedIntoCageId === cageId);
}

/**
 * 連結中 badge 表示用の代表ケージ名。代表ケージが一覧に見つからない場合は
 * `ID <id>` にフォールバックする。
 */
export function resolveRepresentativeName(
  cages: readonly MasterItem[],
  mergedIntoCageId: string,
): string {
  return resolveCageNameById(cages, mergedIntoCageId);
}

/**
 * merge-mode の drag source 可否（advisory）。連結メンバー・代表ケージ・
 * 占有中・無効ケージは source にできない。最終判定は BE（409）が行う。
 */
export function isMergeDragSource(args: {
  cages: readonly MasterItem[];
  cage: MasterItem;
  hasVisibleOccupant: boolean;
}): boolean {
  const { cages, cage, hasVisibleOccupant } = args;
  return (
    cage.mergedIntoCageId == null &&
    !isCageRepresentative(cages, String(cage.id)) &&
    cage.status === "active" &&
    !hasVisibleOccupant
  );
}

/**
 * merge-mode の drop target 可否（advisory）。
 * - target が連結メンバー / 代表ケージ / dragged 自身 / 無効ケージ → 不可
 *   （BE は単一階層のみで既存グループへの member 追加を 409 拒否するため、
 *   代表を target として広告しない。グループ拡大は解除→再連結または
 *   マスタ画面の複数選択で行う）
 * - dragged が merge source でない → 不可
 * - dragged / target のどちらかに可視の入院 occupant → 不可
 */
export function isMergeEligibleTarget(args: {
  cages: readonly MasterItem[];
  draggedCageId: string;
  targetCageId: string;
  hasVisibleOccupant: (cageId: string) => boolean;
}): boolean {
  const { cages, draggedCageId, targetCageId, hasVisibleOccupant } = args;
  if (draggedCageId === targetCageId) return false;
  const dragged = cages.find((c) => String(c.id) === draggedCageId);
  const target = cages.find((c) => String(c.id) === targetCageId);
  if (!dragged || !target) return false;
  if (
    !isMergeDragSource({
      cages,
      cage: dragged,
      hasVisibleOccupant: hasVisibleOccupant(draggedCageId),
    })
  )
    return false;
  if (target.mergedIntoCageId != null) return false;
  if (isCageRepresentative(cages, targetCageId)) return false;
  if (target.status !== "active") return false;
  if (hasVisibleOccupant(draggedCageId) || hasVisibleOccupant(targetCageId)) return false;
  return true;
}

/**
 * merge-mode 中のケージカード視覚状態を一括算出する（draggable / droppable /
 * static mute / drag 中 highlight・mute の5状態）。全て advisory で、最終可否は BE が持つ。
 */
export function cageMergeCardState(args: {
  cages: readonly MasterItem[];
  cage: MasterItem;
  mergeMode: boolean;
  canMerge: boolean;
  /** カードに入院 occupant が載っているか（visible な一覧基準） */
  hasOccupant: boolean;
  /** merge-mode で drag 中の source cage id（未 drag 時は null） */
  activeMergeCageId: string | null;
  hasVisibleOccupant: (cageId: string) => boolean;
}): {
  mergeDraggable: boolean;
  mergeTargetEligible: boolean;
  mergeStaticMuted: boolean;
  mergeDragHighlighted: boolean;
  mergeDragMuted: boolean;
} {
  const { cages, cage, mergeMode, canMerge, hasOccupant, activeMergeCageId, hasVisibleOccupant } =
    args;
  const participates =
    canMerge && isMergeDragSource({ cages, cage, hasVisibleOccupant: hasOccupant });
  const mergeDraggable = mergeMode && participates;
  // drop target 可否（advisory）。drag 未開始時は member・代表・無効ケージを
  // target から除外し、drag 中は source/occupant を含めて再判定する。
  const mergeTargetEligible =
    canMerge &&
    cage.mergedIntoCageId == null &&
    !isCageRepresentative(cages, String(cage.id)) &&
    cage.status === "active" &&
    (activeMergeCageId === null ||
      isMergeEligibleTarget({
        cages,
        draggedCageId: activeMergeCageId,
        targetCageId: String(cage.id),
        hasVisibleOccupant,
      }));
  return {
    mergeDraggable,
    mergeTargetEligible,
    mergeStaticMuted: mergeMode && activeMergeCageId === null && !participates,
    mergeDragHighlighted: mergeMode && activeMergeCageId !== null && mergeTargetEligible,
    mergeDragMuted: mergeMode && activeMergeCageId !== null && !mergeTargetEligible,
  };
}
