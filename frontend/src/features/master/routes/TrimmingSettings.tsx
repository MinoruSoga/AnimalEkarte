// React/Framework
import { useState, useMemo, useCallback, memo, useDeferredValue } from "react";
import { useNavigate } from "react-router";
import { paths } from "@/config/paths";
import { useMasterCRUD } from "@/features/master/hooks/use-master-crud";

// External
import { Plus, Scissors } from "lucide-react";
import { toast } from "sonner";

// Shared
import { TableCell } from "@/components/ui/table";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { PageLayout } from "@/components/shared/PageLayout/PageLayout";
import { NotionFilter } from "@/components/shared/NotionFilter/NotionFilter";
import { MASTER_STATUS_FILTER } from "@/features/master/constants/styles";
import type { ActiveFilter } from "@/components/shared/NotionFilter/types";
import { DataTable } from "@/components/shared/DataTable/DataTable";
import { DataTableRow } from "@/components/shared/DataTable/DataTableRow";
import { ConfirmDialog } from "@/components/shared/ConfirmDialog/ConfirmDialog";
import { NotionStatusPill } from "@/components/shared/StatusPill/NotionStatusPill";
import { PropertyRow } from "@/components/shared/SidePeek/PropertyRow";
import { PropInput } from "@/components/shared/SidePeek/PropInput";
import { MasterSidePanel } from "@/components/shared/SidePeek/MasterSidePanel";
import { PrimaryButton } from "@/components/shared/Form/PrimaryButton";
import { C, STYLE, LAYOUT } from "@/lib/design-tokens";
import {
  useGetTrimmingCourses,
  useCreateTrimmingCourse,
  useUpdateTrimmingCourse,
  useDeleteTrimmingCourse,
  useGetTrimmingOptions,
  useCreateTrimmingOption,
  useUpdateTrimmingOption,
  useDeleteTrimmingOption,
  TARGET_SIZE_LABELS,
  TARGET_SIZE_OPTIONS,
} from "@/features/master/api/trimming";

// Types
import type {
  TrimmingCourse,
  TrimmingOption,
  TargetSize,
  CreateTrimmingCourseRequest,
  UpdateTrimmingCourseRequest,
  CreateTrimmingOptionRequest,
  UpdateTrimmingOptionRequest,
} from "@/features/master/api/trimming";

// ─────────────────────────────────────────────────
// Columns
// ─────────────────────────────────────────────────

const COURSE_COLUMNS = [
  { header: "コース名" },
  { header: "対象サイズ", className: "w-[120px]" },
  { header: "所要時間", className: "w-[100px]" },
  { header: "単価(税込)", className: "w-[110px]", align: "right" as const },
  { header: "ステータス", className: "w-[90px]", align: "right" as const },
];

const OPTION_COLUMNS = [
  { header: "オプション名" },
  { header: "所要時間", className: "w-[100px]" },
  { header: "組合せ可否", className: "w-[110px]", align: "center" as const },
  { header: "単価(税込)", className: "w-[110px]", align: "right" as const },
  { header: "ステータス", className: "w-[90px]", align: "right" as const },
];

// ─────────────────────────────────────────────────
// Hoisted static JSX (rendering-hoist-jsx)
// ─────────────────────────────────────────────────

const TARGET_SIZE_SELECT_ITEMS = [
  <SelectItem key="__none__" value="__none__">指定なし</SelectItem>,
  ...TARGET_SIZE_OPTIONS.map((opt) => (
    <SelectItem key={opt.value} value={opt.value}>
      {opt.label}
    </SelectItem>
  )),
];

// ─────────────────────────────────────────────────
// CombinablePill
// ─────────────────────────────────────────────────

function CombinablePill({ combinable }: { combinable: boolean }) {
  if (combinable) {
    return (
      <span className="inline-flex items-center gap-1 px-2 py-0.5 rounded-sm text-xs bg-[#DDEDEA] text-[#0F7B6C]">
        可
      </span>
    );
  }
  return (
    <span className="inline-flex items-center gap-1 px-2 py-0.5 rounded-sm text-xs bg-[#E3E2E0] text-[#37352F]/60">
      不可
    </span>
  );
}

