// React/Framework
import { useState, useEffect } from "react";
import type { ReactNode } from "react";
import { useNavigate } from "react-router";

// External
import { toast } from "sonner";
import Save from "lucide-react/dist/esm/icons/save";
import Trash2 from "lucide-react/dist/esm/icons/trash-2";
import Plus from "lucide-react/dist/esm/icons/plus";
import Users from "lucide-react/dist/esm/icons/users";
import Building2 from "lucide-react/dist/esm/icons/building-2";
import KeyRound from "lucide-react/dist/esm/icons/key-round";
import Settings2 from "lucide-react/dist/esm/icons/settings-2";

// Internal
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { TableCell } from "@/components/ui/table";
import { PageLayout } from "@/components/shared/PageLayout";
import { SearchFilterBar } from "@/components/shared/SearchFilterBar";
import { DataTable } from "@/components/shared/DataTable";
import { DataTableRow } from "@/components/shared/DataTable";
import { PrimaryButton } from "@/components/shared/Form";
import { StatusBadge } from "@/components/shared/StatusBadge";
import { RowActionButton } from "@/components/shared/RowActionButton";
import { getMasterStatusColor } from "@/utils/status-helpers";
import { useMasterItems } from "@/hooks/use-master-items";
import {
  CATEGORY_CONFIG,
  CATEGORY_ALIAS_MAP,
} from "@/features/master/constants/category-config";
import type { MasterSettingsCategory } from "@/features/master/constants/category-config";
import { C, STYLE } from "@/lib/design-tokens";

// Types
import type { MasterItem } from "@/types";

// ─────────────────────────────────────────────────
// Staff-specific extra state
// ─────────────────────────────────────────────────
interface StaffFormExtras {
  role?: string;
  licenseNumber?: string;
  affiliatedClinic?: string;
  email?: string;
  password?: string;
  userType?: string;
  lastLoginAt?: string;
}

// ─────────────────────────────────────────────────
// Helpers
// ─────────────────────────────────────────────────

/**
 * Resolve the raw category prop (may be camelCase alias) to a
 * canonical MasterSettingsCategory key, or return undefined if unknown.
 */
function resolveCategory(raw: string): MasterSettingsCategory | undefined {
  if (raw in CATEGORY_ALIAS_MAP) {
    return CATEGORY_ALIAS_MAP[raw];
  }
  if (raw in CATEGORY_CONFIG) {
    return raw as MasterSettingsCategory;
  }
  return undefined;
}

/**
 * Map a canonical category key to the camelCase string that
 * useMasterItems() expects (it uses its own internal CATEGORY_MAP).
 * For categories already accepted by the hook in camelCase, pass through.
 * For snake_case aliases, convert back.
 */
function toCamelCaseCategory(cat: MasterSettingsCategory): string {
  const reverseAlias: Partial<Record<MasterSettingsCategory, string>> = {
    trimming_course: "trimmingCourse",
    trimming_option: "trimmingOption",
    diagnosis_category: "diagnosisCategory",
    diagnosis_name: "diagnosisName",
  };
  return reverseAlias[cat] ?? cat;
}

// ─────────────────────────────────────────────────
// Notion-style helpers
// ─────────────────────────────────────────────────
function PropertyRow({
  label,
  required,
  children,
}: {
  label: string;
  required?: boolean;
  children: ReactNode;
}) {
  return (
    <div
      className={`flex gap-2 py-2 px-2 -mx-2 rounded-[3px] ${C.hoverBgLight} transition-colors min-h-[40px] items-center`}
    >
      <div className={`w-[140px] shrink-0 text-sm ${C.text65} select-none`}>
        {label}
        {required && <span className={`${C.textRequired} ml-0.5`}>*</span>}
      </div>
      <div className="flex-1 min-w-0">{children}</div>
    </div>
  );
}

function SectionBlock({
  icon,
  title,
  children,
}: {
  icon: ReactNode;
  title: string;
  children: ReactNode;
}) {
  return (
    <div className="space-y-1">
      <div className={`flex items-center gap-1.5 py-2 border-b ${C.borderLight}`}>
        <span className={C.text40}>{icon}</span>
        <span className={`text-xs ${C.text55} uppercase tracking-wide select-none`}>
          {title}
        </span>
      </div>
      <div className="space-y-0">{children}</div>
    </div>
  );
}

