import { lazy, memo, Suspense, useState, useMemo, useCallback, useTransition } from "react";
import { toast } from "sonner";
import { Button } from "@/components/ui/button";
import { TreatmentTable, TreatmentItem } from "./TreatmentTable";
import { TreatmentDetailedSummary } from "./TreatmentDetailedSummary";
import { useGetTreatments, useCreateTreatment, useUpdateTreatment, useDeleteTreatment } from "../api/treatments";
import { useGetBillingConfirmation, useCreateBillingConfirmation, useCreateBillingReturn } from "../api/billing-confirmation";
import type { CreateTreatmentInput, UpdateTreatmentInput } from "../types";
import { resolveItemTypeFromCategory } from "./TreatmentsTab/treatments-tab-model";
import { usePermission } from "@/hooks/use-permission";
import { useClinicTaxRates } from "@/hooks/use-clinic-tax-rates";
import { CheckCircle2, RotateCcw } from "lucide-react";
import { C, ICON } from "@/lib/design-tokens";
import type { TreatmentMasterItem } from "@/components/shared/TreatmentSearchDialog/TreatmentSearchDialog";
import { calculateBillingTotals } from "@/lib/calculations";
import { formatCurrency } from "@/lib/format/number";
import { useGetRecordExaminations } from "../api/get-record-examinations";
import { useGetPetVaccinations } from "../api/get-pet-vaccinations";
import {
  billCheckExtraLines,
  billCheckPricedExtras,
  isUnbillableMasterPrice,
  type BillCheckExtraLine,
} from "./medical-record-bill-check-model";

const TreatmentSearchDialog = lazy(() =>
  import("@/components/shared/TreatmentSearchDialog/TreatmentSearchDialog").then((m) => ({
    default: m.TreatmentSearchDialog,
  }))
);

interface BillCheckProps {
  isNewRecord?: boolean;
  medicalRecordId?: string;
  petId?: string;
  ownerDiscountRate?: number;
  /** P2-15: 拠点横断で開いたカルテの子リソース操作用。レコード自身の clinicId */
  recordClinicId?: string;
}

function ExtraLinesList({ title, lines }: { title: string; lines: BillCheckExtraLine[] }) {
  if (lines.length === 0) {
    return null;
  }
  return (
    <div className="mt-3">
      <h3 className={`mb-1 text-xs font-bold ${C.text60}`}>{title}</h3>
      <ul className={`divide-y ${C.borderLight} rounded-lg border ${C.borderLight} ${C.bgWhite}`}>
        {lines.map((line) => (
          <li key={line.id} className="flex items-center justify-between gap-3 px-3 py-2 text-sm">
            <span className={C.text}>{line.name}</span>
            <span className={isUnbillableMasterPrice(line.unitPrice) ? C.danger : C.text}>
              {isUnbillableMasterPrice(line.unitPrice) ? "価格未設定" : formatCurrency(line.unitPrice)}
            </span>
          </li>
        ))}
      </ul>
    </div>
  );
}

