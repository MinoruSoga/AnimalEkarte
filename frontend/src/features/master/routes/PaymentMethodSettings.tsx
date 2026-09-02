import { useCallback } from "react";
import { CreditCard } from "lucide-react";
import { usePermission } from "@/hooks/use-permission";
import { useSidePeekDirty } from "@/hooks/use-side-peek-dirty";
import { TableCell } from "@/components/ui/table";
import { DataTableRow } from "@/components/shared/DataTable/DataTableRow";
import { DataTableRowButton } from "@/components/shared/DataTable/DataTableRowButton";
import { RowActionButton } from "@/components/shared/RowActionButton";
import { StatusPill } from "@/components/shared/StatusPill/StatusPill";
import { C, ICON } from "@/lib/design-tokens";
import { MASTER_STATUS_FILTER, MASTER_TABLE_COL } from "../constants/styles";
import { useMasterCRUD } from "../hooks/use-master-crud";
import { useMasterSave } from "../hooks/use-master-save";
import { MasterCRUDPage } from "../components/MasterCRUDPage";
import { PaymentMethodSidePanel } from "../components/PaymentMethodSidePanel";
import type { PaymentMethodFormData } from "../components/payment-method-side-panel-model";
import {
  useGetPaymentMethods,
  useCreatePaymentMethod,
  useUpdatePaymentMethod,
  useDeletePaymentMethod,
} from "../api/payment-method-master";
import type {
  CreatePaymentMethodRequest,
  PaymentMethod,
  UpdatePaymentMethodRequest,
} from "../api/payment-method-master";
import {
  buildPaymentMethodCreateRequest,
  buildPaymentMethodUpdateRequest,
  validatePaymentMethodForm,
} from "./payment-method-settings-model";
import { ResourcePaymentMethod } from "@/types/generated/models";

// ─── Constants ───
const COLUMNS = [
  { header: "名称", className: "flex-1" },
  { header: "ステータス", className: MASTER_TABLE_COL.w100, align: "center" as const },
  { header: "操作", className: MASTER_TABLE_COL.w80, align: "right" as const },
];

// ─── Page ───
export function PaymentMethodSettings() {
  const { canCreate, canEdit, canDelete } = usePermission(ResourcePaymentMethod);
  const { data } = useGetPaymentMethods();
  const createMutation = useCreatePaymentMethod();
  const updateMutation = useUpdatePaymentMethod();
  const deleteMutation = useDeletePaymentMethod();

  const dirty = useSidePeekDirty();
  const crud = useMasterCRUD<PaymentMethod>({
    data,
    deleteMutation,
    entityLabel: "支払方法",
    dirtyGuard: dirty,
    permissions: { canDelete },
  });
  const handleDirtyChange = useCallback((d: boolean) => { if (d) dirty.markDirty(); else dirty.markClean(); }, [dirty]);

  const editingId =
    crud.editTarget !== null && crud.editTarget !== "new" ? crud.editTarget.id : null;
  const { handleSave } = useMasterSave<
    PaymentMethod,
    PaymentMethodFormData,
    CreatePaymentMethodRequest,
    UpdatePaymentMethodRequest
  >({
    crud,
    createMutation,
    updateMutation,
    permissions: { canCreate, canEdit },
    validate: (d) =>
      validatePaymentMethodForm(d, {
        existing: data,
        editingId,
      }),
    toCreateRequest: buildPaymentMethodCreateRequest,
    toUpdateRequest: buildPaymentMethodUpdateRequest,
  });

  return (
    <>
    <MasterCRUDPage
      title="支払方法マスタ"
      icon={<CreditCard className={`${ICON.page} ${C.text}`} />}
      resource={ResourcePaymentMethod}
      entityLabel="支払方法"
      searchPlaceholder="支払方法名で検索..."
      emptyMessage="支払方法が登録されていません"
      crud={crud}
      handleSave={handleSave}
      columns={COLUMNS}
      filterProperties={[MASTER_STATUS_FILTER]}
      renderRow={(item, onEdit, canEdit) => (
        <DataTableRow key={item.id}>
          <TableCell className={`font-medium ${C.text}`}>
            <DataTableRowButton
              aria-label={`詳細: 支払方法 ${item.name} (ID ${item.id})`}
              onClick={() => onEdit(item)}
            >
              {item.name}
            </DataTableRowButton>
          </TableCell>
          <TableCell className="text-center"><StatusPill isActive={item.isActive} /></TableCell>
          <TableCell className="text-right">
            {canEdit ? (
              <RowActionButton
                onClick={() => onEdit(item)}
                aria-label={`支払方法「${item.name}」(ID: ${item.id}) を編集`}
              />
            ) : null}
          </TableCell>
        </DataTableRow>
      )}
      renderSidePanel={(props) => (
        <PaymentMethodSidePanel
          key={props.item?.id ?? "new"}
          {...props}
          onSave={handleSave}
          onDirtyChange={handleDirtyChange}
        />
      )}
    />
    {dirty.discardDialog}
    </>
  );
}
