// React/Framework
import { useState, useCallback, useTransition, memo } from "react";
import { useNavigate, useParams } from "react-router";

// External
import { Package, ArrowLeft, Save } from "lucide-react";

// Internal
import { paths } from "@/config/paths";
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
import { PageLayout } from "@/components/shared/PageLayout/PageLayout";
import { NotionDatePicker } from "@/components/shared/NotionDatePicker/NotionDatePicker";
import { PrimaryButton } from "@/components/shared/Form/PrimaryButton";
import { NavigationBlocker } from "@/components/shared/NavigationBlocker";
import { useUnsavedChanges } from "@/hooks/use-unsaved-changes";
import { handleApiError } from "@/lib/handle-api-error";

// Relative
import {
  useGetInventoryItem,
  useCreateInventoryItem,
  useUpdateInventoryItem,
} from "../api/inventory";

// Types
import type { InventoryItem } from "@/types";
import type {
  CreateInventoryItemRequest,
  UpdateInventoryItemRequest,
} from "../api/types";

const CATEGORY_OPTIONS: { value: InventoryItem["category"]; label: string }[] =
  [
    { value: "medicine", label: "医薬品" },
    { value: "consumable", label: "消耗品" },
    { value: "food", label: "フード" },
    { value: "other", label: "その他" },
  ];

// ─── BasicInfoSection ────────────────────────────────────────────────────────

interface BasicInfoSectionProps {
  defaultName: string | undefined;
  defaultUnit: string | undefined;
  category: InventoryItem["category"];
  existingCategory: InventoryItem["category"] | undefined;
  onCategoryChange: (value: string) => void;
  onMarkDirty: () => void;
}

