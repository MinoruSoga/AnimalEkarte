import { useCallback } from "react";
import { FileText } from "lucide-react";
import { usePermission } from "@/hooks/use-permission";
import { useSidePeekDirty } from "@/hooks/use-side-peek-dirty";
import { TableCell } from "@/components/ui/table";
import { DataTableRow } from "@/components/shared/DataTable/DataTableRow";
import { RowActionButton } from "@/components/shared/RowActionButton";
import { NotionStatusPill } from "@/components/shared/StatusPill/NotionStatusPill";
import { C, ICON } from "@/lib/design-tokens";
import { normalizeKana } from "@/lib/normalize-kana";
import { MASTER_STATUS_FILTER, MASTER_TABLE_COL } from "../constants/styles";
import { useMasterCRUD } from "../hooks/use-master-crud";
import { useMasterSave } from "../hooks/use-master-save";
import { MasterCRUDPage } from "../components/MasterCRUDPage";
import { InterviewTemplateSidePanel } from "../components/InterviewTemplateSidePanel";
import type { InterviewTemplateFormData } from "../components/InterviewTemplateSidePanelModel";
import { useGetInquiryTemplates, useCreateInquiryTemplate, useUpdateInquiryTemplate, useDeleteInquiryTemplate } from "../api/inquiry-templates";
import type {
  InquiryTemplate,
  CreateInquiryTemplateRequest,
  UpdateInquiryTemplateRequest,
} from "../api/inquiry-templates";
import {
  buildInterviewTemplateCreateRequest,
  buildInterviewTemplateUpdateRequest,
} from "./InterviewTemplateSettingsModel";
import { ResourceMasterMedical } from "@/types/generated/models";

// BUG-042: Map English snake_case category codes to Japanese labels
const INQUIRY_CATEGORY_LABELS: Record<string, string> = {
  chief_complaint: "主訴",
  history: "既往歴",
  current_medications: "現在の投薬",
  notes: "メモ/備考",
};

const COLUMNS = [
  { header: "カテゴリ", className: MASTER_TABLE_COL.w150 },
  { header: "タイトル", className: "flex-1" },
  { header: "ステータス", className: MASTER_TABLE_COL.w100, align: "center" as const },
  { header: "操作", className: MASTER_TABLE_COL.w80, align: "right" as const },
];

export function InterviewTemplateSettings() {
  usePermission(ResourceMasterMedical);
  const { data } = useGetInquiryTemplates();
  const createMutation = useCreateInquiryTemplate();
  const updateMutation = useUpdateInquiryTemplate();
  const deleteMutation = useDeleteInquiryTemplate();
  const dirty = useSidePeekDirty();
  const crud = useMasterCRUD<InquiryTemplate>({
    data, deleteMutation, entityLabel: "問診テンプレート",
    searchFilter: (item, lower) => normalizeKana(item.title).toLowerCase().includes(lower) || normalizeKana(item.category).toLowerCase().includes(lower),
    dirtyGuard: dirty,
  });
  const handleDirtyChange = useCallback((d: boolean) => { if (d) dirty.markDirty(); else dirty.markClean(); }, [dirty]);
  const { handleSave } = useMasterSave<InquiryTemplate, InterviewTemplateFormData, CreateInquiryTemplateRequest, UpdateInquiryTemplateRequest>({
    crud, createMutation, updateMutation,
    validate: (d) => {
      if (!d.title.trim()) return "タイトルは必須です";
      if (!d.category.trim()) return "カテゴリは必須です";
      return null;
    },
    toCreateRequest: buildInterviewTemplateCreateRequest,
    toUpdateRequest: buildInterviewTemplateUpdateRequest,
  });

  return (
    <MasterCRUDPage title="問診テンプレートマスタ" icon={<FileText className={`${ICON.page} ${C.text}`} />} resource={ResourceMasterMedical}
      entityLabel="問診テンプレート" searchPlaceholder="カテゴリ、タイトルで検索..." emptyMessage="問診テンプレートが登録されていません"
      crud={crud} handleSave={handleSave} columns={COLUMNS} deleteNameField="title"
      filterProperties={[MASTER_STATUS_FILTER]}
      renderRow={(item, onEdit, canEdit) => (
        <DataTableRow key={item.id} onClick={canEdit ? () => onEdit(item) : undefined}>
          <TableCell className={`text-base ${C.text}`}>{INQUIRY_CATEGORY_LABELS[item.category] ?? item.category}</TableCell>
          <TableCell className={`font-medium text-base ${C.text}`}>{item.title}</TableCell>
          <TableCell className="text-center"><NotionStatusPill isActive={item.isActive} /></TableCell>
          <TableCell className="p-0 text-right">{canEdit ? <RowActionButton onClick={() => onEdit(item)} /> : null}</TableCell>
        </DataTableRow>
      )}
      renderSidePanel={(props) => <InterviewTemplateSidePanel key={props.item?.id ?? "new"} {...props} onDirtyChange={handleDirtyChange} />}
    />
  );
}
