// React/Framework
import { useState, useMemo, useCallback, memo, useDeferredValue, useTransition } from "react";
import { useNavigate, useSearchParams } from "react-router";
import { paths } from "@/config/paths";

// DnD
import {
  DndContext,
  closestCenter,
} from "@dnd-kit/core";
import {
  SortableContext,
  verticalListSortingStrategy,
} from "@dnd-kit/sortable";

// Shared hooks
import { useSortableList } from "@/hooks/useSortableList";

// External
import Plus from "lucide-react/dist/esm/icons/plus";
import FolderTree from "lucide-react/dist/esm/icons/folder-tree";
import ClipboardList from "lucide-react/dist/esm/icons/clipboard-list";
import { toast } from "sonner";

// Internal
import { TableCell } from "@/components/ui/table";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { Tabs, TabsList, TabsTrigger, TabsContent } from "@/components/ui/tabs";
import { PageLayout } from "@/components/shared/PageLayout/PageLayout";
import { SearchFilterBar } from "@/components/shared/SearchFilterBar/SearchFilterBar";
import { DataTable } from "@/components/shared/DataTable/DataTable";
import { SortableDataTableRow } from "@/components/shared/DataTable/SortableDataTableRow";
import { RowActionButton } from "@/components/shared/RowActionButton";
import { ConfirmDialog } from "@/components/shared/ConfirmDialog/ConfirmDialog";
import { NotionStatusPill } from "@/components/shared/StatusPill/NotionStatusPill";
import { PropertyRow } from "@/components/shared/SidePeek/PropertyRow";
import { StatusToggleButton } from "@/components/shared/SidePeek/StatusToggleButton";
import { PropInput } from "@/components/shared/SidePeek/PropInput";
import { MasterSidePanel } from "@/components/shared/SidePeek/MasterSidePanel";
import { PrimaryButton } from "@/components/shared/Form/PrimaryButton";
import { C, STYLE, LAYOUT } from "@/lib/design-tokens";
import {
  useGetDiagnosisCategories,
  useCreateDiagnosisCategory,
  useUpdateDiagnosisCategory,
  useDeleteDiagnosisCategory,
  useReorderDiagnosisCategories,
  useGetDiagnosisNames,
  useCreateDiagnosisName,
  useUpdateDiagnosisName,
  useDeleteDiagnosisName,
  useReorderDiagnosisNames,
} from "@/features/master/api/diagnosis";

// Types
import type { DiagnosisCategory, DiagnosisName } from "@/features/master/api/diagnosis";
import type {
  CreateDiagnosisCategoryRequest,
  UpdateDiagnosisCategoryRequest,
  CreateDiagnosisNameRequest,
  UpdateDiagnosisNameRequest,
} from "@/types/diagnosis";

// ─────────────────────────────────────────────────
// Columns
// ─────────────────────────────────────────────────

const CATEGORY_COLUMNS = [
  { header: "", className: "w-[32px]" },
  { header: "カテゴリ名" },
  { header: "備考", className: "w-[240px]" },
  { header: "ステータス", className: "w-[100px]", align: "center" as const },
  { header: "操作", className: "w-[80px]", align: "right" as const },
];

const NAME_COLUMNS = [
  { header: "", className: "w-[32px]" },
  { header: "所属カテゴリ", className: "w-[160px]" },
  { header: "診断病名" },
  { header: "ステータス", className: "w-[100px]", align: "center" as const },
  { header: "操作", className: "w-[80px]", align: "right" as const },
];

const TABS = [
  { value: "diagnosis_category", label: "診断病名カテゴリ" },
  { value: "diagnosis_name", label: "診断病名" },
] as const;

// ─────────────────────────────────────────────────
// Form state types
// ─────────────────────────────────────────────────

interface DiagnosisCategoryFormData {
  name: string;
  description: string;
  isActive: boolean;
}

interface DiagnosisNameFormData {
  name: string;
  diagnosisCategoryId: string;
  description: string;
  isActive: boolean;
}

// ─────────────────────────────────────────────────
// DiagnosisCategorySidePanel
// ─────────────────────────────────────────────────

interface DiagnosisCategorySidePanelProps {
  item: DiagnosisCategory | null;
  onClose: () => void;
  onSave: (data: DiagnosisCategoryFormData) => void;
  onDeleteRequest: (item: DiagnosisCategory) => void;
}

