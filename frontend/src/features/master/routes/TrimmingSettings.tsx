// React/Framework
import { useState, useMemo, useCallback, memo, useDeferredValue, useTransition } from "react";
import { useNavigate } from "react-router";
import { paths } from "@/config/paths";

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
import { PageLayout } from "@/components/shared/PageLayout";
import { SearchFilterBar } from "@/components/shared/SearchFilterBar";
import { DataTable, DataTableRow } from "@/components/shared/DataTable";
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
  useListTrimmingCourses,
  useCreateTrimmingCourse,
  useUpdateTrimmingCourse,
  useDeleteTrimmingCourse,
  useListTrimmingOptions,
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
  { header: "ステータス", className: "w-[90px]", align: "center" as const },
  { header: "操作", className: "w-[80px]", align: "right" as const },
];

const OPTION_COLUMNS = [
  { header: "オプション名" },
  { header: "所要時間", className: "w-[100px]" },
  { header: "組合せ可否", className: "w-[110px]", align: "center" as const },
  { header: "単価(税込)", className: "w-[110px]", align: "right" as const },
  { header: "ステータス", className: "w-[90px]", align: "center" as const },
  { header: "操作", className: "w-[80px]", align: "right" as const },
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
    <SidePeekPanel>
      <SidePeekToolbar
        isNew={item === null}
        onClose={onClose}
        onDelete={item !== null ? () => onDeleteRequest(item) : undefined}
      />
      <SidePeekBody>
        <div className="pt-4 pb-2">
          <div className={STYLE.pageIcon}>
            <Scissors className={LAYOUT.pageIcon.innerIcon} />
          </div>
        </div>
        <SidePeekTitleInput
          value={formData.name}
          onChange={(v) => setFormData((prev) => ({ ...prev, name: v }))}
        />
        <div className={`${STYLE.sectionDivider} mb-1`} />
        <div className="py-1">
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
        </div>
      </SidePeekBody>
      <SidePeekFooter onCancel={onClose} onSave={() => onSave(formData)} />
    </SidePeekPanel>
  );
});

// ─────────────────────────────────────────────────
// TrimmingCourseTab
// ─────────────────────────────────────────────────

