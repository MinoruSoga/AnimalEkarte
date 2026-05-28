// React/Framework
import { useState, useMemo, memo, useTransition } from "react";
import { useParams, useNavigate, useLocation } from "react-router";

// External
import { useQueryClient } from "@tanstack/react-query";

// Internal
import { PageLayout } from "@/components/shared/PageLayout/PageLayout";

// Shared Hooks
import { useAuth } from "@/hooks/use-auth";
import { usePermission } from "@/hooks/use-permission";

// Relative
import { LoadingFallback, ErrorFallback } from "@/components/shared/DataStates";
import { useGetAccountingDetail } from "../api/get-accounting";
import { ConfirmDialog } from "@/components/shared/ConfirmDialog/ConfirmDialog";
import { paths } from "@/config/paths";
import {
  AccountingDetailColumns,
  AccountingDocumentPreviewDialog,
  AccountingHeaderActions,
  AccountingPrintArea,
  ReadOnlyAccountingBanner,
} from "../components/AccountingDetailPanels";
import { useAccountingCompletionAction } from "./useAccountingCompletionAction";
import { useAccountingDetailState } from "./useAccountingDetailState";
import { useAccountingItemActions } from "./useAccountingItemActions";
import { useAccountingSettlementActions } from "./useAccountingSettlementActions";

// Types
import type { AccountingItem } from "../types";
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

  const {
    baseItems,
    displayItems,
    setLocalItems,
    accounting,
    calculation,
    setCompletedPayment,
    hasInsurance,
    setHasInsurance,
    insuranceRatio,
    setInsuranceRatio,
    paymentSplits,
    setPaymentSplits,
  } = useAccountingDetailState({
    accountingId: id,
    locationSearch: location.search,
    locationState,
    fetchedAccounting,
  });

  // 追加アイテム用State
  const [newItemOpen, setNewItemOpen] = useState(false);
  const [isRefunding, startRefundTransition] = useTransition();

  // BUG-367: 明細兼領収書プレビュー State（旧 previewType 廃止）
  const [previewOpen, setPreviewOpen] = useState(false);

  // BUG-371: 精算済会計の修正確認 / キャンセル確認 モーダル状態
  const [cancelConfirmOpen, setCancelConfirmOpen] = useState(false);
  const [isCancelling, startCancelTransition] = useTransition();

  const {
    editConfirmOpen,
    setEditConfirmOpen,
    confirmCompletedEdit,
    formRef,
    formAction,
  } = useAccountingCompletionAction({
    accountingId: id,
    accounting,
    calculation,
    displayItems,
    hasInsurance,
    insuranceRatio,
    paymentSplits,
    queryClient,
    navigate,
    setCompletedPayment,
  });

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
          onConfirm={confirmCompletedEdit}
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
