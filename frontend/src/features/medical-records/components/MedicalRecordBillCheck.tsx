import React, { lazy, memo, Suspense, useState, useMemo, useCallback } from "react";
import { toast } from "sonner";
import { Button } from "@/components/ui/button";
import { TreatmentTable, TreatmentItem } from "./TreatmentTable";
import { TreatmentDetailedSummary } from "./TreatmentDetailedSummary";
import { useTreatments, useCreateTreatment, useUpdateTreatment } from "../api/treatments";
import { useGetBillingReview, useConfirmBillingReview, useReturnBillingReview } from "../api/billing-review";
import type { CreateTreatmentInput, UpdateTreatmentInput, TreatmentItemType } from "../types";
import { useAuth } from "@/features/auth/hooks/use-auth";
import { CheckCircle2, RotateCcw } from "lucide-react";
import { C } from "@/lib/design-tokens";
import type { TreatmentMasterItem } from "@/components/shared/TreatmentSearchDialog/TreatmentSearchDialog";

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
  const { data: treatments = [] } = useTreatments(medicalRecordId);
  const { data: billingReview } = useGetBillingReview(medicalRecordId);
  const createTreatmentMutation = useCreateTreatment(medicalRecordId);
  const updateTreatmentMutation = useUpdateTreatment(medicalRecordId);
  const confirmMutation = useConfirmBillingReview(medicalRecordId);
  const returnMutation = useReturnBillingReview(medicalRecordId);

  // Treatment[] -> TreatmentItem[] (Backend は全件返すが、会計では selected=true のみ表示する場合が多い)
  // ここでは selected=true のものだけを会計対象とする
  const items: TreatmentItem[] = useMemo(() => {
    return treatments
      .filter(t => t.selected)
      .map(t => ({
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
    // 会計タブでの編集も backend へ同期
    const input: UpdateTreatmentInput = {};
    if (field === "content") input.content = String(value);
    if (field === "memo") input.memo = String(value);
    if (field === "insurance") input.insurance = Boolean(value);
    if (field === "unitPrice") input.unit_price = Number(value);
    if (field === "quantity") input.quantity = Number(value);
    if (field === "discountAmount") input.discount_amount = Number(value);
    if (field === "status") input.status = String(value);

    updateTreatmentMutation.mutate({ treatmentId: String(id), input });
  }, [updateTreatmentMutation]);

  const handleRemoveItem = useCallback((id: number) => {
    // 会計から外す = selected を false にする（データ自体は消さない）
    updateTreatmentMutation.mutate({ 
      treatmentId: String(id), 
      input: { selected: false } 
    });
  }, [updateTreatmentMutation]);

  const handleSelectTreatment = useCallback((item: TreatmentMasterItem) => {
    const nextOrder = treatments.length > 0 ? Math.max(...treatments.map((t) => t.sort_order)) + 1 : 0;
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

  const handleConfirm = useCallback(() => {
    confirmMutation.mutate({
      confirmed_by: Number(user?.id ?? 0),
      memo: "医師確認済み"
    }, {
      onSuccess: () => {
        toast.success("会計確認を完了しました");
      }
    });
  }, [confirmMutation, user]);

  const handleReturn = useCallback(() => {
    returnMutation.mutate({
      return_reason: "医師による差し戻し",
    }, {
      onSuccess: () => {
        toast.success("会計確認を差し戻しました");
      }
    });
  }, [returnMutation]);

  const { subtotal, tax, total } = useMemo(() => {
    // 1. 各明細の合計
    const itemsSubtotal = items.reduce((sum, item) => {
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
  }, [items, ownerDiscountRate, globalDiscountAmount]);

  if (isNewRecord) {
    return (
      <div className={`flex items-center justify-center h-48 text-sm ${C.text40}`}>
        カルテを保存してから会計確認を行えます
      </div>
    );
  }

  const isConfirmed = billingReview?.status === "confirmed";

  return (
    <div className="h-[calc(100vh-220px)] min-h-[500px] flex flex-col gap-3 overflow-y-auto pb-10 pr-1 relative">
      {isConfirmed && (
        <div className="flex items-center gap-2 p-3 bg-green-50 border border-green-100 rounded-lg text-green-700 text-sm mb-1">
          <CheckCircle2 className="size-4" />
          <span>このカルテの会計内容は医師によって確認済みです。</span>
        </div>
      )}

      {/* Items Table */}
      <TreatmentTable
        items={items}
        onUpdate={handleUpdateItem}
        onRemove={handleRemoveItem}
        onOpenSearch={() => setIsSearchOpen(true)}
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

      {/* Action Button */}
      <div className="fixed bottom-6 right-6 z-50 flex gap-2">
        {isConfirmed ? (
          <Button
            variant="outline"
            onClick={handleReturn}
            disabled={returnMutation.isPending}
            className="h-10 text-sm gap-2 border-[#E0E0E0] hover:bg-gray-50"
          >
            <RotateCcw className="size-4" />
            確認を取り消す
          </Button>
        ) : (
          <Button
            onClick={handleConfirm}
            disabled={confirmMutation.isPending || items.length === 0}
            size="sm"
            className="bg-[#2383E2] hover:bg-[#1B6EC2] text-white min-w-[120px] shadow-lg h-10 text-sm border-transparent gap-2"
          >
            <CheckCircle2 className="size-4" />
            チェック完了
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
