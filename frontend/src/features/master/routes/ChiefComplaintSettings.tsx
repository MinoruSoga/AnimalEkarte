import { memo, useState, useCallback, useEffect } from "react";
import { MessageSquareText } from "lucide-react";
import { useSidePeekDirty } from "@/hooks/use-side-peek-dirty";
import { TableCell } from "@/components/ui/table";
import { DataTableRow } from "@/components/shared/DataTable/DataTableRow";
import { RowActionButton } from "@/components/shared/RowActionButton";
import { NotionStatusPill } from "@/components/shared/StatusPill/NotionStatusPill";
import { PropertyRow, StatusToggleButton, MasterSidePanel } from "@/components/shared/SidePeek";
import { C, LAYOUT, ICON } from "@/lib/design-tokens";
import { MASTER_INPUT_CLASS, MASTER_STATUS_FILTER } from "@/features/master/constants/styles";
import { useMasterCRUD } from "@/features/master/hooks/use-master-crud";
import { useMasterSave } from "@/features/master/hooks/use-master-save";
import { MasterCRUDPage } from "@/features/master/components/MasterCRUDPage";
import { useGetChiefComplaintTypes, useCreateChiefComplaintType, useUpdateChiefComplaintType, useDeleteChiefComplaintType } from "@/features/master/api/chief-complaint-types";
import type {
  ChiefComplaintType,
  CreateChiefComplaintTypeRequest,
  UpdateChiefComplaintTypeRequest,
} from "@/features/master/api/chief-complaint-types";
import { ResourceMasterMedical } from "@/types/generated/models";

const COLUMNS = [
  { header: "名称", className: "flex-1" },
  { header: "説明", className: "flex-1" },
  { header: "ステータス", className: "w-[90px]", align: "center" as const },
  { header: "操作", className: "w-[80px]", align: "right" as const },
];

interface FormData { name: string; description: string; isActive: boolean; }

const SidePanel = memo(function SidePanel({
  item, onClose, onSave, onDeleteRequest, readOnly, onDirtyChange,
}: { item: ChiefComplaintType | null; onClose: () => void; onSave: (d: FormData) => void; onDeleteRequest?: (i: ChiefComplaintType) => void; readOnly?: boolean; onDirtyChange?: (dirty: boolean) => void; }) {
  const [f, setF] = useState<FormData>(() => ({
    name: item?.name ?? "", description: item?.description ?? "", isActive: item?.isActive ?? true,
  }));
  const [isDirty, setIsDirty] = useState(false);
  const [nameError, setNameError] = useState("");

  // BUG-380
  useEffect(() => { onDirtyChange?.(isDirty); }, [isDirty, onDirtyChange]);

  const handleTitleChange = useCallback((v: string) => {
    setF((p) => ({ ...p, name: v }));
    setIsDirty(true);
    if (v.trim()) setNameError("");
  }, []);

  const handleDescriptionChange = useCallback((e: React.ChangeEvent<HTMLTextAreaElement>) => {
    setF((p) => ({ ...p, description: e.target.value }));
    setIsDirty(true);
  }, []);

  const handleToggleActive = useCallback(() => {
    setF((p) => ({ ...p, isActive: !p.isActive }));
    setIsDirty(true);
  }, []);

  const handleAction = useCallback(() => {
    if (!f.name.trim()) {
      setNameError("名称を入力してください");
      return;
    }
    setNameError("");
    onSave(f);
    setIsDirty(false);
  }, [f, onSave]);

  const handleClose = useCallback(() => {
    setIsDirty(false);
    onClose();
  }, [onClose]);

  return (
    <MasterSidePanel isNew={item === null} title={f.name}
      onTitleChange={handleTitleChange}
      onClose={handleClose} action={handleAction}
      onDelete={item !== null && onDeleteRequest ? () => onDeleteRequest(item) : undefined}
      icon={<MessageSquareText className={LAYOUT.pageIcon.innerIcon} />}
      isDirty={isDirty}
      titleError={nameError}
      titleMaxLength={100}
      readOnly={readOnly}>
      <StatusToggleButton isActive={f.isActive} onToggle={handleToggleActive} />
      <PropertyRow label="説明">
        <textarea className={`${MASTER_INPUT_CLASS} min-h-[80px] resize-none`}
          value={f.description} onChange={handleDescriptionChange} placeholder="説明を入力" />
      </PropertyRow>
    </MasterSidePanel>
  );
});

export function ChiefComplaintSettings() {
  const { data } = useGetChiefComplaintTypes();
  const createMutation = useCreateChiefComplaintType();
  const updateMutation = useUpdateChiefComplaintType();
  const deleteMutation = useDeleteChiefComplaintType();
  const dirty = useSidePeekDirty();
  const crud = useMasterCRUD<ChiefComplaintType>({ data, deleteMutation, entityLabel: "主訴マスタ", dirtyGuard: dirty });
  const handleDirtyChange = useCallback((d: boolean) => { if (d) dirty.markDirty(); else dirty.markClean(); }, [dirty]);
  const { handleSave } = useMasterSave<ChiefComplaintType, FormData, CreateChiefComplaintTypeRequest, UpdateChiefComplaintTypeRequest>({
    crud, createMutation, updateMutation,
    validate: (d) => (!d.name.trim() ? "名称は必須です" : null),
    toCreateRequest: (d) => ({ name: d.name, description: d.description || undefined }),
    toUpdateRequest: (d) => ({ name: d.name, description: d.description || undefined, is_active: d.isActive }),
  });

  return (
    <MasterCRUDPage title="主訴マスタ" icon={<MessageSquareText className={`${ICON.page} ${C.text}`} />} resource={ResourceMasterMedical}
      entityLabel="主訴マスタ" searchPlaceholder="名称で検索..." emptyMessage="主訴マスタが登録されていません"
      crud={crud} handleSave={handleSave} columns={COLUMNS}
      filterProperties={[MASTER_STATUS_FILTER]}
      renderRow={(item, onEdit, canEdit) => (
        <DataTableRow key={item.id} onClick={canEdit ? () => onEdit(item) : undefined}>
          <TableCell className={`font-medium text-base ${C.text}`}>{item.name}</TableCell>
          <TableCell className={`text-base ${C.text}`}>{item.description}</TableCell>
          <TableCell className="text-center"><NotionStatusPill isActive={item.isActive} /></TableCell>
          <TableCell className="p-0 text-right">{canEdit ? <RowActionButton onClick={() => onEdit(item)} /> : null}</TableCell>
        </DataTableRow>
      )}
      renderSidePanel={(props) => <SidePanel key={props.item?.id ?? "new"} {...props} onDirtyChange={handleDirtyChange} />}
    />
  );
}
