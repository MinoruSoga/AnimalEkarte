import { memo, useState, useMemo, useCallback, useEffect, useLayoutEffect, useRef } from "react";
import { toast } from "sonner";
import { Textarea } from "@/components/ui/textarea";
import { Label } from "@/components/ui/label";
import { Button } from "@/components/ui/button";
import { FormFieldError } from "@/components/shared/FormFieldError/FormFieldError";
import { EstimateForm } from "./EstimateForm";
import { TreatmentTable, TreatmentItem } from "./TreatmentTable";
import { TreatmentDetailedSummary } from "./TreatmentDetailedSummary";
import {
  useGetEstimateByRecord,
  useCreateEstimateRecord,
  useUpdateEstimateRecord,
} from "../api/save-estimate";
import { C } from "@/lib/design-tokens";
import { calcEstimateTotals } from "@/lib/estimate-totals";
import { usePermission } from "@/hooks/use-permission";
import { useClinicTaxRates } from "@/hooks/use-clinic-tax-rates";
import type { Estimate, EstimateItem } from "@/types/generated/models";

// EstimateItem (BE snake_case) → TreatmentItem (UI camelCase) の明示変換。
// 旧実装の `as any` キャストはフィールド名非互換 (name→content, unit_price→unitPrice 等) を
// 隠蔽しており、items が返るケースで項目名・単価が undefined になる潜在バグだった。
function toTreatmentItem(item: EstimateItem, standardTaxRate: number): TreatmentItem {
  return {
    id: item.id,
    content: item.name,
    memo: "",
    is_insurance: item.is_insurance_applicable,
    unitPrice: item.unit_price,
    quantity: item.quantity,
    discountRate: item.discount_rate,
    discountAmount: item.discount_amount,
    taxType: item.tax_type ?? "excluded",
    taxRate: item.tax_rate ?? standardTaxRate,
  };
}

function hydrateFromEstimate(
  estimate: Estimate,
  standardTaxRate: number,
  setters: {
    setSubject: (v: string) => void;
    setComment: (v: string) => void;
    setRemarks: (v: string) => void;
    setGlobalDiscountAmount: (v: number) => void;
    setItems: (v: TreatmentItem[]) => void;
  },
) {
  setters.setSubject(estimate.title ?? "");
  setters.setComment(estimate.comment ?? "");
  setters.setRemarks(estimate.notes ?? "");
  setters.setGlobalDiscountAmount(estimate.discount_amount ?? 0);
  setters.setItems((estimate.items ?? []).map((item) => toTreatmentItem(item, standardTaxRate)));
}

interface MedicalRecordEstimateProps {
  isNewRecord?: boolean;
  ownerDiscountRate?: number;
  medicalRecordId?: string;
  onRegisterSave?: (fn: () => Promise<void>) => void;
}

