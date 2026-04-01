import { lazy, memo, Suspense, useState, useMemo, useCallback, useTransition } from "react";
import { toast } from "sonner";
import { Button } from "@/components/ui/button";
import { handleApiError } from "@/lib/handle-api-error";
import { TreatmentTable, TreatmentItem } from "./TreatmentTable";
import { TreatmentDetailedSummary } from "./TreatmentDetailedSummary";
import { useGetTreatments, useCreateTreatment, useUpdateTreatment, useDeleteTreatment } from "../api/treatments";
import { useGetBillingReview, useConfirmBillingReview, useReturnBillingReview } from "../api/billing-review";
import type { CreateTreatmentInput, UpdateTreatmentInput, TreatmentItemType } from "../types";
import { useAuth } from "@/features/auth/hooks/use-auth";
import { CheckCircle2, RotateCcw } from "lucide-react";
import { C, ICON, STYLE } from "@/lib/design-tokens";
import type { TreatmentMasterItem } from "@/components/shared/TreatmentSearchDialog/TreatmentSearchDialog";
import { calculateBillingTotals } from "@/lib/calculations";

const TreatmentSearchDialog = lazy(() =>
  import("@/components/shared/TreatmentSearchDialog/TreatmentSearchDialog").then((m) => ({
    default: m.TreatmentSearchDialog,
  }))
);

interface BillCheckProps {
  isNewRecord?: boolean;
  medicalRecordId?: string;
  ownerDiscountRate?: number;
}

export const MedicalRecordBillCheck = memo(function MedicalRecordBillCheck({ isNewRecord = false, medicalRecordId = "", ownerDiscountRate = 0 }: BillCheckProps) {
  const { user } = useAuth();
  const [globalDiscountAmount, setGlobalDiscountAmount] = useState(0);
  const [isSearchOpen, setIsSearchOpen] = useState(false);

  // ── API ──
  const { data: treatments = [] } = useGetTreatments(medicalRecordId);
  const { data: billingReview } = useGetBillingReview(medicalRecordId);
  const createTreatmentMutation = useCreateTreatment(medicalRecordId);
  const updateTreatmentMutation = useUpdateTreatment(medicalRecordId);
  const confirmMutation = useConfirmBillingReview(medicalRecordId);
  const returnMutation = useReturnBillingReview(medicalRecordId);

  const [isConfirmPending, startConfirmTransition] = useTransition();
  const userId = user?.id;

  const handleConfirm = useCallback(() => {
    startConfirmTransition(async () => {
      try {
        await confirmMutation.mutateAsync({
          confirmed_by: Number(userId ?? 0),
          memo: "医師確認済み",
        });
        toast.success("会計確認を完了しました");
      } catch (error) {
        handleApiError(error, "会計確認");
      }
    });
  }, [confirmMutation, userId]);

  const handleReturn = useCallback(() => {
    returnMutation.mutate({
      return_reason: "医師による差し戻し",
    }, {
      onSuccess: () => {
        toast.success("会計確認を差し戻しました");
      }
    });
  }, [returnMutation]);

  const items: TreatmentItem[] = useMemo(() => {
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

  const handleUpdateItem = useCallback((id: number, field: keyof TreatmentItem, value: string | number | boolean) => {
    const input: UpdateTreatmentInput = {};
    if (field === "content") input.content = String(value);
    if (field === "memo") input.memo = String(value);
    if (field === "insurance") input.insurance = Boolean(value);
    if (field === "unitPrice") input.unit_price = Number(value);
    if (field === "quantity") input.quantity = Number(value);
    if (field === "discountRate") input.discount_rate = Number(value) / 100;
    if (field === "discountAmount") input.discount_amount = Number(value);
    if (field === "status") input.status = String(value);
    if (field === "selected") input.selected = Boolean(value);

    updateTreatmentMutation.mutate({ treatmentId: String(id), input });
  }, [updateTreatmentMutation]);

  const deleteMutation = useDeleteTreatment(medicalRecordId);

  const handleRemoveItem = useCallback((id: number) => {
    deleteMutation.mutate(String(id));
  }, [deleteMutation]);

  const handleSelectTreatment = useCallback((item: TreatmentMasterItem) => {
    const nextOrder = treatments.length > 0 ? Math.max(...treatments.map(t => t.sort_order)) + 1 : 0;
    const input: CreateTreatmentInput = {
      item_type: (item.category === "薬品" ? "medicine" : item.category === "処置" ? "procedure" : "other") as TreatmentItemType,
      content: item.name,
      memo: item.category,
      unit_price: item.unitPrice,
      quantity: 1,
      selected: true,
      insurance: true,
      discount_amount: 0,
      sort_order: nextOrder,
    };
    createTreatmentMutation.mutate(input);
    setIsSearchOpen(false);
  }, [treatments, createTreatmentMutation]);

  const { subtotal, tax, total } = useMemo(() => {
    const result = calculateBillingTotals(items, ownerDiscountRate, globalDiscountAmount);
    return {
      subtotal: result.subtotal,
      tax: result.tax,
      total: result.total
    };
  }, [items, ownerDiscountRate, globalDiscountAmount]);

  if (isNewRecord) {
    return (
      <div className="flex flex-col items-center justify-center p-12 bg-white rounded-lg border border-dashed text-[#37352F]/40">
        カルテを保存してから会計確認を行えます
      </div>
    );
  }

  const isConfirmed = billingReview?.status === "confirmed";

  return (
    <div className="flex flex-col gap-4 relative h-full">
      <div className="flex-1 min-h-0 flex flex-col">
        <div className="flex items-center justify-between mb-2">
          <h2 className={`text-sm font-bold ${C.text}`}>会計確認 (医師)</h2>
          {isConfirmed ? (
            <div className={`px-2 py-1 rounded bg-[#DDEDEA] text-[#0F7B6C] text-xs font-bold flex items-center gap-1`}>
              <CheckCircle2 className="size-3" />
              確認済み
            </div>
          ) : (
            <div className={`px-2 py-1 rounded bg-[#F1F0EE] text-[#787774] text-xs font-bold`}>
              未確認
            </div>
          )}
        </div>

        <div className="flex-1 min-h-0 bg-white rounded-lg border border-[rgba(55,53,47,0.09)] overflow-hidden flex flex-col">
          <div className="flex-1 min-h-0">
            <TreatmentTable
              items={items}
              onUpdate={handleUpdateItem}
              onRemove={handleRemoveItem}
              onOpenSearch={() => setIsSearchOpen(true)}
              disabled={isConfirmed}
            />
          </div>
        </div>

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

      {/* Action Button */}
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
            disabled={isConfirmPending || items.length === 0}
            onClick={handleConfirm}
            className={`${STYLE.confirmPrimary} min-w-[120px] shadow-lg h-10 text-sm gap-2`}
          >
            <CheckCircle2 className={ICON.action} />
            {isConfirmPending ? "処理中..." : "チェック完了"}
          </Button>
        )}
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
