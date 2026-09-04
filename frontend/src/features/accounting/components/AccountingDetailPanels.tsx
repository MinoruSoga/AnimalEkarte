import { AlertTriangle, EyeOff, Printer } from "lucide-react";
import { Button } from "@/components/ui/button";
import { FormHeaderActions } from "@/components/shared/Form/FormHeaderActions";
import { Card, CardContent } from "@/components/ui/card";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { C, ICON, Z_CLASS } from "@/lib/design-tokens";
import { formatCurrency } from "@/lib/format/number";
import type { TaxType } from "@/types/generated/models";
import type { Accounting, AddAccountingItemInput, PaymentMethod } from "../types";
import { AccountingDocument, type ClinicInfo } from "./AccountingDocument";
import { InsuranceCard } from "./InsuranceCard";
import { ItemListCard } from "./ItemListCard";
import { PaymentCard, type PaymentSplitDraft } from "./PaymentCard";
import { OwnerUnpaidBalanceCard } from "./OwnerUnpaidBalanceCard";
import { RefundSection } from "./RefundSection";

interface AccountingHeaderActionsProps {
  status: Accounting["status"];
  canDelete: boolean;
  isCancelling: boolean;
  onPrint: () => void;
  onCancelClick: () => void;
  onDismiss?: () => void;
  submitLabel?: string;
  submitDisabled?: boolean;
}

export function AccountingHeaderActions({
  status,
  canDelete,
  isCancelling,
  onPrint,
  onCancelClick,
  onDismiss,
  submitLabel,
  submitDisabled,
}: AccountingHeaderActionsProps) {
  if (status !== "completed") {
    return onDismiss ? (
      <FormHeaderActions
        onCancel={onDismiss}
        submitLabel={submitLabel}
        submitDisabled={submitDisabled}
      />
    ) : undefined;
  }

  return (
    <div className="flex gap-2">
      <Button type="button" variant="outline" size="sm" onClick={onPrint}>
        <Printer className={`mr-2 ${ICON.action}`} />
        明細兼領収書
      </Button>
      {canDelete ? (
        <Button
          type="button"
          variant="outline"
          size="sm"
          onClick={onCancelClick}
          className={C.danger}
          disabled={isCancelling}
        >
          会計をキャンセル
        </Button>
      ) : null}
    </div>
  );
}

interface ReadOnlyAccountingBannerProps {
  show: boolean;
  message?: string;
}

export function ReadOnlyAccountingBanner({
  show,
  message = "閲覧専用 — 編集権限がないため変更できません",
}: ReadOnlyAccountingBannerProps) {
  if (!show) return null;

  return (
    <div
      className={`flex items-center gap-2 px-4 py-2.5 rounded-md border mb-4 ${C.bgWarning50} ${C.borderWarning20} ${C.textWarning}`}
      role="status"
      aria-label="閲覧専用モード"
    >
      <EyeOff className={`shrink-0 h-4 w-4 ${C.textWarningIcon}`} aria-hidden="true" />
      <span className="text-sm font-medium">{message}</span>
    </div>
  );
}

interface UngroupedItemsWarningBannerProps {
  show: boolean;
  medicalRecordCount: number;
  trimmingCount: number;
}

// #77: 同日同ペットにまだ会計対象化されていない項目があるとき、取り残し防止のため警告する。
export function UngroupedItemsWarningBanner({
  show,
  medicalRecordCount,
  trimmingCount,
}: UngroupedItemsWarningBannerProps) {
  if (!show || (medicalRecordCount === 0 && trimmingCount === 0)) return null;

  const parts: string[] = [];
  if (medicalRecordCount > 0) parts.push(`診察 ${medicalRecordCount}件`);
  if (trimmingCount > 0) parts.push(`トリミング ${trimmingCount}件`);

  return (
    <div
      className={`flex items-start gap-2 px-4 py-2.5 rounded-md border mb-4 ${C.bgWarning50} ${C.borderWarning20} ${C.textWarning}`}
      role="alert"
      aria-label="同日に会計対象化されていない項目があります"
    >
      <AlertTriangle
        className={`shrink-0 h-4 w-4 mt-0.5 ${C.textWarningIcon}`}
        aria-hidden="true"
      />
      <div className="text-sm">
        <span className="font-medium">
          同日にまだ会計対象化されていない項目があります（{parts.join(" / ")}）。
        </span>
        <span className={`block ${C.text60} mt-0.5`}>
          受付ボードで対象を会計待ちに進めてから会計すると、1会計にまとめられます。
        </span>
      </div>
    </div>
  );
}