const DiagnosisCategorySidePanel = memo(function DiagnosisCategorySidePanel({
  item,
  onClose,
  onSave,
  onDeleteRequest,
}: DiagnosisCategorySidePanelProps) {
  const [formData, setFormData] = useState<DiagnosisCategoryFormData>(() => ({
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
      icon={<FolderTree className={LAYOUT.pageIcon.innerIcon} />}
    >
      <StatusToggleButton
        isActive={formData.isActive}
        onToggle={() => setFormData((prev) => ({ ...prev, isActive: !prev.isActive }))}
      />
      <PropertyRow label="備考">
        <PropInput
          value={formData.description}
          onChange={(v) => setFormData((prev) => ({ ...prev, description: v }))}
          placeholder="補足情報など"
        />
      </PropertyRow>
    </MasterSidePanel>
  );
});

// ─────────────────────────────────────────────────
// DiagnosisNameSidePanel
// ─────────────────────────────────────────────────

interface DiagnosisNameSidePanelProps {
  item: DiagnosisName | null;
  categories: DiagnosisCategory[];
  onClose: () => void;
  onSave: (data: DiagnosisNameFormData) => void;
  onDeleteRequest: (item: DiagnosisName) => void;
}

const DiagnosisNameSidePanel = memo(function DiagnosisNameSidePanel({
  item,
  categories,
  onClose,
  onSave,
  onDeleteRequest,
}: DiagnosisNameSidePanelProps) {
  const [formData, setFormData] = useState<DiagnosisNameFormData>(() => ({
    name: item?.name ?? "",
    diagnosisCategoryId: item
      ? String(item.diagnosisCategoryId)
      : categories[0]
        ? String(categories[0].id)
        : "",
    description: item?.description ?? "",
    isActive: item?.isActive ?? true,
  }));

  // js-cache-function-results: API由来のJSXリストをuseMemoでキャッシュ
  const categorySelectItems = useMemo(
    () => categories.map((cat) => (
      <SelectItem key={cat.id} value={String(cat.id)}>{cat.name}</SelectItem>
    )),
    [categories],
  );

  return (
    <MasterSidePanel
      isNew={item === null}
      title={formData.name}
      onTitleChange={(v) => setFormData((prev) => ({ ...prev, name: v }))}
      onClose={onClose}
      onSave={() => onSave(formData)}
      onDelete={item !== null ? () => onDeleteRequest(item) : undefined}
      icon={<ClipboardList className={LAYOUT.pageIcon.innerIcon} />}
    >
      <StatusToggleButton
        isActive={formData.isActive}
        onToggle={() => setFormData((prev) => ({ ...prev, isActive: !prev.isActive }))}
      />
      <PropertyRow label="カテゴリ">
        <Select
          value={formData.diagnosisCategoryId}
          onValueChange={(v) => setFormData((prev) => ({ ...prev, diagnosisCategoryId: v }))}
        >
          <SelectTrigger className={STYLE.selectCompact}>
            <SelectValue placeholder="カテゴリを選択" />
          </SelectTrigger>
          <SelectContent>
            {categorySelectItems}
          </SelectContent>
        </Select>
      </PropertyRow>
      <PropertyRow label="備考">
        <PropInput
          value={formData.description}
          onChange={(v) => setFormData((prev) => ({ ...prev, description: v }))}
          placeholder="補足情報など"
        />
      </PropertyRow>
    </MasterSidePanel>
  );
});

// ─────────────────────────────────────────────────
// DiagnosisCategoryTab
// ─────────────────────────────────────────────────

interface DiagnosisCategoryTabProps {
  editTarget: DiagnosisCategory | "new" | null;
  onEditTargetChange: (v: DiagnosisCategory | "new" | null) => void;
}

function DiagnosisCategoryTab({ editTarget: _editTarget, onEditTargetChange }: DiagnosisCategoryTabProps) {
  const [searchTerm, setSearchTerm] = useState("");

  const { data: rawCategories } = useGetDiagnosisCategories();
  const reorderMutation = useReorderDiagnosisCategories();

  const { orderedItems: orderedCategories, sensors, handleDragEnd: handleCategoryDragEnd } =
    useSortableList({
      items: rawCategories ?? [],
      onReorder: (newIds) => {
        reorderMutation.mutate({ ids: newIds.map(Number) });
      },
    });

  // rerender-transitions: 検索フィルタを低優先度に遅延
  const deferredSearch = useDeferredValue(searchTerm);

  const filteredItems = useMemo(() => {
    if (!deferredSearch) return orderedCategories;
    const lower = deferredSearch.toLowerCase();
    return orderedCategories.filter((c) => c.name.toLowerCase().includes(lower));
  }, [orderedCategories, deferredSearch]);

  return (
    <div className="flex flex-col gap-4">
      <SearchFilterBar
        searchTerm={searchTerm}
        onSearchChange={setSearchTerm}
        placeholder="カテゴリ名で検索..."
        count={filteredItems.length}
      />

      <DndContext
        sensors={sensors}
        collisionDetection={closestCenter}
        onDragEnd={handleCategoryDragEnd}
      >
        <SortableContext
          items={filteredItems.map((i) => i.id)}
          strategy={verticalListSortingStrategy}
        >
          <DataTable
            columns={CATEGORY_COLUMNS}
            data={filteredItems}
            emptyMessage="診断カテゴリが登録されていません"
            renderRow={(item) => (
              <SortableDataTableRow
                key={item.id}
                id={item.id}
                onClick={() => onEditTargetChange(item)}
              >
                <TableCell className={`font-medium text-sm ${C.text}`}>
                  {item.name}
                </TableCell>
                <TableCell className={`text-sm ${C.text70} truncate max-w-[240px]`}>
                  {item.description || "-"}
                </TableCell>
                <TableCell className="text-center">
                  <NotionStatusPill isActive={item.isActive} />
                </TableCell>
                <TableCell className="p-0 text-right">
                  <RowActionButton onClick={() => onEditTargetChange(item)} />
                </TableCell>
              </SortableDataTableRow>
            )}
          />
        </SortableContext>
      </DndContext>
    </div>
  );
}

// ─────────────────────────────────────────────────
// DiagnosisNameTab
// ─────────────────────────────────────────────────

interface DiagnosisNameTabProps {
  editTarget: DiagnosisName | "new" | null;
  onEditTargetChange: (v: DiagnosisName | "new" | null) => void;
}

function DiagnosisNameTab({ editTarget: _editTarget, onEditTargetChange }: DiagnosisNameTabProps) {
  const [searchTerm, setSearchTerm] = useState("");

  const { data: rawCategories } = useGetDiagnosisCategories();
  const { data: rawNames } = useGetDiagnosisNames();

  const reorderMutation = useReorderDiagnosisNames();

  const { orderedItems: orderedNames, sensors, handleDragEnd: handleNameDragEnd } =
    useSortableList({
      items: rawNames ?? [],
      onReorder: (newIds) => {
        reorderMutation.mutate({ ids: newIds.map(Number) });
      },
    });

  // rerender-transitions: 検索フィルタを低優先度に遅延
  const deferredSearch = useDeferredValue(searchTerm);

  const filteredItems = useMemo(() => {
    if (!deferredSearch) return orderedNames;
    const lower = deferredSearch.toLowerCase();
    return orderedNames.filter((n) => n.name.toLowerCase().includes(lower));
  }, [orderedNames, deferredSearch]);

  const categoryMap = useMemo(
    () => new Map<string, string>((rawCategories ?? []).map((c) => [c.id, c.name])),
    [rawCategories],
  );

  return (
    <div className="flex flex-col gap-4">
      <SearchFilterBar
        searchTerm={searchTerm}
        onSearchChange={setSearchTerm}
        placeholder="診断病名で検索..."
        count={filteredItems.length}
      />

      <DndContext
        sensors={sensors}
        collisionDetection={closestCenter}
        onDragEnd={handleNameDragEnd}
      >
        <SortableContext
          items={filteredItems.map((i) => i.id)}
          strategy={verticalListSortingStrategy}
        >
          <DataTable
            columns={NAME_COLUMNS}
            data={filteredItems}
            emptyMessage="診断病名が登録されていません"
            renderRow={(item) => (
              <SortableDataTableRow
                key={item.id}
                id={item.id}
                onClick={() => onEditTargetChange(item)}
              >
                <TableCell className={`text-sm ${C.text70}`}>
                  {categoryMap.get(item.diagnosisCategoryId) ?? "-"}
                </TableCell>
                <TableCell className={`font-medium text-sm ${C.text}`}>
                  {item.name}
                </TableCell>
                <TableCell className="text-center">
                  <NotionStatusPill isActive={item.isActive} />
                </TableCell>
                <TableCell className="p-0 text-right">
                  <RowActionButton onClick={() => onEditTargetChange(item)} />
                </TableCell>
              </SortableDataTableRow>
            )}
          />
        </SortableContext>
      </DndContext>
    </div>
  );
}

// ─────────────────────────────────────────────────
// DiagnosisSettings (main page)
// ─────────────────────────────────────────────────

export function DiagnosisSettings() {
  const navigate = useNavigate();
  const [searchParams, setSearchParams] = useSearchParams();
  const activeTab = searchParams.get("tab") ?? "diagnosis_category";

  const [categoryEditTarget, setCategoryEditTarget] = useState<DiagnosisCategory | "new" | null>(null);
  const [nameEditTarget, setNameEditTarget] = useState<DiagnosisName | "new" | null>(null);
  const [categoryPendingDelete, setCategoryPendingDelete] = useState<DiagnosisCategory | null>(null);
  const [namePendingDelete, setNamePendingDelete] = useState<DiagnosisName | null>(null);
  const [, startSaveTransition] = useTransition();

  const { data: rawCategories } = useGetDiagnosisCategories();
  const createCategoryMutation = useCreateDiagnosisCategory();
  const updateCategoryMutation = useUpdateDiagnosisCategory();
  const deleteCategoryMutation = useDeleteDiagnosisCategory();

  const createNameMutation = useCreateDiagnosisName();
  const updateNameMutation = useUpdateDiagnosisName();
  const deleteNameMutation = useDeleteDiagnosisName();

  const handleTabChange = useCallback((tab: string) => {
    setSearchParams({ tab });
    setCategoryEditTarget(null);
    setNameEditTarget(null);
  }, [setSearchParams]);

  const handleCategoryClose = useCallback(() => setCategoryEditTarget(null), []);
  const handleNameClose = useCallback(() => setNameEditTarget(null), []);

  const handleCategorySave = useCallback(
    (data: DiagnosisCategoryFormData) => {
      if (!data.name.trim()) {
        toast.error("カテゴリ名は必須です");
        return;
      }
      startSaveTransition(() => {
        if (categoryEditTarget !== null && categoryEditTarget !== "new") {
          const req: UpdateDiagnosisCategoryRequest = {
            name: data.name,
            description: data.description || undefined,
            is_active: data.isActive,
          };
          updateCategoryMutation.mutate(
            { id: categoryEditTarget.id, req },
            {
              onSuccess: () => { toast.success("更新しました"); handleCategoryClose(); },
              onError: () => toast.error("更新に失敗しました"),
            },
          );
        } else {
          const req: CreateDiagnosisCategoryRequest = {
            name: data.name,
            description: data.description || undefined,
            is_active: true,
          };
          createCategoryMutation.mutate(req, {
            onSuccess: () => { toast.success("登録しました"); handleCategoryClose(); },
            onError: () => toast.error("登録に失敗しました"),
          });
        }
      });
    },
    [categoryEditTarget, updateCategoryMutation, createCategoryMutation, handleCategoryClose],
  );

  const handleCategoryDeleteConfirm = useCallback(() => {
    if (!categoryPendingDelete) return;
    deleteCategoryMutation.mutate(categoryPendingDelete.id, {
      onSuccess: () => {
        setCategoryPendingDelete(null);
        handleCategoryClose();
        toast.success("削除しました");
      },
      onError: () => toast.error("削除に失敗しました"),
    });
  }, [categoryPendingDelete, deleteCategoryMutation, handleCategoryClose]);

  const handleNameSave = useCallback(
    (data: DiagnosisNameFormData) => {
      if (!data.name.trim()) {
        toast.error("診断病名は必須です");
        return;
      }
      if (!data.diagnosisCategoryId) {
        toast.error("カテゴリは必須です");
        return;
      }
      startSaveTransition(() => {
        if (nameEditTarget !== null && nameEditTarget !== "new") {
          const req: UpdateDiagnosisNameRequest = {
            name: data.name,
            diagnosis_category_id: Number(data.diagnosisCategoryId),
            description: data.description || undefined,
            is_active: data.isActive,
          };
          updateNameMutation.mutate(
            { id: nameEditTarget.id, req },
            {
              onSuccess: () => { toast.success("更新しました"); handleNameClose(); },
              onError: () => toast.error("更新に失敗しました"),
            },
          );
        } else {
          const req: CreateDiagnosisNameRequest = {
            name: data.name,
            diagnosis_category_id: Number(data.diagnosisCategoryId),
            description: data.description || undefined,
            is_active: true,
          };
          createNameMutation.mutate(req, {
            onSuccess: () => { toast.success("登録しました"); handleNameClose(); },
            onError: () => toast.error("登録に失敗しました"),
          });
        }
      });
    },
    [nameEditTarget, updateNameMutation, createNameMutation, handleNameClose],
  );

  const handleNameDeleteConfirm = useCallback(() => {
    if (!namePendingDelete) return;
    deleteNameMutation.mutate(namePendingDelete.id, {
      onSuccess: () => {
        setNamePendingDelete(null);
        handleNameClose();
        toast.success("削除しました");
      },
      onError: () => toast.error("削除に失敗しました"),
    });
  }, [namePendingDelete, deleteNameMutation, handleNameClose]);

  const categoryPanelItem = categoryEditTarget !== null && categoryEditTarget !== "new" ? categoryEditTarget : null;
  const namePanelItem = nameEditTarget !== null && nameEditTarget !== "new" ? nameEditTarget : null;

  return (
    <>
      <div className="flex h-full">
        <div className="flex-1 min-w-0">
          <PageLayout
            title="診断病名マスタ"
            icon={<ClipboardList className="size-5 text-[#37352F]" />}
            onBack={() => navigate(paths.settings.getHref())}
            maxWidth="max-w-full"
            headerAction={
              <PrimaryButton onClick={() => {
                if (activeTab === "diagnosis_category") setCategoryEditTarget("new");
                else setNameEditTarget("new");
              }}>
                <Plus className="mr-1.5 size-4" />
                新規登録
              </PrimaryButton>
            }
          >
            <Tabs
              value={activeTab}
              onValueChange={handleTabChange}
              className="flex flex-col gap-4"
            >
              <TabsList
                className={`flex h-9 border-b ${C.borderLight} gap-0`}
              >
                {TABS.map((tab) => (
                  <TabsTrigger
                    key={tab.value}
                    value={tab.value}
                    className={`h-9 border-b-2 border-b-transparent px-4 text-sm ${C.text60} outline-none transition-colors cursor-pointer
                      data-[state=active]:border-b-[#37352F] data-[state=active]:text-[#37352F] data-[state=active]:font-medium`}
                  >
                    {tab.label}
                  </TabsTrigger>
                ))}
              </TabsList>
              <TabsContent value="diagnosis_category" className="mt-4">
                <DiagnosisCategoryTab
                  editTarget={categoryEditTarget}
                  onEditTargetChange={setCategoryEditTarget}
                />
              </TabsContent>
              <TabsContent value="diagnosis_name" className="mt-4">
                <DiagnosisNameTab
                  editTarget={nameEditTarget}
                  onEditTargetChange={setNameEditTarget}
                />
              </TabsContent>
            </Tabs>
          </PageLayout>
        </div>

        {activeTab === "diagnosis_category" && categoryEditTarget !== null ? (
          <DiagnosisCategorySidePanel
            key={categoryPanelItem ? String(categoryPanelItem.id) : "new-category"}
            item={categoryPanelItem}
            onClose={handleCategoryClose}
            onSave={handleCategorySave}
            onDeleteRequest={setCategoryPendingDelete}
          />
        ) : null}
        {activeTab === "diagnosis_name" && nameEditTarget !== null ? (
          <DiagnosisNameSidePanel
            key={namePanelItem ? String(namePanelItem.id) : "new-name"}
            item={namePanelItem}
            categories={rawCategories ?? []}
            onClose={handleNameClose}
            onSave={handleNameSave}
            onDeleteRequest={setNamePendingDelete}
          />
        ) : null}
      </div>

      <ConfirmDialog
        open={categoryPendingDelete !== null}
        onClose={() => setCategoryPendingDelete(null)}
        title="診断カテゴリを削除しますか？"
        description={`「${categoryPendingDelete?.name}」を削除します。このカテゴリに属する診断名も影響を受けます。この操作は取り消せません。`}
        confirmLabel="削除"
        variant="destructive"
        onConfirm={handleCategoryDeleteConfirm}
      />
      <ConfirmDialog
        open={namePendingDelete !== null}
        onClose={() => setNamePendingDelete(null)}
        title="診断病名を削除しますか？"
        description={`「${namePendingDelete?.name}」を削除します。この操作は取り消せません。`}
        confirmLabel="削除"
        variant="destructive"
        onConfirm={handleNameDeleteConfirm}
      />
    </>
  );
}
