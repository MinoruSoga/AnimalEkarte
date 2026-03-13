// React/Framework
import type { ReactNode } from "react";
import { useState, useMemo } from "react";
import { useNavigate } from "react-router";

// External
import { Plus, Scissors, X, Trash2 } from "lucide-react";
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
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { PageLayout } from "@/components/shared/PageLayout";
import { SearchFilterBar } from "@/components/shared/SearchFilterBar";
import { DataTable, DataTableRow } from "@/components/shared/DataTable";
import { PrimaryButton } from "@/components/shared/Form/PrimaryButton";
import { ConfirmDialog } from "@/components/shared/ConfirmDialog/ConfirmDialog";
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
// Notion Status Pill
// ─────────────────────────────────────────────────

const STATUS_CONFIG = {
  active: {
    dot: "bg-[#2383E2]",
    label: "有効",
    bg: "bg-[#D3E5EF]",
    text: "text-[#183B56]",
  },
  inactive: {
    dot: "bg-[#37352F]/10",
    label: "無効",
    bg: "bg-[#E3E2E0]",
    text: "text-[#37352F]/60",
  },
} as const;

function NotionStatusPill({ isActive }: { isActive: boolean }) {
  const cfg = STATUS_CONFIG[isActive ? "active" : "inactive"];
  return (
    <span
      className={`inline-flex items-center gap-1.5 px-2 py-0.5 rounded-sm text-xs ${cfg.bg} ${cfg.text}`}
    >
      <span className={`size-[7px] rounded-full ${cfg.dot}`} />
      {cfg.label}
    </span>
  );
}

// Combinable pill
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
// Property Row (Notion-style)
// ─────────────────────────────────────────────────

function PropertyRow({ label, children }: { label: string; children: ReactNode }) {
  return (
    <div
      className={`flex gap-2 py-2 px-2 -mx-2 rounded-[3px] ${C.hoverBgLight} transition-colors min-h-[40px]`}
    >
      <div className="w-[140px] shrink-0 text-sm text-[#37352F]/65 select-none truncate flex items-center">
        {label}
      </div>
      <div className="flex-1 flex items-center">{children}</div>
    </div>
  );
}

// ─────────────────────────────────────────────────
// Inline input for property rows
// ─────────────────────────────────────────────────

function PropInput({
  value,
  onChange,
  placeholder,
  type = "text",
}: {
  value: string;
  onChange: (v: string) => void;
  placeholder?: string;
  type?: string;
}) {
  return (
    <input
      type={type}
      className={`w-full bg-transparent text-sm ${C.text} outline-none border-none px-1.5 py-0.5 rounded-[3px] ${C.hoverBgLight} ${C.focusBgLight} transition-colors ${C.textPlaceholder}`}
      value={value}
      onChange={(e) => onChange(e.target.value)}
      placeholder={placeholder ?? "空"}
    />
  );
}

// ─────────────────────────────────────────────────
// TrimmingCourseTab
// ─────────────────────────────────────────────────

interface CourseFormData {
  name: string;
  price: string;
  targetSize: TargetSize | "";
  duration: string;
  description: string;
  isActive: boolean;
}

const DEFAULT_COURSE_FORM: CourseFormData = {
  name: "",
  price: "",
  targetSize: "",
  duration: "",
  description: "",
  isActive: true,
};

