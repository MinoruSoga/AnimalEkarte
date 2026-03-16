import { useState, useMemo, useCallback, memo, useDeferredValue, useTransition } from "react";
import { useNavigate } from "react-router";
import { paths } from "@/config/paths";

import { Plus, FileText } from "lucide-react";
import { toast } from "sonner";

import { TableCell } from "@/components/ui/table";
import { PageLayout } from "@/components/shared/PageLayout/PageLayout";
import { SearchFilterBar } from "@/components/shared/SearchFilterBar/SearchFilterBar";
import { DataTable } from "@/components/shared/DataTable/DataTable";
import { DataTableRow } from "@/components/shared/DataTable/DataTableRow";
import { RowActionButton } from "@/components/shared/RowActionButton";
import { ConfirmDialog } from "@/components/shared/ConfirmDialog/ConfirmDialog";
import { NotionStatusPill } from "@/components/shared/StatusPill/NotionStatusPill";
import { PropertyRow } from "@/components/shared/SidePeek/PropertyRow";
import { StatusToggleButton } from "@/components/shared/SidePeek/StatusToggleButton";
import { MasterSidePanel } from "@/components/shared/SidePeek/MasterSidePanel";
import { PrimaryButton } from "@/components/shared/Form/PrimaryButton";
import { C, LAYOUT } from "@/lib/design-tokens";
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

// ─────────────────────────────────────────────────
// Constants
// ─────────────────────────────────────────────────

const COLUMNS = [
  { header: "カテゴリ", className: "w-[150px]" },
  { header: "タイトル", className: "flex-1" },
  { header: "ステータス", className: "w-[90px]", align: "center" as const },
  { header: "操作", className: "w-[80px]", align: "right" as const },
];

const INPUT_CLASS = `w-full bg-transparent text-sm ${C.text} outline-none border-none px-1.5 py-0.5 rounded-[3px] ${C.hoverBgLight} ${C.focusBgLight} transition-colors ${C.textPlaceholder}`;

// ─────────────────────────────────────────────────
// Form state types
// ─────────────────────────────────────────────────

interface FormData {
  category: string;
  title: string;
  content: string;
  isActive: boolean;
}

// ─────────────────────────────────────────────────
// SidePanel
// ─────────────────────────────────────────────────

interface SidePanelProps {
  item: InquiryTemplate | null;
  onClose: () => void;
  onSave: (data: FormData) => void;
  onDeleteRequest: (item: InquiryTemplate) => void;
}

const SidePanel = memo(function SidePanel({
  item,
  onClose,
  onSave,
  onDeleteRequest,
}: SidePanelProps) {
  const [formData, setFormData] = useState<FormData>(() => ({
    category: item?.category ?? "",
    title: item?.title ?? "",
    content: item?.content ?? "",
    isActive: item?.isActive ?? true,
  }));

  return (
    <MasterSidePanel
      isNew={item === null}
      title={formData.title}
      onTitleChange={(v) => setFormData((prev) => ({ ...prev, title: v }))}
      onClose={onClose}
      onSave={() => onSave(formData)}
      onDelete={item !== null ? () => onDeleteRequest(item) : undefined}
      icon={<FileText className={LAYOUT.pageIcon.innerIcon} />}
    >
      <StatusToggleButton
        isActive={formData.isActive}
        onToggle={() => setFormData((prev) => ({ ...prev, isActive: !prev.isActive }))}
      />

      <PropertyRow label="カテゴリ">
        <input
          type="text"
          className={INPUT_CLASS}
          value={formData.category}
          onChange={(e) => setFormData((prev) => ({ ...prev, category: e.target.value }))}
          placeholder="カテゴリを入力"
        />
      </PropertyRow>

      <PropertyRow label="テンプレート内容">
        <textarea
          className={`${INPUT_CLASS} min-h-[150px] resize-none`}
          value={formData.content}
          onChange={(e) => setFormData((prev) => ({ ...prev, content: e.target.value }))}
          placeholder="テンプレート内容を入力"
        />
      </PropertyRow>
    </MasterSidePanel>
  );
});

// ─────────────────────────────────────────────────
// Main page
// ─────────────────────────────────────────────────

