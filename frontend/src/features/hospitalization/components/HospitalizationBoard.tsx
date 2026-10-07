// React/Framework
import { memo, useCallback, useMemo } from "react";

// External
import {
  closestCorners,
  DndContext,
  KeyboardSensor,
  PointerSensor,
  useDndContext,
  useSensor,
  useSensors,
  type DragEndEvent,
} from "@dnd-kit/core";
import { useDraggable, useDroppable } from "@dnd-kit/core";
import { Plus, GripVertical, Link2 } from "lucide-react";

// Internal
import { C, ICON, BADGE } from "@/lib/design-tokens";
import { formatDate } from "@/lib/format/date";
import { Card, CardHeader, CardContent } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { getHospitalizationTypeColor } from "@/lib/status-helpers";

// Relative
import { cageKeyboardCoordinateGetter } from "../lib/cage-keyboard-coordinates";
import {
  cageMergeCardState,
  isCageRepresentative,
  isMergeEligibleTarget,
  isMergedMember,
  mergeCageDragId,
  resolveRepresentativeName,
} from "../lib/cage-merge-model";
import { HOSPITALIZATION_STATUS } from "../constants";
import { H_STYLES } from "../lib/styles";

// Types
import type { MasterItem, Hospitalization } from "@/types";

interface HospitalizationBoardProps {
  cages: MasterItem[];
  hospitalizations: Hospitalization[];
  onNavigateToForm: (id?: string) => void;
  onMovePet: (hospitalizationId: string, targetCageId: string) => void;
  canCreate: boolean;
  canEdit: boolean;
  /** EMR-306: true のとき pet-move DnD を止め、ケージ連結 DnD に切り替える */
  mergeMode?: boolean;
  /** master-hospitalization:edit 権限（連結/解除操作のゲート） */
  canMerge?: boolean;
  /**
   * EMR-306 (Grok review MEDIUM-2): merge-mode の占有判定。呼び出し元が
   * admitted/reserved 全件（タブ・ページ非依存）のルックアップを渡す。
   * 未提供時は表示中の入院データにフォールバック（advisory。最終可否は BE）。
   */
  hasMergeOccupant?: (cageId: string) => boolean;
  /** merge-mode で eligible target への drop が確定したとき呼ばれる（dragged, target） */
  onRequestMerge?: (draggedCageId: string, targetCageId: string) => void;
  /** 連結グループの解除要求。常に代表ケージ ID が渡る */
  onRequestUnmerge?: (representativeId: string) => void;
}

interface CageCardProps {
  cage: MasterItem;
  /** EMR-306: merge 判定（代表検出・member 解決）に使う全ケージ一覧 */
  cages: readonly MasterItem[];
  occupant?: Hospitalization;
  /** member cage の場合の代表ケージ名（board 側で解決済み） */
  mergedRepresentativeName?: string;
  hasVisibleOccupant: (cageId: string) => boolean;
  onNavigateToForm: (id?: string) => void;
  onRequestUnmerge?: (representativeId: string) => void;
  canCreate: boolean;
  canEdit: boolean;
  mergeMode: boolean;
  canMerge: boolean;
}