export const MedicalRecordBillCheck = memo(function MedicalRecordBillCheck({
  isNewRecord = false,
  medicalRecordId = "",
  petId,
  ownerDiscountRate = 0,
  recordClinicId,
}: BillCheckProps) {
  const { canEdit, canDelete } = usePermission("medical-records");
  const [globalDiscountAmount, setGlobalDiscountAmount] = useState(0);
  const [isSearchOpen, setIsSearchOpen] = useState(false);

  const { data: treatments = [] } = useGetTreatments(medicalRecordId, recordClinicId);
  const { data: billingConfirmation } = useGetBillingConfirmation(medicalRecordId);
  const { data: examinationResult } = useGetRecordExaminations(
    isNewRecord ? undefined : petId,
    isNewRecord ? undefined : medicalRecordId,
  );
  const { data: vaccinations = [] } = useGetPetVaccinations(isNewRecord ? undefined : petId);
  const createTreatmentMutation = useCreateTreatment(medicalRecordId, recordClinicId);
  const { mutate: updateTreatment } = useUpdateTreatment(medicalRecordId, recordClinicId);
  const confirmMutation = useCreateBillingConfirmation(medicalRecordId);
  const returnMutation = useCreateBillingReturn(medicalRecordId);

  const [isConfirmPending, startConfirmTransition] = useTransition();

  const examLines = useMemo(
    () => billCheckExtraLines(
      "exam",
      (examinationResult?.items ?? []).map((exam) => ({
        id: exam.id,
        name: exam.name,
        price: exam.price,
        medicalRecordId: exam.medicalRecordId,
      })),
      medicalRecordId,
    ),
    [examinationResult?.items, medicalRecordId],
  );
  const vaccinationLines = useMemo(
    () => billCheckExtraLines(
      "vaccination",
      vaccinations.map((item) => ({
        id: item.id,
        name: item.name,
        price: item.price,
        medicalRecordId: item.medicalRecordId,
      })),
      medicalRecordId,
    ),
    [vaccinations, medicalRecordId],
  );
  const extraLines = useMemo(
    () => [...examLines, ...vaccinationLines],
    [examLines, vaccinationLines],
  );
  const pricedExtras = useMemo(() => billCheckPricedExtras(extraLines), [extraLines]);

  const { mutateAsync: confirmBillingAsync } = confirmMutation;
  const handleConfirm = useCallback(() => {
    if (!canEdit) return;
    if (extraLines.some((line) => isUnbillableMasterPrice(line.unitPrice))) {
      toast.error("価格が未設定の検査・接種があるため会計確認できません。マスタの価格を設定してください。");
      return;
    }
    if (treatments.length === 0 && extraLines.length === 0) {
      toast.error("会計確認する明細がありません。診察処置・検査・接種を追加してください。");
      return;
    }
    startConfirmTransition(async () => {
      try {
        await confirmBillingAsync({
          memo: "医師確認済み",
        });
        toast.success("会計確認を完了しました");
      } catch {
        // FE-RC-005: useCreateBillingConfirmation の onError が既に handleApiError
        // でトースト表示済み。ここで再度呼ぶと二重トーストになるため何もしない。
      }
    });
  }, [canEdit, confirmBillingAsync, extraLines, treatments.length]);

  const { mutate: returnBillingFn } = returnMutation;
  const handleReturn = useCallback(() => {
    if (!canEdit) return;
    returnBillingFn({
      return_reason: "医師による差し戻し",
    }, {
      onSuccess: () => {
        toast.success("会計確認を差し戻しました");
      },
      // FE-RC-005: useCreateBillingReturn の onError が既に handleApiError で
      // トースト表示済み。ここで onError を渡すと二重トーストになるため渡さない。
    });
  }, [canEdit, returnBillingFn]);

  const items: TreatmentItem[] = useMemo(() => {
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

    updateTreatment({ treatmentId: String(id), input });
  }, [canEdit, updateTreatment]);

  const { mutate: deleteTreatmentFn } = useDeleteTreatment(medicalRecordId, recordClinicId);

  const handleRemoveItem = useCallback((id: number) => {
    if (!canDelete) return;
    deleteTreatmentFn(String(id));
  }, [canDelete, deleteTreatmentFn]);

  const nextOrder = useMemo(
    () => treatments.reduce((maxOrder, treatment) => Math.max(maxOrder, treatment.sort_order), -1) + 1,
    [treatments],
  );

  const handleSelectTreatment = useCallback((item: TreatmentMasterItem) => {
    if (!canEdit) return;
    const input: CreateTreatmentInput = {
      item_type: resolveItemTypeFromCategory(item.category),
      content: item.name,
      memo: item.category,
      unit_price: item.unitPrice,
      quantity: 1,
      is_selected: true,
      is_insurance: true,
      discount_amount: 0,
      sort_order: nextOrder,
    };
    createTreatmentMutation.mutate(input);
    setIsSearchOpen(false);
  }, [canEdit, nextOrder, createTreatmentMutation]);

  // FE-RC-048: 消費税率はハードコード 0.1 ではなく病院マスタ設定（useClinicTaxRates）を正本にする。
  const { standardTaxRate } = useClinicTaxRates();
  const { subtotal, tax, total } = useMemo(() => {
    const extraItems = pricedExtras.map((line) => ({
      unitPrice: line.unitPrice ?? 0,
      quantity: 1,
      taxType: "excluded" as const,
      taxRate: standardTaxRate,
    }));
    const result = calculateBillingTotals(
      [...items, ...extraItems],
      ownerDiscountRate,
      globalDiscountAmount,
      standardTaxRate,
    );
    return {
      subtotal: result.subtotal,
      tax: result.tax,
      total: result.total
    };
  }, [items, pricedExtras, ownerDiscountRate, globalDiscountAmount, standardTaxRate]);

  if (isNewRecord) {
    return (
      <div className={`flex flex-col items-center justify-center p-12 ${C.bgWhite} rounded-lg border border-dashed ${C.text40}`}>
        カルテを保存してから会計確認を行えます
      </div>
    );
  }

  const isConfirmed = billingConfirmation?.status === "confirmed";

  return (
    <div className="flex flex-col gap-4 relative h-full">
      <div className="flex-1 min-h-0 flex flex-col">
        <div className="flex items-center justify-between mb-2">
          <h2 className={`text-sm font-bold ${C.text}`}>会計確認 (医師)</h2>
          {isConfirmed ? (
            <div className={`px-2 py-1 rounded ${C.bgStatusGreen} ${C.textStatusGreen} text-xs font-bold flex items-center gap-1`}>
              <CheckCircle2 className={ICON.xxs} />
              確認済み
            </div>
          ) : (
            <div className={`px-2 py-1 rounded ${C.bgMuted} ${C.textMuted} text-xs font-bold`}>
              未確認
            </div>
          )}
        </div>

        <div className={`flex-1 min-h-0 ${C.bgWhite} rounded-lg border ${C.borderLight} overflow-hidden flex flex-col`}>
          <div className="flex-1 min-h-0">
            <TreatmentTable
              items={items}
              onUpdate={handleUpdateItem}
              onRemove={handleRemoveItem}
              onOpenSearch={canEdit && !isConfirmed ? () => setIsSearchOpen(true) : undefined}
              disabled={isConfirmed || (!canEdit && !canDelete)}
            />
          </div>
        </div>

        <ExtraLinesList title="検査" lines={examLines} />
        <ExtraLinesList title="予防接種" lines={vaccinationLines} />

        <div className="mt-4">
          <TreatmentDetailedSummary
            subtotal={subtotal}
            tax={tax}
            total={total}
            discountRate={ownerDiscountRate}
            discountAmount={globalDiscountAmount}
            onUpdateDiscountAmount={setGlobalDiscountAmount}
            isDiscountRateReadonly
            disabled={isConfirmed}
          />
        </div>
      </div>

      {canEdit ? (
        <div className="fixed bottom-6 right-6 z-50 flex gap-2">
          {isConfirmed ? (
            <Button
              variant="outline"
              type="button"
              onClick={handleReturn}
              disabled={returnMutation.isPending}
              className={`h-10 text-sm gap-2 border ${C.borderMedium} ${C.hoverBgLight}`}
            >
              <RotateCcw className={ICON.action} />
              確認を取り消す
            </Button>
          ) : (
            <Button
              type="button"
              size="sm"
              disabled={isConfirmPending}
              onClick={handleConfirm}
              className={`${C.bgBrand} ${C.hoverBgBrand} ${C.hoverTextOnBrand} ${C.textOnBrand} rounded-full border-transparent min-w-[120px] h-10 text-sm gap-2 transition-colors`}
            >
              <CheckCircle2 className={ICON.action} />
              {isConfirmPending ? "処理中..." : "チェック完了"}
            </Button>
          )}
        </div>
      ) : null}

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
