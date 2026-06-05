import { AlertTriangle, EyeOff, Printer } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle } from "@/components/ui/dialog";
import { C, ICON } from "@/lib/design-tokens";
import type { PaymentMethod, TaxType } from "@/types/generated/models";
import type { Accounting } from "../types";
import { AccountingDocument } from "./AccountingDocument";
import { InsuranceCard } from "./InsuranceCard";
import { ItemListCard } from "./ItemListCard";
import { PaymentCard, type PaymentSplitDraft } from "./PaymentCard";
import { RefundSection } from "./RefundSection";

interface ClinicInfo {
  name?: string;
  postalCode?: string;
  address?: string;
  phoneNumber?: string;
  registrationNumber?: string;
  invoiceRegistrationNumber?: string;
  standardTaxRate?: number;
  reducedTaxRate?: number;
}

interface AccountingHeaderActionsProps {
  status: Accounting["status"];
  canDelete: boolean;
  isCancelling: boolean;
  onPrint: () => void;
  onCancelClick: () => void;
}

export function AccountingHeaderActions({
  status,
  canDelete,
  isCancelling,
  onPrint,
  onCancelClick,
}: AccountingHeaderActionsProps) {
  if (status !== "completed") return undefined;

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
}

export function ReadOnlyAccountingBanner({ show }: ReadOnlyAccountingBannerProps) {
  if (!show) return null;

  return (
    <div
      className={`flex items-center gap-2 px-4 py-2.5 rounded-md border mb-4 ${C.bgWarning50} ${C.borderWarning20} ${C.textWarning}`}
      role="status"
      aria-label="閲覧専用モード"
    >
      <EyeOff className={`shrink-0 h-4 w-4 ${C.textWarningIcon}`} aria-hidden="true" />
      <span className="text-sm font-medium">閲覧専用 — 編集権限がないため変更できません</span>
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
      <AlertTriangle className={`shrink-0 h-4 w-4 mt-0.5 ${C.textWarningIcon}`} aria-hidden="true" />
      <div className="text-sm">
        <span className="font-medium">同日にまだ会計対象化されていない項目があります（{parts.join(" / ")}）。</span>
        <span className={`block ${C.text60} mt-0.5`}>
          受付ボードで対象を会計待ちに進めてから会計すると、1会計にまとめられます。
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
  onAddItem: (name: string, price: string, category: string, taxRate?: number) => void;
  onDeleteItem: (itemId: string) => void;
  onUpdateItemTax: (itemId: string, taxType: TaxType, taxRate: number) => void;
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
          canEdit={canEdit}
          canDelete={canDelete}
        />
      </div>

      <div className="w-full lg:w-[400px] flex flex-col gap-4 overflow-y-auto">
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
          <div className="shadow-lg transform scale-100 origin-top">
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

export function AccountingPrintArea({
  accounting,
  clinic,
}: AccountingPrintAreaProps) {
  if (!accounting.payment) return null;

  return (
    <div className={`hidden print:block fixed inset-0 ${C.bgWhite} z-[9999] p-0 m-0 w-full h-full`}>
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
