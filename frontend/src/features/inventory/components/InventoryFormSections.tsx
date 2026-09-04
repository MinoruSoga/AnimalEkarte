import { memo, useState } from "react";
import { AlertTriangle } from "lucide-react";
import { FormFieldError } from "@/components/shared/FormFieldError/FormFieldError";
import { DatePicker } from "@/components/shared/DatePicker/DatePicker";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { C } from "@/lib/design-tokens";
import type { InventoryItem } from "@/types";

const CATEGORY_OPTIONS: { value: InventoryItem["category"]; label: string }[] = [
  { value: "medicine", label: "医薬品" },
  { value: "consumable", label: "消耗品" },
  { value: "food", label: "フード" },
  { value: "other", label: "その他" },
];

interface BasicInfoSectionProps {
  defaultName: string | undefined;
  defaultUnit: string | undefined;
  category: InventoryItem["category"];
  existingCategory: InventoryItem["category"] | undefined;
  onCategoryChange: (value: InventoryItem["category"]) => void;
  onMarkDirty: () => void;
  nameError?: string;
  unitError?: string;
}

export const BasicInfoSection = memo(function BasicInfoSection({
  defaultName,
  defaultUnit,
  category,
  existingCategory,
  onCategoryChange,
  onMarkDirty,
  nameError,
  unitError,
}: BasicInfoSectionProps) {
  const [name, setName] = useState(defaultName ?? "");
  const [unit, setUnit] = useState(defaultUnit ?? "");
  return (
    <div className={`${C.bgWhite} rounded-lg border ${C.borderLight} p-6`}>
      <h3 className={`text-base font-medium ${C.text} mb-4`}>基本情報</h3>
      <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
        <div className="sm:col-span-2">
          <Label htmlFor="name" className={`text-sm ${C.text}`}>
            品名 <span className={C.textRequired}>*</span>
          </Label>
          <Input
            id="name"
            name="name"
            value={name}
            onChange={(event) => {
              setName(event.target.value);
              onMarkDirty();
            }}
            placeholder="品名を入力"
            className={`mt-1 ${C.bgWhite} ${C.borderMedium}`}
            aria-invalid={nameError ? true : undefined}
            aria-describedby={nameError ? "name-error" : undefined}
            aria-required="true"
          />
          <FormFieldError id="name-error" message={nameError} />
        </div>
        <div>
          <Label htmlFor="category" className={`text-sm ${C.text}`}>
            カテゴリ <span className={C.textRequired}>*</span>
          </Label>
          <Select
            value={category || (existingCategory ?? "medicine")}
            onValueChange={(value) => {
              const option = CATEGORY_OPTIONS.find((entry) => entry.value === value);
              if (!option) return;
              onMarkDirty();
              onCategoryChange(option.value);
            }}
          >
            <SelectTrigger id="category" className={`mt-1 ${C.bgWhite} ${C.borderMedium}`}>
              <SelectValue placeholder="カテゴリを選択" />
            </SelectTrigger>
            <SelectContent>
              {CATEGORY_OPTIONS.map((option) => (
                <SelectItem key={option.value} value={option.value}>
                  {option.label}
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
        </div>
        <div>
          <Label htmlFor="unit" className={`text-sm ${C.text}`}>
            単位 <span className={C.textRequired}>*</span>
          </Label>
          <Input
            id="unit"
            name="unit"
            value={unit}
            onChange={(event) => {
              setUnit(event.target.value);
              onMarkDirty();
            }}
            placeholder="例: 錠, 本, 袋"
            className={`mt-1 ${C.bgWhite} ${C.borderMedium}`}
            aria-invalid={unitError ? true : undefined}
            aria-describedby={unitError ? "unit-error" : undefined}
            aria-required="true"
          />
          <FormFieldError id="unit-error" message={unitError} />
        </div>
      </div>
    </div>
  );
});

interface StockInfoSectionProps {
  defaultQuantity: number | undefined;
  defaultMinStockLevel: number | undefined;
  defaultLocation: string | undefined;
  resolvedExpiry: string;
  onExpiryChange: (value: string) => void;
  onMarkDirty: () => void;
  quantityError?: string;
  minStockLevelError?: string;
}

export const StockInfoSection = memo(function StockInfoSection({
  defaultQuantity,
  defaultMinStockLevel,
  defaultLocation,
  resolvedExpiry,
  onExpiryChange,
  onMarkDirty,
  quantityError,
  minStockLevelError,
}: StockInfoSectionProps) {
  const [quantity, setQuantity] = useState(String(defaultQuantity ?? 0));
  const [minStockLevel, setMinStockLevel] = useState(String(defaultMinStockLevel ?? 0));
  const parsedQuantity = Number(quantity);
  const parsedMinStockLevel = Number(minStockLevel);
  const stockStatus =
    parsedQuantity <= 0
      ? `在庫切れ — 現在庫数 ${parsedQuantity || 0}、最低在庫数 ${parsedMinStockLevel || 0}`
      : parsedMinStockLevel > 0 && parsedQuantity <= parsedMinStockLevel
        ? `在庫不足（残少）— 現在庫数 ${parsedQuantity}、最低在庫数 ${parsedMinStockLevel}`
        : null;

  return (
    <div className={`${C.bgWhite} rounded-lg border ${C.borderLight} p-6`}>
      <h3 className={`text-base font-medium ${C.text} mb-4`}>在庫情報</h3>
      {stockStatus ? (
        <div
          role="status"
          aria-label="在庫状態"
          className={`mb-4 flex items-center gap-2 text-sm font-medium ${C.danger}`}
        >
          <AlertTriangle className="h-4 w-4 shrink-0" aria-hidden="true" />
          <span>{stockStatus}</span>
        </div>
      ) : null}
      <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
        <div>
          <Label htmlFor="quantity" className={`text-sm ${C.text}`}>
            現在庫数 <span className={C.textRequired}>*</span>
          </Label>
          <Input
            id="quantity"
            name="quantity"
            type="number"
            step="1"
            value={quantity}
            onChange={(event) => {
              setQuantity(event.target.value);
              onMarkDirty();
            }}
            className={`mt-1 ${C.bgWhite} ${C.borderMedium}`}
            aria-invalid={quantityError ? true : undefined}
            aria-describedby={quantityError ? "quantity-error" : undefined}
            aria-required="true"
          />
          <FormFieldError id="quantity-error" message={quantityError} />
        </div>
        <div>
          <Label htmlFor="minStockLevel" className={`text-sm ${C.text}`}>
            最低在庫数 <span className={C.textRequired}>*</span>
          </Label>
          <Input
            id="minStockLevel"
            name="minStockLevel"
            type="number"
            step="1"
            value={minStockLevel}
            onChange={(event) => {
              setMinStockLevel(event.target.value);
              onMarkDirty();
            }}
            className={`mt-1 ${C.bgWhite} ${C.borderMedium}`}
            aria-invalid={minStockLevelError ? true : undefined}
            aria-describedby={minStockLevelError ? "minStockLevel-error" : undefined}
            aria-required="true"
          />
          <FormFieldError id="minStockLevel-error" message={minStockLevelError} />
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
            className={`mt-1 ${C.bgWhite} ${C.borderMedium}`}
          />
        </div>
        <div>
          <Label htmlFor="expiryDate" className={`text-sm ${C.text}`}>
            有効期限
          </Label>
          <input type="hidden" name="expiryDate" value={resolvedExpiry} />
          <DatePicker
            id="expiryDate"
            value={resolvedExpiry}
            onChange={(value) => {
              onMarkDirty();
              onExpiryChange(value);
            }}
            placeholder="有効期限を選択…"
            className={`mt-1 ${C.bgWhite} ${C.borderMedium}`}
          />
        </div>
      </div>
    </div>
  );
});

interface SupplierInfoSectionProps {
  defaultSupplier: string | undefined;
  resolvedLastRestocked: string;
  onLastRestockedChange: (value: string) => void;
  onMarkDirty: () => void;
}

export const SupplierInfoSection = memo(function SupplierInfoSection({
  defaultSupplier,
  resolvedLastRestocked,
  onLastRestockedChange,
  onMarkDirty,
}: SupplierInfoSectionProps) {
  return (
    <div className={`${C.bgWhite} rounded-lg border ${C.borderLight} p-6`}>
      <h3 className={`text-base font-medium ${C.text} mb-4`}>仕入先情報</h3>
      <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
        <div>
          <Label htmlFor="supplier" className={`text-sm ${C.text}`}>
            仕入先
          </Label>
          <Input
            id="supplier"
            name="supplier"
            defaultValue={defaultSupplier}
            placeholder="仕入先名"
            className={`mt-1 ${C.bgWhite} ${C.borderMedium}`}
          />
        </div>
        <div>
          <Label htmlFor="lastRestocked" className={`text-sm ${C.text}`}>
            最終入荷日
          </Label>
          <input type="hidden" name="lastRestocked" value={resolvedLastRestocked} />
          <DatePicker
            id="lastRestocked"
            value={resolvedLastRestocked}
            onChange={(value) => {
              onMarkDirty();
              onLastRestockedChange(value);
            }}
            placeholder="最終入荷日を選択…"
            className={`mt-1 ${C.bgWhite} ${C.borderMedium}`}
          />
        </div>
      </div>
    </div>
  );
});
