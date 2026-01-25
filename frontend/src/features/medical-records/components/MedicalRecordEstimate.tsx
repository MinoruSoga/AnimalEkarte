import React, { useState, useMemo, useCallback } from "react";
import { Button } from "../../../components/ui/button";
import { Textarea } from "../../../components/ui/textarea";
import { Label } from "../../../components/ui/label";
import { EstimateForm } from "./EstimateForm";
import { TreatmentTable, TreatmentItem } from "./TreatmentTable";
import { TreatmentDetailedSummary } from "./TreatmentDetailedSummary";

export function MedicalRecordEstimate({ isNewRecord = false }: { isNewRecord?: boolean }) {
  const [subject, setSubject] = useState("");
  const [comment, setComment] = useState("");
  const [remarks, setRemarks] = useState("");
  const [globalDiscountRate, setGlobalDiscountRate] = useState(0);
  const [globalDiscountAmount, setGlobalDiscountAmount] = useState(0);

  // Mock data for Estimate items
  const [items, setItems] = useState<TreatmentItem[]>(
    isNewRecord
      ? []
      : [
          {
            id: 1,
            content: "recheck(新料金)1",
            memo: "再診科099",
            insurance: true,
            unitPrice: 990,
            quantity: 1,
            discountRate: 0,
            discountAmount: 0,
          },
        ]
  );

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

  const handleUpdateItem = useCallback((id: number, field: keyof TreatmentItem, value: string | number | boolean) => {
    setItems((prev) =>
      prev.map((item) => (item.id === id ? { ...item, [field]: value } : item))
    );
  }, []);

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
        discountRate={globalDiscountRate}
        discountAmount={globalDiscountAmount}
        onUpdateDiscountRate={setGlobalDiscountRate}
        onUpdateDiscountAmount={setGlobalDiscountAmount}
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
      <div className="flex justify-end gap-2 mt-2">
        <Button
          size="sm"
          className="bg-[#37352F] hover:bg-[#37352F]/90 text-white min-w-[70px] h-10 text-sm shadow-sm border-transparent"
        >
          保存
        </Button>
        <Button
          size="sm"
          className="bg-[#37352F] hover:bg-[#37352F]/90 text-white min-w-[70px] h-10 text-sm shadow-sm border-transparent"
        >
          PDF出力
        </Button>
      </div>
    </div>
  );
}
