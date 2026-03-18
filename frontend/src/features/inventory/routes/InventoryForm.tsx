// React/Framework
import { useState, useCallback } from "react";
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
import { PrimaryButton } from "@/components/shared/Form/PrimaryButton";
import { NavigationBlocker } from "@/components/shared/NavigationBlocker";
import { useUnsavedChanges } from "@/hooks/use-unsaved-changes";

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

export function InventoryForm() {
  const navigate = useNavigate();
  const { id } = useParams<{ id: string }>();
  const isEdit = Boolean(id);

  const { isDirty, markDirty, markClean } = useUnsavedChanges();

  const { data: existingItem, isLoading } = useGetInventoryItem(id ?? "");
  const createMutation = useCreateInventoryItem();
  const updateMutation = useUpdateInventoryItem();

  const [category, setCategory] = useState<string>(
    existingItem?.category ?? "medicine"
  );

  const handleBack = useCallback(() => {
    navigate(paths.inventory.getHref());
  }, [navigate]);

  const handleSubmit = useCallback((e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault();
    const formData = new FormData(e.currentTarget);

    const quantityStr = formData.get("quantity") as string;
    const minStockLevelStr = formData.get("minStockLevel") as string;
    const expiryDateStr = formData.get("expiryDate") as string;
    const lastRestockedStr = formData.get("lastRestocked") as string;
    const resolvedCategory = category || "medicine";

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
      updateMutation.mutate(
        { id, req },
        { onSuccess: () => { markClean(); navigate(paths.inventory.getHref()); } }
      );
    } else {
      const req: CreateInventoryItemRequest = {
        name: formData.get("name") as string,
        category: resolvedCategory,
        quantity: quantityStr ? Number(quantityStr) : 0,
        unit: formData.get("unit") as string,
        min_stock_level: minStockLevelStr ? Number(minStockLevelStr) : 0,
        location: (formData.get("location") as string) || undefined,
        expiry_date: expiryDateStr || undefined,
        supplier: (formData.get("supplier") as string) || undefined,
      };
      createMutation.mutate(req, {
        onSuccess: () => { markClean(); navigate(paths.inventory.getHref()); },
      });
    }
  }, [navigate, isEdit, id, category, createMutation, updateMutation, markClean]);

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

  const expiryDateValue = existingItem?.expiry_date
    ? existingItem.expiry_date.slice(0, 10)
    : "";
  const lastRestockedValue = existingItem?.last_restocked
    ? existingItem.last_restocked.slice(0, 10)
    : "";

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
      <NavigationBlocker when={isDirty && !createMutation.isPending && !updateMutation.isPending} />
      <form onSubmit={handleSubmit} onChange={markDirty} className="space-y-6">
        {/* Basic Info */}
        <div className="bg-white rounded-lg border border-[rgba(55,53,47,0.16)] p-6">
          <h3 className="text-base font-medium text-[#37352F] mb-4">
            基本情報
          </h3>
          <div className="grid grid-cols-2 gap-4">
            <div className="col-span-2">
              <Label htmlFor="name" className="text-sm text-[#37352F]">
                品名 <span className="text-red-500">*</span>
              </Label>
              <Input
                id="name"
                name="name"
                defaultValue={existingItem?.name}
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
                value={category || (existingItem?.category ?? "medicine")}
                onValueChange={(v) => { markDirty(); setCategory(v); }}
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
                defaultValue={existingItem?.unit}
                placeholder="例: 錠, 本, 袋"
                className="mt-1"
                required
              />
            </div>
          </div>
        </div>

        {/* Stock Info */}
        <div className="bg-white rounded-lg border border-[rgba(55,53,47,0.16)] p-6">
          <h3 className="text-base font-medium text-[#37352F] mb-4">
            在庫情報
          </h3>
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
                defaultValue={existingItem?.quantity ?? 0}
                className="mt-1"
                required
              />
            </div>
            <div>
              <Label
                htmlFor="minStockLevel"
                className="text-sm text-[#37352F]"
              >
                最低在庫数 <span className="text-red-500">*</span>
              </Label>
              <Input
                id="minStockLevel"
                name="minStockLevel"
                type="number"
                min="0"
                defaultValue={existingItem?.min_stock_level ?? 0}
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
                defaultValue={existingItem?.location}
                placeholder="例: 薬品棚A-1"
                className="mt-1"
              />
            </div>
            <div>
              <Label htmlFor="expiryDate" className="text-sm text-[#37352F]">
                有効期限
              </Label>
              <Input
                id="expiryDate"
                name="expiryDate"
                type="date"
                defaultValue={expiryDateValue}
                className="mt-1"
              />
            </div>
          </div>
        </div>

        {/* Supplier Info */}
        <div className="bg-white rounded-lg border border-[rgba(55,53,47,0.16)] p-6">
          <h3 className="text-base font-medium text-[#37352F] mb-4">
            仕入先情報
          </h3>
          <div className="grid grid-cols-2 gap-4">
            <div>
              <Label htmlFor="supplier" className="text-sm text-[#37352F]">
                仕入先
              </Label>
              <Input
                id="supplier"
                name="supplier"
                defaultValue={existingItem?.supplier}
                placeholder="仕入先名"
                className="mt-1"
              />
            </div>
            <div>
              <Label
                htmlFor="lastRestocked"
                className="text-sm text-[#37352F]"
              >
                最終入荷日
              </Label>
              <Input
                id="lastRestocked"
                name="lastRestocked"
                type="date"
                defaultValue={lastRestockedValue}
                className="mt-1"
              />
            </div>
          </div>
        </div>

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
          <PrimaryButton
            type="submit"
            disabled={createMutation.isPending || updateMutation.isPending}
          >
            <Save className="mr-1.5 size-4" />
            {isEdit ? "更新" : "登録"}
          </PrimaryButton>
        </div>
      </form>
    </PageLayout>
  );
}
