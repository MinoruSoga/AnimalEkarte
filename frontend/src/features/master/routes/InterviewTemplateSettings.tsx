import { memo, useState } from "react";
import { FileText } from "lucide-react";
import { TableCell } from "@/components/ui/table";
import { DataTableRow } from "@/components/shared/DataTable/DataTableRow";
import { RowActionButton } from "@/components/shared/RowActionButton";
import { NotionStatusPill } from "@/components/shared/StatusPill/NotionStatusPill";
import { PropertyRow } from "@/components/shared/SidePeek/PropertyRow";
import { StatusToggleButton } from "@/components/shared/SidePeek/StatusToggleButton";
import { MasterSidePanel } from "@/components/shared/SidePeek/MasterSidePanel";
import { C, LAYOUT } from "@/lib/design-tokens";
import { MASTER_INPUT_CLASS, MASTER_STATUS_FILTER } from "@/features/master/constants/styles";
import { useMasterCRUD } from "@/features/master/hooks/use-master-crud";
import { useMasterSave } from "@/features/master/hooks/use-master-save";
import { MasterCRUDPage } from "@/features/master/components/MasterCRUDPage";
import {
  useGetInquiryTemplates,
  useCreateInquiryTemplate,
  useUpdateInquiryTemplate,
  useDeleteInquiryTemplate,
} from "@/features/master/api/inquiry-templates";
import type {
  InquiryTemplate,
  CreateInquiryTemplateRequest,
  UpdateInquiryTemplateRequest,
} from "@/features/master/api/inquiry-templates";

const COLUMNS = [
  { header: "カテゴリ", className: "w-[150px]" },
  { header: "タイトル", className: "flex-1" },
  { header: "ステータス", className: "w-[90px]", align: "center" as const },
  { header: "操作", className: "w-[80px]", align: "right" as const },
];

interface FormData { category: string; title: string; content: string; isActive: boolean; }

const SidePanel = memo(function SidePanel({
  item, onClose, onSave, onDeleteRequest,
}: { item: InquiryTemplate | null; onClose: () => void; onSave: (d: FormData) => void; onDeleteRequest: (i: InquiryTemplate) => void; }) {
  const [f, setF] = useState<FormData>(() => ({
    category: item?.category ?? "", title: item?.title ?? "", content: item?.content ?? "", isActive: item?.isActive ?? true,
  }));
  return (
    <MasterSidePanel isNew={item === null} title={f.title}
      onTitleChange={(v) => setF((p) => ({ ...p, title: v }))} onClose={onClose} onSave={() => onSave(f)}
      onDelete={item !== null ? () => onDeleteRequest(item) : undefined}
      icon={<FileText className={LAYOUT.pageIcon.innerIcon} />}>
      <StatusToggleButton isActive={f.isActive} onToggle={() => setF((p) => ({ ...p, isActive: !p.isActive }))} />
      <PropertyRow label="カテゴリ">
        <input type="text" className={MASTER_INPUT_CLASS} value={f.category}
          onChange={(e) => setF((p) => ({ ...p, category: e.target.value }))} placeholder="カテゴリを入力" />
      </PropertyRow>
      <PropertyRow label="テンプレート内容">
        <textarea className={`${MASTER_INPUT_CLASS} min-h-[150px] resize-none`} value={f.content}
          onChange={(e) => setF((p) => ({ ...p, content: e.target.value }))} placeholder="テンプレート内容を入力" />
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
    <MasterCRUDPage title="問診テンプレートマスタ" icon={<FileText className="size-5 text-[#37352F]" />}
      entityLabel="問診テンプレート" searchPlaceholder="カテゴリ、タイトルで検索..." emptyMessage="問診テンプレートが登録されていません"
      crud={crud} handleSave={handleSave} columns={COLUMNS} deleteNameField="title"
      filterProperties={[MASTER_STATUS_FILTER]}
      renderRow={(item, onEdit) => (
        <DataTableRow key={item.id} onClick={() => onEdit(item)}>
          <TableCell className={`text-base ${C.text}`}>{item.category}</TableCell>
          <TableCell className={`font-medium text-base ${C.text}`}>{item.title}</TableCell>
          <TableCell className="text-center"><NotionStatusPill isActive={item.isActive} /></TableCell>
          <TableCell className="p-0 text-right"><RowActionButton onClick={() => onEdit(item)} /></TableCell>
        </DataTableRow>
      )}
      renderSidePanel={(props) => <SidePanel key={props.item?.id ?? "new"} {...props} />}
    />
  );
}
