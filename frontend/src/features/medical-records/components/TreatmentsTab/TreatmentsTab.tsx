// React/Framework
import { memo, lazy, Suspense, useState, useCallback, useMemo } from "react";
import { useQueryClient } from "@tanstack/react-query";

// Internal
import { usePermission } from "@/hooks/use-permission";
import { useGetAllMedicinesMaster } from "@/hooks/use-treatment-master";
import { C, STYLE } from "@/lib/design-tokens";
import { QUERY_STALE_TIMES } from "@/lib/react-query";
import { ConfirmDialog } from "@/components/shared/ConfirmDialog/ConfirmDialog";
import type { TreatmentMasterItem } from "@/components/shared/TreatmentSearchDialog/TreatmentSearchDialog";

const TreatmentSearchDialog = lazy(() =>
  import("@/components/shared/TreatmentSearchDialog/TreatmentSearchDialog").then((m) => ({
    default: m.TreatmentSearchDialog,
  }))
);

// Relative
import { useGetTreatments } from "../../api/treatments";
import { useCreateTreatment } from "../../api/treatments";
import { useUpdateTreatment } from "../../api/treatments";
import { useDeleteTreatment } from "../../api/treatments";
import { useReorderTreatments } from "../../api/treatments";
import { useGetVitals } from "../../api/vitals";
import {
  buildDoseCalcInput,
  computeDosePreview,
  fetchMedicineDoseParamsOnce,
  medicineDoseParamsQueryKey,
  resolveLatestVitalWeight,
} from "../../api/medicine-dose-lookup";
import type { CreateTreatmentInput, TreatmentItemType, UpdateTreatmentInput } from "../../types";
import { TreatmentAddControls, TreatmentsTable, TreatmentTotals } from "./TreatmentsTabParts";
import { buildMasterSelectionPayload } from "./treatments-tab-model";
import { computeDoseGate } from "./treatment-row-dose-gate";

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
  const { canCreate, canEdit, canDelete } = usePermission("medical-records");
  // BUG-372: 割引権限（値引額編集制御）
  const { canEdit: canEditDiscount } = usePermission("discount");
  const { data: treatments, isLoading } = useGetTreatments(medicalRecordId, recordClinicId);
  const createMutation = useCreateTreatment(medicalRecordId, recordClinicId);
  const { mutate: createTreatmentFn } = createMutation;
  const updateMutation = useUpdateTreatment(medicalRecordId, recordClinicId);
  const { mutate: updateTreatmentFn } = updateMutation;
  const deleteMutation = useDeleteTreatment(medicalRecordId, recordClinicId);
  const { mutate: deleteTreatmentFn } = deleteMutation;
  const reorderMutation = useReorderTreatments(medicalRecordId, recordClinicId);
  const { mutate: reorderTreatmentsFn } = reorderMutation;
  const queryClient = useQueryClient();

  // #201: 投与量自動計算の入力（当日 vital の体重・薬マスタ一覧）。
  const { data: vitals } = useGetVitals(medicalRecordId, recordClinicId);
  const { data: medicines } = useGetAllMedicinesMaster();
  const resolvedWeight = useMemo(
    () => (vitals ? resolveLatestVitalWeight(vitals) : null),
    [vitals]
  );
  // react-review-201 HIGH-1: オブジェクトリテラルを直接 props に渡すと毎レンダー新規参照になり
  // TreatmentRow の React.memo が無効化される（全行が無関係な state 変更でも再レンダーされる）。
  const doseContext = useMemo(
    () => ({
      medicines,
      petSpecies: petSpecies ?? null,
      weightKg: resolvedWeight?.weightKg ?? null,
    }),
    [medicines, petSpecies, resolvedWeight]
  );

  // マスタ検索ダイアログ
  const [isSearchOpen, setIsSearchOpen] = useState(false);

  // 追加フォームの状態
  const [addItemType, setAddItemType] = useState<TreatmentItemType>("consultation");
  const [addContent, setAddContent] = useState("");
  const [addAdminRoute, setAddAdminRoute] = useState("");
  const [isAdding, setIsAdding] = useState(false);
  // 直前の追加後に最終行へ自動フォーカスするフラグ
  const [focusLastRow, setFocusLastRow] = useState(false);

  // P1-7 (PR #186 review): マスタ選択時の quantity フォールバック(既定値1)自体が
  // per-weight 薬剤の安全域外なら、TreatmentRow.commitQuantity と同じ hard gate を
  // create 前に適用する（確認なしでの不安全用量作成を防ぐ）。
  const [pendingMasterGate, setPendingMasterGate] = useState<{
    payload: CreateTreatmentInput;
    reason: string;
  } | null>(null);

  // sort_order 昇順でソート済みリスト
  const sortedTreatments = useMemo(() => {
    if (!treatments) return [];
    return [...treatments].sort((a, b) => a.sort_order - b.sort_order);
  }, [treatments]);

  // 合計金額 (selected のみ)
  const { selectedSubtotal, selectedCount, finalTotal } = useMemo(() => {
    const selected = sortedTreatments.filter((t) => t.is_selected);
    const sub = selected.reduce(
      (sum, t) => sum + t.unit_price * t.quantity - t.discount_amount,
      0
    );
    // 飼主割引適用
    const ownerDiscount = Math.floor(sub * (ownerDiscountRate / 100));
    const afterDiscount = sub - ownerDiscount;
    const tax = Math.floor(afterDiscount * 0.1);

    return { 
      selectedSubtotal: sub, 
      selectedCount: selected.length,
      finalTotal: afterDiscount + tax
    };
  }, [sortedTreatments, ownerDiscountRate]);

  // 全明細の合計
  const totalSubtotal = useMemo(
    () =>
      sortedTreatments.reduce(
        (sum, t) => sum + t.unit_price * t.quantity - t.discount_amount,
        0
      ),
    [sortedTreatments]
  );

  // ── handlers ──

  const handleUpdate = useCallback(
    (treatmentId: string, input: UpdateTreatmentInput) => {
      if (!canEdit) return;
      updateTreatmentFn({ treatmentId, input });
    },
    [canEdit, updateTreatmentFn]
  );

  const handleDelete = useCallback(
    (treatmentId: string) => {
      if (!canDelete) return;
      deleteTreatmentFn(treatmentId);
    },
    [canDelete, deleteTreatmentFn]
  );

  const handleMoveUp = useCallback(
    (treatmentId: string) => {
      if (!canEdit) return;
      const list = sortedTreatments;
      const idx = list.findIndex((t) => t.id === treatmentId);
      if (idx <= 0) return;
      const newList = list.map((t, i) => {
        if (i === idx - 1) return { id: t.id, sort_order: list[idx].sort_order };
        if (i === idx) return { id: t.id, sort_order: list[idx - 1].sort_order };
        return { id: t.id, sort_order: t.sort_order };
      });
      reorderTreatmentsFn({ treatments: newList });
    },
    [canEdit, sortedTreatments, reorderTreatmentsFn]
  );

  const handleMoveDown = useCallback(
    (treatmentId: string) => {
      if (!canEdit) return;
      const list = sortedTreatments;
      const idx = list.findIndex((t) => t.id === treatmentId);
      if (idx < 0 || idx >= list.length - 1) return;
      const newList = list.map((t, i) => {
        if (i === idx) return { id: t.id, sort_order: list[idx + 1].sort_order };
        if (i === idx + 1) return { id: t.id, sort_order: list[idx].sort_order };
        return { id: t.id, sort_order: t.sort_order };
      });
      reorderTreatmentsFn({ treatments: newList });
    },
    [canEdit, sortedTreatments, reorderTreatmentsFn]
  );

  const handleAddSubmit = useCallback(() => {
    if (!canCreate || !addContent.trim()) return;
    const nextOrder =
      sortedTreatments.length > 0
        ? sortedTreatments[sortedTreatments.length - 1].sort_order + 1
        : 0;
    // 薬品の場合、投与方法を memo に付記する（BE-MEDI-010: 専用カラム実装まで）
    const memoWithRoute =
      addItemType === "medicine" && addAdminRoute
        ? `[投与方法: ${addAdminRoute}]`
        : "";
    createTreatmentFn(
      {
        item_type: addItemType,
        content: addContent.trim(),
        unit_price: 0,
        quantity: 1,
        is_selected: true,
        is_insurance: false,
        discount_amount: 0,
        sort_order: nextOrder,
        memo: memoWithRoute,
      },
      {
        onSuccess: () => {
          setFocusLastRow(true);
          setAddContent("");
          setAddAdminRoute("");
          setIsAdding(false);
        },
      }
    );
  }, [canCreate, addItemType, addContent, addAdminRoute, sortedTreatments, createTreatmentFn]);

  const handleAddCancel = useCallback(() => {
    setAddContent("");
    setAddAdminRoute("");
    setIsAdding(false);
  }, []);

  const handleAddItemTypeChange = useCallback((itemType: TreatmentItemType) => {
    setAddItemType(itemType);
    setAddAdminRoute("");
  }, []);

  const handleSelectFromMaster = useCallback(async (item: TreatmentMasterItem) => {
    if (!canCreate) return;
    const nextOrder =
      sortedTreatments.length > 0
        ? sortedTreatments[sortedTreatments.length - 1].sort_order + 1
        : 0;

    // #201: 薬剤選択時、体重・species・薬マスタの投与量パラメータが揃えば quantity をプリフィルする。
    // いずれか欠けている場合は既定値 1（従来通りの手動入力）にフェイルクローズする。
    let quantity = 1;
    // P1-7: gate 判定用の DoseCalcInput。プリフィル失敗時（fetch エラー等）は null のままとなり、
    // computeDoseGate(null, ...) は常に requiresConfirm=false を返す（既存の fail-closed 挙動を維持）。
    let doseCalcInput = null as ReturnType<typeof buildDoseCalcInput>;
    if (item.medicineId && resolvedWeight) {
      const medicine = medicines?.find((m) => m.id === item.medicineId);
      if (medicine) {
        try {
          const doseParams = await queryClient.fetchQuery({
            queryKey: medicineDoseParamsQueryKey(item.medicineId),
            queryFn: () => fetchMedicineDoseParamsOnce(item.medicineId as string),
            staleTime: QUERY_STALE_TIMES.STATIC,
          });
          doseCalcInput = buildDoseCalcInput(medicine, doseParams, petSpecies, resolvedWeight.weightKg);
          const preview = computeDosePreview(medicine, doseParams, petSpecies, resolvedWeight.weightKg);
          // healthcare-review-201 MEDIUM: 推奨値自体が丸め up 等で安全域を超えている場合、
          // 未編集のまま hard gate を経ずに保存され得るため、その値ではプリフィルしない
          // （既定値 1 の手動入力にフェイルクローズし、獣医の入力を commitQuantity の hard gate に通す）。
          if (preview && !preview.exceedsMax && !preview.belowMin) quantity = preview.quantity;
        } catch {
          // プリフィル取得失敗時は手動入力にフェイルクローズ（quantity は既定値 1 のまま）
        }
      }
    }

    const payload = buildMasterSelectionPayload({ item, quantity, sortOrder: nextOrder });

    // P1-7 (PR #186 review): 実際に submit される quantity（既定値 1 のフォールバック含む）が
    // per-weight 薬剤の安全域外なら、TreatmentRow.commitQuantity と同じ hard gate（確認ダイアログ）
    // を create 前に適用する。既定値 1 のまま無確認で保存することはない。
    const gate = computeDoseGate(doseCalcInput, quantity);
    if (gate.requiresConfirm) {
      setPendingMasterGate({ payload, reason: gate.reason });
      return;
    }

    createTreatmentFn(payload, { onSuccess: () => setFocusLastRow(true) });
  }, [canCreate, sortedTreatments, createTreatmentFn, resolvedWeight, medicines, petSpecies, queryClient]);

  const handleMasterGateConfirm = useCallback(() => {
    if (!pendingMasterGate) return;
    createTreatmentFn(pendingMasterGate.payload, { onSuccess: () => setFocusLastRow(true) });
    setPendingMasterGate(null);
  }, [pendingMasterGate, createTreatmentFn]);

  const handleMasterGateCancel = useCallback(() => {
    setPendingMasterGate(null);
  }, []);

  // ── render ──

  if (isLoading) {
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
          treatments={sortedTreatments}
          isUpdating={updateMutation.isPending || deleteMutation.isPending || reorderMutation.isPending}
          canDelete={canDelete}
          canEditDiscount={canEditDiscount}
          focusLastRow={focusLastRow}
          onUpdate={handleUpdate}
          onDelete={handleDelete}
          onMoveUp={handleMoveUp}
          onMoveDown={handleMoveDown}
          onAutoFocusDone={() => setFocusLastRow(false)}
          doseContext={doseContext}
        />
        <TreatmentAddControls
          canCreate={canCreate}
          isAdding={isAdding}
          isPending={createMutation.isPending}
          addItemType={addItemType}
          addContent={addContent}
          addAdminRoute={addAdminRoute}
          onItemTypeChange={handleAddItemTypeChange}
          onContentChange={setAddContent}
          onAdminRouteChange={setAddAdminRoute}
          onSubmit={handleAddSubmit}
          onCancel={handleAddCancel}
          onOpenSearch={() => setIsSearchOpen(true)}
          onStartAdding={() => setIsAdding(true)}
        />
      </div>

      <Suspense fallback={null}>
        <TreatmentSearchDialog
          open={isSearchOpen}
          onOpenChange={setIsSearchOpen}
          onSelect={handleSelectFromMaster}
        />
      </Suspense>

      {/* P1-7: マスタ選択時の hard gate — 既定値1が安全域外の場合は確認必須（警告への格下げ禁止） */}
      <ConfirmDialog
        open={pendingMasterGate !== null}
        onClose={handleMasterGateCancel}
        onConfirm={handleMasterGateConfirm}
        title="投与量を確認してください"
        description={pendingMasterGate?.reason || ""}
        confirmLabel="この数量で追加する"
        cancelLabel="キャンセル"
        variant="destructive"
        isPending={createMutation.isPending}
      />

      {/* フッター: 合計金額 */}
      <TreatmentTotals
        totalCount={sortedTreatments.length}
        totalSubtotal={totalSubtotal}
        selectedCount={selectedCount}
        selectedSubtotal={selectedSubtotal}
        finalTotal={finalTotal}
        ownerDiscountRate={ownerDiscountRate}
      />
    </div>
  );
});
