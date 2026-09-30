// React/Framework
import { memo, lazy, Suspense, useCallback, useLayoutEffect, useRef, useState } from "react";
import { toast } from "sonner";

// Internal
import { ConfirmDialog } from "@/components/shared/ConfirmDialog/ConfirmDialog";
import { LoadingFallback } from "@/components/shared/DataStates";
import type { TreatmentMasterItem } from "@/components/shared/TreatmentSearchDialog/TreatmentSearchDialog";
import { C, STYLE } from "@/lib/design-tokens";

const TreatmentSearchDialog = lazy(() =>
  import("@/components/shared/TreatmentSearchDialog/TreatmentSearchDialog").then((m) => ({
    default: m.TreatmentSearchDialog,
  })),
);

// Relative
import type { UpdateTreatmentInput } from "../../types";
import { TreatmentAddControls, TreatmentsTable, TreatmentTotals } from "./TreatmentsTabParts";
import { useTreatmentsTab } from "../../hooks/use-treatments-tab";

const DECEASED_TREATMENTS_MESSAGE = "死亡したペットの治療明細は変更できません";

// ── Props ─────────────────────────────────────────────────────────────

interface TreatmentsTabProps {
  medicalRecordId: string;
  ownerDiscountRate?: number;
  /** #201: 投与量自動計算の species 解決に使う free-text ペット種（未設定なら計算はスキップ＝手動） */
  petSpecies?: string | null;
  /** P2-15: 拠点横断で開いたカルテの子リソース操作用。レコード自身の clinicId（同一クリニックなら未設定でも可） */
  recordClinicId?: string;
  /** 死亡ペットのカルテは閲覧専用（VitalsTab と同じ二重ガード方針） */
  isPetDeceased?: boolean;
}

// ── Component ─────────────────────────────────────────────────────────