interface UnbilledBlockingWarningBannerProps {
  show: boolean;
  warnings: ReadonlyArray<{ source: string; code: string; count: number; blocking: boolean }>;
}

/** BUG-013: blocking unbilled warning 中は会計確定を無効化する。 */
export function UnbilledBlockingWarningBanner({
  show,
  warnings,
}: UnbilledBlockingWarningBannerProps) {
  if (!show) return null;
  const blocking = warnings.filter((w) => w.blocking && w.count > 0);
  if (blocking.length === 0) return null;

  const labels = blocking.map((w) => {
    if (w.code === "vaccination_master_unbillable") {
      return `予防接種マスタ未設定/価格不正 ${w.count}件`;
    }
    return `${w.source} ${w.count}件`;
  });

  return (
    <div
      className={`flex items-start gap-2 px-4 py-2.5 rounded-md border mb-4 ${C.bgDanger10} ${C.borderDanger20} ${C.danger}`}
      role="alert"
      aria-label="未請求候補に請求不能な項目があるため会計を確定できません"
    >
      <AlertTriangle className={`shrink-0 h-4 w-4 mt-0.5 ${C.danger}`} aria-hidden="true" />
      <div className="text-sm">
        <span className="font-medium">
          未請求候補に請求不能な項目があるため会計を確定できません（{labels.join(" / ")}）。
        </span>
        <span className={`block ${C.text60} mt-0.5`}>
          ワクチンマスタの価格を整備してから会計してください。有効な処置・トリミングのみでの部分会計は許可されません。
        </span>
      </div>
    </div>
  );
}

interface AccountingCalculationView {
  subtotal: number;
  taxTotal: number;
  totalAmount: number;
  insuranceAmount: number;
  billingAmount: number;
}

interface AccountingDetailColumnsProps {
  accounting: Accounting;
  calculation: AccountingCalculationView;
  accountingId: string | undefined;
  hasInsurance: boolean;
  insuranceRatio: string;
  paymentSplits: PaymentSplitDraft[];
  newItemOpen: boolean;
  isRefunding: boolean;
  canEdit: boolean;
  canCreate: boolean;
  canDelete: boolean;
  onNewItemOpenChange: (open: boolean) => void;
  onAddItem: (input: AddAccountingItemInput) => void;
  onDeleteItem: (itemId: string) => void;
  onUpdateItemTax: (itemId: string, taxType: TaxType, taxRate: number) => void;
  onUpdateItemDiscount: (itemId: string, discountAmount: number) => void;
  onUseInsuranceChange: (enabled: boolean) => void;
  onInsuranceRatioChange: (ratio: string) => void;
  onSplitsChange: (splits: PaymentSplitDraft[]) => void;
  onRefund: (amount: number, reason: string, paymentMethod?: PaymentMethod) => void;
}

