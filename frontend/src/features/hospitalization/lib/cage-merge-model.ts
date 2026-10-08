import type { MasterItem } from "@/types";
import { isCageRepresentative as isCageRepresentativeShared } from "@/lib/cage-merge";
import { resolveCageNameById } from "@/lib/cage-name";

// EMR-306: 代表判定は shared predicate（src/lib/cage-merge.ts）に委譲する。
// 旧来の公開名を維持するため再exportする（Board・master 側の3箇所複写の統一）。
export { isCageRepresentativeShared as isCageRepresentative };

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
 * cageId が連結グループに属するか（構成メンバー or 代表）。
 * EMR-306 (Grok review HIGH-1): 連結グループが絡む入院の入れ替えは FE の2本の
 * 独立 PATCH が片方だけ成功し得るため、送信前の事前検知に使う。
 */
export function isInMergedGroup(
  cages: readonly MasterItem[],
  cageId: string | null | undefined,
): boolean {
  if (cageId == null || cageId === "") return false;
  const cage = cages.find((c) => String(c.id) === cageId);
  if (cage == null) return false;
  return cage.mergedIntoCageId != null || isCageRepresentativeShared(cages, cageId);
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
  /** このケージが占有中か（merge-mode の占有ルックアップ結果） */
  hasOccupant: boolean;
}): boolean {
  const { cages, cage, hasOccupant } = args;
  return (
    cage.mergedIntoCageId == null &&
    !isCageRepresentativeShared(cages, String(cage.id)) &&
    cage.status === "active" &&
    !hasOccupant
  );
}

/**
 * merge-mode の drop target 可否（advisory）。
 * - target が連結メンバー / 代表ケージ / dragged 自身 / 無効ケージ → 不可
 *   （BE は単一階層のみで既存グループへの member 追加を 409 拒否するため、
 *   代表を target として広告しない。グループ拡大は解除→再連結または
 *   マスタ画面の複数選択で行う）
 * - dragged が merge source でない → 不可
 * - dragged / target のどちらかが占有中 → 不可
 */
export function isMergeEligibleTarget(args: {
  cages: readonly MasterItem[];
  draggedCageId: string;
  targetCageId: string;
  /** cageId → 占有中かのルックアップ（admitted/reserved 全件基準が望ましい） */
  isOccupied: (cageId: string) => boolean;
}): boolean {
  const { cages, draggedCageId, targetCageId, isOccupied } = args;
  if (draggedCageId === targetCageId) return false;
  const dragged = cages.find((c) => String(c.id) === draggedCageId);
  const target = cages.find((c) => String(c.id) === targetCageId);
  if (!dragged || !target) return false;
  if (
    !isMergeDragSource({
      cages,
      cage: dragged,
      hasOccupant: isOccupied(draggedCageId),
    })
  )
    return false;
  if (target.mergedIntoCageId != null) return false;
  if (isCageRepresentativeShared(cages, targetCageId)) return false;
  if (target.status !== "active") return false;
  if (isOccupied(draggedCageId) || isOccupied(targetCageId)) return false;
  return true;
}

/**
 * merge-mode 中のケージカード視覚状態を一括算出する（draggable / droppable /
 * static mute・idle eligible / drag 中 highlight・mute）。全て advisory で、
 * 最終可否は BE が持つ。
 */
export function cageMergeCardState(args: {
  cages: readonly MasterItem[];
  cage: MasterItem;
  mergeMode: boolean;
  canMerge: boolean;
  /** カードに入院 occupant が載っているか（merge-mode 占有ルックアップ結果） */
  hasOccupant: boolean;
  /** merge-mode で drag 中の source cage id（未 drag 時は null） */
  activeMergeCageId: string | null;
  /** cageId → 占有中かのルックアップ */
  isOccupied: (cageId: string) => boolean;
}): {
  mergeDraggable: boolean;
  mergeTargetEligible: boolean;
  mergeStaticMuted: boolean;
  /** merge-mode・drag 未開始で source 適格 — 薄い ring で合体可能を常時示す */
  mergeIdleEligible: boolean;
  mergeDragHighlighted: boolean;
  mergeDragMuted: boolean;
} {
  const { cages, cage, mergeMode, canMerge, hasOccupant, activeMergeCageId, isOccupied } = args;
  const eligibleMergeSource = canMerge && isMergeDragSource({ cages, cage, hasOccupant });
  const mergeDraggable = mergeMode && eligibleMergeSource;
  // drop target 可否（advisory）。drag 未開始時は member・代表・無効ケージを
  // target から除外し、drag 中は source/occupant を含めて再判定する。
  const mergeTargetEligible =
    canMerge &&
    cage.mergedIntoCageId == null &&
    !isCageRepresentativeShared(cages, String(cage.id)) &&
    cage.status === "active" &&
    (activeMergeCageId === null ||
      isMergeEligibleTarget({
        cages,
        draggedCageId: activeMergeCageId,
        targetCageId: String(cage.id),
        isOccupied,
      }));
  return {
    mergeDraggable,
    mergeTargetEligible,
    mergeStaticMuted: mergeMode && activeMergeCageId === null && !eligibleMergeSource,
    mergeIdleEligible: mergeMode && activeMergeCageId === null && eligibleMergeSource,
    mergeDragHighlighted: mergeMode && activeMergeCageId !== null && mergeTargetEligible,
    mergeDragMuted: mergeMode && activeMergeCageId !== null && !mergeTargetEligible,
  };
}
