import { memo, useState, useCallback, useEffect } from "react";
import { Shield } from "lucide-react";
import { usePermission } from "@/hooks/use-permission";
import { useSidePeekDirty } from "@/hooks/use-side-peek-dirty";
import { TableCell } from "@/components/ui/table";
import { DataTableRow } from "@/components/shared/DataTable/DataTableRow";
import { RowActionButton } from "@/components/shared/RowActionButton";
import { NotionStatusPill } from "@/components/shared/StatusPill/NotionStatusPill";
import { PropertyRow, StatusToggleButton, PropertyInput, MasterSidePanel } from "@/components/shared/SidePeek";
import { C, LAYOUT, ICON } from "@/lib/design-tokens";
import { MASTER_INPUT_CLASS, MASTER_STATUS_FILTER } from "../constants/styles";
import { useMasterCRUD } from "../hooks/use-master-crud";
import { useMasterSave } from "../hooks/use-master-save";
import { MasterCRUDPage } from "../components/MasterCRUDPage";
import { useGetAllInsurances, useCreateInsurance, useUpdateInsurance, useDeleteInsurance } from "../api/insurances";
import type { Insurance, CreateInsuranceRequest, UpdateInsuranceRequest } from "../api/insurances";
import { ResourceMasterInsurance } from "@/types/generated/models";

const COLUMNS = [
  { header: "名称", className: "flex-1" },
  { header: "補償率", className: "w-[100px]", align: "center" as const },
  { header: "連絡先", className: "w-[140px]" },
  { header: "ステータス", className: "w-[90px]", align: "center" as const },
  { header: "操作", className: "w-[80px]", align: "right" as const },
];

interface InsuranceFormData { name: string; description: string; coverageRate: string; contactPhone: string; isActive: boolean; }

const InsuranceSidePanel = memo(function InsuranceSidePanel({
  item, onClose, onSave, onDeleteRequest, readOnly, onDirtyChange,
}: { item: Insurance | null; onClose: () => void; onSave: (d: InsuranceFormData) => void; onDeleteRequest?: (i: Insurance) => void; readOnly?: boolean; onDirtyChange?: (dirty: boolean) => void; }) {
  const [formData, setFormData] = useState<InsuranceFormData>(() => ({
    name: item?.name ?? "", description: item?.description ?? "",
    coverageRate: item?.coverageRate != null ? String(item.coverageRate) : "0",
    contactPhone: item?.contactPhone ?? "", isActive: item?.isActive ?? true,
  }));
  const [isDirty, setIsDirty] = useState(false);
  const [nameError, setNameError] = useState("");

  // BUG-380
  useEffect(() => { onDirtyChange?.(isDirty); }, [isDirty, onDirtyChange]);
  const setFormDataDirty = useCallback<typeof setFormData>((updater) => {
    setFormData(updater);
    setIsDirty(true);
  }, []);

  const handleTitleChange = useCallback((v: string) => {
    setFormDataDirty((prev) => ({ ...prev, name: v }));
    if (v.trim()) setNameError("");
  }, [setFormDataDirty]);

  const handleCoverageRateChange = useCallback((e: React.ChangeEvent<HTMLInputElement>) => {
    setFormDataDirty((prev) => ({ ...prev, coverageRate: e.target.value }));
  }, [setFormDataDirty]);

  const handleContactPhoneChange = useCallback((e: React.ChangeEvent<HTMLInputElement>) => {
    setFormDataDirty((prev) => ({ ...prev, contactPhone: e.target.value }));
  }, [setFormDataDirty]);

  const handleDescriptionChange = useCallback((v: string) => {
    setFormDataDirty((prev) => ({ ...prev, description: v }));
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
    <MasterSidePanel isNew={item === null} title={formData.name}
      onTitleChange={handleTitleChange}
      onClose={handleClose} action={handleAction}
      onDelete={item !== null && onDeleteRequest ? () => onDeleteRequest(item) : undefined}
      icon={<Shield className={LAYOUT.pageIcon.innerIcon} />}
      isDirty={isDirty}
      titleError={nameError}
      titleMaxLength={100}
      readOnly={readOnly}>
      <StatusToggleButton isActive={formData.isActive} onToggle={handleToggleActive} />
      <PropertyRow label="補償率(%)">
        <input type="number" min={0} max={100} className={MASTER_INPUT_CLASS}
          value={formData.coverageRate} onChange={handleCoverageRateChange} placeholder="0" />
      </PropertyRow>
      <PropertyRow label="連絡先">
        <input type="tel" className={MASTER_INPUT_CLASS}
          value={formData.contactPhone} onChange={handleContactPhoneChange} placeholder="電話番号" />
      </PropertyRow>
      <PropertyRow label="備考">
        <PropertyInput value={formData.description} onChange={handleDescriptionChange} placeholder="補足情報など" />
      </PropertyRow>
    </MasterSidePanel>
  );
});

export function InsuranceSettings() {
  usePermission(ResourceMasterInsurance);
  const { data } = useGetAllInsurances();
  const createMutation = useCreateInsurance();
  const updateMutation = useUpdateInsurance();
  const deleteMutation = useDeleteInsurance();
  const dirty = useSidePeekDirty();
  const crud = useMasterCRUD<Insurance>({ data, deleteMutation, entityLabel: "保険", dirtyGuard: dirty });
  const handleDirtyChange = useCallback((d: boolean) => { if (d) dirty.markDirty(); else dirty.markClean(); }, [dirty]);
  const { handleSave } = useMasterSave<Insurance, InsuranceFormData, CreateInsuranceRequest, UpdateInsuranceRequest>({
    crud, createMutation, updateMutation,
    validate: (d) => (!d.name.trim() ? "名称は必須です" : null),
    toCreateRequest: (d) => ({
      name: d.name, description: d.description || undefined,
      coverage_rate: d.coverageRate !== "" ? Number(d.coverageRate) : 0,
      contact_phone: d.contactPhone || undefined, is_active: d.isActive,
    }),
    toUpdateRequest: (d) => ({
      name: d.name, description: d.description || undefined,
      coverage_rate: d.coverageRate !== "" ? Number(d.coverageRate) : 0,
      contact_phone: d.contactPhone || undefined, is_active: d.isActive,
    }),
  });

  return (
    <MasterCRUDPage title="保険マスタ" icon={<Shield className={`${ICON.page} ${C.text}`} />} resource={ResourceMasterInsurance}
      entityLabel="保険" searchPlaceholder="保険名で検索..." emptyMessage="保険が登録されていません"
      crud={crud} handleSave={handleSave} columns={COLUMNS}
      filterProperties={[MASTER_STATUS_FILTER]}
      renderRow={(item, onEdit, canEdit) => (
        <DataTableRow key={item.id} onClick={canEdit ? () => onEdit(item) : undefined}>
          <TableCell className={`font-medium text-base ${C.text}`}>{item.name}</TableCell>
          <TableCell className={`text-base text-center ${C.text}`}>{item.coverageRate > 0 ? `${item.coverageRate}%` : "-"}</TableCell>
          <TableCell className={`text-base ${C.text70}`}>{item.contactPhone || "-"}</TableCell>
          <TableCell className="text-center"><NotionStatusPill isActive={item.isActive} /></TableCell>
          <TableCell className="p-0 text-right">{canEdit ? <RowActionButton onClick={() => onEdit(item)} /> : null}</TableCell>
        </DataTableRow>
      )}
      renderSidePanel={(props) => <InsuranceSidePanel key={props.item?.id ?? "new"} {...props} onDirtyChange={handleDirtyChange} />}
    />
  );
}
