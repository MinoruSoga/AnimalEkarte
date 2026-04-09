// React/Framework
import { useState, useRef, useMemo, useCallback, memo, useDeferredValue, useEffect } from "react";
import { useNavigate, useSearchParams } from "react-router";
import { paths } from "@/config/paths";
import { useMasterCRUD } from "@/features/master/hooks/use-master-crud";

// DnD
import { DndContext, closestCenter } from "@dnd-kit/core";
import { SortableContext, verticalListSortingStrategy } from "@dnd-kit/sortable";

// Shared hooks
import { useSortableList } from "@/hooks/use-sortable-list";

// External
import Plus from "lucide-react/dist/esm/icons/plus";
import FolderTree from "lucide-react/dist/esm/icons/folder-tree";
import ClipboardList from "lucide-react/dist/esm/icons/clipboard-list";
import { toast } from "sonner";
import { handleApiError } from "@/lib/handle-api-error";

// Internal
import { TableCell } from "@/components/ui/table";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import * as TabsPrimitive from "@radix-ui/react-tabs";
import { PageLayout } from "@/components/shared/PageLayout/PageLayout";
import { NotionFilter } from "@/components/shared/NotionFilter/NotionFilter";
import { DataTable } from "@/components/shared/DataTable/DataTable";
import { SortableDataTableRow } from "@/components/shared/DataTable/SortableDataTableRow";
import { RowActionButton } from "@/components/shared/RowActionButton";
import { ConfirmDialog } from "@/components/shared/ConfirmDialog/ConfirmDialog";
import { NotionStatusPill } from "@/components/shared/StatusPill/NotionStatusPill";
import { PropertyRow } from "@/components/shared/SidePeek/PropertyRow";
import { StatusToggleButton } from "@/components/shared/SidePeek/StatusToggleButton";
import { PropertyInput } from "@/components/shared/SidePeek/PropertyInput";
import { MasterSidePanel } from "@/components/shared/SidePeek/MasterSidePanel";
import { PrimaryButton } from "@/components/shared/Form/PrimaryButton";
import { C, STYLE, LAYOUT, ICON } from "@/lib/design-tokens";
import { useGetDiagnosisCategories, useCreateDiagnosisCategory, useUpdateDiagnosisCategory, useDeleteDiagnosisCategory, useReorderDiagnosisCategories, useGetDiagnosisNames, useCreateDiagnosisName, useUpdateDiagnosisName, useDeleteDiagnosisName, useReorderDiagnosisNames } from "@/features/master/api/diagnosis";

// Types
import type { DiagnosisCategory, DiagnosisName } from "@/features/master/api/diagnosis";
import type {
  CreateDiagnosisCategoryRequest,
  UpdateDiagnosisCategoryRequest,
  CreateDiagnosisNameRequest,
  UpdateDiagnosisNameRequest,
} from "@/types/diagnosis";
import { ResourceMasterMedical } from "@/types/generated/models";
import { usePermission } from "@/features/auth";

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
  onDeleteRequest?: (item: DiagnosisCategory) => void;
  readOnly?: boolean;
}