const CageCard = memo(function CageCard({
  cage,
  cages,
  occupant,
  mergedRepresentativeName,
  hasVisibleOccupant,
  onNavigateToForm,
  onRequestUnmerge,
  canCreate,
  canEdit,
  mergeMode,
  canMerge,
}: CageCardProps) {
  const isDeceased = occupant?.petIsDeceased ?? false;
  // EMR-227: 退院済み入院はケージ移動（書込み）を render 側でも禁止する。
  const isDischarged = occupant?.status === HOSPITALIZATION_STATUS.DISCHARGED;
  const mergedMember = isMergedMember(cage);
  // EMR-306: 代表ケージも連結グループの一員 — カードにも連結中表示を持たせる。
  const isRepresentative = isCageRepresentative(cages, String(cage.id));
  const cageContext = cage.category ? `${cage.category} ${cage.name}` : cage.name;
  const emptyCageActionLabel = `${cageContext}（ケージID: ${cage.id}）の空き枠に入院・ホテルを登録`;
  const mergeDragLabel = `${cageContext}（ケージID: ${cage.id}）を連結元としてドラッグ`;
  const canDrag = !mergeMode && Boolean(occupant) && !isDeceased && !isDischarged && canEdit;
  // FE-RC-044: 詳細への遷移は :127-138 の `詳細` button（aria-label 付き）に一本化する。
  // Card 自体は非フォーカス可能・キーボード操作不可のため、同じ操作を onClick で重複させない。
  const canShowDetailButton = Boolean(occupant) && !isDeceased && canEdit;

  const { active } = useDndContext();
  const activeMergeCageId =
    mergeMode && typeof active?.data.current?.mergeCageId === "string"
      ? active.data.current.mergeCageId
      : null;

  // EMR-306: merge mode では cage card 自体が連結 source になる（patient drag は無効）。
  // draggable / droppable / static mute / drag 中 highlight・mute の5状態は
  // cage-merge-model の純粋関数に集約する。
  const {
    mergeDraggable,
    mergeTargetEligible,
    mergeStaticMuted,
    mergeDragHighlighted,
    mergeDragMuted,
  } = cageMergeCardState({
    cages,
    cage,
    mergeMode,
    canMerge,
    // カード単位の占有も同一ルックアップを使う — 他タブ/別ページの占有のみ持つ
    // ケージが merge source として draggable に見えないよう整合させる。
    hasOccupant: hasVisibleOccupant(String(cage.id)),
    activeMergeCageId,
    hasVisibleOccupant,
  });
  const dragEnabled = mergeMode ? mergeDraggable : canDrag;

  const {
    attributes,
    listeners,
    setNodeRef: setDragRef,
    isDragging,
  } = useDraggable({
    // EMR-306: merge mode では `merge-cage-` prefix の別 id 系を使い、
    // patient drag（hospitalization id）と衝突しない。
    id: mergeMode ? mergeCageDragId(String(cage.id)) : (occupant?.id ?? `empty-${cage.id}`),
    data: mergeMode ? { mergeCageId: String(cage.id) } : { hospitalizationId: occupant?.id },
    disabled: !dragEnabled,
  });

  const { setNodeRef: setDropRef, isOver } = useDroppable({
    id: `cage-${cage.id}`,
    data: { cageId: cage.id },
    // EMR-306: 連結メンバーケージは pet-move drop も受け付けない。
    disabled: mergeMode ? !mergeTargetEligible : !canEdit || mergedMember,
  });

  return (
    <div ref={setDropRef} className="h-full">
      <Card
        ref={setDragRef}
        {...(dragEnabled ? attributes : {})}
        {...(dragEnabled ? listeners : {})}
        {...(mergeDraggable ? { "aria-label": mergeDragLabel } : {})}
        className={`relative flex flex-col h-40 transition-all border touch-none
                  ${
                    occupant
                      ? isDeceased
                        ? `${C.bgPage} border-l-4 ${C.borderPrimary20}`
                        : `${C.bgWhite} border-l-4 ${C.borderLMedicalBlue}`
                      : `${C.bgPage} border-dashed ${C.borderPrimary20}`
                  }
                  ${isDragging ? "opacity-50 scale-95" : "hover:shadow-level1"}
                  ${isOver ? `ring-2 ${C.ringMedicalBlue} ring-offset-2 ${C.bgMedicalBlue5}` : ""}
                  ${dragEnabled ? "cursor-grab active:cursor-grabbing" : "cursor-default"}
                  ${mergeDragHighlighted ? `ring-2 ${C.ringMedicalBlue} ring-offset-2` : ""}
                  ${mergeStaticMuted || mergeDragMuted ? "opacity-50" : ""}
                `}
      >
        <CardHeader
          className={`${H_STYLES.padding.card} pb-0 flex flex-row items-center justify-between space-y-0`}
        >
          <div className="flex items-center gap-1">
            {canDrag ? (
              <div className={`cursor-grab active:cursor-grabbing ${C.text20} ${C.hoverText60}`}>
                <GripVertical className={ICON.action} />
              </div>
            ) : mergeDraggable ? (
              <div className={`cursor-grab active:cursor-grabbing ${C.text20} ${C.hoverText60}`}>
                <Link2 className={ICON.action} />
              </div>
            ) : null}
            <span className={`${H_STYLES.text.sm} font-mono ${C.text60} font-semibold`}>
              {cage.name}
            </span>
          </div>
          {occupant ? (
            <Badge
              variant="outline"
              className={`${getHospitalizationTypeColor(occupant.hospitalizationType)} ${H_STYLES.text.xs} px-1.5 py-0 h-5 border-none`}
            >
              {occupant.hospitalizationType}
            </Badge>
          ) : null}
        </CardHeader>
        <CardContent
          className={`${H_STYLES.padding.card} flex-1 flex flex-col justify-center items-start text-left`}
        >
          {mergedMember ? (
            <div className="mb-1 flex w-full flex-col items-start gap-1">
              <Badge
                variant="secondary"
                className={`${H_STYLES.text.xs} px-1.5 py-0 h-5 font-medium`}
              >
                {`連結中: ${mergedRepresentativeName ?? `ID ${cage.mergedIntoCageId}`}`}
              </Badge>
              {canMerge && onRequestUnmerge ? (
                <button
                  type="button"
                  aria-label={`${mergedRepresentativeName ?? "代表ケージ"}への連結を解除`}
                  className={`flex min-h-8 items-center justify-center text-2xs ${C.textBrand} border ${C.borderBrandLight} rounded px-1.5 ${C.hoverBgBrandLight60} transition-colors`}
                  onClick={(event) => {
                    event.stopPropagation();
                    if (cage.mergedIntoCageId != null) {
                      onRequestUnmerge(cage.mergedIntoCageId);
                    }
                  }}
                >
                  連結解除
                </button>
              ) : null}
            </div>
          ) : isRepresentative ? (
            // EMR-306: 代表ケージにも連結中表示を持たせる（メンバーだけでは
            // グループの見た目が片寄る）。解除はグループ全体に効くため代表からも行える。
            <div className="mb-1 flex w-full flex-col items-start gap-1">
              <Badge
                variant="secondary"
                className={`${H_STYLES.text.xs} px-1.5 py-0 h-5 font-medium`}
              >
                連結中（代表）
              </Badge>
              {canMerge && onRequestUnmerge ? (
                <button
                  type="button"
                  aria-label={`${cage.name}の連結を解除`}
                  className={`flex min-h-8 items-center justify-center text-2xs ${C.textBrand} border ${C.borderBrandLight} rounded px-1.5 ${C.hoverBgBrandLight60} transition-colors`}
                  onClick={(event) => {
                    event.stopPropagation();
                    onRequestUnmerge(String(cage.id));
                  }}
                >
                  連結解除
                </button>
              ) : null}
            </div>
          ) : null}
          {occupant ? (
            <>
              {occupant.startDate ? (
                <div className={`text-xs font-mono ${C.text60} w-full`}>
                  {formatDate(occupant.startDate)}
                </div>
              ) : null}
              <div
                className={`font-bold ${C.text} ${H_STYLES.text.sm} truncate w-full`}
                title={occupant.ownerName}
              >
                {occupant.ownerName}
              </div>
              <div className={`${H_STYLES.text.sm} ${C.text} truncate w-full`}>
                {[occupant.species, occupant.petName].filter(Boolean).join(" ")}
              </div>
              {isDeceased ? (
                // EMR-227: 死亡センチネルは text40 の素テキストではなくバッジで明示する（カードの可読性維持）
                <Badge
                  variant="outline"
                  className={`${BADGE.gray} ${H_STYLES.text.xs} px-1.5 py-0 h-5 font-medium`}
                >
                  死亡
                </Badge>
              ) : null}
              {canShowDetailButton ? (
                <button
                  type="button"
                  aria-label={`${occupant.petName}の詳細`}
                  className={`mt-2 flex min-h-11 items-center justify-center gap-1 text-2xs ${C.textBrand} ${C.bgBrandLight30} border ${C.borderBrandLight} rounded px-1.5 ${C.hoverBgBrandLight60} transition-colors`}
                  onClick={(event) => {
                    event.stopPropagation();
                    onNavigateToForm(occupant.id);
                  }}
                >
                  詳細
                </button>
              ) : null}
            </>
          ) : (
            <div className={`flex flex-col items-center justify-center h-full ${C.text20}`}>
              <span className={H_STYLES.text.sm}>空き</span>
              {canCreate && !mergedMember ? (
                <Button
                  variant="ghost"
                  size="icon"
                  aria-label={emptyCageActionLabel}
                  title={emptyCageActionLabel}
                  className={`h-11 w-11 mt-1 rounded-full ${C.hoverBgPrimary10} ${C.hoverText60}`}
                  onClick={(e) => {
                    e.stopPropagation();
                    onNavigateToForm();
                  }}
                >
                  <Plus className={H_STYLES.button.icon} />
                </Button>
              ) : null}
            </div>
          )}
        </CardContent>
      </Card>
    </div>
  );
});

