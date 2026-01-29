import React, { useState, useMemo, useCallback } from "react";
import { TreatmentSearchDialog } from "@/components/shared/TreatmentSearchDialog";
import type { TreatmentMasterItem } from "@/components/shared/TreatmentSearchDialog";
import { TreatmentTable, TreatmentItem } from "./TreatmentTable";
import { DiagnosisHeader } from "./DiagnosisHeader";
import { TreatmentDetailedSummary } from "./TreatmentDetailedSummary";

export interface DiagnosisPlanProps {
  isNewRecord?: boolean;
  items?: TreatmentItem[];
  setItems?: React.Dispatch<React.SetStateAction<TreatmentItem[]>>;
}

export function MedicalRecordDiagnosisPlan({ isNewRecord = false }: DiagnosisPlanProps) {
  const [policy, setPolicy] = useState("# 治療方針");
  const [diagnosisDetails, setDiagnosisDetails] = useState("# 診断詳細");
  const [isSearchOpen, setIsSearchOpen] = useState(false);
  const [globalDiscountRate, setGlobalDiscountRate] = useState(0);
  const [globalDiscountAmount, setGlobalDiscountAmount] = useState(0);

  // Initial Data
  const [treatmentItems, setTreatmentItems] = useState<TreatmentItem[]>(
    isNewRecord
      ? []
      : [
          {
            id: 1,
            content: "recheck(新料金)",
            memo: "再診料099",
            insurance: true,
            unitPrice: 990,
            quantity: 1,
            discountRate: 0,
            discountAmount: 0,
          },
          {
            id: 2,
            content: "血尿治療Aプラン",
            memo: "血尿治療Aプラン",
            insurance: false,
            unitPrice: 990,
            quantity: 1,
            discountRate: 0,
            discountAmount: 0,
          },
        ]
  );

  const handleRemoveItem = useCallback((id: number) => {
    setTreatmentItems((prev) => prev.filter((item) => item.id !== id));
  }, []);

  const handleUpdateItem = useCallback((id: number, field: keyof TreatmentItem, value: string | number | boolean) => {
    setTreatmentItems((prev) =>
      prev.map((item) => (item.id === id ? { ...item, [field]: value } : item))
    );
  }, []);

  const handleAddRow = useCallback(() => {
    setTreatmentItems((prev) => [
      ...prev,
      {
        id: Date.now(),
        content: "",
        memo: "",
        insurance: true,
        unitPrice: 0,
        quantity: 1,
        discountRate: 0,
        discountAmount: 0,
      },
    ]);
  }, []);

  const handleSelectTreatment = useCallback((item: TreatmentMasterItem) => {
    setTreatmentItems((prev) => [
      ...prev,
      {
        id: Date.now(),
        content: item.name,
        memo: item.category,
        insurance: true,
        unitPrice: item.unitPrice,
        quantity: 1,
        discountRate: 0,
        discountAmount: 0,
      },
    ]);
  }, []);

  // Calculations
  const { subtotal, tax, total } = useMemo(() => {
    const sub = treatmentItems.reduce(
      (sum, item) => sum + item.unitPrice * item.quantity - item.discountAmount,
      0
    );
    const t = Math.floor(sub * 0.1);
    return { subtotal: sub, tax: t, total: sub + t };
  }, [treatmentItems]);

  return (
    <div className="flex flex-col gap-3 h-[calc(100vh-220px)] min-h-[500px]">
      <DiagnosisHeader 
        policy={policy}
        setPolicy={setPolicy}
        diagnosisDetails={diagnosisDetails}
        setDiagnosisDetails={setDiagnosisDetails}
      />

      {/* Bottom Section: Treatment Plan */}
      <div className="flex-1 flex flex-col min-h-0">
        <h2 className="text-sm font-bold text-[#37352F] mb-1.5">治療プラン</h2>

        <TreatmentTable 
          items={treatmentItems}
          onUpdate={handleUpdateItem}
          onRemove={handleRemoveItem}
          onOpenSearch={() => setIsSearchOpen(true)}
          onAddRow={handleAddRow}
        />

        <TreatmentDetailedSummary 
            subtotal={subtotal}
            tax={tax}
            total={total}
            discountRate={globalDiscountRate}
            discountAmount={globalDiscountAmount}
            onUpdateDiscountRate={setGlobalDiscountRate}
            onUpdateDiscountAmount={setGlobalDiscountAmount}
        />
      </div>

      <TreatmentSearchDialog
        open={isSearchOpen}
        onOpenChange={setIsSearchOpen}
        onSelect={handleSelectTreatment}
      />
    </div>
  );
}