// rerender-memo: 基本情報セクションを memo 化してカテゴリ以外の状態変更による
// 不要な再レンダーを防ぐ（onCategoryChange / onMarkDirty は useCallback 済み）
const BasicInfoSection = memo(function BasicInfoSection({
  defaultName,
  defaultUnit,
  category,
  existingCategory,
  onCategoryChange,
  onMarkDirty,
}: BasicInfoSectionProps) {
  return (
    <div className="bg-white rounded-lg border border-[rgba(55,53,47,0.16)] p-6">
      <h3 className="text-base font-medium text-[#37352F] mb-4">基本情報</h3>
      <div className="grid grid-cols-2 gap-4">
        <div className="col-span-2">
          <Label htmlFor="name" className="text-sm text-[#37352F]">
            品名 <span className="text-red-500">*</span>
          </Label>
          <Input
            id="name"
            name="name"
            defaultValue={defaultName}
            placeholder="品名を入力"
            className="mt-1"
            required
          />
        </div>
        <div>
          <Label htmlFor="category" className="text-sm text-[#37352F]">
            カテゴリ <span className="text-red-500">*</span>
          </Label>
          <Select
            value={category || (existingCategory ?? "medicine")}
            onValueChange={(v) => {
              onMarkDirty();
              onCategoryChange(v);
            }}
          >
            <SelectTrigger className="mt-1">
              <SelectValue placeholder="カテゴリを選択" />
            </SelectTrigger>
            <SelectContent>
              {CATEGORY_OPTIONS.map((opt) => (
                <SelectItem key={opt.value} value={opt.value}>
                  {opt.label}
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
        </div>
        <div>
          <Label htmlFor="unit" className="text-sm text-[#37352F]">
            単位 <span className="text-red-500">*</span>
          </Label>
          <Input
            id="unit"
            name="unit"
            defaultValue={defaultUnit}
            placeholder="例: 錠, 本, 袋"
            className="mt-1"
            required
          />
        </div>
      </div>
    </div>
  );
});

// ─── StockInfoSection ─────────────────────────────────────────────────────────

interface StockInfoSectionProps {
  defaultQuantity: number | undefined;
  defaultMinStockLevel: number | undefined;
  defaultLocation: string | undefined;
  resolvedExpiry: string;
  onExpiryChange: (v: string) => void;
  onMarkDirty: () => void;
}

// rerender-memo: 在庫情報セクションを memo 化して仕入先情報や基本情報の変更による
// 不要な再レンダーを防ぐ（onExpiryChange / onMarkDirty は useCallback 済み）
const StockInfoSection = memo(function StockInfoSection({
  defaultQuantity,
  defaultMinStockLevel,
  defaultLocation,
  resolvedExpiry,
  onExpiryChange,
  onMarkDirty,
}: StockInfoSectionProps) {
  return (
    <div className="bg-white rounded-lg border border-[rgba(55,53,47,0.16)] p-6">
      <h3 className="text-base font-medium text-[#37352F] mb-4">在庫情報</h3>
      <div className="grid grid-cols-2 gap-4">
        <div>
          <Label htmlFor="quantity" className="text-sm text-[#37352F]">
            現在庫数 <span className="text-red-500">*</span>
          </Label>
          <Input
            id="quantity"
            name="quantity"
            type="number"
            min="0"
            defaultValue={defaultQuantity ?? 0}
            className="mt-1"
            required
          />
        </div>
        <div>
          <Label htmlFor="minStockLevel" className="text-sm text-[#37352F]">
            最低在庫数 <span className="text-red-500">*</span>
          </Label>
          <Input
            id="minStockLevel"
            name="minStockLevel"
            type="number"
            min="0"
            defaultValue={defaultMinStockLevel ?? 0}
            className="mt-1"
            required
          />
        </div>
        <div>
          <Label htmlFor="location" className="text-sm text-[#37352F]">
            保管場所
          </Label>
          <Input
            id="location"
            name="location"
            defaultValue={defaultLocation}
            placeholder="例: 薬品棚A-1"
            className="mt-1"
          />
        </div>
        <div>
          <Label htmlFor="expiryDate" className="text-sm text-[#37352F]">
            有効期限
          </Label>
          <input type="hidden" name="expiryDate" value={resolvedExpiry} />
          <NotionDatePicker
            id="expiryDate"
            value={resolvedExpiry}
            onChange={(v) => {
              onMarkDirty();
              onExpiryChange(v);
            }}
            placeholder="有効期限を選択…"
            className="mt-1"
          />
        </div>
      </div>
    </div>
  );
});

// ─── SupplierInfoSection ──────────────────────────────────────────────────────

interface SupplierInfoSectionProps {
  defaultSupplier: string | undefined;
  resolvedLastRestocked: string;
  onLastRestockedChange: (v: string) => void;
  onMarkDirty: () => void;
}

// rerender-memo: 仕入先情報セクションを memo 化して他セクションの変更による
// 不要な再レンダーを防ぐ（onLastRestockedChange / onMarkDirty は useCallback 済み）
const SupplierInfoSection = memo(function SupplierInfoSection({
  defaultSupplier,
  resolvedLastRestocked,
  onLastRestockedChange,
  onMarkDirty,
}: SupplierInfoSectionProps) {
  return (
    <div className="bg-white rounded-lg border border-[rgba(55,53,47,0.16)] p-6">
      <h3 className="text-base font-medium text-[#37352F] mb-4">仕入先情報</h3>
      <div className="grid grid-cols-2 gap-4">
        <div>
          <Label htmlFor="supplier" className="text-sm text-[#37352F]">
            仕入先
          </Label>
          <Input
            id="supplier"
            name="supplier"
            defaultValue={defaultSupplier}
            placeholder="仕入先名"
            className="mt-1"
          />
        </div>
        <div>
          <Label htmlFor="lastRestocked" className="text-sm text-[#37352F]">
            最終入荷日
          </Label>
          <input
            type="hidden"
            name="lastRestocked"
            value={resolvedLastRestocked}
          />
          <NotionDatePicker
            id="lastRestocked"
            value={resolvedLastRestocked}
            onChange={(v) => {
              onMarkDirty();
              onLastRestockedChange(v);
            }}
            placeholder="最終入荷日を選択…"
            className="mt-1"
          />
        </div>
      </div>
    </div>
  );
});

// ─── InventoryForm ────────────────────────────────────────────────────────────

export function InventoryForm() {
  const navigate = useNavigate();
  const { id } = useParams<{ id: string }>();
  const isEdit = Boolean(id);

  const { isDirty, markDirty, markClean } = useUnsavedChanges();

  const { data: existingItem, isLoading } = useGetInventoryItem(id ?? "");
  const createMutation = useCreateInventoryItem();
  const updateMutation = useUpdateInventoryItem();

  // rerender-transitions: API 書き込みの pending 状態を useTransition で管理
  // useState(false) + setIsPending パターンを禁止し try-finally 漏れを防ぐ
  const [isSavePending, startSaveTransition] = useTransition();

  const [category, setCategory] = useState<InventoryItem["category"]>(
    (existingItem?.category as InventoryItem["category"]) ?? "medicine"
  );
  const [expiryDate, setExpiryDate] = useState("");
  const [lastRestocked, setLastRestocked] = useState("");
  const resolvedExpiry =
    expiryDate ||
    (existingItem?.expiry_date ? existingItem.expiry_date.slice(0, 10) : "");
  const resolvedLastRestocked =
    lastRestocked ||
    (existingItem?.last_restocked
      ? existingItem.last_restocked.slice(0, 10)
      : "");

  const handleBack = useCallback(() => {
    navigate(paths.inventory.getHref());
  }, [navigate]);

  // rerender-functional-setstate: setCategory は stable setter なので useCallback 内 deps 不要
  const handleCategoryChange = useCallback((value: string) => {
    setCategory(value as InventoryItem["category"]);
  }, []);

  // rerender-functional-setstate: setExpiryDate は stable setter なので useCallback 内 deps 不要
  const handleExpiryChange = useCallback((value: string) => {
    setExpiryDate(value);
  }, []);

  // rerender-functional-setstate: setLastRestocked は stable setter なので useCallback 内 deps 不要
  const handleLastRestockedChange = useCallback((value: string) => {
    setLastRestocked(value);
  }, []);

  const handleSubmit = useCallback(
    (e: React.FormEvent<HTMLFormElement>) => {
      e.preventDefault();
      const formData = new FormData(e.currentTarget);

      const quantityStr = formData.get("quantity") as string;
      const minStockLevelStr = formData.get("minStockLevel") as string;
      const expiryDateStr = formData.get("expiryDate") as string;
      const lastRestockedStr = formData.get("lastRestocked") as string;
      const resolvedCategory = category || "medicine";

      // rerender-transitions: mutateAsync を useTransition で包んで pending を管理
      startSaveTransition(async () => {
        try {
          if (isEdit && id) {
            const req: UpdateInventoryItemRequest = {
              name: formData.get("name") as string,
              category: resolvedCategory,
              quantity: quantityStr ? Number(quantityStr) : undefined,
              unit: formData.get("unit") as string,
              min_stock_level: minStockLevelStr
                ? Number(minStockLevelStr)
                : undefined,
              location: (formData.get("location") as string) || undefined,
              expiry_date: expiryDateStr || undefined,
              supplier: (formData.get("supplier") as string) || undefined,
              last_restocked: lastRestockedStr || undefined,
            };
            await updateMutation.mutateAsync({ id, req });
          } else {
            const req: CreateInventoryItemRequest = {
              name: formData.get("name") as string,
              category: resolvedCategory,
              quantity: quantityStr ? Number(quantityStr) : 0,
              unit: formData.get("unit") as string,
              min_stock_level: minStockLevelStr
                ? Number(minStockLevelStr)
                : 0,
              location: (formData.get("location") as string) || undefined,
              expiry_date: expiryDateStr || undefined,
              supplier: (formData.get("supplier") as string) || undefined,
            };
            await createMutation.mutateAsync(req);
          }
          markClean();
          navigate(paths.inventory.getHref());
        } catch (error) {
          handleApiError(error, "保存");
        }
      });
    },
    // rerender-dependencies: オブジェクト全体でなく primitive を deps に入れる
    [navigate, isEdit, id, category, createMutation, updateMutation, markClean]
  );

  if (isEdit && isLoading) {
    return (
      <PageLayout
        title="在庫編集"
        icon={<Package className="size-5 text-[#37352F]" />}
        maxWidth="max-w-3xl"
      >
        <div className="text-sm text-[#37352F]/60">読み込み中...</div>
      </PageLayout>
    );
  }

  return (
    <PageLayout
      title={isEdit ? "在庫編集" : "在庫登録"}
      icon={<Package className="size-5 text-[#37352F]" />}
      headerAction={
        <Button
          variant="ghost"
          className="h-10 text-sm gap-2"
          onClick={handleBack}
        >
          <ArrowLeft className="size-4" />
          一覧に戻る
        </Button>
      }
      maxWidth="max-w-3xl"
    >
      <NavigationBlocker when={isDirty && !isSavePending} />
      <form onSubmit={handleSubmit} onChange={markDirty} className="space-y-6">
        <BasicInfoSection
          defaultName={existingItem?.name}
          defaultUnit={existingItem?.unit}
          category={category}
          existingCategory={existingItem?.category as InventoryItem["category"]}
          onCategoryChange={handleCategoryChange}
          onMarkDirty={markDirty}
        />

        <StockInfoSection
          defaultQuantity={existingItem?.quantity}
          defaultMinStockLevel={existingItem?.min_stock_level}
          defaultLocation={existingItem?.location}
          resolvedExpiry={resolvedExpiry}
          onExpiryChange={handleExpiryChange}
          onMarkDirty={markDirty}
        />

        <SupplierInfoSection
          defaultSupplier={existingItem?.supplier}
          resolvedLastRestocked={resolvedLastRestocked}
          onLastRestockedChange={handleLastRestockedChange}
          onMarkDirty={markDirty}
        />

        {/* Actions */}
        <div className="flex justify-end gap-3">
          <Button
            type="button"
            variant="outline"
            className="h-10"
            onClick={handleBack}
          >
            キャンセル
          </Button>
          <PrimaryButton type="submit" disabled={isSavePending}>
            <Save className="mr-1.5 size-4" />
            {isEdit ? "更新" : "登録"}
          </PrimaryButton>
        </div>
      </form>
    </PageLayout>
  );
}