const DiagnosisCategorySidePanel = memo(function DiagnosisCategorySidePanel({
  item,
  onClose,
  onSave,
  onDeleteRequest,
  readOnly,
}: DiagnosisCategorySidePanelProps) {
  const [formData, setFormData] = useState<DiagnosisCategoryFormData>(() => ({
    name: item?.name ?? "",
    description: item?.description ?? "",
    isActive: item?.isActive ?? true,
  }));
  const [isDirty, setIsDirty] = useState(false);
  const [nameError, setNameError] = useState("");

  // rerender-dependencies: useRef でオブジェクト deps を回避
  const formDataRef = useRef(formData);
  useEffect(() => { formDataRef.current = formData; }, [formData]);

  const handleTitleChange = useCallback((v: string) => {
    setFormData((prev) => ({ ...prev, name: v }));
    setIsDirty(true);
    if (v.trim()) setNameError("");
  }, []);

  const handleDescriptionChange = useCallback((v: string) => {
    setFormData((prev) => ({ ...prev, description: v }));
    setIsDirty(true);
  }, []);

  const handleToggleActive = useCallback(() => {
    setFormData((prev) => ({ ...prev, isActive: !prev.isActive }));
    setIsDirty(true);
  }, []);

  const handleAction = useCallback(() => {
    const current = formDataRef.current;
    if (!current.name.trim()) {
      setNameError("名称を入力してください");
      return;
    }
    setNameError("");
    onSave(current);
    setIsDirty(false);
  }, [onSave]);

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
      icon={<FolderTree className={LAYOUT.pageIcon.innerIcon} />}
      isDirty={isDirty}
      titleError={nameError}
      titleMaxLength={100}
      readOnly={readOnly}
    >
      <StatusToggleButton
        isActive={formData.isActive}
        onToggle={handleToggleActive}
      />
      <PropertyRow label="備考">
        <PropertyInput
          value={formData.description}
          onChange={handleDescriptionChange}
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
  onDeleteRequest?: (item: DiagnosisName) => void;
  readOnly?: boolean;
}

const DiagnosisNameSidePanel = memo(function DiagnosisNameSidePanel({
  item,
  categories,
  onClose,
  onSave,
  onDeleteRequest,
  readOnly,
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
  const [isDirty, setIsDirty] = useState(false);
  const [nameError, setNameError] = useState("");

  // rerender-dependencies: useRef でオブジェクト deps を回避
  const formDataRef = useRef(formData);
  useEffect(() => { formDataRef.current = formData; }, [formData]);

  const handleTitleChange = useCallback((v: string) => {
    setFormData((prev) => ({ ...prev, name: v }));
    setIsDirty(true);
    if (v.trim()) setNameError("");
  }, []);

  const handleCategoryChange = useCallback((v: string) => {
    setFormData((prev) => ({ ...prev, diagnosisCategoryId: v }));
    setIsDirty(true);
  }, []);

  const handleDescriptionChange = useCallback((v: string) => {
    setFormData((prev) => ({ ...prev, description: v }));
    setIsDirty(true);
  }, []);

  const handleToggleActive = useCallback(() => {
    setFormData((prev) => ({ ...prev, isActive: !prev.isActive }));
    setIsDirty(true);
  }, []);

  const handleAction = useCallback(() => {
    const current = formDataRef.current;
    if (!current.name.trim()) {
      setNameError("診断病名を入力してください");
      return;
    }
    setNameError("");
    onSave(current);
    setIsDirty(false);
  }, [onSave]);

  const handleClose = useCallback(() => {
    setIsDirty(false);
    onClose();
  }, [onClose]);

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
      onTitleChange={handleTitleChange}
      onClose={handleClose}
      action={handleAction}
      onDelete={item !== null && onDeleteRequest ? () => onDeleteRequest(item) : undefined}
      icon={<ClipboardList className={LAYOUT.pageIcon.innerIcon} />}
      isDirty={isDirty}
      titleError={nameError}
      titleMaxLength={100}
      readOnly={readOnly}
    >
      <StatusToggleButton
        isActive={formData.isActive}
        onToggle={handleToggleActive}
      />
      <PropertyRow label="カテゴリ">
        <Select
          value={formData.diagnosisCategoryId}
          onValueChange={handleCategoryChange}
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
        <PropertyInput
          value={formData.description}
          onChange={handleDescriptionChange}
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
  canEdit: boolean;
}

function DiagnosisCategoryTab({ editTarget: _editTarget, onEditTargetChange, canEdit }: DiagnosisCategoryTabProps) {
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
      <NotionFilter
          properties={[]}
          activeFilters={[]}
          onFilterChange={() => {}}
        searchTerm={searchTerm}
        onSearchChange={setSearchTerm}
        searchPlaceholder="カテゴリ名で検索..."
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
                onClick={canEdit ? () => onEditTargetChange(item) : undefined}
              >
                <TableCell className={`font-medium text-base ${C.text}`}>
                  {item.name}
                </TableCell>
                <TableCell className={`text-base ${C.text70} truncate max-w-[240px]`}>
                  {item.description || "-"}
                </TableCell>
                <TableCell className="text-center">
                  <NotionStatusPill isActive={item.isActive} />
                </TableCell>
                <TableCell className="p-0 text-right">
                  {canEdit ? <RowActionButton onClick={() => onEditTargetChange(item)} /> : null}
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
  canEdit: boolean;
}

function DiagnosisNameTab({ editTarget: _editTarget, onEditTargetChange, canEdit }: DiagnosisNameTabProps) {
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
      <NotionFilter
          properties={[]}
          activeFilters={[]}
          onFilterChange={() => {}}
        searchTerm={searchTerm}
        onSearchChange={setSearchTerm}
        searchPlaceholder="診断病名で検索..."
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
                onClick={canEdit ? () => onEditTargetChange(item) : undefined}
              >
                <TableCell className={`text-base ${C.text70}`}>
                  {categoryMap.get(item.diagnosisCategoryId) ?? "-"}
                </TableCell>
                <TableCell className={`font-medium text-base ${C.text}`}>
                  {item.name}
                </TableCell>
                <TableCell className="text-center">
                  <NotionStatusPill isActive={item.isActive} />
                </TableCell>
                <TableCell className="p-0 text-right">
                  {canEdit ? <RowActionButton onClick={() => onEditTargetChange(item)} /> : null}
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
  const { canCreate, canEdit, canDelete } = usePermission(ResourceMasterMedical);
  const [searchParams, setSearchParams] = useSearchParams();
  const activeTab = searchParams.get("tab") ?? "diagnosis_category";

  const { data: rawCategories } = useGetDiagnosisCategories();
  const { data: rawNames } = useGetDiagnosisNames();
  const createCategoryMutation = useCreateDiagnosisCategory();
  const updateCategoryMutation = useUpdateDiagnosisCategory();
  const deleteCategoryMutation = useDeleteDiagnosisCategory();
  const createNameMutation = useCreateDiagnosisName();
  const updateNameMutation = useUpdateDiagnosisName();
  const deleteNameMutation = useDeleteDiagnosisName();

  const catCrud = useMasterCRUD<DiagnosisCategory>({
    data: rawCategories,
    deleteMutation: deleteCategoryMutation,
    entityLabel: "診断カテゴリ",
  });

  const nameCrud = useMasterCRUD<DiagnosisName>({
    data: rawNames,
    deleteMutation: deleteNameMutation,
    entityLabel: "診断病名",
  });

  // rerender-dependencies: destructure methods to avoid object reference instability in useCallback deps
  const catSetEditTarget = catCrud.setEditTarget;
  const catHandleClose = catCrud.handleClose;
  const catStartSave = catCrud.startSaveTransition;
  const catEditTarget = catCrud.editTarget;
  const nameSetEditTarget = nameCrud.setEditTarget;
  const nameHandleClose = nameCrud.handleClose;
  const nameStartSave = nameCrud.startSaveTransition;
  const nameEditTarget = nameCrud.editTarget;

  const handleTabChange = useCallback((tab: string) => {
    setSearchParams({ tab });
    catSetEditTarget(null);
    nameSetEditTarget(null);
  }, [setSearchParams, catSetEditTarget, nameSetEditTarget]);

  const handleCategorySave = useCallback(
    (data: DiagnosisCategoryFormData) => {
      if (!data.name.trim()) {
        toast.error("カテゴリ名は必須です");
        return;
      }
      catStartSave(() => {
        if (catEditTarget !== null && catEditTarget !== "new") {
          const req: UpdateDiagnosisCategoryRequest = {
            name: data.name,
            description: data.description || undefined,
            is_active: data.isActive,
          };
          updateCategoryMutation.mutate(
            { id: catEditTarget.id, req },
            {
              onSuccess: () => { toast.success("更新しました"); catHandleClose(); },
              onError: (error) => handleApiError(error, "更新"),
            },
          );
        } else {
          const req: CreateDiagnosisCategoryRequest = {
            name: data.name,
            description: data.description || undefined,
            is_active: true,
          };
          createCategoryMutation.mutate(req, {
            onSuccess: () => { toast.success("登録しました"); catHandleClose(); },
            onError: (error) => handleApiError(error, "登録"),
          });
        }
      });
    },
    [catEditTarget, updateCategoryMutation, createCategoryMutation, catHandleClose, catStartSave],
  );

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
      nameStartSave(() => {
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
              onSuccess: () => { toast.success("更新しました"); nameHandleClose(); },
              onError: (error) => handleApiError(error, "更新"),
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
            onSuccess: () => { toast.success("登録しました"); nameHandleClose(); },
            onError: (error) => handleApiError(error, "登録"),
          });
        }
      });
    },
    [nameEditTarget, updateNameMutation, createNameMutation, nameHandleClose, nameStartSave],
  );

  return (
    <>
      <div className="flex h-full">
        <div className="flex-1 min-w-0">
          <PageLayout
            title="診断病名マスタ"
            icon={<ClipboardList className={`${ICON.page} ${C.text}`} />}
            resource={ResourceMasterMedical}
            onBack={() => navigate(paths.settings.getHref())}
            maxWidth="max-w-full"
            headerAction={
              canCreate ? (
                <PrimaryButton onClick={() => {
                  if (activeTab === "diagnosis_category") catCrud.handleNew();
                  else nameCrud.handleNew();
                }}>
                  <Plus className={`mr-1.5 ${ICON.action}`} />
                  新規登録
                </PrimaryButton>
              ) : null
            }
          >
            <TabsPrimitive.Root
              value={activeTab}
              onValueChange={handleTabChange}
              className="flex flex-col gap-4"
            >
              <TabsPrimitive.List
                className={`flex h-9 border-b ${C.borderLight} gap-0`}
              >
                {TABS.map((tab) => (
                  <TabsPrimitive.Trigger
                    key={tab.value}
                    value={tab.value}
                    className={`h-9 border-b-2 border-b-transparent px-4 text-base ${C.text60} outline-none transition-colors cursor-pointer
                      ${C.dataActiveBorderB} ${C.dataActiveText} data-[state=active]:font-medium`}
                  >
                    {tab.label}
                  </TabsPrimitive.Trigger>
                ))}
              </TabsPrimitive.List>
              <TabsPrimitive.Content value="diagnosis_category" className="mt-4">
                <DiagnosisCategoryTab
                  editTarget={catCrud.editTarget}
                  onEditTargetChange={catCrud.setEditTarget}
                  canEdit={canEdit}
                />
              </TabsPrimitive.Content>
              <TabsPrimitive.Content value="diagnosis_name" className="mt-4">
                <DiagnosisNameTab
                  editTarget={nameCrud.editTarget}
                  onEditTargetChange={nameCrud.setEditTarget}
                  canEdit={canEdit}
                />
              </TabsPrimitive.Content>
            </TabsPrimitive.Root>
          </PageLayout>
        </div>

        {activeTab === "diagnosis_category" && catCrud.isEditing ? (
          <DiagnosisCategorySidePanel
            key={catCrud.panelItem ? String(catCrud.panelItem.id) : "new-category"}
            item={catCrud.panelItem}
            onClose={catCrud.handleClose}
            onSave={handleCategorySave}
            onDeleteRequest={canDelete ? catCrud.setPendingDelete : undefined}
            readOnly={!canEdit}
          />
        ) : null}
        {activeTab === "diagnosis_name" && nameCrud.isEditing ? (
          <DiagnosisNameSidePanel
            key={nameCrud.panelItem ? String(nameCrud.panelItem.id) : "new-name"}
            item={nameCrud.panelItem}
            categories={rawCategories ?? []}
            onClose={nameCrud.handleClose}
            onSave={handleNameSave}
            onDeleteRequest={canDelete ? nameCrud.setPendingDelete : undefined}
            readOnly={!canEdit}
          />
        ) : null}
      </div>

      <ConfirmDialog
        open={catCrud.pendingDelete !== null}
        onClose={catCrud.handleDeleteCancel}
        title="診断カテゴリを削除しますか？"
        description={`「${catCrud.pendingDelete?.name}」を削除します。このカテゴリに属する診断名も影響を受けます。この操作は取り消せません。`}
        confirmLabel="削除"
        variant="destructive"
        onConfirm={catCrud.handleDeleteConfirm}
      />
      <ConfirmDialog
        open={nameCrud.pendingDelete !== null}
        onClose={nameCrud.handleDeleteCancel}
        title="診断病名を削除しますか？"
        description={`「${nameCrud.pendingDelete?.name}」を削除します。この操作は取り消せません。`}
        confirmLabel="削除"
        variant="destructive"
        onConfirm={nameCrud.handleDeleteConfirm}
      />
    </>
  );
}
