import React, { lazy, Suspense, useState, useMemo, useCallback } from "react";
import type { TreatmentMasterItem } from "@/components/shared/TreatmentSearchDialog/TreatmentSearchDialog";

const TreatmentSearchDialog = lazy(() =>
  import("@/components/shared/TreatmentSearchDialog/TreatmentSearchDialog").then((m) => ({
    default: m.TreatmentSearchDialog,
  }))
);
import { TreatmentTable, TreatmentItem } from "./TreatmentTable";
import { DiagnosisHeader } from "./DiagnosisHeader";
import { ClinicalPlanSection } from "./ClinicalPlanSection";
import { TreatmentDetailedSummary } from "./TreatmentDetailedSummary";

export interface DiagnosisPlanProps {
  isNewRecord?: boolean;
  items?: TreatmentItem[];
  setItems?: React.Dispatch<React.SetStateAction<TreatmentItem[]>>;
  // 制御型props（親フックから状態を受け取る）
  plan?: string;
  setPlan?: (value: string) => void;
  assessment?: string;
  setAssessment?: (value: string) => void;
  diagnosis1CategoryId?: number | null;
  setDiagnosis1CategoryId?: (id: number | null) => void;
  diagnosis1NameId?: number | null;
  setDiagnosis1NameId?: (id: number | null) => void;
  diagnosis2CategoryId?: number | null;
  setDiagnosis2CategoryId?: (id: number | null) => void;
  diagnosis2NameId?: number | null;
  setDiagnosis2NameId?: (id: number | null) => void;
  medicalRecordId?: string;
}

export function MedicalRecordDiagnosisPlan({
  isNewRecord = false,
  plan: planProp,
  setPlan: setPlanProp,
  assessment: assessmentProp,
  setAssessment: setAssessmentProp,
  diagnosis1CategoryId,
  setDiagnosis1CategoryId,
  diagnosis1NameId,
  setDiagnosis1NameId,
  diagnosis2CategoryId,
  setDiagnosis2CategoryId,
  diagnosis2NameId,
  setDiagnosis2NameId,
  medicalRecordId,
}: DiagnosisPlanProps) {
  // 制御型propsが渡された場合はそれを使用し、渡されない場合は内部stateにフォールバック
  const [internalPolicy, setInternalPolicy] = useState("# 治療方針");
  const [internalDiagnosisDetails, setInternalDiagnosisDetails] = useState("# 診断詳細");

  const policy = planProp ?? internalPolicy;
  const setPolicy = setPlanProp ?? setInternalPolicy;
  const diagnosisDetails = assessmentProp ?? internalDiagnosisDetails;
  const setDiagnosisDetails = setAssessmentProp ?? setInternalDiagnosisDetails;
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
        diagnosis1CategoryId={diagnosis1CategoryId}
        setDiagnosis1CategoryId={setDiagnosis1CategoryId}
        diagnosis1NameId={diagnosis1NameId}
        setDiagnosis1NameId={setDiagnosis1NameId}
        diagnosis2CategoryId={diagnosis2CategoryId}
        setDiagnosis2CategoryId={setDiagnosis2CategoryId}
        diagnosis2NameId={diagnosis2NameId}
        setDiagnosis2NameId={setDiagnosis2NameId}
      />

      {!isNewRecord && medicalRecordId ? (
        <ClinicalPlanSection medicalRecordId={medicalRecordId} />
      ) : null}

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

      <Suspense fallback={null}>
        <TreatmentSearchDialog
          open={isSearchOpen}
          onOpenChange={setIsSearchOpen}
          onSelect={handleSelectTreatment}
        />
      </Suspense>
    </div>
  );
}
