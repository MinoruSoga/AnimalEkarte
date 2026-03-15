// React/Framework
import { useState, useMemo } from "react";
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
import {
  Plus,
  FolderTree,
  ClipboardList,
} from "lucide-react";
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
import { PageLayout } from "@/components/shared/PageLayout";
import { SearchFilterBar } from "@/components/shared/SearchFilterBar";
import { DataTable } from "@/components/shared/DataTable";
import { SortableDataTableRow } from "@/components/shared/DataTable/SortableDataTableRow";
import { RowActionButton } from "@/components/shared/RowActionButton";
import { ConfirmDialog } from "@/components/shared/ConfirmDialog/ConfirmDialog";
import { NotionStatusPill } from "@/components/shared/StatusPill/NotionStatusPill";
import {
  PropertyRow,
  PropInput,
  SidePeekPanel,
  SidePeekToolbar,
  SidePeekBody,
  SidePeekTitleInput,
  SidePeekFooter,
} from "@/components/shared/SidePeek";
import { C, STYLE, LAYOUT } from "@/lib/design-tokens";
import {
  useListDiagnosisCategories,
  useCreateDiagnosisCategory,
  useUpdateDiagnosisCategory,
  useDeleteDiagnosisCategory,
  useReorderDiagnosisCategories,
  useListDiagnosisNames,
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
// DiagnosisCategorySidePanel
// ─────────────────────────────────────────────────

function DiagnosisCategorySidePanel({
  item,
  onClose,
  onSave,
  onDeleteRequest,
}: {
  item: DiagnosisCategory | null;
  onClose: () => void;
  onSave: (data: { name: string; description: string; isActive: boolean }) => void;
  onDeleteRequest: () => void;
}) {
  const [formData, setFormData] = useState({
    name: item?.name ?? "",
    description: item?.description ?? "",
    isActive: item?.isActive ?? true,
  });

  return (
    <SidePeekPanel>
      <SidePeekToolbar
        isNew={item === null}
        onClose={onClose}
        onDelete={item !== null ? onDeleteRequest : undefined}
      />
      <SidePeekBody>
        <div className="pt-4 pb-2">
          <div className={STYLE.pageIcon}>
            <FolderTree className={LAYOUT.pageIcon.innerIcon} />
          </div>
        </div>
        <SidePeekTitleInput
          value={formData.name}
          onChange={(v) => setFormData({ ...formData, name: v })}
        />
        <div className={`${STYLE.sectionDivider} mb-1`} />
        <div className="py-1">
          <PropertyRow label="ステータス">
            <button
              type="button"
              onClick={() => setFormData({ ...formData, isActive: !formData.isActive })}
              className={`inline-flex items-center rounded-[3px] ${C.hoverBgLight} transition-colors py-0.5 px-0.5 cursor-pointer`}
            >
              <NotionStatusPill isActive={formData.isActive} />
            </button>
          </PropertyRow>
          <PropertyRow label="備考">
            <PropInput
              value={formData.description}
              onChange={(v) => setFormData({ ...formData, description: v })}
              placeholder="補足情報など"
            />
          </PropertyRow>
        </div>
      </SidePeekBody>
      <SidePeekFooter onCancel={onClose} onSave={() => onSave(formData)} />
    </SidePeekPanel>
  );
}

// ─────────────────────────────────────────────────
// DiagnosisNameSidePanel
// ─────────────────────────────────────────────────

function DiagnosisNameSidePanel({
  item,
  categories,
  onClose,
  onSave,
  onDeleteRequest,
}: {
  item: DiagnosisName | null;
  categories: DiagnosisCategory[];
  onClose: () => void;
  onSave: (data: { name: string; diagnosisCategoryId: string; description: string; isActive: boolean }) => void;
  onDeleteRequest: () => void;
}) {
  const [formData, setFormData] = useState({
    name: item?.name ?? "",
    diagnosisCategoryId: item
      ? String(item.diagnosisCategoryId)
      : categories[0]
        ? String(categories[0].id)
        : "",
    description: item?.description ?? "",
    isActive: item?.isActive ?? true,
  });

  return (
    <SidePeekPanel>
      <SidePeekToolbar
        isNew={item === null}
        onClose={onClose}
        onDelete={item !== null ? onDeleteRequest : undefined}
      />
      <SidePeekBody>
        <div className="pt-4 pb-2">
          <div className={STYLE.pageIcon}>
            <FolderTree className={LAYOUT.pageIcon.innerIcon} />
          </div>
        </div>
        <SidePeekTitleInput
          value={formData.name}
          onChange={(v) => setFormData({ ...formData, name: v })}
        />
        <div className={`${STYLE.sectionDivider} mb-1`} />
        <div className="py-1">
          <PropertyRow label="ステータス">
            <button
              type="button"
              onClick={() => setFormData({ ...formData, isActive: !formData.isActive })}
              className={`inline-flex items-center rounded-[3px] ${C.hoverBgLight} transition-colors py-0.5 px-0.5 cursor-pointer`}
            >
              <NotionStatusPill isActive={formData.isActive} />
            </button>
          </PropertyRow>
          <PropertyRow label="カテゴリ">
            <Select
              value={formData.diagnosisCategoryId}
              onValueChange={(v) => setFormData({ ...formData, diagnosisCategoryId: v })}
            >
              <SelectTrigger className={STYLE.selectCompact}>
                <SelectValue placeholder="カテゴリを選択" />
              </SelectTrigger>
              <SelectContent>
                {categories.map((cat) => (
                  <SelectItem key={cat.id} value={String(cat.id)}>
                    {cat.name}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
          </PropertyRow>
          <PropertyRow label="備考">
            <PropInput
              value={formData.description}
              onChange={(v) => setFormData({ ...formData, description: v })}
              placeholder="補足情報など"
            />
          </PropertyRow>
        </div>
      </SidePeekBody>
      <SidePeekFooter onCancel={onClose} onSave={() => onSave(formData)} />
    </SidePeekPanel>
  );
}

// ─────────────────────────────────────────────────
// DiagnosisCategoryTab
// ─────────────────────────────────────────────────

function DiagnosisCategoryTab() {
  const [selectedItem, setSelectedItem] = useState<DiagnosisCategory | null>(null);
  const [isEditing, setIsEditing] = useState(false);
  const [searchTerm, setSearchTerm] = useState("");
  const [pendingDelete, setPendingDelete] = useState<DiagnosisCategory | null>(null);

  const { data: rawCategories } = useListDiagnosisCategories();
  const createMutation = useCreateDiagnosisCategory();
  const updateMutation = useUpdateDiagnosisCategory();
  const deleteMutation = useDeleteDiagnosisCategory();
  const reorderMutation = useReorderDiagnosisCategories();

  const { orderedItems: orderedCategories, sensors, handleDragEnd: handleCategoryDragEnd } =
    useSortableList({
      items: rawCategories ?? [],
      onReorder: (newIds) => {
        reorderMutation.mutate({ ids: newIds.map(Number) });
      },
    });

  const filteredItems = useMemo(() => {
    if (!searchTerm) return orderedCategories;
    const lower = searchTerm.toLowerCase();
    return orderedCategories.filter((c) => c.name.toLowerCase().includes(lower));
  }, [orderedCategories, searchTerm]);

  const handleEdit = (item: DiagnosisCategory) => {
    setSelectedItem(item);
    setIsEditing(true);
  };

  const handleCreate = () => {
    setSelectedItem(null);
    setIsEditing(true);
  };

  const handleClose = () => {
    setIsEditing(false);
    setSelectedItem(null);
  };

  const handleSave = (data: { name: string; description: string; isActive: boolean }) => {
    if (!data.name.trim()) {
      toast.error("カテゴリ名は必須です");
      return;
    }
    if (selectedItem) {
      const req: UpdateDiagnosisCategoryRequest = {
        name: data.name,
        description: data.description || undefined,
        is_active: data.isActive,
      };
      updateMutation.mutate(
        { id: selectedItem.id, req },
        {
          onSuccess: () => { toast.success("更新しました"); handleClose(); },
          onError: () => toast.error("更新に失敗しました"),
        },
      );
    } else {
      const req: CreateDiagnosisCategoryRequest = {
        name: data.name,
        description: data.description || undefined,
        is_active: true,
      };
      createMutation.mutate(req, {
        onSuccess: () => { toast.success("登録しました"); handleClose(); },
        onError: () => toast.error("登録に失敗しました"),
      });
    }
  };

  const handleDeleteConfirm = () => {
    if (!pendingDelete) return;
    deleteMutation.mutate(pendingDelete.id, {
      onSuccess: () => {
        setPendingDelete(null);
        handleClose();
        toast.success("削除しました");
      },
      onError: () => toast.error("削除に失敗しました"),
    });
  };

  return (
    <>
      <div className="flex h-full">
        {/* Table area */}
        <div className="flex flex-col gap-4 flex-1 min-w-0">
          <div className="flex items-center gap-3">
            <div className="flex-1 min-w-0">
              <SearchFilterBar
                searchTerm={searchTerm}
                onSearchChange={setSearchTerm}
                placeholder="カテゴリ名で検索..."
                count={filteredItems.length}
              />
            </div>
            <button
              type="button"
              onClick={handleCreate}
              className="inline-flex items-center gap-1 text-sm font-medium text-[#2383E2] hover:text-[#1B6EC2] cursor-pointer transition-colors"
            >
              <Plus className="size-4" />
              新規登録
            </button>
          </div>

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
                    onClick={() => handleEdit(item)}
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
                      <RowActionButton onClick={() => handleEdit(item)} />
                    </TableCell>
                  </SortableDataTableRow>
                )}
              />
            </SortableContext>
          </DndContext>
        </div>

        {/* Side peek */}
        {isEditing ? (
          <DiagnosisCategorySidePanel
            key={selectedItem ? String(selectedItem.id) : "new-category"}
            item={selectedItem}
            onClose={handleClose}
            onSave={handleSave}
            onDeleteRequest={() => setPendingDelete(selectedItem)}
          />
        ) : null}
      </div>

      <ConfirmDialog
        open={pendingDelete !== null}
        onClose={() => setPendingDelete(null)}
        title="診断カテゴリを削除しますか？"
        description={`「${pendingDelete?.name}」を削除します。このカテゴリに属する診断名も影響を受けます。この操作は取り消せません。`}
        confirmLabel="削除"
        variant="destructive"
        onConfirm={handleDeleteConfirm}
      />
    </>
  );
}

// ─────────────────────────────────────────────────
// DiagnosisNameTab
// ─────────────────────────────────────────────────

function DiagnosisNameTab() {
  const [selectedItem, setSelectedItem] = useState<DiagnosisName | null>(null);
  const [isEditing, setIsEditing] = useState(false);
  const [searchTerm, setSearchTerm] = useState("");
  const [pendingDelete, setPendingDelete] = useState<DiagnosisName | null>(null);

  const { data: rawCategories } = useListDiagnosisCategories();
  const { data: rawNames } = useListDiagnosisNames();

  const createMutation = useCreateDiagnosisName();
  const updateMutation = useUpdateDiagnosisName();
  const deleteMutation = useDeleteDiagnosisName();
  const reorderMutation = useReorderDiagnosisNames();

  const { orderedItems: orderedNames, sensors, handleDragEnd: handleNameDragEnd } =
    useSortableList({
      items: rawNames ?? [],
      onReorder: (newIds) => {
        reorderMutation.mutate({ ids: newIds.map(Number) });
      },
    });

  const filteredItems = useMemo(() => {
    if (!searchTerm) return orderedNames;
    const lower = searchTerm.toLowerCase();
    return orderedNames.filter((n) => n.name.toLowerCase().includes(lower));
  }, [orderedNames, searchTerm]);

  const categoryMap = useMemo(
    () => new Map<string, string>((rawCategories ?? []).map((c) => [c.id, c.name])),
    [rawCategories],
  );

  const handleEdit = (item: DiagnosisName) => {
    setSelectedItem(item);
    setIsEditing(true);
  };

  const handleCreate = () => {
    setSelectedItem(null);
    setIsEditing(true);
  };

  const handleClose = () => {
    setIsEditing(false);
    setSelectedItem(null);
  };

  const handleSave = (data: {
    name: string;
    diagnosisCategoryId: string;
    description: string;
    isActive: boolean;
  }) => {
    if (!data.name.trim()) {
      toast.error("診断病名は必須です");
      return;
    }
    if (!data.diagnosisCategoryId) {
      toast.error("カテゴリは必須です");
      return;
    }

    if (selectedItem) {
      const req: UpdateDiagnosisNameRequest = {
        name: data.name,
        diagnosis_category_id: Number(data.diagnosisCategoryId),
        description: data.description || undefined,
        is_active: data.isActive,
      };
      updateMutation.mutate(
        { id: selectedItem.id, req },
        {
          onSuccess: () => { toast.success("更新しました"); handleClose(); },
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
      createMutation.mutate(req, {
        onSuccess: () => { toast.success("登録しました"); handleClose(); },
        onError: () => toast.error("登録に失敗しました"),
      });
    }
  };

  const handleDeleteConfirm = () => {
    if (!pendingDelete) return;
    deleteMutation.mutate(pendingDelete.id, {
      onSuccess: () => {
        setPendingDelete(null);
        handleClose();
        toast.success("削除しました");
      },
      onError: () => toast.error("削除に失敗しました"),
    });
  };

  return (
    <>
      <div className="flex h-full">
        {/* Table area */}
        <div className="flex flex-col gap-4 flex-1 min-w-0">
          <div className="flex items-center gap-3">
            <div className="flex-1 min-w-0">
              <SearchFilterBar
                searchTerm={searchTerm}
                onSearchChange={setSearchTerm}
                placeholder="診断病名で検索..."
                count={filteredItems.length}
              />
            </div>
            <button
              type="button"
              onClick={handleCreate}
              className="inline-flex items-center gap-1 text-sm font-medium text-[#2383E2] hover:text-[#1B6EC2] cursor-pointer transition-colors"
            >
              <Plus className="size-4" />
              新規登録
            </button>
          </div>

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
                    onClick={() => handleEdit(item)}
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
                      <RowActionButton onClick={() => handleEdit(item)} />
                    </TableCell>
                  </SortableDataTableRow>
                )}
              />
            </SortableContext>
          </DndContext>
        </div>

        {/* Side peek */}
        {isEditing ? (
          <DiagnosisNameSidePanel
            key={selectedItem ? String(selectedItem.id) : "new-name"}
            item={selectedItem}
            categories={rawCategories ?? []}
            onClose={handleClose}
            onSave={handleSave}
            onDeleteRequest={() => setPendingDelete(selectedItem)}
          />
        ) : null}
      </div>

      <ConfirmDialog
        open={pendingDelete !== null}
        onClose={() => setPendingDelete(null)}
        title="診断病名を削除しますか？"
        description={`「${pendingDelete?.name}」を削除します。この操作は取り消せません。`}
        confirmLabel="削除"
        variant="destructive"
        onConfirm={handleDeleteConfirm}
      />
    </>
  );
}

// ─────────────────────────────────────────────────
// DiagnosisSettings (main page)
// ─────────────────────────────────────────────────

export function DiagnosisSettings() {
  const navigate = useNavigate();
  const [searchParams, setSearchParams] = useSearchParams();
  const activeTab = searchParams.get("tab") ?? "diagnosis_category";

  const handleTabChange = (tab: string) => {
    setSearchParams({ tab });
  };

  return (
    <PageLayout
      title="診断病名マスタ"
      icon={<ClipboardList className="size-5 text-[#37352F]" />}
      onBack={() => navigate(paths.settings.getHref())}
      maxWidth="max-w-full"
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
          <DiagnosisCategoryTab />
        </TabsContent>
        <TabsContent value="diagnosis_name" className="mt-4">
          <DiagnosisNameTab />
        </TabsContent>
      </Tabs>
    </PageLayout>
  );
}
