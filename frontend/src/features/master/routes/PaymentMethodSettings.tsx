import { memo, useState, useCallback, useEffect } from "react";
import { CreditCard } from "lucide-react";
import { useSidePeekDirty } from "@/hooks/use-side-peek-dirty";
import { TableCell } from "@/components/ui/table";
import { DataTableRow } from "@/components/shared/DataTable/DataTableRow";
import { RowActionButton } from "@/components/shared/RowActionButton";
import { NotionStatusPill } from "@/components/shared/StatusPill/NotionStatusPill";
import { StatusToggleButton, MasterSidePanel } from "@/components/shared/SidePeek";
import { C, LAYOUT, ICON } from "@/lib/design-tokens";
import { MASTER_STATUS_FILTER } from "../constants/styles";
import { useMasterCRUD } from "../hooks/use-master-crud";
import { useMasterSave } from "../hooks/use-master-save";
import { usePermission } from "@/hooks/use-permission";
import { MasterCRUDPage } from "../components/MasterCRUDPage";
import {
  useGetPaymentMethods,
  useCreatePaymentMethod,
  useUpdatePaymentMethod,
  useDeletePaymentMethod,
} from "../api/payment-method-master";
import type {
  PaymentMethod,
  CreatePaymentMethodRequest,
  UpdatePaymentMethodRequest,
} from "../api/payment-method-master";
import { ResourcePaymentMethod } from "@/types/generated/models";

// ─── Constants ───
const COLUMNS = [
  { header: "名称", className: "flex-1" },
  { header: "ステータス", className: "w-[90px]", align: "center" as const },
  { header: "操作", className: "w-[80px]", align: "right" as const },
];

// ─── FormData ───
interface PaymentMethodFormData {
  name: string;
  isActive: boolean;
}

// ─── SidePanel ───
const PaymentMethodSidePanel = memo(function PaymentMethodSidePanel({
  item, onClose, onSave, onDeleteRequest, readOnly, onDirtyChange,
}: {
  item: PaymentMethod | null;
  onClose: () => void;
  onSave: (d: PaymentMethodFormData) => void;
  onDeleteRequest?: (i: PaymentMethod) => void;
  readOnly?: boolean;
  onDirtyChange?: (dirty: boolean) => void;
}) {
  const [formData, setFormData] = useState<PaymentMethodFormData>(() => ({
    name: item?.name ?? "",
    isActive: item?.isActive ?? true,
  }));
  const [isDirty, setIsDirty] = useState(false);
  const [nameError, setNameError] = useState("");

  useEffect(() => { onDirtyChange?.(isDirty); }, [isDirty, onDirtyChange]);

  const setFormDataDirty = useCallback<typeof setFormData>((updater) => {
    setFormData(updater);
    setIsDirty(true);
  }, []);

  const handleTitleChange = useCallback((v: string) => {
    setFormDataDirty((prev) => ({ ...prev, name: v }));
    if (v.trim()) setNameError("");
  }, [setFormDataDirty]);

  const handleToggleActive = useCallback(() => {
    setFormDataDirty((prev) => ({ ...prev, isActive: !prev.isActive }));
  }, [setFormDataDirty]);

  const handleAction = useCallback(() => {
    if (!formData.name.trim()) {
      setNameError("名称を入力してください");
      return;
    }
    setNameError("");
    onSave(formData);
    setIsDirty(false);
  }, [formData, onSave]);

  const handleClose = useCallback(() => {
    setIsDirty(false);
    onClose();
  }, [onClose]);

  return (
    <MasterSidePanel
      isNew={item === null}
      title={formData.name}
      onTitleChange={handleTitleChange}
      onClose={handleClose}
      action={handleAction}
      onDelete={item !== null && onDeleteRequest ? () => onDeleteRequest(item) : undefined}
      icon={<CreditCard className={LAYOUT.pageIcon.innerIcon} />}
      isDirty={isDirty}
      titleError={nameError}
      titleMaxLength={100}
      readOnly={readOnly}
    >
      <StatusToggleButton isActive={formData.isActive} onToggle={handleToggleActive} />
    </MasterSidePanel>
  );
});

// ─── Page ───
export function PaymentMethodSettings() {
  usePermission(ResourcePaymentMethod);
  const { data } = useGetPaymentMethods();
  const createMutation = useCreatePaymentMethod();
  const updateMutation = useUpdatePaymentMethod();
  const deleteMutation = useDeletePaymentMethod();

  const dirty = useSidePeekDirty();
  const crud = useMasterCRUD<PaymentMethod>({ data, deleteMutation, entityLabel: "支払方法", dirtyGuard: dirty });
  const handleDirtyChange = useCallback((d: boolean) => { if (d) dirty.markDirty(); else dirty.markClean(); }, [dirty]);

  const { handleSave } = useMasterSave<PaymentMethod, PaymentMethodFormData, CreatePaymentMethodRequest, UpdatePaymentMethodRequest>({
    crud,
    createMutation,
    updateMutation,
    validate: (d) => (!d.name.trim() ? "名称は必須です" : null),
    toCreateRequest: (d) => ({ name: d.name, is_active: d.isActive }),
    toUpdateRequest: (d) => ({ name: d.name, is_active: d.isActive }),
  });

  return (
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
        <DataTableRow key={item.id} onClick={canEdit ? () => onEdit(item) : undefined}>
          <TableCell className={`font-medium text-base ${C.text}`}>{item.name}</TableCell>
          <TableCell className="text-center"><NotionStatusPill isActive={item.isActive} /></TableCell>
          <TableCell className="p-0 text-right">{canEdit ? <RowActionButton onClick={() => onEdit(item)} /> : null}</TableCell>
        </DataTableRow>
      )}
      renderSidePanel={(props) => (
        <PaymentMethodSidePanel key={props.item?.id ?? "new"} {...props} onDirtyChange={handleDirtyChange} />
      )}
    />
  );
}