function TrimmingCourseTab() {
  // null=closed, "new"=create mode, TrimmingCourse=edit mode
  const [editTarget, setEditTarget] = useState<TrimmingCourse | "new" | null>(null);
  const [searchTerm, setSearchTerm] = useState("");
  const [pendingDelete, setPendingDelete] = useState<TrimmingCourse | null>(null);
  const [, startSaveTransition] = useTransition();

  const { data: rawCourses } = useListTrimmingCourses();
  const createMutation = useCreateTrimmingCourse();
  const updateMutation = useUpdateTrimmingCourse();
  const deleteMutation = useDeleteTrimmingCourse();

  // rerender-transitions: 検索フィルタを低優先度に遅延
  const deferredSearch = useDeferredValue(searchTerm);

  const filteredItems = useMemo(() => {
    const courses = rawCourses ?? [];
    if (!deferredSearch) return courses;
    const lower = deferredSearch.toLowerCase();
    return courses.filter((c) => c.name.toLowerCase().includes(lower));
  }, [rawCourses, deferredSearch]);

  const handleClose = useCallback(() => setEditTarget(null), []);

  const handleSave = useCallback(
    (data: CourseFormData) => {
      if (!data.name.trim()) {
        toast.error("コース名は必須です");
        return;
      }

      const priceValue = data.price !== "" ? Number(data.price) : null;

      // rerender-transitions: API書き込みを非緊急マーク
      startSaveTransition(() => {
        if (editTarget !== null && editTarget !== "new") {
          const req: UpdateTrimmingCourseRequest = {
            name: data.name,
            price: priceValue,
            target_size: data.targetSize !== "" ? data.targetSize : null,
            duration: data.duration !== "" ? Number(data.duration) : null,
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
    [editTarget, updateMutation, createMutation, handleClose],
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
        {/* ── Left: List ── */}
        <div className="flex-1 min-w-0 flex flex-col gap-4">
          <div className="flex items-center gap-3">
            <div className="flex-1 min-w-0">
              <SearchFilterBar
                searchTerm={searchTerm}
                onSearchChange={setSearchTerm}
                placeholder="コース名で検索..."
                count={filteredItems.length}
              />
            </div>
            <button
              type="button"
              onClick={() => setEditTarget("new")}
              className="inline-flex items-center gap-1 text-sm font-medium text-[#2383E2] hover:text-[#1B6EC2] cursor-pointer transition-colors"
            >
              <Plus className="size-4" />
              新規登録
            </button>
          </div>

          <DataTable
            columns={COURSE_COLUMNS}
            data={filteredItems}
            emptyMessage="トリミングコースが登録されていません"
            renderRow={(item) => (
              <DataTableRow key={item.id} onClick={() => setEditTarget(item)}>
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

        {/* ── Right: Side Peek ── */}
        {editTarget !== null ? (
          <TrimmingCourseSidePanel
            key={panelItem ? String(panelItem.id) : "new-trimming-course"}
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
        title="トリミングコースを削除しますか？"
        description={`「${pendingDelete?.name}」を削除します。この操作は取り消せません。`}
        confirmLabel="削除"
        variant="destructive"
        onConfirm={handleDeleteConfirm}
      />
    </>
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
    <SidePeekPanel>
      <SidePeekToolbar
        isNew={item === null}
        onClose={onClose}
        onDelete={item !== null ? () => onDeleteRequest(item) : undefined}
      />
      <SidePeekBody>
        <div className="pt-4 pb-2">
          <div className={STYLE.pageIcon}>
            <Scissors className={LAYOUT.pageIcon.innerIcon} />
          </div>
        </div>
        <SidePeekTitleInput
          value={formData.name}
          onChange={(v) => setFormData((prev) => ({ ...prev, name: v }))}
        />
        <div className={`${STYLE.sectionDivider} mb-1`} />
        <div className="py-1">
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
        </div>
      </SidePeekBody>
      <SidePeekFooter onCancel={onClose} onSave={() => onSave(formData)} />
    </SidePeekPanel>
  );
});

// ─────────────────────────────────────────────────
// TrimmingOptionTab
// ─────────────────────────────────────────────────

function TrimmingOptionTab() {
  // null=closed, "new"=create mode, TrimmingOption=edit mode
  const [editTarget, setEditTarget] = useState<TrimmingOption | "new" | null>(null);
  const [searchTerm, setSearchTerm] = useState("");
  const [pendingDelete, setPendingDelete] = useState<TrimmingOption | null>(null);
  const [, startSaveTransition] = useTransition();

  const { data: rawOptions } = useListTrimmingOptions();
  const createMutation = useCreateTrimmingOption();
  const updateMutation = useUpdateTrimmingOption();
  const deleteMutation = useDeleteTrimmingOption();

  // rerender-transitions: 検索フィルタを低優先度に遅延
  const deferredSearch = useDeferredValue(searchTerm);

  const filteredItems = useMemo(() => {
    const options = rawOptions ?? [];
    if (!deferredSearch) return options;
    const lower = deferredSearch.toLowerCase();
    return options.filter((o) => o.name.toLowerCase().includes(lower));
  }, [rawOptions, deferredSearch]);

  const handleClose = useCallback(() => setEditTarget(null), []);

  const handleSave = useCallback(
    (data: OptionFormData) => {
      if (!data.name.trim()) {
        toast.error("オプション名は必須です");
        return;
      }

      const priceValue = data.price !== "" ? Number(data.price) : null;

      // rerender-transitions: API書き込みを非緊急マーク
      startSaveTransition(() => {
        if (editTarget !== null && editTarget !== "new") {
          const req: UpdateTrimmingOptionRequest = {
            name: data.name,
            price: priceValue,
            duration: data.duration !== "" ? Number(data.duration) : null,
            combinable: data.combinable,
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
    [editTarget, updateMutation, createMutation, handleClose],
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
        {/* ── Left: List ── */}
        <div className="flex-1 min-w-0 flex flex-col gap-4">
          <div className="flex items-center gap-3">
            <div className="flex-1 min-w-0">
              <SearchFilterBar
                searchTerm={searchTerm}
                onSearchChange={setSearchTerm}
                placeholder="オプション名で検索..."
                count={filteredItems.length}
              />
            </div>
            <button
              type="button"
              onClick={() => setEditTarget("new")}
              className="inline-flex items-center gap-1 text-sm font-medium text-[#2383E2] hover:text-[#1B6EC2] cursor-pointer transition-colors"
            >
              <Plus className="size-4" />
              新規登録
            </button>
          </div>

          <DataTable
            columns={OPTION_COLUMNS}
            data={filteredItems}
            emptyMessage="トリミングオプションが登録されていません"
            renderRow={(item) => (
              <DataTableRow key={item.id} onClick={() => setEditTarget(item)}>
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

        {/* ── Right: Side Peek ── */}
        {editTarget !== null ? (
          <TrimmingOptionSidePanel
            key={panelItem ? String(panelItem.id) : "new-trimming-option"}
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
        title="トリミングオプションを削除しますか？"
        description={`「${pendingDelete?.name}」を削除します。この操作は取り消せません。`}
        confirmLabel="削除"
        variant="destructive"
        onConfirm={handleDeleteConfirm}
      />
    </>
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

  return (
    <PageLayout
      title="トリミングマスタ"
      icon={<Scissors className="size-5 text-[#37352F]" />}
      onBack={() => navigate(paths.settings.getHref())}
      maxWidth="max-w-full"
    >
      <div className="flex flex-col gap-4">
        <Tabs value={activeTab} onValueChange={setActiveTab}>
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
            <TrimmingCourseTab />
          </TabsContent>
          <TabsContent value="option" className="mt-4">
            <TrimmingOptionTab />
          </TabsContent>
        </Tabs>
      </div>
    </PageLayout>
  );
}
