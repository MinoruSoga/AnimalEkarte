import { useState, useMemo, memo, useTransition, lazy, Suspense } from "react";
import { useParams, useNavigate, useLocation } from "react-router";
import { useQueryClient } from "@tanstack/react-query";

import { PageLayout } from "@/components/shared/PageLayout/PageLayout";
import { ConfirmDialog } from "@/components/shared/ConfirmDialog/ConfirmDialog";
import { LoadingFallback, ErrorFallback } from "@/components/shared/DataStates";
import { useAuth } from "@/hooks/use-auth";
import { useGetCashRegisterCloses } from "@/hooks/use-cash-register-closes";
import { usePermission } from "@/hooks/use-permission";
import { paths } from "@/config/paths";
import { C } from "@/lib/design-tokens";
import { todayJSTISO } from "@/lib/jst-date";
import {
  ResourceAccounting,
  ResourceAccountingCancel,
  ResourceAccountingPostCloseEdit,
  ResourceCashRegisterClose,
} from "@/types/generated/models";

import { useGetAccountingDetail } from "../api/get-accounting";
import {
  AccountingDetailColumns,
  AccountingDocumentPreviewDialog,
  AccountingHeaderActions,
  AccountingPrintArea,
  ReadOnlyAccountingBanner,
  UngroupedItemsWarningBanner,
  UnbilledBlockingWarningBanner,
} from "../components/AccountingDetailPanels";
import { isPaymentSubmitDisabled, isPostCloseSubmitBlocked } from "../components/accounting-detail-model";
import { useAccountingCompletionAction } from "../hooks/use-accounting-completion-action";
import { useAccountingDetailState } from "../hooks/use-accounting-detail-state";
import { useAccountingItemActions } from "../hooks/use-accounting-item-actions";
import { useAccountingSettlementActions } from "../hooks/use-accounting-settlement-actions";
import type { AccountingItem } from "../types";

const CreditCorrectionDialog = lazy(() =>
  import("../components/CreditCorrectionDialog").then((m) => ({ default: m.CreditCorrectionDialog })),
);

interface AccountingDetailProps {
  invoiceRegistrationNumber?: string;
}