// ─────────────────────────────────────────────────
// Compact text input for property rows
// ─────────────────────────────────────────────────
interface PropInputProps {
  value: string;
  onChange: (v: string) => void;
  placeholder?: string;
  type?: string;
}

function PropInput({ value, onChange, placeholder, type = "text" }: PropInputProps) {
  return (
    <input
      type={type}
      value={value}
      onChange={(e) => onChange(e.target.value)}
      placeholder={placeholder}
      className={`${STYLE.propertyInput}`}
    />
  );
}

// ─────────────────────────────────────────────────
// Props
// ─────────────────────────────────────────────────
interface SettingsPageProps {
  category?: string;
  /** When true, skip PageLayout wrapper and render list content only (for tabbed pages) */
  embedded?: boolean;
}

// ─────────────────────────────────────────────────
// Main Component
// ─────────────────────────────────────────────────
export function Settings({ category: propCategory, embedded = false }: SettingsPageProps) {
  const navigate = useNavigate();
  const rawCategory = propCategory ?? "examination";
  const resolvedCategory = resolveCategory(rawCategory);
  const config = resolvedCategory ? CATEGORY_CONFIG[resolvedCategory] : undefined;

  const isStaff = rawCategory === "staff";
  const hookCategory = resolvedCategory ? toCamelCaseCategory(resolvedCategory) : rawCategory;

  const [isEditing, setIsEditing] = useState(false);
  const [selectedItem, setSelectedItem] = useState<MasterItem | null>(null);
  const [searchTerm, setSearchTerm] = useState("");
  const [formData, setFormData] = useState<Partial<MasterItem>>({});
  const [staffExtras, setStaffExtras] = useState<StaffFormExtras>({});

  const { data: filteredItems, add, update, remove } = useMasterItems(
    hookCategory,
    searchTerm
  );

  const pageTitle = config?.label ?? "マスタ";
  const showPrice = config?.showPrice ?? (!isStaff && rawCategory !== "cage" && rawCategory !== "insurance");
  const showCode = config?.showCode ?? false;
  const showCategory = config?.showCategory ?? false;
  const labels = config?.labels ?? { code: "コード", name: "名称", category: "カテゴリ" };

  // Reset editing state when category changes
  useEffect(() => {
    setIsEditing(false);
    setSelectedItem(null);
    setSearchTerm("");
    setFormData({});
    setStaffExtras({});
  }, [rawCategory]);

  const handleEdit = (item: MasterItem) => {
    setSelectedItem(item);
    setFormData({ ...item });
    setIsEditing(true);
    if (isStaff) {
      setStaffExtras({ role: item.category ?? "" });
    }
  };

  const handleCreate = () => {
    setSelectedItem(null);
    setFormData({ status: "active", price: 0 });
    setStaffExtras({});
    setIsEditing(true);
  };

  const handleCloseEdit = () => {
    setIsEditing(false);
    setSelectedItem(null);
    setFormData({});
    setStaffExtras({});
  };

  const handleSave = () => {
    if (!formData.name) {
      toast.error("名称は必須です");
      return;
    }
    if (showCode && !formData.code) {
      toast.error("コードは必須です");
      return;
    }

    // For staff, merge extras into formData description/category
    const payload: Partial<MasterItem> = isStaff
      ? {
          ...formData,
          category: staffExtras.role ?? formData.category,
          description: [
            staffExtras.licenseNumber ? `資格番号:${staffExtras.licenseNumber}` : "",
            staffExtras.affiliatedClinic ? `所属:${staffExtras.affiliatedClinic}` : "",
            staffExtras.email ? `mail:${staffExtras.email}` : "",
          ]
            .filter(Boolean)
            .join(",") || formData.description,
        }
      : formData;

    if (selectedItem) {
      update(selectedItem.id, payload, {
        onSuccess: () => {
          toast.success("更新しました");
          handleCloseEdit();
        },
      });
    } else {
      add(payload as Omit<MasterItem, "id">, {
        onSuccess: () => {
          toast.success("登録しました");
          handleCloseEdit();
        },
      });
    }
  };

  const handleDelete = () => {
    if (!selectedItem) return;
    if (window.confirm("本当に削除しますか？")) {
      remove(selectedItem.id, {
        onSuccess: () => {
          handleCloseEdit();
          toast.success("削除しました");
        },
      });
    }
  };

  // ── Edit Form ──────────────────────────────────
  if (isEditing) {
    return (
      <PageLayout
        title={selectedItem ? `${pageTitle} 編集` : `${pageTitle} 新規登録`}
        onBack={handleCloseEdit}
        maxWidth="max-w-2xl"
        align="left"
      >
        <div
          className={`bg-white rounded-lg border ${C.borderLight} px-6 py-4 space-y-4 mt-2`}
        >
          {isStaff ? (
            // ── Staff edit form ──
            <>
              <SectionBlock
                icon={<Users className="size-3.5" />}
                title="スタッフ詳細"
              >
                <PropertyRow label={labels.name} required>
                  <PropInput
                    value={formData.name ?? ""}
                    onChange={(v) => setFormData({ ...formData, name: v })}
                    placeholder={config?.namePlaceholder ?? "山田 太郎"}
                  />
                </PropertyRow>
                <PropertyRow label={labels.code}>
                  <PropInput
                    value={formData.code ?? ""}
                    onChange={(v) => setFormData({ ...formData, code: v })}
                    placeholder={config?.codePlaceholder ?? "ST-001"}
                  />
                </PropertyRow>
                <PropertyRow label="職種">
                  <PropInput
                    value={staffExtras.role ?? ""}
                    onChange={(v) => setStaffExtras({ ...staffExtras, role: v })}
                    placeholder="獣医師"
                  />
                </PropertyRow>
                <PropertyRow label="資格番号">
                  <PropInput
                    value={staffExtras.licenseNumber ?? ""}
                    onChange={(v) =>
                      setStaffExtras({ ...staffExtras, licenseNumber: v })
                    }
                    placeholder="例: 12345"
                  />
                </PropertyRow>
              </SectionBlock>

              <SectionBlock
                icon={<Building2 className="size-3.5" />}
                title="所属情報"
              >
                <PropertyRow label="所属医院">
                  <PropInput
                    value={staffExtras.affiliatedClinic ?? ""}
                    onChange={(v) =>
                      setStaffExtras({ ...staffExtras, affiliatedClinic: v })
                    }
                    placeholder="○○動物病院"
                  />
                </PropertyRow>
              </SectionBlock>

              <SectionBlock
                icon={<KeyRound className="size-3.5" />}
                title="アカウント発行・権限設定"
              >
                <PropertyRow label="メールアドレス">
                  <PropInput
                    value={staffExtras.email ?? ""}
                    onChange={(v) => setStaffExtras({ ...staffExtras, email: v })}
                    placeholder="example@clinic.jp"
                    type="email"
                  />
                </PropertyRow>
                <PropertyRow label="パスワード">
                  <PropInput
                    value={staffExtras.password ?? ""}
                    onChange={(v) =>
                      setStaffExtras({ ...staffExtras, password: v })
                    }
                    placeholder="••••••••"
                    type="password"
                  />
                </PropertyRow>
                <PropertyRow label="ユーザー種別">
                  <Select
                    value={staffExtras.userType ?? ""}
                    onValueChange={(v) =>
                      setStaffExtras({ ...staffExtras, userType: v })
                    }
                  >
                    <SelectTrigger className={`${STYLE.selectCompact} h-8`}>
                      <SelectValue placeholder="選択してください" />
                    </SelectTrigger>
                    <SelectContent>
                      <SelectItem value="system_admin">運営管理者</SelectItem>
                      <SelectItem value="clinic_admin">医院管理者</SelectItem>
                      <SelectItem value="staff">スタッフ</SelectItem>
                    </SelectContent>
                  </Select>
                </PropertyRow>
              </SectionBlock>

              <SectionBlock
                icon={<Settings2 className="size-3.5" />}
                title="ステータス"
              >
                <PropertyRow label="ステータス">
                  <div className="flex items-center gap-6">
                    <label
                      className={`flex items-center gap-2 cursor-pointer text-sm ${C.text}`}
                    >
                      <input
                        type="radio"
                        name="status"
                        checked={formData.status !== "inactive"}
                        onChange={() =>
                          setFormData({ ...formData, status: "active" })
                        }
                        className="w-4 h-4"
                      />
                      <span>有効</span>
                    </label>
                    <label
                      className={`flex items-center gap-2 cursor-pointer text-sm ${C.text}`}
                    >
                      <input
                        type="radio"
                        name="status"
                        checked={formData.status === "inactive"}
                        onChange={() =>
                          setFormData({ ...formData, status: "inactive" })
                        }
                        className="w-4 h-4"
                      />
                      <span>無効</span>
                    </label>
                  </div>
                </PropertyRow>
              </SectionBlock>
            </>
          ) : (
            // ── Generic edit form ──
            <>
              {showCode && (
                <PropertyRow label={labels.code} required>
                  <PropInput
                    value={formData.code ?? ""}
                    onChange={(v) => setFormData({ ...formData, code: v })}
                    placeholder={config?.codePlaceholder ?? "EX-001"}
                  />
                </PropertyRow>
              )}
              <PropertyRow label={labels.name} required>
                <PropInput
                  value={formData.name ?? ""}
                  onChange={(v) => setFormData({ ...formData, name: v })}
                  placeholder={config?.namePlaceholder ?? "名称"}
                />
              </PropertyRow>
              {showCategory && (
                <PropertyRow label={labels.category}>
                  <PropInput
                    value={formData.category ?? ""}
                    onChange={(v) => setFormData({ ...formData, category: v })}
                    placeholder="分類"
                  />
                </PropertyRow>
              )}
              {showPrice && (
                <PropertyRow label="単価(税込)">
                  <input
                    type="number"
                    value={formData.price ?? 0}
                    onChange={(e) =>
                      setFormData({ ...formData, price: Number(e.target.value) })
                    }
                    placeholder="0"
                    className={`${STYLE.propertyInput} w-32`}
                  />
                </PropertyRow>
              )}
              <PropertyRow label="備考 / 詳細">
                <PropInput
                  value={formData.description ?? ""}
                  onChange={(v) => setFormData({ ...formData, description: v })}
                  placeholder="補足情報など"
                />
              </PropertyRow>
              <PropertyRow label="ステータス">
                <div className="flex items-center gap-6">
                  <label
                    className={`flex items-center gap-2 cursor-pointer text-sm ${C.text}`}
                  >
                    <input
                      type="radio"
                      name="status"
                      checked={formData.status !== "inactive"}
                      onChange={() =>
                        setFormData({ ...formData, status: "active" })
                      }
                      className="w-4 h-4"
                    />
                    <span>有効</span>
                  </label>
                  <label
                    className={`flex items-center gap-2 cursor-pointer text-sm ${C.text}`}
                  >
                    <input
                      type="radio"
                      name="status"
                      checked={formData.status === "inactive"}
                      onChange={() =>
                        setFormData({ ...formData, status: "inactive" })
                      }
                      className="w-4 h-4"
                    />
                    <span>無効</span>
                  </label>
                </div>
              </PropertyRow>
            </>
          )}

          {/* Footer actions */}
          <div className={`flex justify-between pt-4 border-t ${C.borderLight}`}>
            {selectedItem ? (
              <Button
                variant="ghost"
                size="sm"
                onClick={handleDelete}
                className={`h-10 text-sm text-[#E03E3E] hover:bg-[#E03E3E]/10 hover:text-[#E03E3E]`}
              >
                <Trash2 className="mr-1.5 size-4" />
                削除
              </Button>
            ) : (
              <div />
            )}
            <div className="flex gap-2">
              <Button
                variant="outline"
                size="sm"
                onClick={handleCloseEdit}
                className={`h-10 text-sm bg-white ${C.text} ${C.borderMedium}`}
              >
                キャンセル
              </Button>
              <PrimaryButton onClick={handleSave} className="gap-2">
                <Save className="size-4" />
                保存
              </PrimaryButton>
            </div>
          </div>
        </div>
      </PageLayout>
    );
  }

  // ── List Columns ──────────────────────────────────
  const columns = isStaff
    ? [
        { header: "氏名" },
        { header: "職種", className: "w-[120px]" },
        { header: "所属医院", className: "w-[140px]" },
        { header: "メールアドレス", className: "w-[200px]" },
        { header: "ステータス", className: "w-[100px]", align: "center" as const },
        { header: "操作", className: "w-[80px]", align: "right" as const },
      ]
    : [
        ...(showCode
          ? [{ header: labels.code, className: "w-[120px]" }]
          : []),
        { header: labels.name },
        ...(showCategory
          ? [{ header: labels.category, className: "w-[120px]" }]
          : []),
        ...(showPrice
          ? [
              {
                header: "単価(税込)",
                className: "w-[110px]",
                align: "right" as const,
              },
            ]
          : []),
        { header: "ステータス", className: "w-[100px]", align: "center" as const },
        { header: "操作", className: "w-[80px]", align: "right" as const },
      ];

  // ── List content (shared between standalone and embedded) ──
  const listContent = (
    <div className="flex flex-col gap-4">
      <div className="flex items-center gap-3">
        <div className="flex-1 min-w-0">
          <SearchFilterBar
            searchTerm={searchTerm}
            onSearchChange={setSearchTerm}
            placeholder={`${labels.name}で検索...`}
            count={filteredItems.length}
          />
        </div>
        <PrimaryButton onClick={handleCreate}>
          <Plus className="mr-1.5 size-4" />
          新規登録
        </PrimaryButton>
      </div>

      <DataTable
        columns={columns}
        data={filteredItems}
        emptyMessage="データが見つかりません"
        renderRow={(item) =>
          isStaff ? (
            <DataTableRow key={item.id} onClick={() => handleEdit(item)}>
              <TableCell className={`font-medium text-sm ${C.text} py-2`}>
                {item.name}
              </TableCell>
              <TableCell className={`text-sm ${C.text} py-2`}>
                {item.category ?? "-"}
              </TableCell>
              <TableCell className={`text-sm ${C.text70} py-2`}>-</TableCell>
              <TableCell className={`text-sm ${C.text70} py-2`}>-</TableCell>
              <TableCell className="text-center py-2">
                <StatusBadge colorClass={getMasterStatusColor(item.status)}>
                  {item.status === "active" ? "有効" : "無効"}
                </StatusBadge>
              </TableCell>
              <TableCell className="text-right py-2">
                <RowActionButton onClick={() => handleEdit(item)} />
              </TableCell>
            </DataTableRow>
          ) : (
            <DataTableRow key={item.id} onClick={() => handleEdit(item)}>
              {showCode && (
                <TableCell className={`font-mono text-sm ${C.text80} py-2`}>
                  {item.code}
                </TableCell>
              )}
              <TableCell className={`font-medium text-sm ${C.text} py-2`}>
                {item.name}
              </TableCell>
              {showCategory && (
                <TableCell className={`text-sm ${C.text} py-2`}>
                  {item.category ?? "-"}
                </TableCell>
              )}
              {showPrice && (
                <TableCell className={`text-right font-mono text-sm ${C.text} py-2`}>
                  {item.price ? `¥${item.price.toLocaleString()}` : "-"}
                </TableCell>
              )}
              <TableCell className="text-center py-2">
                <StatusBadge colorClass={getMasterStatusColor(item.status)}>
                  {item.status === "active" ? "有効" : "無効"}
                </StatusBadge>
              </TableCell>
              <TableCell className="text-right py-2">
                <RowActionButton onClick={() => handleEdit(item)} />
              </TableCell>
            </DataTableRow>
          )
        }
      />
    </div>
  );

  // ── Embedded mode (for tabbed pages) ──────────
  if (embedded) {
    return listContent;
  }

  // ── Standalone list view ──────────────────────
  return (
    <PageLayout
      title={pageTitle}
      icon={<Settings2 className="size-5 text-[#37352F]" />}
      onBack={() => navigate("/settings")}
      maxWidth="max-w-full"
    >
      {listContent}
    </PageLayout>
  );
}
