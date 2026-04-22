import { memo, useState, useCallback, useEffect } from "react";
import { FileText } from "lucide-react";
import { usePermission } from "@/hooks/use-permission";
import { useSidePeekDirty } from "@/hooks/use-side-peek-dirty";
import { TableCell } from "@/components/ui/table";
import { DataTableRow } from "@/components/shared/DataTable/DataTableRow";
import { RowActionButton } from "@/components/shared/RowActionButton";
import { NotionStatusPill } from "@/components/shared/StatusPill/NotionStatusPill";
import { PropertyRow, StatusToggleButton, MasterSidePanel } from "@/components/shared/SidePeek";
import { C, LAYOUT, ICON } from "@/lib/design-tokens";
import { MASTER_INPUT_CLASS, MASTER_STATUS_FILTER } from "../constants/styles";
import { useMasterCRUD } from "../hooks/use-master-crud";
import { useMasterSave } from "../hooks/use-master-save";
import { MasterCRUDPage } from "../components/MasterCRUDPage";
import { useGetInquiryTemplates, useCreateInquiryTemplate, useUpdateInquiryTemplate, useDeleteInquiryTemplate } from "../api/inquiry-templates";
import type {
  InquiryTemplate,
  CreateInquiryTemplateRequest,
  UpdateInquiryTemplateRequest,
} from "../api/inquiry-templates";
import { ResourceMasterMedical } from "@/types/generated/models";

// BUG-042: Map English snake_case category codes to Japanese labels
const INQUIRY_CATEGORY_LABELS: Record<string, string> = {
  chief_complaint: "主訴",
  history: "既往歴",
  current_medications: "現在の投薬",
  notes: "メモ/備考",
};

const COLUMNS = [
  { header: "カテゴリ", className: "w-[150px]" },
  { header: "タイトル", className: "flex-1" },
  { header: "ステータス", className: "w-[90px]", align: "center" as const },
  { header: "操作", className: "w-[80px]", align: "right" as const },
];

interface FormData { category: string; title: string; content: string; isActive: boolean; }

const SidePanel = memo(function SidePanel({
  item, onClose, onSave, onDeleteRequest, readOnly, onDirtyChange,
}: { item: InquiryTemplate | null; onClose: () => void; onSave: (d: FormData) => void; onDeleteRequest?: (i: InquiryTemplate) => void; readOnly?: boolean; onDirtyChange?: (dirty: boolean) => void; }) {
  const [formData, setFormData] = useState<FormData>(() => ({
    category: item?.category ?? "", title: item?.title ?? "", content: item?.content ?? "", isActive: item?.isActive ?? true,
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
    setFormDataDirty((prev) => ({ ...prev, title: v }));
    if (v.trim()) setNameError("");
  }, [setFormDataDirty]);

  const handleCategoryChange = useCallback((e: React.ChangeEvent<HTMLInputElement>) => {
    setFormDataDirty((prev) => ({ ...prev, category: e.target.value }));
  }, [setFormDataDirty]);

  const handleContentChange = useCallback((e: React.ChangeEvent<HTMLTextAreaElement>) => {
    setFormDataDirty((prev) => ({ ...prev, content: e.target.value }));
  }, [setFormDataDirty]);

  const handleToggleActive = useCallback(() => {
    setFormDataDirty((prev) => ({ ...prev, isActive: !prev.isActive }));
  }, [setFormDataDirty]);

  const handleSave = useCallback(() => {
    if (!formData.title.trim()) {
      setNameError("タイトルを入力してください");
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
    <MasterSidePanel isNew={item === null} title={formData.title}
      onTitleChange={handleTitleChange} onClose={handleClose} action={handleSave}
      onDelete={item !== null && onDeleteRequest ? () => onDeleteRequest(item) : undefined}
      isDirty={isDirty}
      icon={<FileText className={LAYOUT.pageIcon.innerIcon} />}
      titleError={nameError}
      titleMaxLength={100}
      readOnly={readOnly}>
      <StatusToggleButton isActive={formData.isActive} onToggle={handleToggleActive} />
      <PropertyRow label="カテゴリ">
        <input type="text" className={MASTER_INPUT_CLASS} value={formData.category}
          onChange={handleCategoryChange} placeholder="カテゴリを入力" />
      </PropertyRow>
      <PropertyRow label="テンプレート内容">
        <textarea className={`${MASTER_INPUT_CLASS} min-h-[150px] resize-none`} value={formData.content}
          onChange={handleContentChange} placeholder="テンプレート内容を入力" />
      </PropertyRow>
    </MasterSidePanel>
  );
});

export function InterviewTemplateSettings() {
  usePermission(ResourceMasterMedical);
  const { data } = useGetInquiryTemplates();
  const createMutation = useCreateInquiryTemplate();
  const updateMutation = useUpdateInquiryTemplate();
  const deleteMutation = useDeleteInquiryTemplate();
  const dirty = useSidePeekDirty();
  const crud = useMasterCRUD<InquiryTemplate>({
    data, deleteMutation, entityLabel: "問診テンプレート",
    searchFilter: (item, lower) => item.title.toLowerCase().includes(lower) || item.category.toLowerCase().includes(lower),
    dirtyGuard: dirty,
  });
  const handleDirtyChange = useCallback((d: boolean) => { if (d) dirty.markDirty(); else dirty.markClean(); }, [dirty]);
  const { handleSave } = useMasterSave<InquiryTemplate, FormData, CreateInquiryTemplateRequest, UpdateInquiryTemplateRequest>({
    crud, createMutation, updateMutation,
    validate: (d) => {
      if (!d.title.trim()) return "タイトルは必須です";
      if (!d.category.trim()) return "カテゴリは必須です";
      return null;
    },
    toCreateRequest: (d) => ({ category: d.category, title: d.title, content: d.content }),
    toUpdateRequest: (d) => ({ category: d.category, title: d.title, content: d.content, is_active: d.isActive }),
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
      renderSidePanel={(props) => <SidePanel key={props.item?.id ?? "new"} {...props} onDirtyChange={handleDirtyChange} />}
    />
  );
}
