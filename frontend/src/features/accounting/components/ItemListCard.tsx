import { memo, useCallback, useDeferredValue, useMemo, useState } from "react";
import { Plus } from "lucide-react";

import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Dialog, DialogContent, DialogDescription, DialogHeader, DialogTitle, DialogTrigger } from "@/components/ui/dialog";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { EmptyState } from "@/components/shared/DataStates";
import { DataTableRowButton } from "@/components/shared/DataTable/DataTableRowButton";
import { FormFieldError } from "@/components/shared/FormFieldError/FormFieldError";
import { C, ICON, LAYOUT } from "@/lib/design-tokens";
import { DEFAULT_STANDARD_TAX_RATE, DEFAULT_REDUCED_TAX_RATE } from "@/constants/tax";
import { formatCurrency } from "@/lib/format/number";
import { normalizeKana } from "@/lib/normalize-kana";
import { CATEGORY_LABELS } from "@/constants/item-category";
import type { TaxType } from "@/types/generated/models";

import { useGetAllMerchandiseItems } from "../api/get-merchandise-items";
import type { FrontendMerchandiseItem } from "../api/get-merchandise-items";
import type { AccountingItem, AddAccountingItemInput, ItemCategory } from "../types";

import { AccountingItemRow } from "./AccountingItemRow";

const MERCHANDISE_CATEGORY_OPTIONS = [
  { value: "all", label: "すべて" },
  { value: "food", label: "フード" },
  { value: "goods", label: "物販" },
  { value: "other", label: "その他" },
];

const MERCHANDISE_CATEGORY_SELECT_ITEMS = MERCHANDISE_CATEGORY_OPTIONS.map((o) => (
  <SelectItem key={o.value} value={o.value}>{o.label}</SelectItem>
));

const MANUAL_CATEGORY_SELECT_ITEMS = Object.entries(CATEGORY_LABELS).map(([value, label]) => (
  <SelectItem key={value} value={value}>{label}</SelectItem>
));

// FE-RC-049: 単価カラム(bigint 系)の実務上の妥当な上限。DB 制約ではなく入力ミス検出用。
const MAX_MANUAL_ITEM_UNIT_PRICE = 999_999_999;
// FE-RC-049: 「その他」区分の理由欄の最大文字数。
const MANUAL_OTHER_REASON_MAX_LENGTH = 500;

function getManualPriceError(value: string): string | null {
  const priceNum = parseInt(value, 10);
  if (isNaN(priceNum) || priceNum < 0) {
    return "単価は0以上の整数で入力してください";
  }
  if (priceNum > MAX_MANUAL_ITEM_UNIT_PRICE) {
    return "単価は999,999,999円以下で入力してください";
  }
  return null;
}

interface ItemListCardProps {
  items: AccountingItem[];
  subtotal: number;
  taxTotal: number;
  totalAmount: number;
  newItemOpen: boolean;
  onNewItemOpenChange: (open: boolean) => void;
  onAddItem: (input: AddAccountingItemInput) => void;
  onDeleteItem: (id: string) => void;
  accountingId?: string;
  onUpdateItemTax?: (itemId: string, taxType: TaxType, taxRate: number) => void;
  onUpdateItemDiscount?: (itemId: string, discountAmount: number) => void;
  canEdit: boolean;
  canDelete: boolean;
}

