// React/Framework
import { useCallback, memo } from "react";
import { useNavigate, useParams } from "react-router";

// External
import { Package, ArrowLeft, Save } from "lucide-react";

// Internal
import { C, ICON } from "@/lib/design-tokens";
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
import { SubmitButton } from "@/components/shared/Form/SubmitButton";
import { NavigationBlocker } from "@/components/shared/NavigationBlocker";
import { useUnsavedChanges } from "@/hooks/use-unsaved-changes";
import { useEffect } from "react";

// Relative
import { useInventoryForm } from "../hooks/use-inventory-form";
import { usePermission } from "@/features/auth";

// Types
import type { InventoryItem } from "@/types";

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
      <h3 className={`text-base font-medium ${C.text} mb-4`}>基本情報</h3>
      <div className="grid grid-cols-2 gap-4">
        <div className="col-span-2">
          <Label htmlFor="name" className={`text-sm ${C.text}`}>
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
          <Label htmlFor="category" className={`text-sm ${C.text}`}>
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
          <Label htmlFor="unit" className={`text-sm ${C.text}`}>
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
      <h3 className={`text-base font-medium ${C.text} mb-4`}>在庫情報</h3>
      <div className="grid grid-cols-2 gap-4">
        <div>
          <Label htmlFor="quantity" className={`text-sm ${C.text}`}>
            現在庫数 <span className="text-red-500">*</span>
          </Label>
          <Input
            id="quantity"
            name="quantity"
            type="number"
            min="0"
            step="1"
            defaultValue={defaultQuantity ?? 0}
            className="mt-1"
            required
          />
        </div>
        <div>
          <Label htmlFor="minStockLevel" className={`text-sm ${C.text}`}>
            最低在庫数 <span className="text-red-500">*</span>
          </Label>
          <Input
            id="minStockLevel"
            name="minStockLevel"
            type="number"
            min="0"
            step="1"
            defaultValue={defaultMinStockLevel ?? 0}
            className="mt-1"
            required
          />
        </div>
        <div>
          <Label htmlFor="location" className={`text-sm ${C.text}`}>
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
          <Label htmlFor="expiryDate" className={`text-sm ${C.text}`}>
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
      <h3 className={`text-base font-medium ${C.text} mb-4`}>仕入先情報</h3>
      <div className="grid grid-cols-2 gap-4">
        <div>
          <Label htmlFor="supplier" className={`text-sm ${C.text}`}>
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
          <Label htmlFor="lastRestocked" className={`text-sm ${C.text}`}>
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
  const { canEdit } = usePermission("inventory");

  const {
    isEdit,
    isLoading,
    existingItem,
    category,
    setCategory,
    resolvedExpiry,
    setExpiryDate,
    resolvedLastRestocked,
    setLastRestocked,
    formAction,
    formState,
    isPending,
  } = useInventoryForm(id);

  const { isDirty, markDirty, markClean } = useUnsavedChanges();

  // React 19 Action の成功を検知して遷移
  useEffect(() => {
    if (formState.success) {
      markClean();
      navigate(paths.inventory.getHref());
    }
  }, [formState.success, formState.timestamp, navigate, markClean]);

  const handleBack = useCallback(() => {
    navigate(paths.inventory.getHref());
  }, [navigate]);

  // rerender-functional-setstate: setCategory は stable setter なので useCallback 内 deps 不要
  const handleCategoryChange = useCallback((value: string) => {
    setCategory(value as InventoryItem["category"]);
  }, [setCategory]);

  // rerender-functional-setstate: setExpiryDate は stable setter なので useCallback 内 deps 不要
  const handleExpiryChange = useCallback((value: string) => {
    setExpiryDate(value);
  }, [setExpiryDate]);

  // rerender-functional-setstate: setLastRestocked は stable setter なので useCallback 内 deps 不要
  const handleLastRestockedChange = useCallback((value: string) => {
    setLastRestocked(value);
  }, [setLastRestocked]);

  if (isEdit && isLoading) {
    return (
      <PageLayout
        title="在庫編集"
        icon={<Package className={`${ICON.page} ${C.text}`} />}
        maxWidth="max-w-3xl"
      >
        <div className={`text-sm ${C.text60}`}>読み込み中...</div>
      </PageLayout>
    );
  }

  return (
    <PageLayout
      title={isEdit ? "在庫編集" : "在庫登録"}
      icon={<Package className={`${ICON.page} ${C.text}`} />}
      headerAction={
        <Button
          variant="ghost"
          type="button"
          className="h-10 text-sm gap-2"
          onClick={handleBack}
        >
          <ArrowLeft className={ICON.action} />
          一覧に戻る
        </Button>
      }
      maxWidth="max-w-3xl"
    >
      <NavigationBlocker when={isDirty && !isPending} />
      <form action={formAction} onChange={markDirty} className="space-y-6">
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
          {canEdit ? (
            <SubmitButton className="h-10">
              <Save className={`mr-1.5 ${ICON.action}`} />
              {isEdit ? "更新" : "登録"}
            </SubmitButton>
          ) : null}
        </div>
      </form>
    </PageLayout>
  );
}
