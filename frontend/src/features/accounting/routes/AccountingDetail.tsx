// React/Framework
import { useState, useMemo, memo, useTransition, useActionState, useEffect, useRef } from "react";
import { useParams, useNavigate, useLocation } from "react-router";

// External
import { useQueryClient, useQuery } from "@tanstack/react-query";
import { toast } from "sonner";

// Internal
import { PageLayout } from "@/components/shared/PageLayout/PageLayout";

// Shared Hooks
import { useAuth } from "@/hooks/use-auth";
import { usePermission } from "@/hooks/use-permission";

// Relative
import { LoadingFallback, ErrorFallback } from "@/components/shared/DataStates";
import { useGetAccountingDetail } from "../api/get-accounting";
import { createAccounting } from "../api/create-accounting";
import { updateAccounting } from "../api/update-accounting";
import { createBillingItem } from "../api/create-billing-item";
import { getUnbilledItems } from "../api/get-unbilled-items";
import { useGetPet } from "@/hooks/use-pet";
import { handleApiError } from "@/lib/handle-api-error";
import { calculateBillingTotals } from "@/lib/calculations";
import { ConfirmDialog } from "@/components/shared/ConfirmDialog/ConfirmDialog";
import { paths } from "@/config/paths";
import {
  AccountingDetailColumns,
  AccountingDocumentPreviewDialog,
  AccountingHeaderActions,
  AccountingPrintArea,
  ReadOnlyAccountingBanner,
} from "../components/AccountingDetailPanels";
import type { PaymentSplitDraft } from "../components/PaymentCard";
import { createInitialPaymentSplits, type AccountingFormState } from "./AccountingDetailModel";
import { useAccountingItemActions } from "./useAccountingItemActions";
import { useAccountingSettlementActions } from "./useAccountingSettlementActions";

// Types
import type { Accounting, AccountingItem, PaymentInfo, PaymentMethod } from "../types";
import { ResourceAccounting } from "@/types/generated/models";

// ── メインコンポーネント ──────────────────────────────────

interface AccountingDetailProps {
  invoiceRegistrationNumber?: string;
}

