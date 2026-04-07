import { memo, useState, useCallback } from "react";
import { FileText } from "lucide-react";
import { TableCell } from "@/components/ui/table";
import { DataTableRow } from "@/components/shared/DataTable/DataTableRow";
import { RowActionButton } from "@/components/shared/RowActionButton";
import { NotionStatusPill } from "@/components/shared/StatusPill/NotionStatusPill";
import { PropertyRow } from "@/components/shared/SidePeek/PropertyRow";
import { StatusToggleButton } from "@/components/shared/SidePeek/StatusToggleButton";
import { MasterSidePanel } from "@/components/shared/SidePeek/MasterSidePanel";
import { C, LAYOUT, ICON } from "@/lib/design-tokens";
import { MASTER_INPUT_CLASS, MASTER_STATUS_FILTER } from "@/features/master/constants/styles";
import { useMasterCRUD } from "@/features/master/hooks/use-master-crud";
import { useMasterSave } from "@/features/master/hooks/use-master-save";
import { MasterCRUDPage } from "@/features/master/components/MasterCRUDPage";
import { useGetInquiryTemplates, useCreateInquiryTemplate, useUpdateInquiryTemplate, useDeleteInquiryTemplate } from "@/features/master/api/inquiry-templates";
import type {
  InquiryTemplate,
  CreateInquiryTemplateRequest,
  UpdateInquiryTemplateRequest,
} from "@/features/master/api/inquiry-templates";
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
  item, onClose, onSave, onDeleteRequest,
}: { item: InquiryTemplate | null; onClose: () => void; onSave: (d: FormData) => void; onDeleteRequest?: (i: InquiryTemplate) => void; }) {
  const [f, setF] = useState<FormData>(() => ({
    category: item?.category ?? "", title: item?.title ?? "", content: item?.content ?? "", isActive: item?.isActive ?? true,
  }));
  const [isDirty, setIsDirty] = useState(false);
  const [nameError, setNameError] = useState("");

  const handleTitleChange = useCallback((v: string) => {
    setF((p) => ({ ...p, title: v }));
    setIsDirty(true);
    if (v.trim()) setNameError("");
  }, []);

  const handleCategoryChange = useCallback((e: React.ChangeEvent<HTMLInputElement>) => {
    setF((p) => ({ ...p, category: e.target.value }));
    setIsDirty(true);
  }, []);

  const handleContentChange = useCallback((e: React.ChangeEvent<HTMLTextAreaElement>) => {
    setF((p) => ({ ...p, content: e.target.value }));
    setIsDirty(true);
  }, []);

  const handleToggleActive = useCallback(() => {
    setF((p) => ({ ...p, isActive: !p.isActive }));
    setIsDirty(true);
  }, []);

  const handleSave = useCallback(() => {
    if (!f.title.trim()) {
      setNameError("タイトルを入力してください");
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
    <MasterSidePanel isNew={item === null} title={f.title}
      onTitleChange={handleTitleChange} onClose={handleClose} action={handleSave}
      onDelete={item !== null && onDeleteRequest ? () => onDeleteRequest(item) : undefined}
      isDirty={isDirty}
      icon={<FileText className={LAYOUT.pageIcon.innerIcon} />}
      titleError={nameError}
      titleMaxLength={100}>
      <StatusToggleButton isActive={f.isActive} onToggle={handleToggleActive} />
      <PropertyRow label="カテゴリ">
        <input type="text" className={MASTER_INPUT_CLASS} value={f.category}
          onChange={handleCategoryChange} placeholder="カテゴリを入力" />
      </PropertyRow>
      <PropertyRow label="テンプレート内容">
        <textarea className={`${MASTER_INPUT_CLASS} min-h-[150px] resize-none`} value={f.content}
          onChange={handleContentChange} placeholder="テンプレート内容を入力" />
      </PropertyRow>
    </MasterSidePanel>
  );
});

export function InterviewTemplateSettings() {
  const { data } = useGetInquiryTemplates();
  const createMutation = useCreateInquiryTemplate();
  const updateMutation = useUpdateInquiryTemplate();
  const deleteMutation = useDeleteInquiryTemplate();
  const crud = useMasterCRUD<InquiryTemplate>({
    data, deleteMutation, entityLabel: "問診テンプレート",
    searchFilter: (item, lower) => item.title.toLowerCase().includes(lower) || item.category.toLowerCase().includes(lower),
  });
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
        <DataTableRow key={item.id} onClick={() => onEdit(item)}>
          <TableCell className={`text-base ${C.text}`}>{INQUIRY_CATEGORY_LABELS[item.category] ?? item.category}</TableCell>
          <TableCell className={`font-medium text-base ${C.text}`}>{item.title}</TableCell>
          <TableCell className="text-center"><NotionStatusPill isActive={item.isActive} /></TableCell>
          <TableCell className="p-0 text-right">{canEdit ? <RowActionButton onClick={() => onEdit(item)} /> : null}</TableCell>
        </DataTableRow>
      )}
      renderSidePanel={(props) => <SidePanel key={props.item?.id ?? "new"} {...props} />}
    />
  );
}