function TrimmingCourseTab() {
  const [isEditing, setIsEditing] = useState(false);
  const [selectedItem, setSelectedItem] = useState<TrimmingCourse | null>(null);
  const [searchTerm, setSearchTerm] = useState("");
  const [pendingDelete, setPendingDelete] = useState<TrimmingCourse | null>(null);
  const [formData, setFormData] = useState<CourseFormData>(DEFAULT_COURSE_FORM);

  const { data: rawCourses } = useListTrimmingCourses();
  const createMutation = useCreateTrimmingCourse();
  const updateMutation = useUpdateTrimmingCourse();
  const deleteMutation = useDeleteTrimmingCourse();

  const filteredItems = useMemo(() => {
    const courses = rawCourses ?? [];
    if (!searchTerm) return courses;
    const lower = searchTerm.toLowerCase();
    return courses.filter((c) => c.name.toLowerCase().includes(lower));
  }, [rawCourses, searchTerm]);

  const handleEdit = (item: TrimmingCourse) => {
    setSelectedItem(item);
    setFormData({
      name: item.name,
      price: item.price != null ? String(item.price) : "",
      targetSize: item.targetSize ?? "",
      duration: item.duration,
      description: item.description,
      isActive: item.isActive,
    });
    setIsEditing(true);
  };

  const handleCreate = () => {
    setSelectedItem(null);
    setFormData(DEFAULT_COURSE_FORM);
    setIsEditing(true);
  };

  const handleCloseEdit = () => {
    setIsEditing(false);
    setSelectedItem(null);
    setFormData(DEFAULT_COURSE_FORM);
  };

  const handleSave = () => {
    if (!formData.name.trim()) {
      toast.error("コース名は必須です");
      return;
    }

    const priceValue = formData.price !== "" ? Number(formData.price) : null;

    if (selectedItem) {
      const req: UpdateTrimmingCourseRequest = {
        name: formData.name,
        price: priceValue,
        target_size: formData.targetSize !== "" ? formData.targetSize : null,
        duration: formData.duration || undefined,
        description: formData.description || undefined,
        is_active: formData.isActive,
      };
      updateMutation.mutate(
        { id: selectedItem.id, req },
        {
          onSuccess: () => {
            toast.success("更新しました");
            setIsEditing(false);
          },
          onError: () => toast.error("更新に失敗しました"),
        },
      );
    } else {
      const req: CreateTrimmingCourseRequest = {
        name: formData.name,
        price: priceValue,
        target_size: formData.targetSize !== "" ? formData.targetSize : null,
        duration: formData.duration || undefined,
        description: formData.description || undefined,
        is_active: true,
      };
      createMutation.mutate(req, {
        onSuccess: () => {
          toast.success("登録しました");
          setIsEditing(false);
        },
        onError: () => toast.error("登録に失敗しました"),
      });
    }
  };

  const handleDeleteConfirm = () => {
    if (!pendingDelete) return;
    deleteMutation.mutate(pendingDelete.id, {
      onSuccess: () => {
        setPendingDelete(null);
        setIsEditing(false);
        toast.success("削除しました");
      },
      onError: () => toast.error("削除に失敗しました"),
    });
  };

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
            <PrimaryButton onClick={handleCreate}>
              <Plus className="mr-1.5 size-4" />
              新規登録
            </PrimaryButton>
          </div>

          <DataTable
            columns={COURSE_COLUMNS}
            data={filteredItems}
            emptyMessage="トリミングコースが登録されていません"
            renderRow={(item) => (
              <DataTableRow key={item.id} onClick={() => handleEdit(item)}>
                <TableCell className={`font-medium text-sm ${C.text} py-2.5`}>
                  {item.name}
                </TableCell>
                <TableCell className={`text-sm ${C.text70} py-2.5`}>
                  {item.targetSize ? TARGET_SIZE_LABELS[item.targetSize] : "-"}
                </TableCell>
                <TableCell className={`text-sm ${C.text70} py-2.5`}>
                  {item.duration || "-"}
                </TableCell>
                <TableCell className={`text-right font-mono text-sm ${C.text} py-2.5`}>
                  {item.price != null ? `¥${item.price.toLocaleString()}` : "-"}
                </TableCell>
                <TableCell className="text-right py-2.5">
                  <span className="inline-flex items-center gap-1.5">
                    <span className={`size-[7px] rounded-full ${item.isActive ? "bg-[#2383E2]" : "bg-[#37352F]/20"}`} />
                    <span className={`text-sm ${item.isActive ? "text-[#37352F]/65" : "text-[#37352F]/35"}`}>
                      {item.isActive ? "有効" : "無効"}
                    </span>
                  </span>
                </TableCell>
              </DataTableRow>
            )}
          />
          <button
            type="button"
            onClick={handleCreate}
            className="flex items-center gap-1.5 w-full px-3 py-2 text-sm text-[#37352F]/40 hover:text-[#37352F]/65 hover:bg-[rgba(55,53,47,0.04)] transition-colors rounded"
          >
            <Plus className="size-3.5" />
            新しいコースを追加...
          </button>
        </div>

        {/* ── Right: Side Peek ── */}
        {isEditing && (
          <div
            className={`${STYLE.sidePeekPanel} ${LAYOUT.sidePeek.width} shrink-0 flex flex-col`}
          >
            {/* Toolbar */}
            <div className={STYLE.sidePeekToolbar}>
              <span className="text-xs text-[#37352F]/35 pl-1 select-none">
                {selectedItem ? "編集" : "新規作成"}
              </span>
              <div className="flex items-center gap-1">
                {selectedItem ? (
                  <button
                    type="button"
                    onClick={() => setPendingDelete(selectedItem)}
                    className={`${STYLE.sidePeekToolbarBtn} cursor-pointer text-[#EB5757] hover:bg-[#EB5757]/10`}
                  >
                    <Trash2 className="size-4" />
                  </button>
                ) : null}
                <button
                  type="button"
                  onClick={handleCloseEdit}
                  className={`${STYLE.sidePeekToolbarBtn} cursor-pointer`}
                >
                  <X className="size-4" />
                </button>
              </div>
            </div>

            {/* Body */}
            <div className={STYLE.sidePeekBody}>
              <div className="px-16 pb-8">
                <div className="pt-4 pb-2">
                  <div className={STYLE.pageIcon}>
                    <Scissors className={LAYOUT.pageIcon.innerIcon} />
                  </div>
                </div>

                <div className="pb-1 mb-4">
                  <input
                    type="text"
                    className={`w-full bg-transparent ${C.text} placeholder:text-[rgba(55,53,47,0.15)] outline-none border-none p-0`}
                    style={{
                      fontSize: LAYOUT.pageTitle.fontSize,
                      fontWeight: LAYOUT.pageTitle.fontWeight,
                      lineHeight: LAYOUT.pageTitle.lineHeight,
                    }}
                    value={formData.name}
                    onChange={(e) => setFormData({ ...formData, name: e.target.value })}
                    placeholder="無題"
                  />
                </div>

                <div className={`${STYLE.sectionDivider} mb-1`} />

                <div className="py-1">
                  <PropertyRow label="ステータス">
                    <button
                      type="button"
                      onClick={() =>
                        setFormData({ ...formData, isActive: !formData.isActive })
                      }
                      className={`inline-flex items-center rounded-[3px] ${C.hoverBgLight} transition-colors py-0.5 px-0.5 cursor-pointer`}
                    >
                      <NotionStatusPill isActive={formData.isActive} />
                    </button>
                  </PropertyRow>

                  <PropertyRow label="対象サイズ">
                    <Select
                      value={formData.targetSize}
                      onValueChange={(v) =>
                        setFormData({ ...formData, targetSize: v as TargetSize | "" })
                      }
                    >
                      <SelectTrigger className={STYLE.selectCompact}>
                        <SelectValue placeholder="選択" />
                      </SelectTrigger>
                      <SelectContent>
                        <SelectItem value="">指定なし</SelectItem>
                        {TARGET_SIZE_OPTIONS.map((opt) => (
                          <SelectItem key={opt.value} value={opt.value}>
                            {opt.label}
                          </SelectItem>
                        ))}
                      </SelectContent>
                    </Select>
                  </PropertyRow>

                  <PropertyRow label="所要時間">
                    <PropInput
                      value={formData.duration}
                      onChange={(v) => setFormData({ ...formData, duration: v })}
                      placeholder="例: 90分"
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
                        onChange={(e) => setFormData({ ...formData, price: e.target.value })}
                        placeholder="0"
                      />
                    </div>
                  </PropertyRow>

                  <PropertyRow label="備考">
                    <PropInput
                      value={formData.description}
                      onChange={(v) => setFormData({ ...formData, description: v })}
                      placeholder="補足情報など"
                    />
                  </PropertyRow>
                </div>
              </div>
            </div>

            {/* Footer */}
            <div className={STYLE.sidePeekFooter}>
              <button type="button" onClick={handleCloseEdit} className={STYLE.sidePeekCancelBtn}>
                キャンセル
              </button>
              <button type="button" onClick={handleSave} className={STYLE.sidePeekSaveBtn}>
                保存
              </button>
            </div>
          </div>
        )}
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
// TrimmingOptionTab
// ─────────────────────────────────────────────────

interface OptionFormData {
  name: string;
  price: string;
  duration: string;
  combinable: boolean;
  description: string;
  isActive: boolean;
}

const DEFAULT_OPTION_FORM: OptionFormData = {
  name: "",
  price: "",
  duration: "",
  combinable: true,
  description: "",
  isActive: true,
};

function TrimmingOptionTab() {
  const [isEditing, setIsEditing] = useState(false);
  const [selectedItem, setSelectedItem] = useState<TrimmingOption | null>(null);
  const [searchTerm, setSearchTerm] = useState("");
  const [pendingDelete, setPendingDelete] = useState<TrimmingOption | null>(null);
  const [formData, setFormData] = useState<OptionFormData>(DEFAULT_OPTION_FORM);

  const { data: rawOptions } = useListTrimmingOptions();
  const createMutation = useCreateTrimmingOption();
  const updateMutation = useUpdateTrimmingOption();
  const deleteMutation = useDeleteTrimmingOption();

  const filteredItems = useMemo(() => {
    const options = rawOptions ?? [];
    if (!searchTerm) return options;
    const lower = searchTerm.toLowerCase();
    return options.filter((o) => o.name.toLowerCase().includes(lower));
  }, [rawOptions, searchTerm]);

  const handleEdit = (item: TrimmingOption) => {
    setSelectedItem(item);
    setFormData({
      name: item.name,
      price: item.price != null ? String(item.price) : "",
      duration: item.duration,
      combinable: item.combinable,
      description: item.description,
      isActive: item.isActive,
    });
    setIsEditing(true);
  };

  const handleCreate = () => {
    setSelectedItem(null);
    setFormData(DEFAULT_OPTION_FORM);
    setIsEditing(true);
  };

  const handleCloseEdit = () => {
    setIsEditing(false);
    setSelectedItem(null);
    setFormData(DEFAULT_OPTION_FORM);
  };

  const handleSave = () => {
    if (!formData.name.trim()) {
      toast.error("オプション名は必須です");
      return;
    }

    const priceValue = formData.price !== "" ? Number(formData.price) : null;

    if (selectedItem) {
      const req: UpdateTrimmingOptionRequest = {
        name: formData.name,
        price: priceValue,
        duration: formData.duration || undefined,
        combinable: formData.combinable,
        description: formData.description || undefined,
        is_active: formData.isActive,
      };
      updateMutation.mutate(
        { id: selectedItem.id, req },
        {
          onSuccess: () => {
            toast.success("更新しました");
            setIsEditing(false);
          },
          onError: () => toast.error("更新に失敗しました"),
        },
      );
    } else {
      const req: CreateTrimmingOptionRequest = {
        name: formData.name,
        price: priceValue,
        duration: formData.duration || undefined,
        combinable: formData.combinable,
        description: formData.description || undefined,
        is_active: true,
      };
      createMutation.mutate(req, {
        onSuccess: () => {
          toast.success("登録しました");
          setIsEditing(false);
        },
        onError: () => toast.error("登録に失敗しました"),
      });
    }
  };

  const handleDeleteConfirm = () => {
    if (!pendingDelete) return;
    deleteMutation.mutate(pendingDelete.id, {
      onSuccess: () => {
        setPendingDelete(null);
        setIsEditing(false);
        toast.success("削除しました");
      },
      onError: () => toast.error("削除に失敗しました"),
    });
  };

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
            <PrimaryButton onClick={handleCreate}>
              <Plus className="mr-1.5 size-4" />
              新規登録
            </PrimaryButton>
          </div>

          <DataTable
            columns={OPTION_COLUMNS}
            data={filteredItems}
            emptyMessage="トリミングオプションが登録されていません"
            renderRow={(item) => (
              <DataTableRow key={item.id} onClick={() => handleEdit(item)}>
                <TableCell className={`font-medium text-sm ${C.text} py-2.5`}>
                  {item.name}
                </TableCell>
                <TableCell className={`text-sm ${C.text70} py-2.5`}>
                  {item.duration || "-"}
                </TableCell>
                <TableCell className="text-center py-2.5">
                  <CombinablePill combinable={item.combinable} />
                </TableCell>
                <TableCell className={`text-right font-mono text-sm ${C.text} py-2.5`}>
                  {item.price != null ? `¥${item.price.toLocaleString()}` : "-"}
                </TableCell>
                <TableCell className="text-right py-2.5">
                  <span className="inline-flex items-center gap-1.5">
                    <span className={`size-[7px] rounded-full ${item.isActive ? "bg-[#2383E2]" : "bg-[#37352F]/20"}`} />
                    <span className={`text-sm ${item.isActive ? "text-[#37352F]/65" : "text-[#37352F]/35"}`}>
                      {item.isActive ? "有効" : "無効"}
                    </span>
                  </span>
                </TableCell>
              </DataTableRow>
            )}
          />
          <button
            type="button"
            onClick={handleCreate}
            className="flex items-center gap-1.5 w-full px-3 py-2 text-sm text-[#37352F]/40 hover:text-[#37352F]/65 hover:bg-[rgba(55,53,47,0.04)] transition-colors rounded"
          >
            <Plus className="size-3.5" />
            新しいオプションを追加...
          </button>
        </div>

        {/* ── Right: Side Peek ── */}
        {isEditing && (
          <div
            className={`${STYLE.sidePeekPanel} ${LAYOUT.sidePeek.width} shrink-0 flex flex-col`}
          >
            {/* Toolbar */}
            <div className={STYLE.sidePeekToolbar}>
              <span className="text-xs text-[#37352F]/35 pl-1 select-none">
                {selectedItem ? "編集" : "新規作成"}
              </span>
              <div className="flex items-center gap-1">
                {selectedItem ? (
                  <button
                    type="button"
                    onClick={() => setPendingDelete(selectedItem)}
                    className={`${STYLE.sidePeekToolbarBtn} cursor-pointer text-[#EB5757] hover:bg-[#EB5757]/10`}
                  >
                    <Trash2 className="size-4" />
                  </button>
                ) : null}
                <button
                  type="button"
                  onClick={handleCloseEdit}
                  className={`${STYLE.sidePeekToolbarBtn} cursor-pointer`}
                >
                  <X className="size-4" />
                </button>
              </div>
            </div>

            {/* Body */}
            <div className={STYLE.sidePeekBody}>
              <div className="px-16 pb-8">
                <div className="pt-4 pb-2">
                  <div className={STYLE.pageIcon}>
                    <Scissors className={LAYOUT.pageIcon.innerIcon} />
                  </div>
                </div>

                <div className="pb-1 mb-4">
                  <input
                    type="text"
                    className={`w-full bg-transparent ${C.text} placeholder:text-[rgba(55,53,47,0.15)] outline-none border-none p-0`}
                    style={{
                      fontSize: LAYOUT.pageTitle.fontSize,
                      fontWeight: LAYOUT.pageTitle.fontWeight,
                      lineHeight: LAYOUT.pageTitle.lineHeight,
                    }}
                    value={formData.name}
                    onChange={(e) => setFormData({ ...formData, name: e.target.value })}
                    placeholder="無題"
                  />
                </div>

                <div className={`${STYLE.sectionDivider} mb-1`} />

                <div className="py-1">
                  <PropertyRow label="ステータス">
                    <button
                      type="button"
                      onClick={() =>
                        setFormData({ ...formData, isActive: !formData.isActive })
                      }
                      className={`inline-flex items-center rounded-[3px] ${C.hoverBgLight} transition-colors py-0.5 px-0.5 cursor-pointer`}
                    >
                      <NotionStatusPill isActive={formData.isActive} />
                    </button>
                  </PropertyRow>

                  <PropertyRow label="所要時間">
                    <PropInput
                      value={formData.duration}
                      onChange={(v) => setFormData({ ...formData, duration: v })}
                      placeholder="例: 30分"
                    />
                  </PropertyRow>

                  <PropertyRow label="組合せ可否">
                    <button
                      type="button"
                      onClick={() =>
                        setFormData({ ...formData, combinable: !formData.combinable })
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
                        onChange={(e) => setFormData({ ...formData, price: e.target.value })}
                        placeholder="0"
                      />
                    </div>
                  </PropertyRow>

                  <PropertyRow label="備考">
                    <PropInput
                      value={formData.description}
                      onChange={(v) => setFormData({ ...formData, description: v })}
                      placeholder="補足情報など"
                    />
                  </PropertyRow>
                </div>
              </div>
            </div>

            {/* Footer */}
            <div className={STYLE.sidePeekFooter}>
              <button type="button" onClick={handleCloseEdit} className={STYLE.sidePeekCancelBtn}>
                キャンセル
              </button>
              <button type="button" onClick={handleSave} className={STYLE.sidePeekSaveBtn}>
                保存
              </button>
            </div>
          </div>
        )}
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
      onBack={() => navigate("/settings")}
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
