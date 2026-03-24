import React, { memo, useState, useMemo, useCallback, useEffect, useTransition } from "react";
import { toast } from "sonner";
import { Textarea } from "@/components/ui/textarea";
import { Label } from "@/components/ui/label";
import { Button } from "@/components/ui/button";
import { EstimateForm } from "./EstimateForm";
import { TreatmentTable, TreatmentItem } from "./TreatmentTable";
import { TreatmentDetailedSummary } from "./TreatmentDetailedSummary";
import {
  useGetEstimateByRecord,
  useCreateEstimateRecord,
  useUpdateEstimateRecord,
} from "../api/save-estimate";

interface MedicalRecordEstimateProps {
  isNewRecord?: boolean;
  ownerDiscountRate?: number;
  medicalRecordId?: string;
}

export const MedicalRecordEstimate = memo(function MedicalRecordEstimate({
  isNewRecord = false,
  ownerDiscountRate = 0,
  medicalRecordId,
}: MedicalRecordEstimateProps) {
  const [subject, setSubject] = useState("");
  const [comment, setComment] = useState("");
  const [remarks, setRemarks] = useState("");
  const [globalDiscountAmount, setGlobalDiscountAmount] = useState(0);
  const [isSavePending, startSaveTransition] = useTransition();

  const [items, setItems] = useState<TreatmentItem[]>([]);

  // Load existing estimate
  const { data: existingEstimate } = useGetEstimateByRecord(
    isNewRecord ? undefined : medicalRecordId,
  );
  const createEstimate = useCreateEstimateRecord(medicalRecordId ?? "");
  const updateEstimate = useUpdateEstimateRecord(
    existingEstimate?.id ?? 0,
    medicalRecordId ?? "",
  );

  // Populate form from existing estimate
  useEffect(() => {
    if (existingEstimate) {
      setSubject(existingEstimate.title ?? "");
      setComment(existingEstimate.comment ?? "");
      setRemarks(existingEstimate.notes ?? "");
      setGlobalDiscountAmount(existingEstimate.discount_amount ?? 0);
    }
  }, [existingEstimate]);

  const handleAddItem = useCallback(() => {
    setItems((prev) => [
      ...prev,
      {
        id: Date.now(),
        content: "",
        memo: "",
        insurance: false,
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
    const t = Math.floor(sub * 0.1);
    return { subtotal: sub, tax: t, total: sub + t };
  }, [items]);

  const handleSave = useCallback(() => {
    if (!medicalRecordId) {
      toast.error("カルテを保存してから見積書を保存してください");
      return;
    }
    if (!subject.trim()) {
      toast.error("件名を入力してください");
      return;
    }

    const payload = {
      title: subject,
      subtotal,
      tax_total: tax,
      total_amount: total - globalDiscountAmount,
      discount_amount: globalDiscountAmount,
      comment,
      notes: remarks,
      medical_record_id: Number(medicalRecordId),
    };

    startSaveTransition(async () => {
      try {
        if (existingEstimate) {
          await updateEstimate.mutateAsync(payload);
        } else {
          await createEstimate.mutateAsync(payload);
        }
        toast.success("見積書を保存しました");
      } catch {
        toast.error("保存に失敗しました");
      }
    });
  }, [
    medicalRecordId,
    subject,
    subtotal,
    tax,
    total,
    globalDiscountAmount,
    comment,
    remarks,
    existingEstimate,
    updateEstimate,
    createEstimate,
  ]);

  const handlePdfExport = useCallback(() => {
    toast.info("PDF出力機能は準備中です");
  }, []);

  return (
    <div className="h-[calc(100vh-220px)] min-h-[500px] flex flex-col gap-3 overflow-y-auto pb-10 pr-1">
      {/* Subject */}
      <EstimateForm subject={subject} onSubjectChange={setSubject} />

      {/* Items Table */}
      <TreatmentTable
        items={items}
        onUpdate={handleUpdateItem}
        onRemove={handleRemoveItem}
        onAddRow={handleAddItem}
        showStatus={false}
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
      />

      {/* Comments & Remarks */}
      <div className="grid grid-cols-2 gap-3">
        <div className="flex flex-col gap-1">
          <Label className="text-sm font-medium text-[#37352F]/60">
            コメント
          </Label>
          <Textarea
            value={comment}
            onChange={(e) => setComment(e.target.value)}
            className="bg-white border-[rgba(55,53,47,0.16)] min-h-[60px] resize-none p-2 text-sm text-[#37352F]"
          />
        </div>
        <div className="flex flex-col gap-1">
          <Label className="text-sm font-medium text-[#37352F]/60">
            備考
          </Label>
          <Textarea
            value={remarks}
            onChange={(e) => setRemarks(e.target.value)}
            className="bg-white border-[rgba(55,53,47,0.16)] min-h-[60px] resize-none p-2 text-sm text-[#37352F]"
          />
        </div>
      </div>

      {/* Action Buttons */}
      <div className="flex justify-end gap-2 pt-2">
        <Button
          variant="outline"
          className="h-10 px-4 text-sm border-[rgba(55,53,47,0.16)] text-[#37352F] hover:bg-[#F7F6F3]"
          onClick={handlePdfExport}
          disabled={isSavePending}
        >
          PDF出力
        </Button>
        <Button
          className="h-10 px-4 text-sm bg-[#37352F] text-white hover:bg-[#37352F]/90"
          onClick={handleSave}
          disabled={isSavePending || isNewRecord}
        >
          {isSavePending ? "保存中..." : "保存"}
        </Button>
      </div>
    </div>
  );
});
