import { lazy, memo, Suspense, useState, useMemo, useCallback } from "react";
import type { TreatmentMasterItem } from "@/components/shared/TreatmentSearchDialog/TreatmentSearchDialog";

const TreatmentSearchDialog = lazy(() =>
  import("@/components/shared/TreatmentSearchDialog/TreatmentSearchDialog").then((m) => ({
    default: m.TreatmentSearchDialog,
  }))
);
import { TreatmentTable, TreatmentItem } from "./TreatmentTable";
import { DiagnosisHeader } from "./DiagnosisHeader";
import { ClinicalPlanSection } from "./ClinicalPlanSection/ClinicalPlanSection";
import { TreatmentDetailedSummary } from "./TreatmentDetailedSummary";
import { useGetClinicalPlan } from "../api/clinical-plan";
import { useGetTreatments, useCreateTreatment, useUpdateTreatment, useDeleteTreatment } from "../api/treatments";
import type { TreatmentItemType, UpdateTreatmentInput } from "../types";
import { usePermission } from "@/hooks/use-permission";
import { C, LAYOUT } from "@/lib/design-tokens";
import { calculateBillingTotals } from "@/lib/calculations";

export interface DiagnosisPlanProps {
  isNewRecord?: boolean;
  chiefComplaint?: string;
  // 制御型props（親フックから状態を受け取る）— clinical_plan 3欄の単一 owner
  physicalExam: string;
  setPhysicalExam: (value: string) => void;
  plan: string;
  setPlan: (value: string) => void;
  assessment: string;
  setAssessment: (value: string) => void;
  diagnosis1CategoryId: number | null;
  setDiagnosis1CategoryId: (id: number | null) => void;
  diagnosis1NameId: number | null;
  setDiagnosis1NameId: (id: number | null) => void;
  diagnosis2CategoryId: number | null;
  setDiagnosis2CategoryId: (id: number | null) => void;
  diagnosis2NameId: number | null;
  setDiagnosis2NameId: (id: number | null) => void;
  medicalRecordId?: string;
  ownerDiscountRate?: number;
  diagnosis1NameIdError?: string | null;
  /** P2-15: 拠点横断で開いたカルテの子リソース操作用。レコード自身の clinicId */
  recordClinicId?: string;
}