export const TreatmentsTab = memo(function TreatmentsTab({
  medicalRecordId,
  ownerDiscountRate = 0,
  petSpecies,
  recordClinicId,
  isPetDeceased = false,
}: TreatmentsTabProps) {
  const t = useTreatmentsTab({ medicalRecordId, ownerDiscountRate, petSpecies, recordClinicId });
  // rerender-memo: hook のハンドラは useCallback で安定済み。ガードで包む際は分割代入して
  // 正確な deps を宣言する（`t` 自体を dep にすると毎レンダー新規参照になり行 memo が壊れる）。
  const {
    handleUpdate,
    handleDelete,
    handleMoveUp,
    handleMoveDown,
    handleAddSubmit,
    handleSelectFromMaster,
  } = t;

  const [deletingId, setDeletingId] = useState<string | null>(null);
  const isPetDeceasedRef = useRef(isPetDeceased);
  useLayoutEffect(() => {
    isPetDeceasedRef.current = isPetDeceased;
  }, [isPetDeceased]);

  // 死亡ペットの mutation は UI 側で非活性化するのに加え、コールバック側でも必ず止める
  // （VitalsTab の isPetDeceasedRef パターンに揃える）。ref 経由なのでハンドラは安定参照を保つ。
  const isDeceasedWriteBlocked = useCallback((): boolean => {
    if (isPetDeceasedRef.current !== true) return false;
    toast.error(DECEASED_TREATMENTS_MESSAGE);
    return true;
  }, []);

  const guardedUpdate = useCallback(
    (treatmentId: string, input: UpdateTreatmentInput) => {
      if (isDeceasedWriteBlocked()) return;
      handleUpdate(treatmentId, input);
    },
    [handleUpdate, isDeceasedWriteBlocked],
  );

  const guardedMoveUp = useCallback(
    (treatmentId: string) => {
      if (isDeceasedWriteBlocked()) return;
      handleMoveUp(treatmentId);
    },
    [handleMoveUp, isDeceasedWriteBlocked],
  );

  const guardedMoveDown = useCallback(
    (treatmentId: string) => {
      if (isDeceasedWriteBlocked()) return;
      handleMoveDown(treatmentId);
    },
    [handleMoveDown, isDeceasedWriteBlocked],
  );

  const guardedAddSubmit = useCallback(() => {
    if (isDeceasedWriteBlocked()) return;
    handleAddSubmit();
  }, [handleAddSubmit, isDeceasedWriteBlocked]);

  const guardedSelectFromMaster = useCallback(
    (item: TreatmentMasterItem) => {
      if (isDeceasedWriteBlocked()) return;
      void handleSelectFromMaster(item);
    },
    [handleSelectFromMaster, isDeceasedWriteBlocked],
  );

  const handleDeleteConfirm = useCallback(() => {
    if (isDeceasedWriteBlocked()) return;
    if (!deletingId) return;
    // canDelete の権限チェックは hook 側の handleDelete が fail-closed で担う。
    handleDelete(deletingId);
    setDeletingId(null);
  }, [deletingId, handleDelete, isDeceasedWriteBlocked]);

  // ── render ──

  if (t.isLoading) {
    return <LoadingFallback />;
  }

  return (
    <div className="flex flex-col gap-3 pb-24 flex-1 min-h-0 overflow-y-auto relative">
      {/* テーブル */}
      <div className={`${STYLE.tableContainer} overflow-x-auto`}>
        <TreatmentsTable
          treatments={t.sortedTreatments}
          isUpdating={t.isMutating || isPetDeceased}
          canDelete={Boolean(t.canDelete && !isPetDeceased)}
          canEditDiscount={t.canEditDiscount}
          focusLastRow={t.focusLastRow}
          onUpdate={guardedUpdate}
          onDelete={setDeletingId}
          onMoveUp={guardedMoveUp}
          onMoveDown={guardedMoveDown}
          onAutoFocusDone={t.handleAutoFocusDone}
          doseContext={t.doseContext}
        />
        <TreatmentAddControls
          canCreate={Boolean(t.canCreate && !isPetDeceased)}
          isAdding={t.isAdding}
          isPending={t.createIsPending}
          addItemType={t.addItemType}
          addContent={t.addContent}
          addAdminRoute={t.addAdminRoute}
          onItemTypeChange={t.handleAddItemTypeChange}
          onContentChange={t.setAddContent}
          onAdminRouteChange={t.setAddAdminRoute}
          onSubmit={guardedAddSubmit}
          onCancel={t.handleAddCancel}
          onOpenSearch={t.handleOpenSearch}
          onStartAdding={t.handleStartAdding}
        />
      </div>

      <Suspense fallback={null}>
        <TreatmentSearchDialog
          open={t.isSearchOpen}
          onOpenChange={t.setIsSearchOpen}
          onSelect={guardedSelectFromMaster}
        />
      </Suspense>

      <ConfirmDialog
        open={deletingId !== null}
        onClose={() => setDeletingId(null)}
        onConfirm={handleDeleteConfirm}
        title="治療明細を削除しますか？"
        description="この治療明細を削除します。この操作は元に戻せません。"
        confirmLabel="削除する"
        cancelLabel="キャンセル"
        variant="destructive"
      />

      {t.masterDoseBlockReason ? (
        <div
          role="alert"
          className={`rounded-xs border px-3 py-2 text-sm font-semibold ${C.borderDanger20} ${C.bgDanger8} ${C.danger}`}
        >
          <div>⚠ {t.masterDoseBlockReason}</div>
          {t.pendingMasterLookupItem ? (
            <button
              type="button"
              className={`mt-2 inline-flex min-h-11 min-w-11 items-center text-sm font-medium underline ${C.danger}`}
              onClick={t.handleRetryMasterDoseLookup}
              aria-label="投与量パラメータの取得を再試行する"
            >
              再試行する
            </button>
          ) : null}
        </div>
      ) : null}

      {/* フッター: 合計金額 */}
      <TreatmentTotals
        totalCount={t.sortedTreatments.length}
        totalSubtotal={t.totalSubtotal}
        selectedCount={t.selectedCount}
        selectedSubtotal={t.selectedSubtotal}
        finalTotal={t.finalTotal}
        ownerDiscountRate={ownerDiscountRate}
      />
    </div>
  );
});