export function AccountingDetailColumns({
  accounting,
  calculation,
  accountingId,
  hasInsurance,
  insuranceRatio,
  paymentSplits,
  newItemOpen,
  isRefunding,
  canEdit,
  canCreate,
  canDelete,
  onNewItemOpenChange,
  onAddItem,
  onDeleteItem,
  onUpdateItemTax,
  onUpdateItemDiscount,
  onUseInsuranceChange,
  onInsuranceRatioChange,
  onSplitsChange,
  onRefund,
}: AccountingDetailColumnsProps) {
  return (
    <div className="flex flex-col lg:flex-row gap-6 h-[calc(100vh-140px)]">
      <div className="flex-1 flex flex-col gap-4 overflow-hidden">
        <ItemListCard
          items={accounting.items}
          subtotal={calculation.subtotal}
          taxTotal={calculation.taxTotal}
          totalAmount={calculation.totalAmount}
          newItemOpen={newItemOpen}
          onNewItemOpenChange={onNewItemOpenChange}
          onAddItem={onAddItem}
          onDeleteItem={onDeleteItem}
          accountingId={accountingId}
          onUpdateItemTax={onUpdateItemTax}
          onUpdateItemDiscount={onUpdateItemDiscount}
          canEdit={canEdit}
          canDelete={canDelete}
        />
      </div>

      <div className="w-full lg:w-[400px] flex flex-col gap-4 overflow-y-auto">
        {/* BUG-007: 当該会計のクレジット訂正差額など、この会計固有の未収 */}
        {accounting.outstandingAmount != null && accounting.outstandingAmount > 0 ? (
          <Card data-testid="billing-outstanding-amount">
            <CardContent className="p-4 flex items-center justify-between">
              <div className="flex flex-col gap-0.5">
                <span className={`text-sm font-medium ${C.text60}`}>この会計の未収残高</span>
                <span className={`text-xs ${C.text40}`}>
                  支払額と請求額の差額（クレジット訂正などを含む）
                </span>
              </div>
              <span className={`text-xl font-bold ${C.danger}`}>
                {formatCurrency(accounting.outstandingAmount)}
              </span>
            </CardContent>
          </Card>
        ) : null}

        {/* #182: 飼主の未納残高（未納がある場合のみ表示） */}
        {/* P2-11: 拠点横断で開いた会計の場合、残高は accounting.clinicId のクリニックで解決する */}
        <OwnerUnpaidBalanceCard ownerId={accounting.ownerId} clinicId={accounting.clinicId} />

        <InsuranceCard
          useInsurance={hasInsurance}
          onUseInsuranceChange={onUseInsuranceChange}
          insuranceRatio={insuranceRatio}
          onInsuranceRatioChange={onInsuranceRatioChange}
          insuranceAmount={calculation.insuranceAmount}
        />

        <PaymentCard
          billingAmount={calculation.billingAmount}
          paymentSplits={paymentSplits}
          onSplitsChange={onSplitsChange}
          isCompleted={accounting.status === "completed"}
          canEdit={canEdit}
          canCreate={canCreate}
          isEditMode={Boolean(accountingId)}
        />

        {accountingId && accounting.status === "completed" ? (
          <RefundSection
            accountingId={accountingId}
            totalAmount={accounting.payment?.totalAmount ?? 0}
            paymentSplits={accounting.paymentSplits ?? []}
            isRefunding={isRefunding}
            onRefund={onRefund}
            canEdit={canEdit}
          />
        ) : null}
      </div>
    </div>
  );
}

interface AccountingDocumentPreviewDialogProps {
  open: boolean;
  accounting: Accounting;
  clinic: ClinicInfo | null;
  onOpenChange: (open: boolean) => void;
}

export function AccountingDocumentPreviewDialog({
  open,
  accounting,
  clinic,
  onOpenChange,
}: AccountingDocumentPreviewDialogProps) {
  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-w-4xl h-[90vh] overflow-hidden flex flex-col">
        <DialogHeader>
          <DialogTitle>明細兼領収書プレビュー</DialogTitle>
          <DialogDescription>印刷イメージを確認できます。</DialogDescription>
        </DialogHeader>
        <div className={`flex-1 ${C.bgActive} overflow-auto p-8 flex items-center justify-center`}>
          <div className="shadow-level1 transform scale-100 origin-top">
            {accounting.payment ? (
              <AccountingDocument
                accounting={accounting}
                paymentInfo={accounting.payment}
                clinic={clinic}
              />
            ) : null}
          </div>
        </div>
        <DialogFooter className="gap-2">
          <Button type="button" variant="outline" onClick={() => onOpenChange(false)}>
            閉じる
          </Button>
          <Button type="button" onClick={() => window.print()}>
            <Printer className={`mr-2 ${ICON.action}`} />
            印刷する
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}

interface AccountingPrintAreaProps {
  accounting: Accounting;
  clinic: ClinicInfo | null;
}

export function AccountingPrintArea({ accounting, clinic }: AccountingPrintAreaProps) {
  if (!accounting.payment) return null;

  return (
    <div
      className={`hidden print:block fixed inset-0 ${C.bgWhite} ${Z_CLASS.overlay} p-0 m-0 w-full h-full`}
    >
      <style type="text/css" media="print">
        {`
          @page { size: auto; margin: 0; }
          body { margin: 0; -webkit-print-color-adjust: exact; }
        `}
      </style>
      <div className="p-8">
        <AccountingDocument
          accounting={accounting}
          paymentInfo={accounting.payment}
          clinic={clinic}
        />
      </div>
    </div>
  );
}