export const MedicalRecordEstimate = memo(function MedicalRecordEstimate({
  isNewRecord = false,
  ownerDiscountRate = 0,
  medicalRecordId,
  onRegisterSave,
}: MedicalRecordEstimateProps) {
  const [subject, setSubject] = useState("");
  const [subjectError, setSubjectError] = useState<string>("");
  const [comment, setComment] = useState("");
  const [remarks, setRemarks] = useState("");
  const [globalDiscountAmount, setGlobalDiscountAmount] = useState(0);
  const [items, setItems] = useState<TreatmentItem[]>([]);

  const { canEdit } = usePermission("medical-records");
  const { canCreate: canCreateDiscount, canEdit: canEditDiscount } = usePermission("discount");
  // FE-RC-048: 消費税率はハードコード 0.1 ではなく病院マスタ設定を正本にする。
  const { standardTaxRate } = useClinicTaxRates();

  // Load existing estimate
  const { data: existingEstimate } = useGetEstimateByRecord(
    isNewRecord ? undefined : medicalRecordId,
  );
  const createEstimate = useCreateEstimateRecord(medicalRecordId ?? "");
  // BUG-016: id は mutate 時に渡す（hook 引数の id クロージャに依存しない）
  const updateEstimate = useUpdateEstimateRecord(medicalRecordId ?? "");

  // create 成功直後も PATCH できるよう id を ref で保持
  const estimateIdRef = useRef<number | null>(null);

  // PR665: UI 行 id → 現在の見積に保存済みと証明された item id の対応。
  // 取得時（行 id = 保存済み id）と保存応答（sort_order + name で突合）でのみ追加する。
  const persistedItemIdByRowIdRef = useRef(new Map<number, number>());
  const boundEstimateIdRef = useRef<number | null>(null);
  useLayoutEffect(() => {
    if (existingEstimate?.id == null) return;
    estimateIdRef.current = existingEstimate.id;
    if (boundEstimateIdRef.current === existingEstimate.id) return;
    persistedItemIdByRowIdRef.current = new Map(
      (existingEstimate.items ?? []).map((item) => [item.id, item.id]),
    );
    boundEstimateIdRef.current = existingEstimate.id;
  }, [existingEstimate]);

  // id が変わったときだけ hydrate（refetch の参照変更で入力を巻き戻さない）
  const [hydratedEstimateId, setHydratedEstimateId] = useState<number | null>(null);
  const canChangeDiscount =
    (existingEstimate?.id ?? hydratedEstimateId ?? 0) > 0 ? canEditDiscount : canCreateDiscount;
  if (existingEstimate?.id != null && existingEstimate.id !== hydratedEstimateId) {
    setHydratedEstimateId(existingEstimate.id);
    hydrateFromEstimate(existingEstimate, standardTaxRate, {
      setSubject,
      setComment,
      setRemarks,
      setGlobalDiscountAmount,
      setItems,
    });
  }

  const handleAddItem = useCallback(() => {
    setItems((prev) => [
      ...prev,
      {
        id: Date.now(),
        content: "",
        memo: "",
        is_insurance: false,
        unitPrice: 0,
        quantity: 1,
        discountRate: 0,
        discountAmount: 0,
        taxType: "included",
        taxRate: standardTaxRate,
      },
    ]);
  }, [standardTaxRate]);

  const handleRemoveItem = useCallback((id: number) => {
    setItems((prev) => prev.filter((item) => item.id !== id));
  }, []);

  const handleUpdateItem = useCallback(
    (id: number, field: keyof TreatmentItem, value: string | number | boolean) => {
      setItems((prev) => prev.map((item) => (item.id === id ? { ...item, [field]: value } : item)));
    },
    [],
  );

  // PR665: 表示・payload・検証は同じ「内容入力済み行」述語を共有する。
  // 空名行は保存されないため、合計や割引検証に寄与してはならない。
  const activeItems = useMemo(() => items.filter((item) => item.content.trim()), [items]);
  const { subtotal, tax, total } = useMemo(() => {
    const t = calcEstimateTotals(
      activeItems.map((item) => ({
        unitPrice: item.unitPrice,
        quantity: item.quantity,
        discountAmount: item.discountAmount,
        taxType: item.taxType,
        taxRate: item.taxRate,
      })),
    );
    return { subtotal: t.subtotal, tax: t.tax, total: t.total };
  }, [activeItems]);

  /**
   * BUG-016: メイン保存 post-save から await される。
   * startTransition で包まない（await 不能・2回目以降の偽成功の温床）。
   * 成功トーストは実 API 成功時のみ。
   */
  const handleSave = useCallback(async (): Promise<void> => {
    if (!medicalRecordId) {
      toast.error("カルテを保存してから見積書を保存してください");
      throw new Error("medicalRecordId missing");
    }
    if (!subject.trim()) {
      setSubjectError("件名を入力してください");
      throw new Error("subject required");
    }
    setSubjectError("");

    // PR665: ヘッダー割引は mutation 前に検証する（負・非安全整数・明細総額超過は拒否）。
    if (!Number.isSafeInteger(globalDiscountAmount) || globalDiscountAmount < 0) {
      toast.error("値引額は0以上の整数で入力してください");
      throw new Error("invalid discount amount");
    }
    if (activeItems.length > 0 && globalDiscountAmount > total) {
      toast.error("値引額が合計金額を超えています");
      throw new Error("discount exceeds total");
    }
    // 行なし置換は Create と同じレガシー契約: 正味合計 0、有効な割引メタデータは保持
    // し、明細総額に対する超過チェックのみ適用しない。
    const netTotal = activeItems.length === 0 ? 0 : total - globalDiscountAmount;

    const sentRows = activeItems.map((item, index) => {
      const persistedId = persistedItemIdByRowIdRef.current.get(item.id);
      const row = {
        name: item.content.trim(),
        category: "other",
        unit_price: Number(item.unitPrice) || 0,
        quantity: Number(item.quantity) || 1,
        discount_rate: Number(item.discountRate) || 0,
        discount_amount: Number(item.discountAmount) || 0,
        is_insurance_applicable: Boolean(item.is_insurance),
        tax_type: item.taxType ?? "excluded",
        tax_rate: item.taxRate ?? standardTaxRate,
        sort_order: index,
      };
      return {
        rowId: item.id,
        row: persistedId === undefined ? row : { ...row, existing_item_id: persistedId },
      };
    });

    const payload = {
      title: subject,
      subtotal,
      tax_total: tax,
      total_amount: netTotal,
      discount_amount: globalDiscountAmount,
      comment,
      notes: remarks,
      medical_record_id: Number(medicalRecordId),
      items: sentRows.map((sent) => sent.row),
    };

    try {
      const knownId = estimateIdRef.current ?? existingEstimate?.id ?? null;
      let saved: Estimate | undefined;
      let created = false;
      if (knownId != null && knownId > 0) {
        saved = await updateEstimate.mutateAsync({ id: knownId, payload });
      } else {
        saved = await createEstimate.mutateAsync(payload);
        created = true;
      }
      const isCurrentEstimate =
        boundEstimateIdRef.current === null || boundEstimateIdRef.current === saved.id;
      if (isCurrentEstimate) {
        estimateIdRef.current = saved.id;
        if (created) setHydratedEstimateId(saved.id);
      }
      // PR665: 応答の items を送信行と突合（sort_order + name）。配列位置や
      // 編集可能フィールドだけに頼らず、await 中の編集を壊さないよう map のみ更新する。
      if (Array.isArray(saved?.items) && isCurrentEstimate) {
        const nextBindings = new Map(persistedItemIdByRowIdRef.current);
        for (const responseItem of saved.items) {
          if (typeof responseItem?.id !== "number") continue;
          const match = sentRows.find(
            (sent) =>
              sent.row.sort_order === responseItem.sort_order &&
              sent.row.name === responseItem.name,
          );
          if (match) {
            nextBindings.set(match.rowId, responseItem.id);
          }
        }
        persistedItemIdByRowIdRef.current = nextBindings;
        boundEstimateIdRef.current = saved.id;
      }
      toast.success("見積書を保存しました");
    } catch (error) {
      // API エラーは mutation onError が handleApiError 済み。ここでは再通知しない。
      throw error instanceof Error ? error : new Error("見積書の保存に失敗しました");
    }
  }, [
    medicalRecordId,
    subject,
    subtotal,
    tax,
    total,
    globalDiscountAmount,
    comment,
    remarks,
    existingEstimate?.id,
    updateEstimate,
    createEstimate,
    activeItems,
    standardTaxRate,
  ]);

  useEffect(() => {
    if (!onRegisterSave) return;
    onRegisterSave(handleSave);
  }, [onRegisterSave, handleSave]);

  const handlePdfExport = useCallback(() => {
    toast.info("PDF出力機能は準備中です");
  }, []);

  return (
    <div className="flex flex-col gap-3 flex-1 min-h-0 overflow-y-auto relative pb-24 pr-1">
      {/* Subject */}
      <EstimateForm subject={subject} onSubjectChange={setSubject} canEdit={canEdit} />
      <FormFieldError message={subjectError} />

      {/* Items Table */}
      <TreatmentTable
        items={items}
        onUpdate={handleUpdateItem}
        onRemove={handleRemoveItem}
        onAddRow={handleAddItem}
        showStatus={false}
        disabled={!canEdit}
      />

      {/* Summary Table */}
      <TreatmentDetailedSummary
        subtotal={subtotal}
        tax={tax}
        total={activeItems.length === 0 ? 0 : Math.max(total - globalDiscountAmount, 0)}
        discountRate={ownerDiscountRate}
        discountAmount={globalDiscountAmount}
        onUpdateDiscountAmount={setGlobalDiscountAmount}
        isDiscountRateReadonly
        disabled={!canEdit || !canChangeDiscount}
      />

      {/* Comments & Remarks */}
      <div className="grid grid-cols-1 gap-3">
        <div className="flex flex-col gap-1">
          <Label className={`text-sm font-medium ${C.text60}`}>コメント</Label>
          <Textarea
            value={comment}
            onChange={(e) => setComment(e.target.value)}
            className={`${C.bgWhite} ${C.borderMedium} min-h-[60px] resize-none p-2 text-sm ${C.text}`}
            disabled={!canEdit}
          />
        </div>
        <div className="flex flex-col gap-1">
          <Label className={`text-sm font-medium ${C.text60}`}>備考</Label>
          <Textarea
            value={remarks}
            onChange={(e) => setRemarks(e.target.value)}
            className={`${C.bgWhite} ${C.borderMedium} min-h-[60px] resize-none p-2 text-sm ${C.text}`}
            disabled={!canEdit}
          />
        </div>
      </div>

      {/* Action Buttons */}
      <div className="flex justify-end gap-2 pt-2">
        <Button
          variant="outline"
          className={`h-10 px-4 text-sm ${C.borderMedium} ${C.text} ${C.hoverBgPage}`}
          onClick={handlePdfExport}
        >
          PDF出力
        </Button>
      </div>
    </div>
  );
});