export function InterviewTemplateSettings() {
  const navigate = useNavigate();

  const [editTarget, setEditTarget] = useState<InquiryTemplate | "new" | null>(null);
  const [searchTerm, setSearchTerm] = useState("");
  const [pendingDelete, setPendingDelete] = useState<InquiryTemplate | null>(null);
  const [, startSaveTransition] = useTransition();

  const { data: items = [] } = useGetInquiryTemplates();
  const createMutation = useCreateInquiryTemplate();
  const updateMutation = useUpdateInquiryTemplate();
  const deleteMutation = useDeleteInquiryTemplate();

  const deferredSearch = useDeferredValue(searchTerm);

  const filteredItems = useMemo(() => {
    if (!deferredSearch) return items;
    const lower = deferredSearch.toLowerCase();
    return items.filter(
      (item) =>
        item.title.toLowerCase().includes(lower) || item.category.toLowerCase().includes(lower)
    );
  }, [items, deferredSearch]);

  const handleClose = useCallback(() => setEditTarget(null), []);

  const handleSave = useCallback(
    (data: FormData) => {
      if (!data.title.trim()) {
        toast.error("タイトルは必須です");
        return;
      }
      if (!data.category.trim()) {
        toast.error("カテゴリは必須です");
        return;
      }

      startSaveTransition(() => {
        if (editTarget !== null && editTarget !== "new") {
          const req: UpdateInquiryTemplateRequest = {
            category: data.category,
            title: data.title,
            content: data.content,
            is_active: data.isActive,
          };
          updateMutation.mutate(
            { id: editTarget.id, req },
            {
              onSuccess: () => {
                toast.success("更新しました");
                handleClose();
              },
              onError: () => toast.error("更新に失敗しました"),
            }
          );
        } else {
          const req: CreateInquiryTemplateRequest = {
            category: data.category,
            title: data.title,
            content: data.content,
          };
          createMutation.mutate(req, {
            onSuccess: () => {
              toast.success("登録しました");
              handleClose();
            },
            onError: () => toast.error("登録に失敗しました"),
          });
        }
      });
    },
    [editTarget, updateMutation, createMutation, handleClose]
  );

  const handleDeleteConfirm = useCallback(() => {
    if (!pendingDelete) return;
    deleteMutation.mutate(pendingDelete.id, {
      onSuccess: () => {
        setPendingDelete(null);
        handleClose();
        toast.success("削除しました");
      },
      onError: () => toast.error("削除に失敗しました"),
    });
  }, [pendingDelete, deleteMutation, handleClose]);

  const panelItem = editTarget !== null && editTarget !== "new" ? editTarget : null;

  return (
    <>
      <div className="flex h-full">
        <div className="flex-1 min-w-0">
          <PageLayout
            title="問診テンプレートマスタ"
            icon={<FileText className="size-5 text-[#37352F]" />}
            onBack={() => navigate(paths.settings.getHref())}
            maxWidth="max-w-full"
            headerAction={
              <PrimaryButton onClick={() => setEditTarget("new")}>
                <Plus className="mr-1.5 size-4" />
                新規登録
              </PrimaryButton>
            }
          >
            <div className="flex flex-col gap-4">
              <SearchFilterBar
                searchTerm={searchTerm}
                onSearchChange={setSearchTerm}
                placeholder="カテゴリ、タイトルで検索..."
                count={filteredItems.length}
              />

              <DataTable
                columns={COLUMNS}
                data={filteredItems}
                emptyMessage="問診テンプレートが登録されていません"
                renderRow={(item) => (
                  <DataTableRow key={item.id} onClick={() => setEditTarget(item)}>
                    <TableCell className={`text-sm ${C.text}`}>{item.category}</TableCell>
                    <TableCell className={`font-medium text-sm ${C.text}`}>
                      {item.title}
                    </TableCell>
                    <TableCell className="text-center">
                      <NotionStatusPill isActive={item.isActive} />
                    </TableCell>
                    <TableCell className="p-0 text-right">
                      <RowActionButton onClick={() => setEditTarget(item)} />
                    </TableCell>
                  </DataTableRow>
                )}
              />
            </div>
          </PageLayout>
        </div>

        {editTarget !== null ? (
          <SidePanel
            key={panelItem ? String(panelItem.id) : "new"}
            item={panelItem}
            onClose={handleClose}
            onSave={handleSave}
            onDeleteRequest={setPendingDelete}
          />
        ) : null}
      </div>

      <ConfirmDialog
        open={pendingDelete !== null}
        onClose={() => setPendingDelete(null)}
        title="問診テンプレートを削除しますか？"
        description={`「${pendingDelete?.title}」を削除します。この操作は取り消せません。`}
        confirmLabel="削除"
        variant="destructive"
        onConfirm={handleDeleteConfirm}
      />
    </>
  );
}
