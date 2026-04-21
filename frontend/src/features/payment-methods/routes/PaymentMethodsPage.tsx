import { useActionState, useState, useCallback, memo } from "react";
import { useNavigate } from "react-router";
import { Plus, Trash2, Pencil, CreditCard } from "lucide-react";
import { toast } from "sonner";
import { C, ICON, STYLE } from "@/lib/design-tokens";
import { SubmitButton } from "@/components/shared/Form/SubmitButton";
import { PrimaryButton } from "@/components/shared/Form/PrimaryButton";
import { PageLayout } from "@/components/shared/PageLayout/PageLayout";
import { LoadingFallback, ErrorFallback } from "@/components/shared/DataStates";
import { handleApiError } from "@/lib/handle-api-error";
import { paths } from "@/config/paths";
import type { PaymentMethodMaster } from "@/types/generated/models";
import { ResourcePaymentMethod } from "@/types/generated/models";
import { useGetPaymentMethods } from "../api/get-payment-methods";
import { useCreatePaymentMethod } from "../api/create-payment-method";
import { useUpdatePaymentMethod } from "../api/update-payment-method";
import { useDeletePaymentMethod } from "../api/delete-payment-method";

const PaymentMethodRow = memo(function PaymentMethodRow({
  method,
  onEdit,
  onDelete,
}: {
  method: PaymentMethodMaster;
  onEdit: (m: PaymentMethodMaster) => void;
  onDelete: (id: number) => void;
}) {
  const handleEdit = useCallback(() => onEdit(method), [method, onEdit]);
  const handleDelete = useCallback(() => onDelete(method.id), [method.id, onDelete]);

  return (
    <div
      className={`flex items-center justify-between px-4 py-3 border-b ${C.borderLight} last:border-b-0`}
    >
      <div className="flex items-center gap-3">
        <span className={`text-base font-medium ${C.text}`}>{method.name}</span>
        <span
          className={`text-base px-2 py-0.5 rounded-full ${method.is_active ? `${C.bgStatusGreen} ${C.textStatusGreen}` : `${C.bgInactive} ${C.text50}`}`}
        >
          {method.is_active ? "有効" : "無効"}
        </span>
      </div>
      <div className="flex items-center gap-1">
        <button
          type="button"
          onClick={handleEdit}
          aria-label="編集"
          className={`size-8 flex items-center justify-center rounded-[3px] ${C.text50} ${C.hoverText} ${C.hoverBgLight} transition-colors`}
        >
          <Pencil className="size-4" />
        </button>
        <button
          type="button"
          onClick={handleDelete}
          aria-label="削除"
          className={`size-8 flex items-center justify-center rounded-[3px] ${C.text50} ${C.hoverTextDanger} ${C.hoverBgDanger5} transition-colors`}
        >
          <Trash2 className="size-4" />
        </button>
      </div>
    </div>
  );
});

interface CreateFormProps {
  onClose: () => void;
}

function CreateForm({ onClose }: CreateFormProps) {
  const createMutation = useCreatePaymentMethod();

  const [, formAction] = useActionState(async (_prev: null, formData: FormData) => {
    try {
      await createMutation.mutateAsync({
        name: formData.get("name") as string,
        is_active: formData.get("is_active") === "on",
      });
      toast.success("支払方法を追加しました");
      onClose();
    } catch (error) {
      handleApiError(error, "支払方法の追加");
    }
    return null;
  }, null);

  return (
    <form action={formAction} className={`p-4 mb-4 rounded-lg border ${C.borderMedium} space-y-3`}>
      <p className={`text-base font-medium ${C.text}`}>新しい支払方法</p>
      <div>
        <label htmlFor="pm_name" className={STYLE.formLabel}>
          名称
        </label>
        <input
          id="pm_name"
          name="name"
          type="text"
          className={`${STYLE.formInput} mt-1 w-full rounded-[4px] border px-3`}
          placeholder="例: クレジットカード"
          required
        />
      </div>
      <label className="flex items-center gap-2 cursor-pointer">
        <input type="checkbox" name="is_active" defaultChecked className="rounded" />
        <span className={`text-base ${C.text}`}>有効</span>
      </label>
      <div className="flex justify-end gap-2">
        <button
          type="button"
          onClick={onClose}
          className={`px-4 py-2 text-base ${C.text60} ${C.hoverBgLight} rounded-[4px] transition-colors`}
        >
          キャンセル
        </button>
        <SubmitButton>追加</SubmitButton>
      </div>
    </form>
  );
}