export const ItemListCard = memo(function ItemListCard({
  items,
  subtotal,
  taxTotal,
  totalAmount,
  newItemOpen,
  onNewItemOpenChange,
  onAddItem,
  onDeleteItem,
  accountingId,
  onUpdateItemTax,
  onUpdateItemDiscount,
  canEdit,
  canDelete,
}: ItemListCardProps) {
  const [categoryFilter, setCategoryFilter] = useState("all");
  const [merchandiseSearch, setMerchandiseSearch] = useState("");
  const deferredMerchandiseSearch = useDeferredValue(merchandiseSearch);
  const [addMode, setAddMode] = useState<"master" | "manual">("master");
  const [manualName, setManualName] = useState("");
  const [manualPrice, setManualPrice] = useState("");
  const [manualCategory, setManualCategory] = useState("");
  const [manualOtherReason, setManualOtherReason] = useState("");
  const [manualPriceError, setManualPriceError] = useState("");

  const { data: merchandiseItems = [] } = useGetAllMerchandiseItems();

  const filteredMerchandise = useMemo(() => {
    let result = merchandiseItems.filter((item) => item.isActive);
    if (categoryFilter !== "all") {
      result = result.filter((item) => item.category === categoryFilter);
    }
    if (deferredMerchandiseSearch) {
      const normalizedTerm = normalizeKana(deferredMerchandiseSearch).toLowerCase();
      result = result.filter((item) => normalizeKana(item.name).toLowerCase().includes(normalizedTerm));
    }
    return result;
  }, [merchandiseItems, categoryFilter, deferredMerchandiseSearch]);

  const handleSelectMerchandise = useCallback(
    (item: FrontendMerchandiseItem) => {
      onAddItem({
        name: item.name,
        price: String(item.unitPrice),
        category: item.category,
        taxRate: item.taxRate,
        merchandiseItemId: item.id,
      });
    },
    [onAddItem],
  );

  const handleManualCategoryChange = useCallback((value: string) => {
    setManualCategory(value);
    if (value !== "other") {
      setManualOtherReason("");
    }
  }, []);

  const handleAddManualItem = useCallback(() => {
    const name = manualName.trim();
    const otherReason = manualOtherReason.trim();
    if (!name || !manualPrice || !manualCategory || (manualCategory === "other" && !otherReason)) {
      return;
    }

    const priceError = getManualPriceError(manualPrice);
    if (priceError !== null) {
      setManualPriceError(priceError);
      return;
    }

    setManualPriceError("");
    onAddItem({
      name,
      price: manualPrice,
      category: manualCategory,
      ...(manualCategory === "other" ? { otherReason } : {}),
    });
    setManualName("");
    setManualPrice("");
    setManualCategory("");
    setManualOtherReason("");
    onNewItemOpenChange(false);
  }, [manualName, manualPrice, manualCategory, manualOtherReason, onAddItem, onNewItemOpenChange]);

  const itemRows = useMemo(
    () =>
      items.map((item) => (
        <AccountingItemRow
          key={item.id}
          item={item}
          accountingId={accountingId}
          canEdit={canEdit}
          canDelete={canDelete}
          onDeleteItem={onDeleteItem}
          onUpdateItemTax={onUpdateItemTax}
          onUpdateItemDiscount={onUpdateItemDiscount}
        />
      )),
    [items, accountingId, onDeleteItem, onUpdateItemTax, onUpdateItemDiscount, canEdit, canDelete],
  );

  return (
    <Card className="flex-1 flex flex-col overflow-hidden">
      <CardHeader className="flex flex-row items-center justify-between py-4 border-b shrink-0">
        <CardTitle className="text-base font-medium">明細一覧</CardTitle>
        {canEdit ? (
          <Dialog open={newItemOpen} onOpenChange={onNewItemOpenChange}>
            <DialogTrigger asChild>
              <Button type="button" variant="outline" size="sm" className="h-9">
                <Plus className={`mr-2 ${ICON.action}`} />
                物販・その他追加
              </Button>
            </DialogTrigger>
            <DialogContent className={`${LAYOUT.modal.md} max-h-[70vh] flex flex-col`}>
              <DialogHeader>
                <DialogTitle>物販・その他追加</DialogTitle>
                <DialogDescription>マスタから選択するか、手動入力で追加できます。</DialogDescription>
              </DialogHeader>
              <div className="flex gap-0 border-b">
                <button
                  type="button"
                  onClick={() => setAddMode("master")}
                  className={`px-4 py-2 text-sm font-medium border-b-2 transition-colors ${
                    addMode === "master"
                      ? `${C.borderBrand} ${C.textBrand}`
                      : `border-transparent ${C.text50} ${C.hoverText}`
                  }`}
                >
                  マスタから選択
                </button>
                <button
                  type="button"
                  onClick={() => setAddMode("manual")}
                  className={`px-4 py-2 text-sm font-medium border-b-2 transition-colors ${
                    addMode === "manual"
                      ? `${C.borderBrand} ${C.textBrand}`
                      : `border-transparent ${C.text50} ${C.hoverText}`
                  }`}
                >
                  手動入力
                </button>
              </div>

              {addMode === "master" ? (
                <>
                  <div className="flex items-center gap-2 py-2">
                    <Input
                      autoFocus
                      aria-label="品目名検索"
                      value={merchandiseSearch}
                      onChange={(e) => setMerchandiseSearch(e.target.value)}
                      placeholder="品目名で検索..."
                      className="flex-1 h-11"
                    />
                    <Select value={categoryFilter} onValueChange={setCategoryFilter}>
                      <SelectTrigger aria-label="品目カテゴリ" className="w-[120px] h-11">
                        <SelectValue />
                      </SelectTrigger>
                      <SelectContent>
                        {MERCHANDISE_CATEGORY_SELECT_ITEMS}
                      </SelectContent>
                    </Select>
                  </div>
                  <div className="flex-1 overflow-auto min-h-[200px] border rounded-md">
                    {filteredMerchandise.length > 0 ? (
                      <table className="w-full">
                        <thead>
                          <tr className={`border-b ${C.bgPage} text-2xs font-semibold`}>
                            <TableHead>品目名</TableHead>
                            <TableHead className="w-[70px]">区分</TableHead>
                            <TableHead className="w-[90px] text-right">単価</TableHead>
                            <TableHead className="w-[60px] text-right">税率</TableHead>
                            <TableHead className="w-[80px]">操作</TableHead>
                          </tr>
                        </thead>
                        <tbody>
                          {filteredMerchandise.map((item) => (
                            <tr key={item.id} className={`border-b ${C.hoverBgLight} transition-colors`}>
                              <TableCell className="font-medium">{item.name}</TableCell>
                              <TableCell className={C.text50}>
                                {CATEGORY_LABELS[item.category as ItemCategory] ?? item.category}
                              </TableCell>
                              <TableCell className="text-right font-mono">
                                {formatCurrency(item.unitPrice)}
                              </TableCell>
                              <TableCell className={`text-right ${C.text50}`}>
                                {item.taxRate === DEFAULT_STANDARD_TAX_RATE ? "10%" : item.taxRate === DEFAULT_REDUCED_TAX_RATE ? "8%" : `${item.taxRate * 100}%`}
                              </TableCell>
                              <TableCell>
                                <DataTableRowButton
                                  aria-label={`追加: ${item.name} (ID ${item.id})`}
                                  onClick={() => handleSelectMerchandise(item)}
                                >
                                  追加
                                </DataTableRowButton>
                              </TableCell>
                            </tr>
                          ))}
                        </tbody>
                      </table>
                    ) : (
                      <div className="flex items-center justify-center h-full">
                        <EmptyState message="該当する品目がありません" />
                      </div>
                    )}
                  </div>
                </>
              ) : (
                <div className="flex flex-col gap-4 py-4">
                  <div className="flex flex-col gap-1.5">
                    <Label htmlFor="manual-name" className="text-sm">品目名 <span className={C.textRequired}>*</span></Label>
                    <Input
                      id="manual-name"
                      autoFocus
                      value={manualName}
                      onChange={(e) => setManualName(e.target.value)}
                      placeholder="例: 診察料（その他）"
                      className="h-9"
                    />
                  </div>
                  <div className="flex flex-col gap-1.5">
                    <Label htmlFor="manual-category" className="text-sm">カテゴリ <span className={C.textRequired}>*</span></Label>
                    <Select value={manualCategory} onValueChange={handleManualCategoryChange}>
                      <SelectTrigger id="manual-category" aria-label="カテゴリ">
                        <SelectValue placeholder="カテゴリを選択" />
                      </SelectTrigger>
                      <SelectContent>
                        {MANUAL_CATEGORY_SELECT_ITEMS}
                      </SelectContent>
                    </Select>
                  </div>
                  {manualCategory === "other" ? (
                    <div className="flex flex-col gap-1.5">
                      <Label htmlFor="manual-other-reason" className="text-sm">その他理由 <span className={C.textRequired}>*</span></Label>
                      <Input
                        id="manual-other-reason"
                        value={manualOtherReason}
                        onChange={(e) => setManualOtherReason(e.target.value)}
                        maxLength={MANUAL_OTHER_REASON_MAX_LENGTH}
                        placeholder="例: 締め時に分類を確認"
                        className="h-9"
                      />
                    </div>
                  ) : null}
                  <div className="flex flex-col gap-1.5">
                    <Label htmlFor="manual-price" className="text-sm">単価（円）<span className={C.textRequired}>*</span></Label>
                    <Input
                      id="manual-price"
                      type="number"
                      step={1}
                      min={0}
                      value={manualPrice}
                      onChange={(e) => setManualPrice(e.target.value)}
                      placeholder="例: 3000"
                      className="h-9"
                    />
                    <FormFieldError message={manualPriceError} />
                  </div>
                  <Button
                    type="button"
                    className="w-full"
                    disabled={
                      !manualName.trim() ||
                      !manualPrice ||
                      !manualCategory ||
                      (manualCategory === "other" && !manualOtherReason.trim())
                    }
                    onClick={handleAddManualItem}
                  >
                    追加する
                  </Button>
                </div>
              )}
            </DialogContent>
          </Dialog>
        ) : null}
      </CardHeader>
      <CardContent className="p-0 overflow-auto flex-1">
        <Table>
          <TableHeader className={`sticky top-0 ${C.bgPage} z-10 shadow-level1`}>
            <TableRow>
              <TableHead className="w-[100px]">区分</TableHead>
              <TableHead className="min-w-[180px]">項目名</TableHead>
              <TableHead className="text-right w-[90px]">単価</TableHead>
              <TableHead className="text-center w-[60px]">数量</TableHead>
              <TableHead className="text-center w-[90px]">割引</TableHead>
              <TableHead className="w-[100px] text-center">課税区分</TableHead>
              <TableHead className="text-center w-[70px]">税率</TableHead>
              <TableHead className="text-right w-[80px] whitespace-nowrap">税額</TableHead>
              <TableHead className="text-center w-[60px]">保険</TableHead>
              <TableHead className="text-right w-[100px]">金額</TableHead>
              <TableHead className="w-[50px]" />
            </TableRow>
          </TableHeader>
          <TableBody>{itemRows}</TableBody>
        </Table>
      </CardContent>
      <div className={`p-4 ${C.bgPage} border-t flex justify-end gap-6 text-sm`}>
        <span>税抜小計: {formatCurrency(subtotal)}</span>
        <span>消費税: {formatCurrency(taxTotal)}</span>
        <span className="font-bold text-xl">
          合計: {formatCurrency(totalAmount)}
        </span>
      </div>
    </Card>
  );
});
