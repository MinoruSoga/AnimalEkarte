import { useState, useMemo, useCallback, memo, useDeferredValue, useTransition } from "react";
import { useNavigate } from "react-router";
import { paths } from "@/config/paths";

import { Plus, MessageSquareText } from "lucide-react";
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
  useGetChiefComplaintCategories,
  useCreateChiefComplaintCategory,
  useUpdateChiefComplaintCategory,
  useDeleteChiefComplaintCategory,
} from "@/features/master/api/chief-complaint-categories";

import type {
  ChiefComplaintCategory,
  CreateChiefComplaintCategoryRequest,
  UpdateChiefComplaintCategoryRequest,
} from "@/features/master/api/chief-complaint-categories";

// ─────────────────────────────────────────────────
// Constants
// ─────────────────────────────────────────────────

const COLUMNS = [
  { header: "名称", className: "flex-1" },
  { header: "説明", className: "flex-1" },
  { header: "ステータス", className: "w-[90px]", align: "center" as const },
  { header: "操作", className: "w-[80px]", align: "right" as const },
];

const INPUT_CLASS = `w-full bg-transparent text-sm ${C.text} outline-none border-none px-1.5 py-0.5 rounded-[3px] ${C.hoverBgLight} ${C.focusBgLight} transition-colors ${C.textPlaceholder}`;

// ─────────────────────────────────────────────────
// Form state types
// ─────────────────────────────────────────────────

interface FormData {
  name: string;
  description: string;
  isActive: boolean;
}

// ─────────────────────────────────────────────────
// SidePanel
// ─────────────────────────────────────────────────

interface SidePanelProps {
  item: ChiefComplaintCategory | null;
  onClose: () => void;
  onSave: (data: FormData) => void;
  onDeleteRequest: (item: ChiefComplaintCategory) => void;
}

const SidePanel = memo(function SidePanel({
  item,
  onClose,
  onSave,
  onDeleteRequest,
}: SidePanelProps) {
  const [formData, setFormData] = useState<FormData>(() => ({
    name: item?.name ?? "",
    description: item?.description ?? "",
    isActive: item?.isActive ?? true,
  }));

  return (
    <MasterSidePanel
      isNew={item === null}
      title={formData.name}
      onTitleChange={(v) => setFormData((prev) => ({ ...prev, name: v }))}
      onClose={onClose}
      onSave={() => onSave(formData)}
      onDelete={item !== null ? () => onDeleteRequest(item) : undefined}
      icon={<MessageSquareText className={LAYOUT.pageIcon.innerIcon} />}
    >
      <StatusToggleButton
        isActive={formData.isActive}
        onToggle={() => setFormData((prev) => ({ ...prev, isActive: !prev.isActive }))}
      />

      <PropertyRow label="説明">
        <textarea
          className={`${INPUT_CLASS} min-h-[80px] resize-none`}
          value={formData.description}
          onChange={(e) => setFormData((prev) => ({ ...prev, description: e.target.value }))}
          placeholder="説明を入力"
        />
      </PropertyRow>
    </MasterSidePanel>
  );
});

// ─────────────────────────────────────────────────
// Main page
// ─────────────────────────────────────────────────

export function ChiefComplaintSettings() {
  const navigate = useNavigate();

  const [editTarget, setEditTarget] = useState<ChiefComplaintCategory | "new" | null>(null);
  const [searchTerm, setSearchTerm] = useState("");
  const [pendingDelete, setPendingDelete] = useState<ChiefComplaintCategory | null>(null);
  const [, startSaveTransition] = useTransition();

  const { data: items = [] } = useGetChiefComplaintCategories();
  const createMutation = useCreateChiefComplaintCategory();
  const updateMutation = useUpdateChiefComplaintCategory();
  const deleteMutation = useDeleteChiefComplaintCategory();

  const deferredSearch = useDeferredValue(searchTerm);

  const filteredItems = useMemo(() => {
    if (!deferredSearch) return items;
    const lower = deferredSearch.toLowerCase();
    return items.filter((item) => item.name.toLowerCase().includes(lower));
  }, [items, deferredSearch]);

  const handleClose = useCallback(() => setEditTarget(null), []);

  const handleSave = useCallback(
    (data: FormData) => {
      if (!data.name.trim()) {
        toast.error("名称は必須です");
        return;
      }

      startSaveTransition(() => {
        if (editTarget !== null && editTarget !== "new") {
          const req: UpdateChiefComplaintCategoryRequest = {
            name: data.name,
            description: data.description || undefined,
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
          const req: CreateChiefComplaintCategoryRequest = {
            name: data.name,
            description: data.description || undefined,
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
            title="主訴マスタ"
            icon={<MessageSquareText className="size-5 text-[#37352F]" />}
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
                placeholder="名称で検索..."
                count={filteredItems.length}
              />

              <DataTable
                columns={COLUMNS}
                data={filteredItems}
                emptyMessage="主訴マスタが登録されていません"
                renderRow={(item) => (
                  <DataTableRow key={item.id} onClick={() => setEditTarget(item)}>
                    <TableCell className={`font-medium text-sm ${C.text}`}>
                      {item.name}
                    </TableCell>
                    <TableCell className={`text-sm ${C.text}`}>
                      {item.description}
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
        title="主訴マスタを削除しますか？"
        description={`「${pendingDelete?.name}」を削除します。この操作は取り消せません。`}
        confirmLabel="削除"
        variant="destructive"
        onConfirm={handleDeleteConfirm}
      />
    </>
  );
}
