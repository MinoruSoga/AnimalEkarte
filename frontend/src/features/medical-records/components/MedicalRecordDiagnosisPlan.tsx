import React, { lazy, memo, Suspense, useState, useMemo, useCallback } from "react";
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
import { useGetTreatments, useCreateTreatment, useUpdateTreatment, useDeleteTreatment } from "../api/treatments";
import type { TreatmentItemType, UpdateTreatmentInput } from "../types";

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
  ownerDiscountRate?: number;
}

export const MedicalRecordDiagnosisPlan = memo(function MedicalRecordDiagnosisPlan({
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
  ownerDiscountRate = 0,
}: DiagnosisPlanProps) {
  // 制御型propsが渡された場合はそれを使用し、渡されない場合は内部stateにフォールバック
  const [internalPolicy, setInternalPolicy] = useState("# 治療方針");
  const [internalDiagnosisDetails, setInternalDiagnosisDetails] = useState("# 診断詳細");

  const policy = planProp ?? internalPolicy;
  const setPolicy = setPlanProp ?? setInternalPolicy;
  const diagnosisDetails = assessmentProp ?? internalDiagnosisDetails;
  const setDiagnosisDetails = setAssessmentProp ?? setInternalDiagnosisDetails;
  const [isSearchOpen, setIsSearchOpen] = useState(false);
  const [globalDiscountAmount, setGlobalDiscountAmount] = useState(0);

  // ── API ──
  const { data: treatments = [] } = useGetTreatments(medicalRecordId ?? "");
  const createMutation = useCreateTreatment(medicalRecordId ?? "");
  const updateMutation = useUpdateTreatment(medicalRecordId ?? "");
  const deleteMutation = useDeleteTreatment(medicalRecordId ?? "");

  // Treatment[] (Backend) -> TreatmentItem[] (Generic Table) 変換
  const treatmentItems: TreatmentItem[] = useMemo(() => {
    return treatments.map(t => ({
      id: Number(t.id),
      content: t.content,
      memo: t.memo,
      insurance: t.insurance,
      unitPrice: t.unit_price,
      quantity: t.quantity,
      discountRate: t.discount_rate,
      discountAmount: t.discount_amount,
      status: t.status,
      selected: t.selected
    }));
  }, [treatments]);

  const handleRemoveItem = useCallback((id: number) => {
    deleteMutation.mutate(String(id));
  }, [deleteMutation]);

  const handleUpdateItem = useCallback((id: number, field: keyof TreatmentItem, value: string | number | boolean) => {
    const input: UpdateTreatmentInput = {};
    if (field === "content") input.content = String(value);
    if (field === "memo") input.memo = String(value);
    if (field === "insurance") input.insurance = Boolean(value);
    if (field === "unitPrice") input.unit_price = Number(value);
    if (field === "quantity") input.quantity = Number(value);
    if (field === "discountAmount") input.discount_amount = Number(value);
    if (field === "status") input.status = String(value);
    if (field === "selected") input.selected = Boolean(value);

    updateMutation.mutate({ treatmentId: String(id), input });
  }, [updateMutation]);

  const handleAddRow = useCallback(() => {
    const nextOrder = treatments.length > 0 ? Math.max(...treatments.map(t => t.sort_order)) + 1 : 0;
    createMutation.mutate({
      item_type: "other" as TreatmentItemType,
      content: "",
      unit_price: 0,
      quantity: 1,
      selected: true,
      insurance: false,
      discount_amount: 0,
      sort_order: nextOrder,
    });
  }, [treatments, createMutation]);

  const handleSelectTreatment = useCallback((item: TreatmentMasterItem) => {
    const nextOrder = treatments.length > 0 ? Math.max(...treatments.map(t => t.sort_order)) + 1 : 0;
    createMutation.mutate({
      item_type: (item.category === "薬品" ? "medicine" : item.category === "処置" ? "procedure" : "other") as TreatmentItemType,
      content: item.name,
      memo: item.category,
      unit_price: item.unitPrice,
      quantity: 1,
      selected: true,
      insurance: true,
      discount_amount: 0,
      sort_order: nextOrder,
    });
  }, [treatments, createMutation]);

  // Calculations
  const { subtotal, tax, total } = useMemo(() => {
    // 1. 各明細の合計
    const itemsSubtotal = treatmentItems.reduce((sum, item) => {
      const price = Number(item.unitPrice) || 0;
      const qty = Number(item.quantity) || 0;
      const itemDiscount = Number(item.discountAmount) || 0;
      return sum + (price * qty - itemDiscount);
    }, 0);

    // 2. 飼主割引（パーセント）を適用
    const ownerDiscountAmount = Math.floor(itemsSubtotal * (ownerDiscountRate / 100));
    const afterOwnerDiscount = itemsSubtotal - ownerDiscountAmount;

    // 3. 全体値引き（絶対額）を適用
    const afterGlobalDiscount = Math.max(0, afterOwnerDiscount - globalDiscountAmount);

    // 4. 消費税 (10%)
    const taxAmount = Math.floor(afterGlobalDiscount * 0.1);

    return { 
      subtotal: itemsSubtotal, 
      tax: taxAmount, 
      total: afterGlobalDiscount + taxAmount 
    };
  }, [treatmentItems, ownerDiscountRate, globalDiscountAmount]);

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

        {isNewRecord ? (
          <div className="flex-1 flex items-center justify-center border border-dashed rounded-lg text-sm text-[#37352F]/40">
            カルテを保存してから治療プランを作成できます
          </div>
        ) : (
          <TreatmentTable 
            items={treatmentItems}
            onUpdate={handleUpdateItem}
            onRemove={handleRemoveItem}
            onOpenSearch={() => setIsSearchOpen(true)}
            onAddRow={handleAddRow}
            showStatus={true}
          />
        )}

        <TreatmentDetailedSummary
            subtotal={subtotal}
            tax={tax}
            total={total}
            discountRate={ownerDiscountRate}
            discountAmount={globalDiscountAmount}
            onUpdateDiscountAmount={setGlobalDiscountAmount}
            isDiscountRateReadonly
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
});
