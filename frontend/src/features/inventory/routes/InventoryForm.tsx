// React/Framework
import { useNavigate, useParams } from "react-router";

// External
import { Package, ArrowLeft, Save } from "lucide-react";

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
import { PageLayout } from "@/components/shared/PageLayout";
import { PrimaryButton } from "@/components/shared/Form";

// Relative
import { MOCK_INVENTORY_ITEMS } from "../api";

// Types
import type { InventoryItem } from "@/types";

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

  // Get existing item for edit mode
  const existingItem = isEdit
    ? MOCK_INVENTORY_ITEMS.find((item) => item.id === id)
    : null;

  const handleBack = () => {
    navigate("/inventory");
  };

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    // TODO: Implement actual save logic
    navigate("/inventory");
  };

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
      <form onSubmit={handleSubmit} className="space-y-6">
        {/* Basic Info */}
        <div className="bg-white rounded-lg border border-[rgba(55,53,47,0.16)] p-6">
          <h3 className="text-base font-medium text-[#37352F] mb-4">基本情報</h3>
          <div className="grid grid-cols-2 gap-4">
            <div className="col-span-2">
              <Label htmlFor="name" className="text-sm text-[#37352F]">
                品名 <span className="text-red-500">*</span>
              </Label>
              <Input
                id="name"
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
              <Select defaultValue={existingItem?.category ?? "medicine"}>
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
          <h3 className="text-base font-medium text-[#37352F] mb-4">在庫情報</h3>
          <div className="grid grid-cols-2 gap-4">
            <div>
              <Label htmlFor="quantity" className="text-sm text-[#37352F]">
                現在庫数 <span className="text-red-500">*</span>
              </Label>
              <Input
                id="quantity"
                type="number"
                min="0"
                defaultValue={existingItem?.quantity ?? 0}
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
                type="number"
                min="0"
                defaultValue={existingItem?.minStockLevel ?? 0}
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
                type="date"
                defaultValue={existingItem?.expiryDate?.replace(/\//g, "-")}
                className="mt-1"
              />
            </div>
          </div>
        </div>

        {/* Supplier Info */}
        <div className="bg-white rounded-lg border border-[rgba(55,53,47,0.16)] p-6">
          <h3 className="text-base font-medium text-[#37352F] mb-4">仕入先情報</h3>
          <div className="grid grid-cols-2 gap-4">
            <div>
              <Label htmlFor="supplier" className="text-sm text-[#37352F]">
                仕入先
              </Label>
              <Input
                id="supplier"
                defaultValue={existingItem?.supplier}
                placeholder="仕入先名"
                className="mt-1"
              />
            </div>
            <div>
              <Label htmlFor="lastRestocked" className="text-sm text-[#37352F]">
                最終入荷日
              </Label>
              <Input
                id="lastRestocked"
                type="date"
                defaultValue={existingItem?.lastRestocked?.replace(/\//g, "-")}
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
          <PrimaryButton type="submit">
            <Save className="mr-1.5 size-4" />
            {isEdit ? "更新" : "登録"}
          </PrimaryButton>
        </div>
      </form>
    </PageLayout>
  );
}
