// React/Framework
import { memo, lazy, Suspense, useState, useCallback, useMemo } from "react";

// Internal
import { usePermission } from "@/hooks/use-permission";
import { C, STYLE } from "@/lib/design-tokens";
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
import type { TreatmentItemType, UpdateTreatmentInput } from "../../types";
import { TreatmentAddControls, TreatmentsTable, TreatmentTotals } from "./TreatmentsTabParts";

// ── Props ─────────────────────────────────────────────────────────────

interface TreatmentsTabProps {
  medicalRecordId: string;
  ownerDiscountRate?: number;
}

// ── Component ─────────────────────────────────────────────────────────

export const TreatmentsTab = memo(function TreatmentsTab({ medicalRecordId, ownerDiscountRate = 0 }: TreatmentsTabProps) {
  const { canCreate, canEdit, canDelete } = usePermission("medical-records");
  // BUG-372: 割引権限（値引額編集制御）
  const { canEdit: canEditDiscount } = usePermission("discount");
  const { data: treatments, isLoading } = useGetTreatments(medicalRecordId);
  const createMutation = useCreateTreatment(medicalRecordId);
  const { mutate: createTreatmentFn } = createMutation;
  const updateMutation = useUpdateTreatment(medicalRecordId);
  const { mutate: updateTreatmentFn } = updateMutation;
  const deleteMutation = useDeleteTreatment(medicalRecordId);
  const { mutate: deleteTreatmentFn } = deleteMutation;
  const reorderMutation = useReorderTreatments(medicalRecordId);
  const { mutate: reorderTreatmentsFn } = reorderMutation;

  // マスタ検索ダイアログ
  const [isSearchOpen, setIsSearchOpen] = useState(false);

  // 追加フォームの状態
  const [addItemType, setAddItemType] = useState<TreatmentItemType>("consultation");
  const [addContent, setAddContent] = useState("");
  const [addAdminRoute, setAddAdminRoute] = useState("");
  const [isAdding, setIsAdding] = useState(false);
  // 直前の追加後に最終行へ自動フォーカスするフラグ
  const [focusLastRow, setFocusLastRow] = useState(false);

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

  const handleSelectFromMaster = useCallback((item: TreatmentMasterItem) => {
    if (!canCreate) return;
    const nextOrder =
      sortedTreatments.length > 0
        ? sortedTreatments[sortedTreatments.length - 1].sort_order + 1
        : 0;
    const itemType: TreatmentItemType =
      item.category === "薬剤" ? "medicine"
        : item.category === "処置" ? "procedure"
        : item.category === "診察" ? "consultation"
        : "other";
    createTreatmentFn(
      {
        item_type: itemType,
        content: item.name,
        unit_price: item.unitPrice,
        quantity: 1,
        is_selected: true,
        is_insurance: false,
        discount_amount: 0,
        sort_order: nextOrder,
        memo: "",
      },
      { onSuccess: () => setFocusLastRow(true) },
    );
  }, [canCreate, sortedTreatments, createTreatmentFn]);

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