export const AccountingDetail = memo(function AccountingDetail({ invoiceRegistrationNumber }: AccountingDetailProps) {
  const { id } = useParams();
  const navigate = useNavigate();
  const location = useLocation();
  const queryClient = useQueryClient();
  const [, startItemUpdateTransition] = useTransition();
  const [, startAddItemTransition] = useTransition();
  const [, startDeleteItemTransition] = useTransition();

  const locationState = location.state as { accountingItems?: AccountingItem[] } | null;

  // 既存データは API から取得
  const { data: fetchedAccounting, isLoading } = useGetAccountingDetail(id);

  const {
    baseItems,
    displayItems,
    setLocalItems,
    accounting,
    ungroupedSummary,
    unbilledWarnings,
    hasBlockingUnbilledWarning,
    blocksNewAccountingSubmit,
    deceasedPetBlockMessage,
    unbilledDetailsError,
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

  // #115: 締め後編集理由
  const [postCloseReason, setPostCloseReason] = useState("");

  // 追加アイテム用State
  const [newItemOpen, setNewItemOpen] = useState(false);
  const [isRefunding, startRefundTransition] = useTransition();

  // BUG-367: 明細兼領収書プレビュー State（旧 previewType 廃止）
  const [previewOpen, setPreviewOpen] = useState(false);

  // BUG-371: 精算済会計の修正確認 / キャンセル確認 モーダル状態
  const [cancelConfirmOpen, setCancelConfirmOpen] = useState(false);
  const [isCancelling, startCancelTransition] = useTransition();

  // clinic 情報（AccountingDocument に props 注入）
  const { user } = useAuth();
  const { canEdit, canCreate, canDelete } = usePermission("accounting");

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
    postCloseReason,
    blockCreateReason: deceasedPetBlockMessage,
    permissions: { canCreate, canEdit },
  });
  // #118: キャンセルボタン用の専用権限（accounting-cancel の edit = キャンセル可否）
  const { canEdit: canCancelAccounting } = usePermission(ResourceAccountingCancel);
  // #115: 締め後編集専用権限
  const { canEdit: canPostCloseEdit } = usePermission(ResourceAccountingPostCloseEdit);
  const { canView: canViewCashRegisterClose } = usePermission(ResourceCashRegisterClose);
  const hasAccountingMutationPermission = id ? canEdit : canCreate;
  // レジ締め状態を検証できない場合は、締め後編集を誤って許可しないよう fail closed にする。
  // BUG-013: blocking unbilled / details 未取得・失敗中は新規会計確定を無効化。
  const canSubmit =
    hasAccountingMutationPermission && canViewCashRegisterClose && !blocksNewAccountingSubmit;

  // #115: billing の scheduled_date が締め済みか確認
  const scheduledDateStr = accounting?.scheduledDate
    ? accounting.scheduledDate.slice(0, 10)
    : todayJSTISO();
  // #115: scheduled_date が締め済みか、その 1 日分（AM/PM/EMG）を BE 契約（start_date/end_date）で問い合わせる
  const { data: closesData } = useGetCashRegisterCloses(
    scheduledDateStr ? { start_date: scheduledDateStr, end_date: scheduledDateStr } : undefined,
    Boolean(scheduledDateStr && hasAccountingMutationPermission && canViewCashRegisterClose),
  );
  const isScheduledDateClosed = Boolean(
    scheduledDateStr && closesData?.data.some((c) => c.closeDate?.slice(0, 10) === scheduledDateStr),
  );
  const clinicForDocument = useMemo(() => {
    const baseClinic = user?.clinic ?? null;
    if (!baseClinic) return null;
    return {
      ...baseClinic,
      invoiceRegistrationNumber,
    };
  // rerender-dependencies: user?.clinic（オブジェクト）の代わりに user（安定参照）を deps に使用
  }, [user, invoiceRegistrationNumber]);

  const { handleAddItem, handleDeleteItem, handleUpdateItemTax, handleUpdateItemDiscount } = useAccountingItemActions({
    accountingId: id,
    accountingStatus: accounting?.status,
    postCloseReason,
    canPostCloseEdit,
    isScheduledDateClosed,
    baseItems,
    queryClient,
    setLocalItems,
    setNewItemOpen,
    startAddItemTransition,
    startDeleteItemTransition,
    startItemUpdateTransition,
    permissions: { canCreate, canEdit, canDelete },
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

  const readOnlyMessage = deceasedPetBlockMessage
    ? deceasedPetBlockMessage
    : !hasAccountingMutationPermission
      ? "閲覧専用 — 編集権限がないため変更できません"
      : !canViewCashRegisterClose
        ? "閲覧専用 — レジ締め状態を確認する権限がないため変更できません"
        : undefined;

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
              canDelete={canCancelAccounting}
              isCancelling={isCancelling}
              onPrint={handlePrint}
              onCancelClick={() => setCancelConfirmOpen(true)}
              onDismiss={() => navigate(paths.accounting.getHref())}
              submitLabel={canSubmit ? (accounting.status === "completed" ? "修正を保存する" : "会計を確定する") : undefined}
              submitDisabled={
                isPaymentSubmitDisabled(calculation.billingAmount, paymentSplits)
                || isPostCloseSubmitBlocked({
                  isScheduledDateClosed,
                  canPostCloseEdit: Boolean(canPostCloseEdit),
                  postCloseReason,
                })
              }
            />
          }
        >
          <ReadOnlyAccountingBanner
            show={readOnlyMessage !== undefined}
            message={readOnlyMessage}
          />
          <UngroupedItemsWarningBanner
            show={Boolean(!id && !deceasedPetBlockMessage && ungroupedSummary?.hasUngrouped)}
            medicalRecordCount={ungroupedSummary?.medicalRecordCount ?? 0}
            trimmingCount={ungroupedSummary?.trimmingCount ?? 0}
          />
          <UnbilledBlockingWarningBanner
            show={Boolean(
              !id && !deceasedPetBlockMessage && (hasBlockingUnbilledWarning || unbilledDetailsError),
            )}
            warnings={
              unbilledDetailsError
                ? [{ source: "unbilled", code: "unbilled_details_unavailable", count: 1, blocking: true }]
                : unbilledWarnings
            }
          />
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
              canEdit={Boolean(canEdit && canViewCashRegisterClose && !blocksNewAccountingSubmit)}
              canCreate={Boolean(canCreate && canViewCashRegisterClose && !blocksNewAccountingSubmit)}
              canDelete={Boolean(canDelete && canViewCashRegisterClose)}
              onNewItemOpenChange={setNewItemOpen}
              onAddItem={handleAddItem}
              onDeleteItem={handleDeleteItem}
              onUpdateItemTax={handleUpdateItemTax}
              onUpdateItemDiscount={handleUpdateItemDiscount}
              onUseInsuranceChange={setHasInsurance}
              onInsuranceRatioChange={setInsuranceRatio}
              onSplitsChange={setPaymentSplits}
              onRefund={handleRefund}
            />
          </fieldset>

          {/* #189: 確定済みカード金額の確定後訂正（専用フロー・監査付き）。確定済み + 訂正権限がある場合のみ導線を表示。
              CreditCorrectionDialog は確定済み(completed)かつカード支払いが無ければ自動で null を返す。 */}
          {id && canPostCloseEdit && canSubmit ? (
            <div className="px-4 pb-4">
              <Suspense fallback={null}>
                <CreditCorrectionDialog accounting={accounting} isPostClose={isScheduledDateClosed} />
              </Suspense>
            </div>
          ) : null}

          {isScheduledDateClosed && canSubmit && !canPostCloseEdit ? (
            <div className={`px-4 pb-4 text-sm font-semibold ${C.danger}`} role="status">
              レジ締め済み期間の会計確定には締め後編集権限（accounting-post-close-edit）が必要です。
            </div>
          ) : null}
          {/* #115 / BUG-009 / BUG-004: 締め後または確定済み会計の明細修正理由（権限あり時） */}
          {(isScheduledDateClosed || accounting.status === "completed") && canSubmit && canPostCloseEdit ? (
            <div className="px-4 pb-4">
              <label htmlFor="postCloseReason" className={`block text-sm font-semibold ${C.danger} mb-1`}>
                {isScheduledDateClosed
                  ? "⚠ レジ締め済み期間の編集 — 修正理由（必須）"
                  : "⚠ 確定済み会計の明細修正 — 修正理由（必須）"}
              </label>
              <textarea
                id="postCloseReason"
                value={postCloseReason}
                onChange={(e) => setPostCloseReason(e.target.value)}
                className="w-full border rounded p-2 text-sm resize-none"
                rows={2}
                placeholder="例: 入力金額の誤りのため修正"
              />
            </div>
          ) : null}
          {/* 確定済みだが締め後編集権限が無い場合は拒否理由を明示（BUG-009） */}
          {accounting.status === "completed" && canSubmit && !canPostCloseEdit ? (
            <div className={`px-4 pb-4 text-sm font-semibold ${C.danger}`} role="status">
              確定済み会計の明細修正には締め後編集権限（accounting-post-close-edit）が必要です。カード金額の訂正以外はクレジット訂正導線または権限付与を確認してください。
            </div>
          ) : null}

          <AccountingDocumentPreviewDialog
            open={previewOpen}
            accounting={accounting}
            clinic={clinicForDocument}
            onOpenChange={setPreviewOpen}
          />

          <ConfirmDialog
            open={editConfirmOpen}
            onClose={() => setEditConfirmOpen(false)}
            title="精算済みの会計を修正します"
            description="この操作は会計データに変更を加えます。よろしいですか?"
            confirmLabel="修正する"
            cancelLabel="キャンセル"
            onConfirm={confirmCompletedEdit}
          />

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
