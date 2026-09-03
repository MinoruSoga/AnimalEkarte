// React/Framework
import { memo, lazy, Suspense } from "react";

// Internal
import { C, STYLE } from "@/lib/design-tokens";

const TreatmentSearchDialog = lazy(() =>
  import("@/components/shared/TreatmentSearchDialog/TreatmentSearchDialog").then((m) => ({
    default: m.TreatmentSearchDialog,
  }))
);

// Relative
import { TreatmentAddControls, TreatmentsTable, TreatmentTotals } from "./TreatmentsTabParts";
import { useTreatmentsTab } from "../../hooks/use-treatments-tab";

// ── Props ─────────────────────────────────────────────────────────────

interface TreatmentsTabProps {
  medicalRecordId: string;
  ownerDiscountRate?: number;
  /** #201: 投与量自動計算の species 解決に使う free-text ペット種（未設定なら計算はスキップ＝手動） */
  petSpecies?: string | null;
  /** P2-15: 拠点横断で開いたカルテの子リソース操作用。レコード自身の clinicId（同一クリニックなら未設定でも可） */
  recordClinicId?: string;
}

// ── Component ─────────────────────────────────────────────────────────

export const TreatmentsTab = memo(function TreatmentsTab({
  medicalRecordId,
  ownerDiscountRate = 0,
  petSpecies,
  recordClinicId,
}: TreatmentsTabProps) {
  const t = useTreatmentsTab({ medicalRecordId, ownerDiscountRate, petSpecies, recordClinicId });

  // ── render ──

  if (t.isLoading) {
    return (
      <div className={`flex items-center justify-center h-48 text-sm ${C.text40}`}>
        読み込み中...
      </div>
    );
  }

  return (
    <div className="flex flex-col gap-3 pb-24">
      {/* テーブル */}
      <div className={`${STYLE.tableContainer} overflow-x-auto`}>
        <TreatmentsTable
          treatments={t.sortedTreatments}
          isUpdating={t.isMutating}
          canDelete={t.canDelete}
          canEditDiscount={t.canEditDiscount}
          focusLastRow={t.focusLastRow}
          onUpdate={t.handleUpdate}
          onDelete={t.handleDelete}
          onMoveUp={t.handleMoveUp}
          onMoveDown={t.handleMoveDown}
          onAutoFocusDone={t.handleAutoFocusDone}
          doseContext={t.doseContext}
        />
        <TreatmentAddControls
          canCreate={t.canCreate}
          isAdding={t.isAdding}
          isPending={t.createIsPending}
          addItemType={t.addItemType}
          addContent={t.addContent}
          addAdminRoute={t.addAdminRoute}
          onItemTypeChange={t.handleAddItemTypeChange}
          onContentChange={t.setAddContent}
          onAdminRouteChange={t.setAddAdminRoute}
          onSubmit={t.handleAddSubmit}
          onCancel={t.handleAddCancel}
          onOpenSearch={t.handleOpenSearch}
          onStartAdding={t.handleStartAdding}
        />
      </div>

      <Suspense fallback={null}>
        <TreatmentSearchDialog
          open={t.isSearchOpen}
          onOpenChange={t.setIsSearchOpen}
          onSelect={t.handleSelectFromMaster}
        />
      </Suspense>

      {t.masterDoseBlockReason ? (
        <div
          role="alert"
          className={`rounded-xs border px-3 py-2 text-sm font-semibold ${C.borderDanger20} ${C.bgDanger8} ${C.danger}`}
        >
          <div>⚠ {t.masterDoseBlockReason}</div>
          {t.pendingMasterLookupItem ? (
            <button
              type="button"
              className={`mt-2 text-sm font-medium underline ${C.danger}`}
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
