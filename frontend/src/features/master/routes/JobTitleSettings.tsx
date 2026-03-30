import { memo, useState } from "react";
import { Briefcase } from "lucide-react";
import { TableCell } from "@/components/ui/table";
import { DataTableRow } from "@/components/shared/DataTable/DataTableRow";
import { RowActionButton } from "@/components/shared/RowActionButton";
import { NotionStatusPill } from "@/components/shared/StatusPill/NotionStatusPill";
import { PropertyRow } from "@/components/shared/SidePeek/PropertyRow";
import { StatusToggleButton } from "@/components/shared/SidePeek/StatusToggleButton";
import { PropertyInput } from "@/components/shared/SidePeek/PropertyInput";
import { MasterSidePanel } from "@/components/shared/SidePeek/MasterSidePanel";
import { C, LAYOUT, ICON } from "@/lib/design-tokens";
import { MASTER_STATUS_FILTER } from "@/features/master/constants/styles";
import { useMasterCRUD } from "@/features/master/hooks/use-master-crud";
import { useMasterSave } from "@/features/master/hooks/use-master-save";
import { MasterCRUDPage } from "@/features/master/components/MasterCRUDPage";
import {
  useGetAllJobTitles,
  useCreateJobTitle,
  useUpdateJobTitle,
  useDeleteJobTitle,
} from "@/features/master/api/job-titles";
import type { JobTitle, CreateJobTitleRequest, UpdateJobTitleRequest } from "@/features/master/api/job-titles";

// ─── Constants ───
const COLUMNS = [
  { header: "名称", className: "flex-1" },
  { header: "説明", className: "flex-1" },
  { header: "ステータス", className: "w-[90px]", align: "center" as const },
  { header: "操作", className: "w-[80px]", align: "right" as const },
];

// ─── FormData ───
interface JobTitleFormData {
  name: string;
  description: string;
  isActive: boolean;
}

// ─── SidePanel ───
const JobTitleSidePanel = memo(function JobTitleSidePanel({
  item, onClose, onSave, onDeleteRequest,
}: {
  item: JobTitle | null;
  onClose: () => void;
  onSave: (d: JobTitleFormData) => void;
  onDeleteRequest: (i: JobTitle) => void;
}) {
  const [f, setF] = useState<JobTitleFormData>(() => ({
    name: item?.name ?? "",
    description: item?.description ?? "",
    isActive: item?.isActive ?? true,
  }));

  return (
    <MasterSidePanel
      isNew={item === null}
      title={f.name}
      onTitleChange={(v) => setF((p) => ({ ...p, name: v }))}
      onClose={onClose}
      action={() => onSave(f)}
      onDelete={item !== null ? () => onDeleteRequest(item) : undefined}
      icon={<Briefcase className={LAYOUT.pageIcon.innerIcon} />}
    >
      <StatusToggleButton isActive={f.isActive} onToggle={() => setF((p) => ({ ...p, isActive: !p.isActive }))} />
      <PropertyRow label="説明">
        <PropertyInput value={f.description} onChange={(v) => setF((p) => ({ ...p, description: v }))} placeholder="説明を入力" />
      </PropertyRow>
    </MasterSidePanel>
  );
});

// ─── Page ───
export function JobTitleSettings() {
  const { data } = useGetAllJobTitles();
  const createMutation = useCreateJobTitle();
  const updateMutation = useUpdateJobTitle();
  const deleteMutation = useDeleteJobTitle();

  const crud = useMasterCRUD<JobTitle>({ data, deleteMutation, entityLabel: "役職" });

  const { handleSave } = useMasterSave<JobTitle, JobTitleFormData, CreateJobTitleRequest, UpdateJobTitleRequest>({
    crud,
    createMutation,
    updateMutation,
    validate: (d) => (!d.name.trim() ? "名称は必須です" : null),
    toCreateRequest: (d) => ({ name: d.name, description: d.description || undefined, is_active: d.isActive }),
    toUpdateRequest: (d) => ({ name: d.name, description: d.description || undefined, is_active: d.isActive }),
  });

  return (
    <MasterCRUDPage
      title="役職マスタ"
      icon={<Briefcase className={`${ICON.page} text-[#37352F]`} />}
      entityLabel="役職"
      searchPlaceholder="役職名で検索..."
      emptyMessage="役職が登録されていません"
      crud={crud}
      handleSave={handleSave}
      columns={COLUMNS}
      filterProperties={[MASTER_STATUS_FILTER]}
      renderRow={(item, onEdit) => (
        <DataTableRow key={item.id} onClick={() => onEdit(item)}>
          <TableCell className={`font-medium text-base ${C.text}`}>{item.name}</TableCell>
          <TableCell className={`text-base ${C.text}`}>{item.description || "-"}</TableCell>
          <TableCell className="text-center"><NotionStatusPill isActive={item.isActive} /></TableCell>
          <TableCell className="p-0 text-right"><RowActionButton onClick={() => onEdit(item)} /></TableCell>
        </DataTableRow>
      )}
      renderSidePanel={(props) => <JobTitleSidePanel key={props.item?.id ?? "new"} {...props} />}
    />
  );
}
