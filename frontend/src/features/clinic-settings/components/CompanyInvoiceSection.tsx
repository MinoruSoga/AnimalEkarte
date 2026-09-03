import { memo, useActionState } from "react";
import { FileText } from "lucide-react";
import { toast } from "sonner";

import { SubmitButton } from "@/components/shared/Form/SubmitButton";
import { handleApiError } from "@/lib/handle-api-error";
import { C, STYLE } from "@/lib/design-tokens";
import { useGetCompany, useUpdateCompany } from "@/hooks/use-company";

interface CompanyInvoiceFormProps {
  invoiceRegistrationNumber: string;
  canEdit: boolean;
}

// updatedAt をキーに親から remount させることで、保存成功後に再取得した最新値へ
// defaultValue を追従させる（useEffect 同期ではなく key によるリセット）。
const CompanyInvoiceForm = memo(function CompanyInvoiceForm({
  invoiceRegistrationNumber,
  canEdit,
}: CompanyInvoiceFormProps) {
  const updateMutation = useUpdateCompany();

  const [, formAction] = useActionState(async (_prev: null, formData: FormData) => {
    try {
      await updateMutation.mutateAsync({
        invoice_registration_number: String(formData.get("invoice_registration_number") ?? ""),
      });
      toast.success("インボイス登録番号を更新しました");
    } catch (error) {
      handleApiError(error, "インボイス登録番号の更新");
    }
    return null;
  }, null);

  return (
    <form action={formAction} className="flex items-end gap-3">
      <fieldset disabled={!canEdit} className="flex-1 border-0 p-0 m-0 min-w-0">
        <label htmlFor="invoice_registration_number" className={STYLE.formLabel}>
          インボイス登録番号
        </label>
        <input
          id="invoice_registration_number"
          name="invoice_registration_number"
          type="text"
          defaultValue={invoiceRegistrationNumber}
          placeholder="例: T1234567890123"
          className={`${STYLE.formInput} mt-1 w-full rounded-xs border px-3`}
        />
      </fieldset>
      {canEdit ? <SubmitButton>保存</SubmitButton> : null}
    </form>
  );
});

interface CompanyInvoiceSectionProps {
  canEdit: boolean;
}

// 会計伝票・領収書はこの値を AccountingDetailPage 経由で表示するため、
// 保存成功時は useUpdateCompany の invalidate によって即座に反映される。
export function CompanyInvoiceSection({ canEdit }: CompanyInvoiceSectionProps) {
  const { data: company } = useGetCompany();

  return (
    <section className={`bg-white rounded-lg border ${C.borderLight} p-6`}>
      <div className="flex items-center gap-2 mb-4">
        <FileText className={`size-4 ${C.text}`} />
        <h2 className={`text-base font-semibold ${C.text}`}>法人情報（インボイス）</h2>
      </div>
      {company ? (
        <CompanyInvoiceForm
          key={company.updatedAt}
          invoiceRegistrationNumber={company.invoiceRegistrationNumber}
          canEdit={canEdit}
        />
      ) : null}
    </section>
  );
}
