// React/Framework
import { useState, useRef, useMemo, useCallback, memo, useDeferredValue, useEffect } from "react";
import { useNavigate, useSearchParams } from "react-router";
import { paths } from "@/config/paths";
import { useMasterCRUD } from "../hooks/use-master-crud";
import { useMasterSave } from "../hooks/use-master-save";

// DnD
import { DndContext, closestCenter } from "@dnd-kit/core";
import { SortableContext, verticalListSortingStrategy } from "@dnd-kit/sortable";

// Shared hooks
import { useSortableList } from "@/hooks/use-sortable-list";
import { useSidePeekDirty } from "@/hooks/use-side-peek-dirty";

// External
import Plus from "lucide-react/dist/esm/icons/plus";
import FolderTree from "lucide-react/dist/esm/icons/folder-tree";
import ClipboardList from "lucide-react/dist/esm/icons/clipboard-list";

// Internal
import { TableCell } from "@/components/ui/table";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { UnifiedTabs, UnifiedTabsContent } from "@/components/shared/UnifiedTabs";
import { PageLayout } from "@/components/shared/PageLayout/PageLayout";
import { NotionFilter } from "@/components/shared/NotionFilter/NotionFilter";
import { DataTable } from "@/components/shared/DataTable/DataTable";
import { SortableDataTableRow } from "@/components/shared/DataTable/SortableDataTableRow";
import { RowActionButton } from "@/components/shared/RowActionButton";
import { ConfirmDialog } from "@/components/shared/ConfirmDialog/ConfirmDialog";
import { NotionStatusPill } from "@/components/shared/StatusPill/NotionStatusPill";
import { PropertyRow, StatusToggleButton, PropertyInput, MasterSidePanel } from "@/components/shared/SidePeek";
import { FormFieldError } from "@/components/shared/FormFieldError";
import { PrimaryButton } from "@/components/shared/Form/PrimaryButton";
import { C, STYLE, LAYOUT, ICON } from "@/lib/design-tokens";
import { useGetDiagnosisTypes, useCreateDiagnosisType, useUpdateDiagnosisType, useDeleteDiagnosisType, useReorderDiagnosisTypes, useGetDiagnosisNames, useCreateDiagnosisName, useUpdateDiagnosisName, useDeleteDiagnosisName, useReorderDiagnosisNames } from "../api/diagnosis";

// Types
import type { DiagnosisType, DiagnosisName } from "../api/diagnosis";
import type {
  CreateDiagnosisTypeRequest,
  UpdateDiagnosisTypeRequest,
  CreateDiagnosisNameRequest,
  UpdateDiagnosisNameRequest,
} from "@/types/diagnosis";
import { ResourceMasterMedical } from "@/types/generated/models";
import { usePermission } from "@/hooks/use-permission";

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
  { value: "diagnosis_type", label: "診断病名カテゴリ" },
  { value: "diagnosis_name", label: "診断病名" },
] as const;

// ─────────────────────────────────────────────────
// Form state types
// ─────────────────────────────────────────────────

interface DiagnosisTypeFormData {
  name: string;
  description: string;
  isActive: boolean;
}

interface DiagnosisNameFormData {
  name: string;
  diagnosisTypeId: string;
  description: string;
  isActive: boolean;
}

// ─────────────────────────────────────────────────
// DiagnosisTypeSidePanel
// ─────────────────────────────────────────────────

interface DiagnosisTypeSidePanelProps {
  item: DiagnosisType | null;
  onClose: () => void;
  onSave: (data: DiagnosisTypeFormData) => void;
  onDeleteRequest?: (item: DiagnosisType) => void;
  readOnly?: boolean;
  onDirtyChange?: (dirty: boolean) => void;
}