// ─────────────────────────────────────────────────
// CourseFormData
// ─────────────────────────────────────────────────

interface CourseFormData {
  name: string;
  price: string;
  targetSize: TargetSize | "";
  duration: string;
  description: string;
  isActive: boolean;
}

// ─────────────────────────────────────────────────
// TrimmingCourseSidePanel
// ─────────────────────────────────────────────────

interface TrimmingCourseSidePanelProps {
  item: TrimmingCourse | null;
  onClose: () => void;
  onSave: (data: CourseFormData) => void;
  onDeleteRequest: (item: TrimmingCourse) => void;
}

const TrimmingCourseSidePanel = memo(function TrimmingCourseSidePanel({
  item,
  onClose,
  onSave,
  onDeleteRequest,
}: TrimmingCourseSidePanelProps) {
  const [formData, setFormData] = useState<CourseFormData>(() => ({
    name: item?.name ?? "",
    price: item?.price != null ? String(item.price) : "",
    targetSize: item?.targetSize ?? "",
    duration: item?.duration != null ? String(item.duration) : "",
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
      icon={<Scissors className={LAYOUT.pageIcon.innerIcon} />}
    >
      <PropertyRow label="ステータス">
        <button
          type="button"
          onClick={() =>
            setFormData((prev) => ({ ...prev, isActive: !prev.isActive }))
          }
          className={`inline-flex items-center rounded-[3px] ${C.hoverBgLight} transition-colors py-0.5 px-0.5 cursor-pointer`}
        >
          <NotionStatusPill isActive={formData.isActive} />
        </button>
      </PropertyRow>

      <PropertyRow label="対象サイズ">
        <Select
          value={formData.targetSize || "__none__"}
          onValueChange={(v) =>
            setFormData((prev) => ({
              ...prev,
              targetSize: v === "__none__" ? "" : (v as TargetSize),
            }))
          }
        >
          <SelectTrigger className={STYLE.selectCompact}>
            <SelectValue placeholder="選択" />
          </SelectTrigger>
          <SelectContent>{TARGET_SIZE_SELECT_ITEMS}</SelectContent>
        </Select>
      </PropertyRow>

      <PropertyRow label="所要時間(分)">
        <PropInput
          type="number"
          value={formData.duration}
          onChange={(v) => setFormData((prev) => ({ ...prev, duration: v }))}
          placeholder="90"
        />
      </PropertyRow>

      <PropertyRow label="単価(税込)">
        <div className="flex items-center gap-1">
          <span className={`text-sm ${C.text65} select-none`}>¥</span>
          <input
            type="number"
            min={0}
            className={`w-32 bg-transparent text-sm ${C.text} outline-none border-none px-1.5 py-0.5 rounded-[3px] ${C.hoverBgLight} ${C.focusBgLight} transition-colors ${C.textPlaceholder}`}
            value={formData.price}
            onChange={(e) =>
              setFormData((prev) => ({ ...prev, price: e.target.value }))
            }
            placeholder="0"
          />
        </div>
      </PropertyRow>

      <PropertyRow label="備考">
        <PropInput
          value={formData.description}
          onChange={(v) =>
            setFormData((prev) => ({ ...prev, description: v }))
          }
          placeholder="補足情報など"
        />
      </PropertyRow>
    </MasterSidePanel>
  );
});

// ─────────────────────────────────────────────────
// TrimmingCourseTab
// ─────────────────────────────────────────────────

interface TrimmingCourseTabProps {
  editTarget: TrimmingCourse | "new" | null;
  onEditTargetChange: (v: TrimmingCourse | "new" | null) => void;
}

function TrimmingCourseTab({ editTarget: _editTarget, onEditTargetChange }: TrimmingCourseTabProps) {
  const [searchTerm, setSearchTerm] = useState("");
  const [activeFilters, setActiveFilters] = useState<ActiveFilter[]>([]);

  const { data: rawCourses } = useGetTrimmingCourses();

  // rerender-transitions: 検索フィルタを低優先度に遅延
  const deferredSearch = useDeferredValue(searchTerm);

  const filteredItems = useMemo(() => {
    let items = rawCourses ?? [];
    for (const f of activeFilters) {
      if (f.key === "status" && typeof f.value === "string") {
        const want = f.value === "active";
        items = items.filter((c) => f.condition === "is" ? c.isActive === want : c.isActive !== want);
      }
    }
    if (deferredSearch) {
      const lower = deferredSearch.toLowerCase();
      items = items.filter((c) => c.name.toLowerCase().includes(lower));
    }
    return items;
  }, [rawCourses, activeFilters, deferredSearch]);

  return (
    <div className="flex flex-col gap-4">
      <NotionFilter
        properties={[MASTER_STATUS_FILTER]}
        activeFilters={activeFilters}
        onFilterChange={setActiveFilters}
        searchTerm={searchTerm}
        onSearchChange={setSearchTerm}
        searchPlaceholder="コース名で検索..."
        count={filteredItems.length}
      />

      <DataTable
        columns={COURSE_COLUMNS}
        data={filteredItems}
        emptyMessage="トリミングコースが登録されていません"
        renderRow={(item) => (
          <DataTableRow key={item.id} onClick={() => onEditTargetChange(item)}>
            <TableCell className={`font-medium text-sm ${C.text}`}>
              {item.name}
            </TableCell>
            <TableCell className={`text-sm ${C.text70}`}>
              {item.targetSize ? TARGET_SIZE_LABELS[item.targetSize] : "-"}
            </TableCell>
            <TableCell className={`text-sm ${C.text70}`}>
              {item.duration != null ? `${item.duration}分` : "-"}
            </TableCell>
            <TableCell className={`text-right font-mono text-sm ${C.text}`}>
              {item.price != null ? `¥${item.price.toLocaleString()}` : "-"}
            </TableCell>
            <TableCell className="text-right">
              <NotionStatusPill isActive={item.isActive} />
            </TableCell>
          </DataTableRow>
        )}
      />
    </div>
  );
}

// ─────────────────────────────────────────────────
// OptionFormData
// ─────────────────────────────────────────────────

interface OptionFormData {
  name: string;
  price: string;
  duration: string;
  combinable: boolean;
  description: string;
  isActive: boolean;
}

// ─────────────────────────────────────────────────
// TrimmingOptionSidePanel
// ─────────────────────────────────────────────────

interface TrimmingOptionSidePanelProps {
  item: TrimmingOption | null;
  onClose: () => void;
  onSave: (data: OptionFormData) => void;
  onDeleteRequest: (item: TrimmingOption) => void;
}

const TrimmingOptionSidePanel = memo(function TrimmingOptionSidePanel({
  item,
  onClose,
  onSave,
  onDeleteRequest,
}: TrimmingOptionSidePanelProps) {
  const [formData, setFormData] = useState<OptionFormData>(() => ({
    name: item?.name ?? "",
    price: item?.price != null ? String(item.price) : "",
    duration: item?.duration != null ? String(item.duration) : "",
    combinable: item?.combinable ?? true,
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
      icon={<Scissors className={LAYOUT.pageIcon.innerIcon} />}
    >
      <PropertyRow label="ステータス">
        <button
          type="button"
          onClick={() =>
            setFormData((prev) => ({ ...prev, isActive: !prev.isActive }))
          }
          className={`inline-flex items-center rounded-[3px] ${C.hoverBgLight} transition-colors py-0.5 px-0.5 cursor-pointer`}
        >
          <NotionStatusPill isActive={formData.isActive} />
        </button>
      </PropertyRow>

      <PropertyRow label="所要時間(分)">
        <PropInput
          type="number"
          value={formData.duration}
          onChange={(v) => setFormData((prev) => ({ ...prev, duration: v }))}
          placeholder="30"
        />
      </PropertyRow>

      <PropertyRow label="組合せ可否">
        <button
          type="button"
          onClick={() =>
            setFormData((prev) => ({ ...prev, combinable: !prev.combinable }))
          }
          className={`inline-flex items-center rounded-[3px] ${C.hoverBgLight} transition-colors py-0.5 px-0.5 cursor-pointer`}
        >
          <CombinablePill combinable={formData.combinable} />
        </button>
      </PropertyRow>

      <PropertyRow label="単価(税込)">
        <div className="flex items-center gap-1">
          <span className={`text-sm ${C.text65} select-none`}>¥</span>
          <input
            type="number"
            min={0}
            className={`w-32 bg-transparent text-sm ${C.text} outline-none border-none px-1.5 py-0.5 rounded-[3px] ${C.hoverBgLight} ${C.focusBgLight} transition-colors ${C.textPlaceholder}`}
            value={formData.price}
            onChange={(e) =>
              setFormData((prev) => ({ ...prev, price: e.target.value }))
            }
            placeholder="0"
          />
        </div>
      </PropertyRow>

      <PropertyRow label="備考">
        <PropInput
          value={formData.description}
          onChange={(v) =>
            setFormData((prev) => ({ ...prev, description: v }))
          }
          placeholder="補足情報など"
        />
      </PropertyRow>
    </MasterSidePanel>
  );
});

// ─────────────────────────────────────────────────
// TrimmingOptionTab
// ─────────────────────────────────────────────────

interface TrimmingOptionTabProps {
  editTarget: TrimmingOption | "new" | null;
  onEditTargetChange: (v: TrimmingOption | "new" | null) => void;
}

function TrimmingOptionTab({ editTarget: _editTarget, onEditTargetChange }: TrimmingOptionTabProps) {
  const [searchTerm, setSearchTerm] = useState("");
  const [activeFilters, setActiveFilters] = useState<ActiveFilter[]>([]);

  const { data: rawOptions } = useGetTrimmingOptions();

  // rerender-transitions: 検索フィルタを低優先度に遅延
  const deferredSearch = useDeferredValue(searchTerm);

  const filteredItems = useMemo(() => {
    let items = rawOptions ?? [];
    for (const f of activeFilters) {
      if (f.key === "status" && typeof f.value === "string") {
        const want = f.value === "active";
        items = items.filter((o) => f.condition === "is" ? o.isActive === want : o.isActive !== want);
      }
    }
    if (deferredSearch) {
      const lower = deferredSearch.toLowerCase();
      items = items.filter((o) => o.name.toLowerCase().includes(lower));
    }
    return items;
  }, [rawOptions, activeFilters, deferredSearch]);

  return (
    <div className="flex flex-col gap-4">
      <NotionFilter
        properties={[MASTER_STATUS_FILTER]}
        activeFilters={activeFilters}
        onFilterChange={setActiveFilters}
        searchTerm={searchTerm}
        onSearchChange={setSearchTerm}
        searchPlaceholder="オプション名で検索..."
        count={filteredItems.length}
      />

      <DataTable
        columns={OPTION_COLUMNS}
        data={filteredItems}
        emptyMessage="トリミングオプションが登録されていません"
        renderRow={(item) => (
          <DataTableRow key={item.id} onClick={() => onEditTargetChange(item)}>
            <TableCell className={`font-medium text-sm ${C.text}`}>
              {item.name}
            </TableCell>
            <TableCell className={`text-sm ${C.text70}`}>
              {item.duration != null ? `${item.duration}分` : "-"}
            </TableCell>
            <TableCell className="text-center">
              <CombinablePill combinable={item.combinable} />
            </TableCell>
            <TableCell className={`text-right font-mono text-sm ${C.text}`}>
              {item.price != null ? `¥${item.price.toLocaleString()}` : "-"}
            </TableCell>
            <TableCell className="text-right">
              <NotionStatusPill isActive={item.isActive} />
            </TableCell>
          </DataTableRow>
        )}
      />
    </div>
  );
}

// ─────────────────────────────────────────────────
// TrimmingSettings (main page)
// ─────────────────────────────────────────────────

const TABS = [
  { value: "course", label: "コース" },
  { value: "option", label: "オプション" },
] as const;

export function TrimmingSettings() {
  const navigate = useNavigate();
  const [activeTab, setActiveTab] = useState<string>("course");

  const createCourseMutation = useCreateTrimmingCourse();
  const updateCourseMutation = useUpdateTrimmingCourse();
  const deleteCourseMutation = useDeleteTrimmingCourse();
  const createOptionMutation = useCreateTrimmingOption();
  const updateOptionMutation = useUpdateTrimmingOption();
  const deleteOptionMutation = useDeleteTrimmingOption();

  const courseCrud = useMasterCRUD<TrimmingCourse>({
    data: undefined,
    deleteMutation: deleteCourseMutation,
    entityLabel: "トリミングコース",
  });

  const optionCrud = useMasterCRUD<TrimmingOption>({
    data: undefined,
    deleteMutation: deleteOptionMutation,
    entityLabel: "トリミングオプション",
  });

  const handleTabChange = useCallback((tab: string) => {
    setActiveTab(tab);
    courseCrud.setEditTarget(null);
    optionCrud.setEditTarget(null);
  }, [courseCrud.setEditTarget, optionCrud.setEditTarget]);

  const handleCourseSave = useCallback(
    (data: CourseFormData) => {
      if (!data.name.trim()) {
        toast.error("コース名は必須です");
        return;
      }
      const priceValue = data.price !== "" ? Number(data.price) : null;
      courseCrud.startSaveTransition(() => {
        if (courseCrud.editTarget !== null && courseCrud.editTarget !== "new") {
          const req: UpdateTrimmingCourseRequest = {
            name: data.name,
            price: priceValue,
            target_size: data.targetSize !== "" ? data.targetSize : null,
            duration: data.duration !== "" ? Number(data.duration) : null,
            description: data.description || undefined,
            is_active: data.isActive,
          };
          updateCourseMutation.mutate(
            { id: courseCrud.editTarget.id, req },
            {
              onSuccess: () => { toast.success("更新しました"); courseCrud.handleClose(); },
              onError: () => toast.error("更新に失敗しました"),
            },
          );
        } else {
          const req: CreateTrimmingCourseRequest = {
            name: data.name,
            price: priceValue,
            target_size: data.targetSize !== "" ? data.targetSize : null,
            duration: data.duration !== "" ? Number(data.duration) : null,
            description: data.description || undefined,
            is_active: true,
          };
          createCourseMutation.mutate(req, {
            onSuccess: () => { toast.success("登録しました"); courseCrud.handleClose(); },
            onError: () => toast.error("登録に失敗しました"),
          });
        }
      });
    },
    [courseCrud.editTarget, updateCourseMutation, createCourseMutation, courseCrud.handleClose, courseCrud.startSaveTransition],
  );

  const handleOptionSave = useCallback(
    (data: OptionFormData) => {
      if (!data.name.trim()) {
        toast.error("オプション名は必須です");
        return;
      }
      const priceValue = data.price !== "" ? Number(data.price) : null;
      optionCrud.startSaveTransition(() => {
        if (optionCrud.editTarget !== null && optionCrud.editTarget !== "new") {
          const req: UpdateTrimmingOptionRequest = {
            name: data.name,
            price: priceValue,
            duration: data.duration !== "" ? Number(data.duration) : null,
            combinable: data.combinable,
            description: data.description || undefined,
            is_active: data.isActive,
          };
          updateOptionMutation.mutate(
            { id: optionCrud.editTarget.id, req },
            {
              onSuccess: () => { toast.success("更新しました"); optionCrud.handleClose(); },
              onError: () => toast.error("更新に失敗しました"),
            },
          );
        } else {
          const req: CreateTrimmingOptionRequest = {
            name: data.name,
            price: priceValue,
            duration: data.duration !== "" ? Number(data.duration) : null,
            combinable: data.combinable,
            description: data.description || undefined,
            is_active: true,
          };
          createOptionMutation.mutate(req, {
            onSuccess: () => { toast.success("登録しました"); optionCrud.handleClose(); },
            onError: () => toast.error("登録に失敗しました"),
          });
        }
      });
    },
    [optionCrud.editTarget, updateOptionMutation, createOptionMutation, optionCrud.handleClose, optionCrud.startSaveTransition],
  );

  return (
    <>
      <div className="flex h-full">
        <div className="flex-1 min-w-0">
          <PageLayout
            title="トリミングマスタ"
            icon={<Scissors className="size-5 text-[#37352F]" />}
            onBack={() => navigate(paths.settings.getHref())}
            maxWidth="max-w-full"
            headerAction={
              <PrimaryButton onClick={() => {
                if (activeTab === "course") courseCrud.handleNew();
                else optionCrud.handleNew();
              }}>
                <Plus className="mr-1.5 size-4" />
                新規登録
              </PrimaryButton>
            }
          >
            <div className="flex flex-col gap-4">
              <Tabs value={activeTab} onValueChange={handleTabChange}>
                <TabsList
                  className={`h-9 bg-transparent border-b ${C.borderLight} rounded-none w-full justify-start gap-0 p-0`}
                >
                  {TABS.map((tab) => (
                    <TabsTrigger
                      key={tab.value}
                      value={tab.value}
                      className={`h-9 rounded-none border-b-2 border-transparent px-4 text-sm ${C.text60}
                        data-[state=active]:border-[#37352F] data-[state=active]:${C.text}
                        data-[state=active]:shadow-none data-[state=active]:bg-transparent`}
                    >
                      {tab.label}
                    </TabsTrigger>
                  ))}
                </TabsList>
                <TabsContent value="course" className="mt-4">
                  <TrimmingCourseTab
                    editTarget={courseCrud.editTarget}
                    onEditTargetChange={courseCrud.setEditTarget}
                  />
                </TabsContent>
                <TabsContent value="option" className="mt-4">
                  <TrimmingOptionTab
                    editTarget={optionCrud.editTarget}
                    onEditTargetChange={optionCrud.setEditTarget}
                  />
                </TabsContent>
              </Tabs>
            </div>
          </PageLayout>
        </div>

        {activeTab === "course" && courseCrud.isEditing ? (
          <TrimmingCourseSidePanel
            key={courseCrud.panelItem ? String(courseCrud.panelItem.id) : "new-trimming-course"}
            item={courseCrud.panelItem}
            onClose={courseCrud.handleClose}
            onSave={handleCourseSave}
            onDeleteRequest={courseCrud.setPendingDelete}
          />
        ) : null}
        {activeTab === "option" && optionCrud.isEditing ? (
          <TrimmingOptionSidePanel
            key={optionCrud.panelItem ? String(optionCrud.panelItem.id) : "new-trimming-option"}
            item={optionCrud.panelItem}
            onClose={optionCrud.handleClose}
            onSave={handleOptionSave}
            onDeleteRequest={optionCrud.setPendingDelete}
          />
        ) : null}
      </div>

      <ConfirmDialog
        open={courseCrud.pendingDelete !== null}
        onClose={courseCrud.handleDeleteCancel}
        title="トリミングコースを削除しますか？"
        description={`「${courseCrud.pendingDelete?.name}」を削除します。この操作は取り消せません。`}
        confirmLabel="削除"
        variant="destructive"
        onConfirm={courseCrud.handleDeleteConfirm}
      />
      <ConfirmDialog
        open={optionCrud.pendingDelete !== null}
        onClose={optionCrud.handleDeleteCancel}
        title="トリミングオプションを削除しますか？"
        description={`「${optionCrud.pendingDelete?.name}」を削除します。この操作は取り消せません。`}
        confirmLabel="削除"
        variant="destructive"
        onConfirm={optionCrud.handleDeleteConfirm}
      />
    </>
  );
}
