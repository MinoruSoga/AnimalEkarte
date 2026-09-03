import { useActionState, useState, useCallback } from "react";
import { useNavigate } from "react-router";
import { Calculator } from "lucide-react";
import { toast } from "sonner";
import { C, ICON, LAYOUT } from "@/lib/design-tokens";
import { FormHeaderActions } from "@/components/shared/Form/FormHeaderActions";
import { PageLayout } from "@/components/shared/PageLayout/PageLayout";
import { paths } from "@/config/paths";
import { getFormString, getFormEnum } from "@/lib/form-data";
import { useCurrentClinicName } from "@/hooks/use-current-clinic-name";
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from "@/components/ui/alert-dialog";
import { useGetCashRegisterPreview } from "../api/get-cash-register-preview";
import { useCreateCashRegisterClose } from "../api/create-cash-register-close";
import { PERIOD_LABELS, PERIOD_OPTIONS, type CashRegisterPeriod } from "../constants";
import {
  CashRegisterClosePreview,
  CashRegisterCloseTargetSection,
} from "../components/CashRegisterClosePanels";
import { useCashRegisterCloseForm } from "../hooks/use-cash-register-close-form";
import { ResourceCashRegisterClose } from "@/types/generated/models";

// FE-RC-036: FormData の未検証キャストを避けるための enum ガード。
function isCashRegisterPeriod(value: string): value is CashRegisterPeriod {
  return (PERIOD_OPTIONS as readonly string[]).includes(value);
}

export function CashRegisterClosePage() {
  const navigate = useNavigate();
  const { date, period, previewEnabled, previewNonce, handleDateChange, handlePeriodChange, enablePreview } =
    useCashRegisterCloseForm();
  const clinicName = useCurrentClinicName();
  const [actualCash, setActualCash] = useState<string>("");
  const [showConfirm, setShowConfirm] = useState(false);
  const [pendingFormData, setPendingFormData] = useState<FormData | null>(null);

  const { data: preview, isLoading: previewLoading } = useGetCashRegisterPreview(
    date,
    period,
    previewEnabled,
    previewNonce,
  );
  const createMutation = useCreateCashRegisterClose();
  const { mutateAsync, isPending } = createMutation;

  const handleActualCashChange = useCallback((e: React.ChangeEvent<HTMLInputElement>) => {
    setActualCash(e.target.value);
  }, []);

  const [, formAction] = useActionState(async (_prev: null, formData: FormData) => {
    const cash = Number(formData.get("actual_cash"));
    if (!Number.isFinite(cash) || cash < 0) {
      toast.error("実際のレジ現金は0以上の金額を入力してください");
      return null;
    }
    setPendingFormData(formData);
    setShowConfirm(true);
    return null;
  }, null);

  const handleConfirmClose = useCallback(async () => {
    if (!pendingFormData) return;
    const closeDate = getFormString(pendingFormData, "close_date");
    const closePeriod = getFormEnum(pendingFormData, "period", isCashRegisterPeriod);
    if (!closeDate || !closePeriod) {
      toast.error("締め対象の日付・区分が不正です");
      setShowConfirm(false);
      return;
    }
    try {
      await mutateAsync({
        date: closeDate,
        period: closePeriod,
        actual_cash: Number(getFormString(pendingFormData, "actual_cash")),
        memo: getFormString(pendingFormData, "memo") || undefined,
      });
      toast.success("締めを実行しました");
      setShowConfirm(false);
      setPendingFormData(null);
      setActualCash("");
    } catch {
      // FE-RC-005: useCreateCashRegisterClose.onError が既に handleApiError で通知済み。
      setShowConfirm(false);
    }
  }, [pendingFormData, mutateAsync]);

  const handleCancelConfirm = useCallback(() => {
    setShowConfirm(false);
    setPendingFormData(null);
  }, []);

  const periodLabel = PERIOD_LABELS[period];
  const goHistory = useCallback(() => {
    navigate(paths.accounting.closeHistory.getHref());
  }, [navigate]);

  return (
    <PageLayout
      title="レジ締め"
      resource={ResourceCashRegisterClose}
      icon={<Calculator className={`${ICON.page} ${C.text}`} />}
      maxWidth={LAYOUT.pageContentMaxWidth.full}
      headerAction={
        <FormHeaderActions
          onCancel={goHistory}
          submitLabel="締める"
          submitFormId="cash-register-close-form"
          submitDisabled={actualCash === "" || Number(actualCash) < 0}
        />
      }
    >
      <div className="space-y-6">
        <CashRegisterCloseTargetSection
          date={date}
          period={period}
          onDateChange={handleDateChange}
          onPeriodChange={handlePeriodChange}
          onEnablePreview={enablePreview}
        />

        {previewLoading ? (
          <div className="flex items-center justify-center py-8">
            <p className={`text-base ${C.text50}`}>読み込み中...</p>
          </div>
        ) : null}

        {preview !== undefined && !previewLoading ? (
          <CashRegisterClosePreview
            preview={preview}
            period={period}
            periodLabel={periodLabel}
            clinicName={clinicName}
            actualCash={actualCash}
            formAction={formAction}
            onActualCashChange={handleActualCashChange}
            onCancel={goHistory}
          />
        ) : null}

        <AlertDialog open={showConfirm} onOpenChange={setShowConfirm}>
          <AlertDialogContent>
            <AlertDialogHeader>
              <AlertDialogTitle>締めを実行しますか？</AlertDialogTitle>
              <AlertDialogDescription>
                {date} {periodLabel} の締めを実行します。この操作は取り消せません。
              </AlertDialogDescription>
            </AlertDialogHeader>
            <AlertDialogFooter>
              <AlertDialogCancel onClick={handleCancelConfirm} disabled={isPending}>
                キャンセル
              </AlertDialogCancel>
              <AlertDialogAction
                onClick={handleConfirmClose}
                disabled={isPending}
                className={`${C.bgBrand} ${C.textOnBrand} ${C.hoverBgBrand} ${C.hoverTextOnBrand} rounded-full`}
              >
                {isPending ? "処理中..." : "締める"}
              </AlertDialogAction>
            </AlertDialogFooter>
          </AlertDialogContent>
        </AlertDialog>
      </div>
    </PageLayout>
  );
}