interface EditFormProps {
  method: PaymentMethodMaster;
  onClose: () => void;
}

function EditForm({ method, onClose }: EditFormProps) {
  const updateMutation = useUpdatePaymentMethod();

  const [, formAction] = useActionState(async (_prev: null, formData: FormData) => {
    try {
      await updateMutation.mutateAsync({
        id: method.id,
        data: {
          name: formData.get("name") as string,
          is_active: formData.get("is_active") === "on",
        },
      });
      toast.success("支払方法を更新しました");
      onClose();
    } catch (error) {
      handleApiError(error, "支払方法の更新");
    }
    return null;
  }, null);

  return (
    <form action={formAction} className={`p-4 mb-4 rounded-lg border ${C.borderMedium} space-y-3`}>
      <p className={`text-base font-medium ${C.text}`}>編集: {method.name}</p>
      <div>
        <label htmlFor="edit_pm_name" className={STYLE.formLabel}>
          名称
        </label>
        <input
          id="edit_pm_name"
          name="name"
          type="text"
          defaultValue={method.name}
          className={`${STYLE.formInput} mt-1 w-full rounded-[4px] border px-3`}
          required
        />
      </div>
      <label className="flex items-center gap-2 cursor-pointer">
        <input
          type="checkbox"
          name="is_active"
          defaultChecked={method.is_active}
          className="rounded"
        />
        <span className={`text-base ${C.text}`}>有効</span>
      </label>
      <div className="flex justify-end gap-2">
        <button
          type="button"
          onClick={onClose}
          className={`px-4 py-2 text-base ${C.text60} ${C.hoverBgLight} rounded-[4px] transition-colors`}
        >
          キャンセル
        </button>
        <SubmitButton>保存</SubmitButton>
      </div>
    </form>
  );
}

export function PaymentMethodsPage() {
  const navigate = useNavigate();
  const { data: methods, isLoading, isError } = useGetPaymentMethods();
  const deleteMutation = useDeletePaymentMethod();
  const [showCreateForm, setShowCreateForm] = useState(false);
  const [editTarget, setEditTarget] = useState<PaymentMethodMaster | null>(null);

  const handleEdit = useCallback((m: PaymentMethodMaster) => {
    setShowCreateForm(false);
    setEditTarget(m);
  }, []);

  const handleDelete = useCallback(
    async (id: number) => {
      try {
        await deleteMutation.mutateAsync(id);
        toast.success("支払方法を削除しました");
      } catch (error) {
        handleApiError(error, "支払方法の削除");
      }
    },
    [deleteMutation],
  );

  const handleShowCreate = useCallback(() => {
    setEditTarget(null);
    setShowCreateForm(true);
  }, []);

  const handleCloseCreate = useCallback(() => setShowCreateForm(false), []);
  const handleCloseEdit = useCallback(() => setEditTarget(null), []);

  return (
    <PageLayout
      title="支払方法マスタ"
      icon={<CreditCard className={`${ICON.page} ${C.text}`} />}
      resource={ResourcePaymentMethod}
      onBack={() => navigate(paths.settings.getHref())}
      maxWidth="max-w-2xl"
      headerAction={
        <PrimaryButton onClick={handleShowCreate}>
          <Plus className={`mr-1.5 ${ICON.action}`} />
          追加
        </PrimaryButton>
      }
    >
      {isLoading ? <LoadingFallback /> : null}
      {isError ? <ErrorFallback /> : null}
      {!isLoading && !isError ? (
        <div className="flex flex-col gap-4">
          {showCreateForm ? <CreateForm onClose={handleCloseCreate} /> : null}
          {editTarget !== null ? (
            <EditForm key={editTarget.id} method={editTarget} onClose={handleCloseEdit} />
          ) : null}

          <div className={`bg-white rounded-lg border ${C.borderLight}`}>
            {methods && methods.length > 0 ? (
              methods.map((method) => (
                <PaymentMethodRow
                  key={method.id}
                  method={method}
                  onEdit={handleEdit}
                  onDelete={handleDelete}
                />
              ))
            ) : (
              <p className={`text-base ${C.text50} py-8 text-center`}>
                支払方法は登録されていません
              </p>
            )}
          </div>
        </div>
      ) : null}
    </PageLayout>
  );
}