export const HospitalizationBoard = memo(function HospitalizationBoard({
  cages,
  hospitalizations,
  onNavigateToForm,
  onMovePet,
  canCreate,
  canEdit,
  mergeMode = false,
  canMerge = false,
  hasMergeOccupant,
  onRequestMerge,
  onRequestUnmerge,
}: HospitalizationBoardProps) {
  const sensors = useSensors(
    useSensor(PointerSensor, { activationConstraint: { distance: 8 } }),
    useSensor(KeyboardSensor, { coordinateGetter: cageKeyboardCoordinateGetter }),
  );

  // js-index-maps: cageId → Hospitalization の Map を事前構築（O(n)）しレンダーループ内で O(1) 検索
  // 親コンポーネントがタブに応じて既にフィルタリング済みのデータを渡すため、ここでは再フィルタしない
  const occupantByCageId = useMemo(
    () => new Map(hospitalizations.map((h) => [h.cageId, h])),
    [hospitalizations],
  );

  // EMR-306: merge-mode の占有判定は hasMergeOccupant（admitted/reserved 全件・
  // 表示タブ/ページ非依存）を優先し、未提供時は表示中データへフォールバックする。
  const hasVisibleOccupant = useCallback(
    (cageId: string) => hasMergeOccupant?.(cageId) ?? occupantByCageId.has(cageId),
    [hasMergeOccupant, occupantByCageId],
  );

  const handleDragEnd = useCallback(
    (event: DragEndEvent) => {
      const { active, over } = event;
      if (!over) return;

      // EMR-306: merge mode では cage merge drop を確認ダイアログ要求に変換する。
      if (mergeMode) {
        if (!canMerge || !onRequestMerge) return;
        const draggedCageId = active.data.current?.mergeCageId;
        if (typeof draggedCageId !== "string" || !draggedCageId) return;
        const overId = over.id;
        if (typeof overId !== "string" || !overId.startsWith("cage-")) return;
        const targetCageId = overId.slice("cage-".length);
        const eligible = isMergeEligibleTarget({
          cages,
          draggedCageId,
          targetCageId,
          hasVisibleOccupant,
        });
        if (!eligible) return;
        onRequestMerge(draggedCageId, targetCageId);
        return;
      }

      if (!canEdit) return;

      // FE-RC-037: dnd-kit の data/id は unknown 相当のため、無検証キャストせず typeof で確認する。
      const hospitalizationId = active.data.current?.hospitalizationId;
      if (typeof hospitalizationId !== "string" || !hospitalizationId) return;

      const overId = over.id;
      if (typeof overId !== "string" || !overId.startsWith("cage-")) return;

      const targetCageId = overId.replace("cage-", "");
      // EMR-306: 連結メンバーケージへの pet-move drop を拒否する（droppable disabled の二重ガード）。
      const targetCage = cages.find((c) => String(c.id) === targetCageId);
      if (targetCage != null && targetCage.mergedIntoCageId != null) return;
      onMovePet(hospitalizationId, targetCageId);
    },
    [canEdit, canMerge, mergeMode, cages, hasVisibleOccupant, onMovePet, onRequestMerge],
  );

  // Group cages by category (Area)
  const cagesByArea = cages.reduce(
    (acc, cage) => {
      const area = cage.category || "その他";
      if (!acc[area]) acc[area] = [];
      acc[area].push(cage);
      return acc;
    },
    {} as Record<string, MasterItem[]>,
  );

  return (
    <DndContext sensors={sensors} collisionDetection={closestCorners} onDragEnd={handleDragEnd}>
      <div className="space-y-6 pb-4">
        {Object.entries(cagesByArea).map(([area, areaCages]) => (
          // No min-w-[800px]: reflow 1/2/3+ columns by container width (BUG-458).
          <div key={area} className="min-w-0">
            <h3
              className={`${H_STYLES.text.lg} font-bold ${C.text} mb-3 border-b pb-1 ${C.borderPrimary10}`}
            >
              {area}
            </h3>
            <div className="grid grid-cols-1 sm:grid-cols-2 md:grid-cols-3 lg:grid-cols-4 xl:grid-cols-5 gap-4">
              {areaCages.map((cage) => {
                const occupant = occupantByCageId.get(String(cage.id));
                const mergedRepName =
                  cage.mergedIntoCageId != null
                    ? resolveRepresentativeName(cages, cage.mergedIntoCageId)
                    : undefined;
                return (
                  <CageCard
                    key={cage.id}
                    cage={cage}
                    cages={cages}
                    occupant={occupant}
                    mergedRepresentativeName={mergedRepName}
                    hasVisibleOccupant={hasVisibleOccupant}
                    onNavigateToForm={onNavigateToForm}
                    onRequestUnmerge={onRequestUnmerge}
                    canCreate={canCreate}
                    canEdit={canEdit}
                    mergeMode={mergeMode}
                    canMerge={canMerge}
                  />
                );
              })}
            </div>
          </div>
        ))}
      </div>
    </DndContext>
  );
});
