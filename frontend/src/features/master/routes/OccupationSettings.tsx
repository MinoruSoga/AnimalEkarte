import { memo, useState, useCallback, useEffect } from "react";
import { Briefcase } from "lucide-react";
import { useSidePeekDirty } from "@/hooks/use-side-peek-dirty";
import { TableCell } from "@/components/ui/table";
import { DataTableRow } from "@/components/shared/DataTable/DataTableRow";
import { RowActionButton } from "@/components/shared/RowActionButton";
import { NotionStatusPill } from "@/components/shared/StatusPill/NotionStatusPill";
import { PropertyRow, StatusToggleButton, PropertyInput, MasterSidePanel } from "@/components/shared/SidePeek";
import { C, LAYOUT, ICON } from "@/lib/design-tokens";
import { MASTER_STATUS_FILTER } from "../constants/styles";
import { useMasterCRUD } from "../hooks/use-master-crud";
import { useMasterSave } from "../hooks/use-master-save";
import { usePermission } from "@/hooks/use-permission";
import { MasterCRUDPage } from "../components/MasterCRUDPage";
import { useGetAllOccupations, useCreateOccupation, useUpdateOccupation, useDeleteOccupation } from "../api/occupations";
import type { Occupation, CreateOccupationRequest, UpdateOccupationRequest } from "../api/occupations";
import { ResourceMasterStaff } from "@/types/generated/models";

// ─── Constants ───
const COLUMNS = [
  { header: "名称", className: "flex-1" },
  { header: "説明", className: "flex-1" },
  { header: "ステータス", className: "w-[90px]", align: "center" as const },
  { header: "操作", className: "w-[80px]", align: "right" as const },
];

// ─── FormData ───
interface OccupationFormData {
  name: string;
  description: string;
  isActive: boolean;
}

// ─── SidePanel ───
const OccupationSidePanel = memo(function OccupationSidePanel({
  item, onClose, onSave, onDeleteRequest, readOnly, onDirtyChange,
}: {
  item: Occupation | null;
  onClose: () => void;
  onSave: (d: OccupationFormData) => void;
  onDeleteRequest?: (i: Occupation) => void;
  readOnly?: boolean;
  onDirtyChange?: (dirty: boolean) => void;
}) {
  const [formData, setFormData] = useState<OccupationFormData>(() => ({
    name: item?.name ?? "",
    description: item?.description ?? "",
    isActive: item?.isActive ?? true,
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
    <MasterSidePanel
      isNew={item === null}
      title={formData.name}
      onTitleChange={handleTitleChange}
      onClose={handleClose}
      action={handleAction}
      onDelete={item !== null && onDeleteRequest ? () => onDeleteRequest(item) : undefined}
      icon={<Briefcase className={LAYOUT.pageIcon.innerIcon} />}
      isDirty={isDirty}
      titleError={nameError}
      titleMaxLength={100}
      readOnly={readOnly}
    >
      <StatusToggleButton isActive={formData.isActive} onToggle={handleToggleActive} />
      <PropertyRow label="説明">
        <PropertyInput value={formData.description} onChange={handleDescriptionChange} placeholder="説明を入力" />
      </PropertyRow>
    </MasterSidePanel>
  );
});

// ─── Page ───
export function OccupationSettings() {
  usePermission(ResourceMasterStaff);
  const { data } = useGetAllOccupations();
  const createMutation = useCreateOccupation();
  const updateMutation = useUpdateOccupation();
  const deleteMutation = useDeleteOccupation();

  const dirty = useSidePeekDirty();
  const crud = useMasterCRUD<Occupation>({ data, deleteMutation, entityLabel: "職種", dirtyGuard: dirty });
  const handleDirtyChange = useCallback((d: boolean) => { if (d) dirty.markDirty(); else dirty.markClean(); }, [dirty]);

  const { handleSave } = useMasterSave<Occupation, OccupationFormData, CreateOccupationRequest, UpdateOccupationRequest>({
    crud,
    createMutation,
    updateMutation,
    validate: (d) => (!d.name.trim() ? "名称は必須です" : null),
    toCreateRequest: (d) => ({ name: d.name, description: d.description || undefined, is_active: d.isActive }),
    toUpdateRequest: (d) => ({ name: d.name, description: d.description || undefined, is_active: d.isActive }),
  });

  return (
    <MasterCRUDPage
      title="職種マスタ"
      icon={<Briefcase className={`${ICON.page} ${C.text}`} />}
      resource={ResourceMasterStaff}
      entityLabel="職種"
      searchPlaceholder="職種名で検索..."
      emptyMessage="職種が登録されていません"
      crud={crud}
      handleSave={handleSave}
      columns={COLUMNS}
      filterProperties={[MASTER_STATUS_FILTER]}
      renderRow={(item, onEdit, canEdit) => (
        <DataTableRow key={item.id} onClick={canEdit ? () => onEdit(item) : undefined}>
          <TableCell className={`font-medium text-base ${C.text}`}>{item.name}</TableCell>
          <TableCell className={`text-base ${C.text}`}>{item.description || "-"}</TableCell>
          <TableCell className="text-center"><NotionStatusPill isActive={item.isActive} /></TableCell>
          <TableCell className="p-0 text-right">{canEdit ? <RowActionButton onClick={() => onEdit(item)} /> : null}</TableCell>
        </DataTableRow>
      )}
      renderSidePanel={(props) => <OccupationSidePanel key={props.item?.id ?? "new"} {...props} onDirtyChange={handleDirtyChange} />}
    />
  );
}
