// React/Framework
import { memo, lazy, Suspense, useState, useCallback, useMemo } from "react";

// External
import { Plus, Search } from "lucide-react";

// Internal
import { Button } from "@/components/ui/button";
import { usePermission } from "@/hooks/use-permission";
import { C, STYLE, ICON } from "@/lib/design-tokens";
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
import { TreatmentRow } from "./TreatmentRow";

// ── 静的定数 ───────────────────────────────────────────────────────────

const TABLE_HEADER = (
  <thead>
    <tr className={`border-b ${C.borderLight} ${C.bgPage30} h-10`}>
      <th className={`px-3 text-left text-xs font-medium ${C.text70} w-10`}></th>
      <th className={`px-3 text-left text-xs font-medium ${C.text70} w-24`}>種別</th>
      <th className={`px-3 text-left text-xs font-medium ${C.text70}`}>内容</th>
      <th className={`px-3 text-center text-xs font-medium ${C.text70} w-16`}>保険</th>
      <th className={`px-3 text-right text-xs font-medium ${C.text70} w-28`}>単価</th>
      <th className={`px-3 text-right text-xs font-medium ${C.text70} w-20`}>数量</th>
      <th className={`px-3 text-right text-xs font-medium ${C.text70} w-28`}>値引き</th>
      <th className={`px-3 text-right text-xs font-medium ${C.text70} w-28`}>小計</th>
      <th className={`px-3 text-left text-xs font-medium ${C.text70}`}>メモ</th>
      <th className={`px-2 text-right text-xs font-medium ${C.text70} w-28`}></th>
    </tr>
  </thead>
);

const ITEM_TYPE_OPTIONS: { value: TreatmentItemType; label: string }[] = [
  { value: "consultation", label: "診療" },
  { value: "procedure", label: "処置" },
  { value: "medicine", label: "薬品" },
  { value: "other", label: "その他" },
];

const ADMIN_ROUTE_OPTIONS = [
  { value: "", label: "投与方法を選択" },
  { value: "経口", label: "経口" },
  { value: "注射", label: "注射" },
  { value: "外用", label: "外用" },
  { value: "点眼", label: "点眼" },
  { value: "点耳", label: "点耳" },
  { value: "吸入", label: "吸入" },
  { value: "その他", label: "その他" },
] as const;

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
        <table className="w-full">
          {TABLE_HEADER}
          <tbody>
            {sortedTreatments.length === 0 ? (
              <tr>
                <td
                  colSpan={10}
                  className={STYLE.tableEmptySm}
                >
                  治療明細がありません。下の「明細を追加」ボタンから追加してください。
                </td>
              </tr>
            ) : (
              sortedTreatments.map((treatment, idx) => (
                <TreatmentRow
                  key={treatment.id}
                  treatment={treatment}
                  isFirst={idx === 0}
                  isLast={idx === sortedTreatments.length - 1}
                  onUpdate={handleUpdate}
                  onDelete={handleDelete}
                  onMoveUp={handleMoveUp}
                  onMoveDown={handleMoveDown}
                  isUpdating={
                    updateMutation.isPending ||
                    deleteMutation.isPending ||
                    reorderMutation.isPending
                  }
                  canDelete={canDelete}
                  canEditDiscount={canEditDiscount}
                  autoFocusQuantity={focusLastRow && idx === sortedTreatments.length - 1}
                  onAutoFocusDone={focusLastRow && idx === sortedTreatments.length - 1 ? () => setFocusLastRow(false) : undefined}
                />
              ))
            )}
          </tbody>
        </table>

        {/* インライン追加フォーム */}
        {isAdding ? (
          <div className={`flex items-center gap-2 px-3 py-2 border-t ${C.borderLight} ${C.bgPage30}`}>
            <select
              value={addItemType}
              onChange={(e) => {
                setAddItemType(e.target.value as TreatmentItemType);
                setAddAdminRoute("");
              }}
              className={`h-8 text-sm rounded-[3px] border ${C.borderMedium} ${C.bgWhite} px-2 ${C.text}`}
            >
              {ITEM_TYPE_OPTIONS.map((opt) => (
                <option key={opt.value} value={opt.value}>
                  {opt.label}
                </option>
              ))}
            </select>
            {addItemType === "medicine" ? (
              <select
                value={addAdminRoute}
                onChange={(e) => setAddAdminRoute(e.target.value)}
                className={`h-8 text-sm rounded-[3px] border ${C.borderMedium} ${C.bgWhite} px-2 ${C.text}`}
              >
                {ADMIN_ROUTE_OPTIONS.map((opt) => (
                  <option key={opt.value} value={opt.value}>
                    {opt.label}
                  </option>
                ))}
              </select>
            ) : null}
            <input
              autoFocus
              type="text"
              placeholder="内容を入力..."
              value={addContent}
              onChange={(e) => setAddContent(e.target.value)}
              onKeyDown={(e) => {
                if (e.key === "Enter") handleAddSubmit();
                if (e.key === "Escape") handleAddCancel();
              }}
              className={`flex-1 h-8 text-sm border ${C.borderMedium} rounded-[3px] px-2 ${C.bgWhite} ${C.text} outline-none ${C.focusBorderAccent}`}
            />
            <Button
              size="sm"
              className={`${STYLE.btnPrimary} h-8 text-xs px-3`}
              onClick={handleAddSubmit}
              disabled={createMutation.isPending || !addContent.trim()}
            >
              追加
            </Button>
            <Button
              size="sm"
              variant="outline"
              className={`h-8 text-xs px-3 ${C.borderMedium}`}
              onClick={handleAddCancel}
            >
              キャンセル
            </Button>
          </div>
        ) : (
          canCreate ? (
            <div className="flex items-center gap-0">
              <button
                className={STYLE.inlineAddBtn}
                onClick={() => setIsSearchOpen(true)}
              >
                <Search className={`${ICON.xs}`} />
                <span>マスタから追加</span>
              </button>
              <button
                className={STYLE.inlineAddBtn}
                onClick={() => setIsAdding(true)}
              >
                <Plus className={`${ICON.xs}`} />
                <span>手入力で追加</span>
              </button>
            </div>
          ) : null
        )}
      </div>

      <Suspense fallback={null}>
        <TreatmentSearchDialog
          open={isSearchOpen}
          onOpenChange={setIsSearchOpen}
          onSelect={handleSelectFromMaster}
        />
      </Suspense>

      {/* フッター: 合計金額 */}
      <div className={`${C.bgWhite} border ${C.borderLight} rounded-[4px] px-4 py-3`}>
        <div className="flex flex-col gap-1.5">
          {/* 全明細合計 */}
          <div className={`flex items-center justify-between text-sm ${C.text60}`}>
            <span>全明細合計 ({sortedTreatments.length}件)</span>
            <span className="font-mono">¥{totalSubtotal.toLocaleString()}</span>
          </div>
          {/* 選択済み合計 */}
          <div className={`flex items-center justify-between text-sm font-medium ${C.text}`}>
            <span>選択済み合計 ({selectedCount}件)</span>
            <span className="font-mono text-base">
              ¥{selectedSubtotal.toLocaleString()}
            </span>
          </div>
          {/* 税込 */}
          <div
            className={`flex items-center justify-between text-sm pt-1.5 border-t ${C.borderLight}`}
          >
            <span className={C.text60}>税込合計 (10% {ownerDiscountRate > 0 ? `飼主割引${ownerDiscountRate}%適用後` : ""})</span>
            <span className={`font-mono font-semibold ${C.textCostBlue}`}>
              ¥{finalTotal.toLocaleString()}
            </span>
          </div>
        </div>
      </div>
    </div>
  );
});