const DiagnosisTypeSidePanel = memo(function DiagnosisTypeSidePanel({
  item,
  onClose,
  onSave,
  onDeleteRequest,
  readOnly,
  onDirtyChange,
}: DiagnosisTypeSidePanelProps) {
  const [formData, setFormData] = useState<DiagnosisTypeFormData>(() => ({
    name: item?.name ?? "",
    description: item?.description ?? "",
    isActive: item?.isActive ?? true,
  }));
  const [isDirty, setIsDirty] = useState(false);
  const [nameError, setNameError] = useState("");

  // BUG-380
  useEffect(() => { onDirtyChange?.(isDirty); }, [isDirty, onDirtyChange]);
  const setFormDataDirty = useCallback<typeof setFormData>((updater) => {
    setFormData(updater);
    setIsDirty(true);
  }, []);

  // rerender-dependencies: useRef でオブジェクト deps を回避
  const formDataRef = useRef(formData);
  useEffect(() => { formDataRef.current = formData; }, [formData]);

  const handleTitleChange = useCallback((v: string) => {
    setFormDataDirty((prev) => ({ ...prev, name: v }));
    if (v.trim()) setNameError("");
  }, [setFormDataDirty]);

  const handleDescriptionChange = useCallback((v: string) => {
    setFormDataDirty((prev) => ({ ...prev, description: v }));
  }, [setFormDataDirty]);

  const handleToggleActive = useCallback(() => {
    setFormDataDirty((prev) => ({ ...prev, isActive: !prev.isActive }));
  }, [setFormDataDirty]);

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
  categories: DiagnosisType[];
  onClose: () => void;
  onSave: (data: DiagnosisNameFormData) => void;
  onDeleteRequest?: (item: DiagnosisName) => void;
  readOnly?: boolean;
  onDirtyChange?: (dirty: boolean) => void;
}

const DiagnosisNameSidePanel = memo(function DiagnosisNameSidePanel({
  item,
  categories,
  onClose,
  onSave,
  onDeleteRequest,
  readOnly,
  onDirtyChange,
}: DiagnosisNameSidePanelProps) {
  const [formData, setFormData] = useState<DiagnosisNameFormData>(() => ({
    name: item?.name ?? "",
    diagnosisTypeId: item
      ? String(item.diagnosisTypeId)
      : categories[0]
        ? String(categories[0].id)
        : "",
    description: item?.description ?? "",
    isActive: item?.isActive ?? true,
  }));
  const [isDirty, setIsDirty] = useState(false);
  const [nameError, setNameError] = useState("");
  const [categoryError, setCategoryError] = useState("");

  // BUG-380
  useEffect(() => { onDirtyChange?.(isDirty); }, [isDirty, onDirtyChange]);
  const setFormDataDirty = useCallback<typeof setFormData>((updater) => {
    setFormData(updater);
    setIsDirty(true);
  }, []);

  // rerender-dependencies: useRef でオブジェクト deps を回避
  const formDataRef = useRef(formData);
  useEffect(() => { formDataRef.current = formData; }, [formData]);

  const handleTitleChange = useCallback((v: string) => {
    setFormDataDirty((prev) => ({ ...prev, name: v }));
    if (v.trim()) setNameError("");
  }, [setFormDataDirty]);

  const handleCategoryChange = useCallback((v: string) => {
    setFormDataDirty((prev) => ({ ...prev, diagnosisTypeId: v }));
    if (v) setCategoryError("");
  }, [setFormDataDirty]);

  const handleDescriptionChange = useCallback((v: string) => {
    setFormDataDirty((prev) => ({ ...prev, description: v }));
  }, [setFormDataDirty]);

  const handleToggleActive = useCallback(() => {
    setFormDataDirty((prev) => ({ ...prev, isActive: !prev.isActive }));
  }, [setFormDataDirty]);

  const handleAction = useCallback(() => {
    const current = formDataRef.current;
    if (!current.name.trim()) {
      setNameError("診断病名を入力してください");
      return;
    }
    if (!current.diagnosisTypeId) {
      setCategoryError("カテゴリを選択してください");
      return;
    }
    setNameError("");
    setCategoryError("");
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
          value={formData.diagnosisTypeId}
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
      <FormFieldError message={categoryError} />
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
// DiagnosisTypeTab
// ─────────────────────────────────────────────────

interface DiagnosisTypeTabProps {
  editTarget: DiagnosisType | "new" | null;
  onEditTargetChange: (v: DiagnosisType | "new" | null) => void;
  canEdit: boolean;
}

function DiagnosisTypeTab({ editTarget: _editTarget, onEditTargetChange, canEdit }: DiagnosisTypeTabProps) {
  const [searchTerm, setSearchTerm] = useState("");

  const { data: rawCategories } = useGetDiagnosisTypes();
  const reorderMutation = useReorderDiagnosisTypes();

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

  const { data: rawCategories } = useGetDiagnosisTypes();
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
                  {categoryMap.get(item.diagnosisTypeId) ?? "-"}
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
  const activeTab = searchParams.get("tab") ?? "diagnosis_type";

  const { data: rawCategories } = useGetDiagnosisTypes();
  const { data: rawNames } = useGetDiagnosisNames();
  const createCategoryMutation = useCreateDiagnosisType();
  const updateCategoryMutation = useUpdateDiagnosisType();
  const deleteCategoryMutation = useDeleteDiagnosisType();
  const createNameMutation = useCreateDiagnosisName();
  const updateNameMutation = useUpdateDiagnosisName();
  const deleteNameMutation = useDeleteDiagnosisName();

  // BUG-380: タブ間共有の未保存ガード
  const dirty = useSidePeekDirty();
  const handleDirtyChange = useCallback((d: boolean) => { if (d) dirty.markDirty(); else dirty.markClean(); }, [dirty]);

  const catCrud = useMasterCRUD<DiagnosisType>({
    data: rawCategories,
    deleteMutation: deleteCategoryMutation,
    entityLabel: "診断カテゴリ",
    dirtyGuard: dirty,
  });

  const nameCrud = useMasterCRUD<DiagnosisName>({
    data: rawNames,
    deleteMutation: deleteNameMutation,
    entityLabel: "診断病名",
    dirtyGuard: dirty,
  });

  // rerender-dependencies: destructure methods to avoid object reference instability in useCallback deps
  const catSetEditTarget = catCrud.setEditTarget;
  const nameSetEditTarget = nameCrud.setEditTarget;

  const tabItems = TABS;

  const handleTabChange = useCallback((tab: string) => {
    // BUG-380: タブ切替前に未保存破棄を確認
    if (!dirty.confirmDiscard()) return;
    setSearchParams({ tab });
    catSetEditTarget(null);
    nameSetEditTarget(null);
  }, [setSearchParams, catSetEditTarget, nameSetEditTarget, dirty]);

  const catSave = useMasterSave({
    crud: catCrud,
    createMutation: createCategoryMutation,
    updateMutation: updateCategoryMutation,
    validate: (data: DiagnosisTypeFormData) => data.name.trim() ? null : "名称を入力してください",
    toCreateRequest: (data: DiagnosisTypeFormData): CreateDiagnosisTypeRequest => ({
      name: data.name,
      description: data.description || undefined,
      is_active: true,
    }),
    toUpdateRequest: (data: DiagnosisTypeFormData): UpdateDiagnosisTypeRequest => ({
      name: data.name,
      description: data.description || undefined,
      is_active: data.isActive,
    }),
  });

  const nameSave = useMasterSave({
    crud: nameCrud,
    createMutation: createNameMutation,
    updateMutation: updateNameMutation,
    validate: (data: DiagnosisNameFormData) => {
      if (!data.name.trim()) return "診断病名を入力してください";
      if (!data.diagnosisTypeId) return "カテゴリを選択してください";
      return null;
    },
    toCreateRequest: (data: DiagnosisNameFormData): CreateDiagnosisNameRequest => ({
      name: data.name,
      diagnosis_type_id: Number(data.diagnosisTypeId),
      description: data.description || undefined,
      is_active: true,
    }),
    toUpdateRequest: (data: DiagnosisNameFormData): UpdateDiagnosisNameRequest => ({
      name: data.name,
      diagnosis_type_id: Number(data.diagnosisTypeId),
      description: data.description || undefined,
      is_active: data.isActive,
    }),
  });

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
                  if (activeTab === "diagnosis_type") catCrud.handleNew();
                  else nameCrud.handleNew();
                }}>
                  <Plus className={`mr-1.5 ${ICON.action}`} />
                  新規登録
                </PrimaryButton>
              ) : null
            }
          >
            <UnifiedTabs
              items={tabItems}
              value={activeTab}
              onValueChange={handleTabChange}
              className="flex flex-col gap-4"
            >
              <UnifiedTabsContent value="diagnosis_type" className="mt-4">
                <DiagnosisTypeTab
                  editTarget={catCrud.editTarget}
                  onEditTargetChange={catCrud.setEditTarget}
                  canEdit={canEdit}
                />
              </UnifiedTabsContent>
              <UnifiedTabsContent value="diagnosis_name" className="mt-4">
                <DiagnosisNameTab
                  editTarget={nameCrud.editTarget}
                  onEditTargetChange={nameCrud.setEditTarget}
                  canEdit={canEdit}
                />
              </UnifiedTabsContent>
            </UnifiedTabs>
          </PageLayout>
        </div>

        {activeTab === "diagnosis_type" && catCrud.isEditing ? (
          <DiagnosisTypeSidePanel
            key={catCrud.panelItem ? String(catCrud.panelItem.id) : "new-category"}
            item={catCrud.panelItem}
            onClose={catCrud.handleClose}
            onSave={catSave.handleSave}
            onDeleteRequest={canDelete ? catCrud.setPendingDelete : undefined}
            readOnly={!canEdit}
            onDirtyChange={handleDirtyChange}
          />
        ) : null}
        {activeTab === "diagnosis_name" && nameCrud.isEditing ? (
          <DiagnosisNameSidePanel
            key={nameCrud.panelItem ? String(nameCrud.panelItem.id) : "new-name"}
            item={nameCrud.panelItem}
            categories={rawCategories ?? []}
            onClose={nameCrud.handleClose}
            onSave={nameSave.handleSave}
            onDeleteRequest={canDelete ? nameCrud.setPendingDelete : undefined}
            readOnly={!canEdit}
            onDirtyChange={handleDirtyChange}
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