export const MedicalRecordDiagnosisPlan = memo(function MedicalRecordDiagnosisPlan({
  isNewRecord = false,
  chiefComplaint,
  physicalExam,
  setPhysicalExam,
  plan,
  setPlan,
  assessment,
  setAssessment,
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
  diagnosis1NameIdError,
  recordClinicId,
}: DiagnosisPlanProps) {
  const { canCreate, canEdit, canDelete } = usePermission("medical-records");
  const [isSearchOpen, setIsSearchOpen] = useState(false);
  const [globalDiscountAmount, setGlobalDiscountAmount] = useState(0);

  // ── API ──
  const { data: treatments = [] } = useGetTreatments(medicalRecordId ?? "", recordClinicId);
  const { data: clinicalPlan } = useGetClinicalPlan(medicalRecordId ?? "", recordClinicId);
  const createMutation = useCreateTreatment(medicalRecordId ?? "", recordClinicId);
  const { mutate: createTreatmentFn } = createMutation;
  const { mutate: updateTreatmentFn } = useUpdateTreatment(medicalRecordId ?? "", recordClinicId);
  const { mutate: deleteTreatmentFn } = useDeleteTreatment(medicalRecordId ?? "", recordClinicId);

  // Treatment[] (Backend) -> TreatmentItem[] (Generic Table) 変換
  const treatmentItems: TreatmentItem[] = useMemo(() => {
    return treatments.map(t => ({
      id: Number(t.id),
      content: t.content,
      memo: t.memo,
      is_insurance: t.is_insurance,
      unitPrice: t.unit_price,
      quantity: t.quantity,
      discountRate: t.discount_rate,
      discountAmount: t.discount_amount,
      status: t.status,
      is_selected: t.is_selected
    }));
  }, [treatments]);

  // rerender-dependencies: treatments 配列を deps から除外するため nextOrder を useMemo で事前計算
  const nextOrder = useMemo(
    () => treatments.length > 0 ? Math.max(...treatments.map(t => t.sort_order)) + 1 : 0,
    [treatments],
  );

  const handleRemoveItem = useCallback((id: number) => {
    if (!canDelete) return;
    deleteTreatmentFn(String(id));
  }, [canDelete, deleteTreatmentFn]);

  const handleUpdateItem = useCallback((id: number, field: keyof TreatmentItem, value: string | number | boolean) => {
    if (!canEdit) return;
    const input: UpdateTreatmentInput = {};
    if (field === "content") input.content = String(value);
    if (field === "memo") input.memo = String(value);
    if (field === "is_insurance") input.is_insurance = Boolean(value);
    if (field === "unitPrice") input.unit_price = Number(value);
    if (field === "quantity") input.quantity = Number(value);
    if (field === "discountRate") input.discount_rate = Number(value) / 100;
    if (field === "discountAmount") input.discount_amount = Number(value);
    if (field === "status") input.status = String(value);
    if (field === "is_selected") input.is_selected = Boolean(value);

    updateTreatmentFn({ treatmentId: String(id), input });
  }, [canEdit, updateTreatmentFn]);

  const handleAddRow = useCallback(() => {
    if (!canCreate) return;
    createTreatmentFn({
      item_type: "other" as TreatmentItemType,
      content: "",
      unit_price: 0,
      quantity: 1,
      is_selected: true,
      is_insurance: false,
      discount_amount: 0,
      sort_order: nextOrder,
    });
  }, [canCreate, nextOrder, createTreatmentFn]);

  const handleSelectTreatment = useCallback((item: TreatmentMasterItem) => {
    if (!canCreate) return;
    createTreatmentFn({
      item_type: (item.category === "薬品" ? "medicine" : item.category === "処置" ? "procedure" : "other") as TreatmentItemType,
      content: item.name,
      memo: item.category,
      unit_price: item.unitPrice,
      quantity: 1,
      is_selected: true,
      is_insurance: true,
      discount_amount: 0,
      sort_order: nextOrder,
    });
  }, [canCreate, nextOrder, createTreatmentFn]);

  // Calculations
  const { subtotal, tax, total } = useMemo(() => {
    const result = calculateBillingTotals(treatmentItems, ownerDiscountRate, globalDiscountAmount);
    return {
      subtotal: result.subtotal,
      tax: result.tax,
      total: result.total
    };
  }, [treatmentItems, ownerDiscountRate, globalDiscountAmount]);

  return (
    <div className={`gap-3 ${LAYOUT.fullHeight}`}>
      <DiagnosisHeader
        chiefComplaint={chiefComplaint}
        physicalExam={physicalExam}
        setPhysicalExam={setPhysicalExam}
        diagnosisDetails={assessment}
        setDiagnosisDetails={setAssessment}
        diagnosis1CategoryId={diagnosis1CategoryId}
        setDiagnosis1CategoryId={setDiagnosis1CategoryId}
        diagnosis1NameId={diagnosis1NameId}
        setDiagnosis1NameId={setDiagnosis1NameId}
        diagnosis2CategoryId={diagnosis2CategoryId}
        setDiagnosis2CategoryId={setDiagnosis2CategoryId}
        diagnosis2NameId={diagnosis2NameId}
        setDiagnosis2NameId={setDiagnosis2NameId}
        diagnosis1NameIdError={diagnosis1NameIdError}
        selectedDiagnosisType={clinicalPlan?.diagnosis_type}
        selectedDiagnosisName={clinicalPlan?.diagnosis_name}
        selectedDiagnosis2Type={clinicalPlan?.diagnosis_2_type}
        selectedDiagnosis2Name={clinicalPlan?.diagnosis_2_name}
      />

      {!isNewRecord && medicalRecordId ? (
        <div className="shrink-0">
          <ClinicalPlanSection
            medicalRecordId={medicalRecordId}
            canEdit={canEdit}
            recordClinicId={recordClinicId}
            physicalExam={physicalExam}
            onPhysicalExamChange={setPhysicalExam}
            diagnosisDetails={assessment}
            onDiagnosisDetailsChange={setAssessment}
            treatmentPolicy={plan}
            onTreatmentPolicyChange={setPlan}
            diagnosisTypeId={diagnosis1CategoryId}
            onDiagnosisTypeIdChange={setDiagnosis1CategoryId}
            diagnosisNameId={diagnosis1NameId}
            onDiagnosisNameIdChange={setDiagnosis1NameId}
          />
        </div>
      ) : null}

      {/* Bottom Section: Treatment Plan */}
      <div className="flex-1 min-h-0 flex flex-col">
        <h2 className={`text-sm font-bold ${C.text} mb-2`}>治療プラン</h2>

        <div className={`flex-1 min-h-0 flex flex-col ${C.bgWhite} rounded-lg border ${C.borderLight} overflow-hidden`}>
          {isNewRecord ? (
            <div className={`flex-1 flex items-center justify-center border border-dashed rounded-lg text-sm ${C.text40}`}>
              カルテを保存してから治療プランを作成できます
            </div>
          ) : (
            <div className="flex-1 min-h-0">
              <TreatmentTable
                items={treatmentItems}
                onUpdate={handleUpdateItem}
                onRemove={handleRemoveItem}
                onOpenSearch={canCreate ? () => setIsSearchOpen(true) : undefined}
                onAddRow={canCreate ? handleAddRow : undefined}
                showStatus={true}
                disabled={!canEdit && !canCreate ? !canDelete : false}
              />
            </div>
          )}
        </div>

        <div className="shrink-0 mt-2">
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