export const AccountingDetail = memo(function AccountingDetail({ invoiceRegistrationNumber }: AccountingDetailProps) {
  const { id } = useParams();
  const navigate = useNavigate();
  const location = useLocation();
  const queryClient = useQueryClient();
  const [, startTaxUpdateTransition] = useTransition();
  const [, startAddItemTransition] = useTransition();

  const locationState = location.state as { accountingItems?: AccountingItem[] } | null;

  // 既存データは API から取得
  const { data: fetchedAccounting, isLoading } = useGetAccountingDetail(id);

  // 新規作成時のペット情報取得（URL の petId から owner 情報を解決するため）
  const newPetId = useMemo(() => {
    if (id) return "";
    return new URLSearchParams(location.search).get("petId") ?? "";
  }, [id, location.search]);
  const { data: newPetData } = useGetPet(newPetId);

  // 新規作成の場合は location.state から items を引き継ぐ（派生データ）
  const baseAccounting = useMemo<Accounting | null>(() => {
    if (id) {
      return fetchedAccounting ?? null;
    }
    const stateItems = locationState?.accountingItems ?? [];
    return {
      id: "acc_new",
      ownerId: newPetData?.ownerId ?? "",
      ownerName: newPetData?.ownerName ?? "飼い主様",
      petId: newPetId,
      petName: newPetData?.name ?? "ペット",
      petSpecies: newPetData?.species ?? "犬",
      status: "waiting",
      scheduledDate: new Date().toISOString().split("T")[0],
      items: stateItems,
      payment: undefined,
      totalRefundedAmount: 0,
    };
  }, [id, fetchedAccounting, locationState, newPetId, newPetData]);

  // baseAccounting.items の安定参照 — useCallback deps に配列オブジェクトを渡さないための useMemo
  const baseItems = useMemo(() => baseAccounting?.items ?? [], [baseAccounting]);

  // 新規作成時: 未請求の治療明細をペットIDで取得してローカル明細を初期化
  const { data: unbilledItems } = useQuery({
    queryKey: ["unbilledItems", newPetId],
    queryFn: () => getUnbilledItems(newPetId),
    enabled: !id && !!newPetId,
    staleTime: 30_000,
  });

  // ユーザー操作による追加・削除を管理するローカル明細
  const [localItems, setLocalItems] = useState<AccountingItem[] | null>(null);

  // 未請求明細が取得できたらローカル明細を初期化（ユーザー未編集の場合のみ）
  useEffect(() => {
    if (unbilledItems && unbilledItems.length > 0 && localItems === null) {
      setLocalItems(unbilledItems);
    }
  }, [unbilledItems, localItems]);

  // 表示する明細: ローカル編集があればそちら優先
  const displayItems = useMemo(
    () => localItems ?? baseItems,
    [localItems, baseItems],
  );

  // 決済完了フラグ（API 更新後に画面に反映するためのローカル状態）
  const [completedPayment, setCompletedPayment] = useState<PaymentInfo | null>(null);

  // 実際に表示する accounting (payment は completedPayment 優先)
  const accounting: Accounting | null = useMemo(() => {
    if (!baseAccounting) return null;
    return {
      ...baseAccounting,
      items: displayItems,
      payment: completedPayment ?? baseAccounting.payment,
      status: completedPayment ? "completed" : baseAccounting.status,
    };
  }, [baseAccounting, displayItems, completedPayment]);

  // 保険・決済の状態（API データ取得後に同期）
  const [hasInsurance, setHasInsurance] = useState(() => (baseAccounting?.payment?.insuranceAmount ?? 0) < 0);
  const [insuranceRatio, setInsuranceRatio] = useState(() => baseAccounting?.payment?.insuranceRatio?.toString() ?? "0.5");

  const [paymentSplits, setPaymentSplits] = useState<PaymentSplitDraft[]>(() => createInitialPaymentSplits(baseAccounting));

  // 金額計算
  const calculation = useMemo(() => {
    if (!accounting) return null;

    const billingResult = calculateBillingTotals(
      accounting.items,
      0, // ownerDiscountRate (accounting detail supports specific line discounts instead)
      0, // globalDiscountAmount
      0.10, // taxRate
      hasInsurance ? parseFloat(insuranceRatio) : 0
    );

    return {
      subtotal: billingResult.subtotal,
      taxTotal: billingResult.tax,
      totalAmount: billingResult.total,
      insuranceAmount: billingResult.insuranceAmount,
      billingAmount: billingResult.billingAmount,
    };
  }, [accounting, hasInsurance, insuranceRatio]);

  // 追加アイテム用State
  const [newItemOpen, setNewItemOpen] = useState(false);
  const [isRefunding, startRefundTransition] = useTransition();

  // BUG-367: 明細兼領収書プレビュー State（旧 previewType 廃止）
  const [previewOpen, setPreviewOpen] = useState(false);

  // BUG-371: 精算済会計の修正確認 / キャンセル確認 モーダル状態
  const [editConfirmOpen, setEditConfirmOpen] = useState(false);
  const [cancelConfirmOpen, setCancelConfirmOpen] = useState(false);
  const [isCancelling, startCancelTransition] = useTransition();
  // editConfirmedRef: 確認モーダル OK 後の formAction 再実行用フラグ
  const editConfirmedRef = useRef(false);
  const formRef = useRef<HTMLFormElement>(null);

  /**
   * React 19 useActionState を使用した会計確定アクション
   */
  const [formState, formAction, _isPending] = useActionState(
    async (_prevState: AccountingFormState, _formData: FormData): Promise<AccountingFormState> => {
      if (!accounting || !calculation) return { success: false, timestamp: Date.now() };

      // BUG-371: 精算済会計の修正時は確認モーダル経由を強制
      // editConfirmed フラグが未設定の場合はモーダルを出して処理を中断
      if (accounting.status === "completed" && !editConfirmedRef.current) {
        setEditConfirmOpen(true);
        return { success: false, timestamp: Date.now() };
      }
      editConfirmedRef.current = false;

      // 決済 splits をリクエスト形式に変換
      const builtSplits = paymentSplits
        .filter((s) => s.amount && parseInt(s.amount, 10) > 0)
        .map((s) => {
          const amount = parseInt(s.amount, 10);
          const received = s.method === "cash" ? parseInt(s.receivedAmount || "0", 10) : 0;
          return {
            method: s.method as PaymentMethod,
            amount,
            received_amount: received,
            change_amount: s.method === "cash" ? Math.max(0, received - amount) : 0,
          };
        });

      // 代表支払方法 (cash > credit_card > electronic_money)
      const repMethod: PaymentMethod =
        builtSplits.some((s) => s.method === "cash") ? "cash" :
        builtSplits.some((s) => s.method === "credit_card") ? "credit_card" :
        "electronic_money";

      const cashSplit = builtSplits.find((s) => s.method === "cash");
      const totalReceived = cashSplit ? cashSplit.received_amount : calculation.billingAmount;
      const totalChange = cashSplit ? cashSplit.change_amount : 0;

      const paymentInfo: PaymentInfo = {
        subtotal: calculation.subtotal,
        taxTotal: calculation.taxTotal,
        totalAmount: calculation.totalAmount,
        insuranceAmount: calculation.insuranceAmount,
        discountAmount: 0,
        billingAmount: calculation.billingAmount,
        receivedAmount: totalReceived,
        changeAmount: totalChange,
        method: repMethod,
        insuranceRatio: hasInsurance ? parseFloat(insuranceRatio) : undefined,
      };

      try {
        if (!id) {
          // 新規会計: まず作成してから完了状態に更新
          const created = await createAccounting({
            pet_id: Number(accounting.petId),
            owner_id: Number(accounting.ownerId),
            scheduled_date: accounting.scheduledDate
              ? `${accounting.scheduledDate}T00:00:00Z`
              : new Date().toISOString(),
            subtotal: calculation.subtotal,
            tax_total: calculation.taxTotal,
            total_amount: calculation.totalAmount,
          });
          // BUG-ACCOUNTING-NEW-ITEMS-LOST: 新規作成後に明細を保存する
          for (const item of displayItems) {
            await createBillingItem({
              billing_id: Number(created.id),
              category: item.category,
              name: item.name,
              unit_price: item.unitPrice,
              quantity: item.quantity,
              tax_type: item.taxType,
              tax_rate: item.taxRate,
              is_insurance_applicable: item.isInsuranceApplicable,
              source: item.source,
            });
          }
          await updateAccounting(created.id, {
            status: "completed",
            subtotal: calculation.subtotal,
            tax_total: calculation.taxTotal,
            total_amount: calculation.totalAmount,
            insurance_ratio: hasInsurance ? parseFloat(insuranceRatio) : null,
            insurance_amount:
              calculation.insuranceAmount !== 0 ? calculation.insuranceAmount : null,
            billing_amount: calculation.billingAmount,
            received_amount: totalReceived,
            change_amount: totalChange,
            payment_method: repMethod,
            payment_splits: builtSplits,
            completed_at: new Date().toISOString(),
          });
          queryClient.invalidateQueries({ queryKey: ["accountings"] });
          toast.success("会計を登録・完了しました");
          navigate(paths.accounting.detail.getHref(created.id));
        } else {
          // 既存会計: 直接更新
          await updateAccounting(id, {
            status: "completed",
            subtotal: calculation.subtotal,
            tax_total: calculation.taxTotal,
            total_amount: calculation.totalAmount,
            insurance_ratio: hasInsurance ? parseFloat(insuranceRatio) : null,
            insurance_amount:
              calculation.insuranceAmount !== 0 ? calculation.insuranceAmount : null,
            billing_amount: calculation.billingAmount,
            received_amount: totalReceived,
            change_amount: totalChange,
            payment_method: repMethod,
            payment_splits: builtSplits,
            completed_at: new Date().toISOString(),
          });
          setCompletedPayment(paymentInfo);
          queryClient.invalidateQueries({ queryKey: ["accountings"] });
          queryClient.invalidateQueries({ queryKey: ["accounting", id] });
          queryClient.invalidateQueries({ queryKey: ["accounting-detail", id] });
          toast.success("会計を完了しました");
        }
        return { success: true, timestamp: Date.now() };
      } catch (error) {
        handleApiError(error, "会計の処理");
        return { success: false, timestamp: Date.now() };
      }
    },
    { success: false, timestamp: 0 }
  );

  // --- Focus Management (Accessibility) ---
  useEffect(() => {
    // If validation fails or amount is insufficient
    if (formState.success === false) {
      const element = document.getElementById("receivedAmount");
      if (element) {
        element.focus();
        element.scrollIntoView({ behavior: "smooth", block: "center" });
      }
    }
  }, [formState.success, formState.timestamp]);

  // clinic 情報（AccountingDocument に props 注入）
  const { user } = useAuth();
  const { canEdit, canCreate, canDelete } = usePermission("accounting");
  const canSubmit = id ? canEdit : canCreate;
  const clinicForDocument = useMemo(() => {
    const baseClinic = user?.clinic ?? null;
    if (!baseClinic) return null;
    return {
      ...baseClinic,
      invoiceRegistrationNumber,
    };
  // rerender-dependencies: user?.clinic（オブジェクト）の代わりに user（安定参照）を deps に使用
  }, [user, invoiceRegistrationNumber]);

  const { handleAddItem, handleDeleteItem, handleUpdateItemTax } = useAccountingItemActions({
    accountingId: id,
    baseItems,
    queryClient,
    setLocalItems,
    setNewItemOpen,
    startAddItemTransition,
    startTaxUpdateTransition,
  });

  const { handleCancelConfirm, handlePrint, handleRefund } = useAccountingSettlementActions({
    accountingId: id,
    navigate,
    queryClient,
    setCancelConfirmOpen,
    setPreviewOpen,
    startCancelTransition,
    startRefundTransition,
  });

  if (id && isLoading) return <LoadingFallback />;
  if (!accounting || !calculation) return <ErrorFallback message="データが見つかりません" />;

  return (
    <>
    <form ref={formRef} action={formAction}>
      <PageLayout
        className="print:hidden"
        title="会計精算"
        resource={ResourceAccounting}
        description={`受付No: ${accounting.id} | ${accounting.ownerName}様 - ${accounting.petName}ちゃん`}
        onBack={() => navigate(paths.accounting.getHref())}
        headerAction={
          <AccountingHeaderActions
            status={accounting.status}
            canDelete={canDelete}
            isCancelling={isCancelling}
            onPrint={handlePrint}
            onCancelClick={() => setCancelConfirmOpen(true)}
          />
        }
      >
        <ReadOnlyAccountingBanner show={Boolean(id && !canEdit)} />
        <fieldset disabled={!canSubmit} className="border-0 p-0 m-0 min-w-0">
        <AccountingDetailColumns
          accounting={accounting}
          calculation={calculation}
          accountingId={id}
          hasInsurance={hasInsurance}
          insuranceRatio={insuranceRatio}
          paymentSplits={paymentSplits}
          newItemOpen={newItemOpen}
          isRefunding={isRefunding}
          canEdit={canEdit}
          canCreate={canCreate}
          canDelete={canDelete}
          onNewItemOpenChange={setNewItemOpen}
          onAddItem={handleAddItem}
          onDeleteItem={handleDeleteItem}
          onUpdateItemTax={handleUpdateItemTax}
          onUseInsuranceChange={setHasInsurance}
          onInsuranceRatioChange={setInsuranceRatio}
          onSplitsChange={setPaymentSplits}
          onRefund={handleRefund}
        />
        </fieldset>

        <AccountingDocumentPreviewDialog
          open={previewOpen}
          accounting={accounting}
          clinic={clinicForDocument}
          onOpenChange={setPreviewOpen}
        />

        {/* BUG-371: 精算済修正の確認モーダル */}
        <ConfirmDialog
          open={editConfirmOpen}
          onClose={() => setEditConfirmOpen(false)}
          title="精算済みの会計を修正します"
          description="この操作は会計データに変更を加えます。よろしいですか?"
          confirmLabel="修正する"
          cancelLabel="キャンセル"
          onConfirm={() => {
            editConfirmedRef.current = true;
            setEditConfirmOpen(false);
            requestAnimationFrame(() => {
              formRef.current?.requestSubmit();
            });
          }}
        />

        {/* BUG-371: 会計キャンセル確認モーダル */}
        <ConfirmDialog
          open={cancelConfirmOpen}
          onClose={() => setCancelConfirmOpen(false)}
          title="この会計をキャンセルします"
          description="元に戻せません。キャンセルされた会計はステータスが「cancelled」になります。"
          confirmLabel="キャンセルする"
          cancelLabel="戻る"
          variant="destructive"
          isPending={isCancelling}
          onConfirm={handleCancelConfirm}
        />
      </PageLayout>
    </form>

      <AccountingPrintArea accounting={accounting} clinic={clinicForDocument} />
    </>
  );
});
