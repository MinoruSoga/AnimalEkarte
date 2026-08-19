import { memo, useState, useMemo, useCallback, useEffect, useRef } from "react";
import { toast } from "sonner";
import { Textarea } from "@/components/ui/textarea";
import { Label } from "@/components/ui/label";
import { Button } from "@/components/ui/button";
import { FormFieldError } from "@/components/shared/FormFieldError/FormFieldError";
import { EstimateForm } from "./EstimateForm";
import { TreatmentTable, TreatmentItem } from "./TreatmentTable";
import { TreatmentDetailedSummary } from "./TreatmentDetailedSummary";
import { useGetEstimateByRecord, useCreateEstimateRecord, useUpdateEstimateRecord } from "../api/save-estimate";
import { C } from "@/lib/design-tokens";
import { usePermission } from "@/hooks/use-permission";
import type { Estimate, EstimateItem } from "@/types/generated/models";

// EstimateItem (BE snake_case) → TreatmentItem (UI camelCase) の明示変換。
// 旧実装の `as any` キャストはフィールド名非互換 (name→content, unit_price→unitPrice 等) を
// 隠蔽しており、items が返るケースで項目名・単価が undefined になる潜在バグだった。
function toTreatmentItem(item: EstimateItem): TreatmentItem {
  return {
    id: item.id,
    content: item.name,
    memo: "",
    is_insurance: item.is_insurance_applicable,
    unitPrice: item.unit_price,
    quantity: item.quantity,
    discountRate: item.discount_rate,
    discountAmount: item.discount_amount,
  };
}

function hydrateFromEstimate(
  estimate: Estimate,
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
  setters.setItems((estimate.items ?? []).map(toTreatmentItem));
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

  // Load existing estimate
  const { data: existingEstimate } = useGetEstimateByRecord(
    isNewRecord ? undefined : medicalRecordId,
  );
  const createEstimate = useCreateEstimateRecord(medicalRecordId ?? "");
  // BUG-016: id は mutate 時に渡す（hook 引数の id クロージャに依存しない）
  const updateEstimate = useUpdateEstimateRecord(medicalRecordId ?? "");

  // create 成功直後も PATCH できるよう id を ref で保持
  const estimateIdRef = useRef<number | null>(null);

  // id が変わったときだけ hydrate（refetch の参照変更で入力を巻き戻さない）
  const [hydratedEstimateId, setHydratedEstimateId] = useState<number | null>(null);
  if (existingEstimate?.id != null && existingEstimate.id !== hydratedEstimateId) {
    setHydratedEstimateId(existingEstimate.id);
    estimateIdRef.current = existingEstimate.id;
    hydrateFromEstimate(existingEstimate, {
      setSubject,
      setComment,
      setRemarks,
      setGlobalDiscountAmount,
      setItems,
    });
  } else if (existingEstimate?.id != null) {
    estimateIdRef.current = existingEstimate.id;
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
      },
    ]);
  }, []);

  const handleRemoveItem = useCallback((id: number) => {
    setItems((prev) => prev.filter((item) => item.id !== id));
  }, []);

  const handleUpdateItem = useCallback(
    (id: number, field: keyof TreatmentItem, value: string | number | boolean) => {
      setItems((prev) =>
        prev.map((item) => (item.id === id ? { ...item, [field]: value } : item)),
      );
    },
    [],
  );

  // Calculate totals
  const { subtotal, tax, total } = useMemo(() => {
    const sub = items.reduce((sum, item) => {
      const price = Number(item.unitPrice) || 0;
      const qty = Number(item.quantity) || 0;
      const discount = Number(item.discountAmount) || 0;
      return sum + (price * qty - discount);
    }, 0);
    const taxAmount = Math.floor(sub * 0.1);
    return { subtotal: sub, tax: taxAmount, total: sub + taxAmount };
  }, [items]);

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

    const payload = {
      title: subject,
      subtotal,
      tax_total: tax,
      total_amount: total - globalDiscountAmount,
      discount_amount: globalDiscountAmount,
      comment,
      notes: remarks,
      medical_record_id: Number(medicalRecordId),
      items: items
        .filter((item) => item.content.trim())
        .map((item, index) => ({
          name: item.content,
          category: "other",
          unit_price: Number(item.unitPrice) || 0,
          quantity: Number(item.quantity) || 1,
          discount_rate: Number(item.discountRate) || 0,
          discount_amount: Number(item.discountAmount) || 0,
          is_insurance_applicable: Boolean(item.is_insurance),
          sort_order: index,
        })),
    };

    try {
      const knownId = estimateIdRef.current ?? existingEstimate?.id ?? null;
      if (knownId != null && knownId > 0) {
        const updated = await updateEstimate.mutateAsync({ id: knownId, payload });
        estimateIdRef.current = updated.id;
      } else {
        const created = await createEstimate.mutateAsync(payload);
        estimateIdRef.current = created.id;
        setHydratedEstimateId(created.id);
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
  ]);

  useEffect(() => {
    if (!onRegisterSave) return;
    onRegisterSave(handleSave);
  }, [onRegisterSave, handleSave]);

  const handlePdfExport = useCallback(() => {
    toast.info("PDF出力機能は準備中です");
  }, []);

  return (
    <div className="h-[calc(100vh-220px)] min-h-[500px] flex flex-col gap-3 overflow-y-auto pb-24 pr-1">
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
        total={total}
        discountRate={ownerDiscountRate}
        discountAmount={globalDiscountAmount}
        onUpdateDiscountAmount={setGlobalDiscountAmount}
        isDiscountRateReadonly
        disabled={!canEdit}
      />

      {/* Comments & Remarks */}
      <div className="grid grid-cols-2 gap-3">
        <div className="flex flex-col gap-1">
          <Label className={`text-sm font-medium ${C.text60}`}>
            コメント
          </Label>
          <Textarea
            value={comment}
            onChange={(e) => setComment(e.target.value)}
            className={`${C.bgWhite} ${C.borderMedium} min-h-[60px] resize-none p-2 text-sm ${C.text}`}
            disabled={!canEdit}
          />
        </div>
        <div className="flex flex-col gap-1">
          <Label className={`text-sm font-medium ${C.text60}`}>
            備考
          </Label>
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
