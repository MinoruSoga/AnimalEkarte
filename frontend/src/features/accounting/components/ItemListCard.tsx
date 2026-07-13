import { memo, useCallback, useDeferredValue, useMemo, useState } from "react";
import { normalizeKana } from "@/lib/normalize-kana";
import { Plus } from "lucide-react";

import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Dialog, DialogContent, DialogDescription, DialogHeader, DialogTitle, DialogTrigger } from "@/components/ui/dialog";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { Table, TableBody, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { EmptyState } from "@/components/shared/DataStates";
import { FormFieldError } from "@/components/shared/FormFieldError/FormFieldError";
import { C, ICON, LAYOUT } from "@/lib/design-tokens";
import { DEFAULT_STANDARD_TAX_RATE, DEFAULT_REDUCED_TAX_RATE } from "@/constants/tax";
import { formatCurrency } from "@/utils/format/number";
import { CATEGORY_LABELS } from "@/constants/item-category";
import type { TaxType } from "@/types/generated/models";

import { useGetAllMerchandiseItems } from "../api/get-merchandise-items";
import type { FrontendMerchandiseItem } from "../api/get-merchandise-items";
import type { AccountingItem, ItemCategory } from "../types";

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

function getManualPriceError(value: string): string | null {
  const priceNum = parseInt(value, 10);
  if (isNaN(priceNum) || priceNum < 0) {
    return "単価は0以上の整数で入力してください";
  }
  if (priceNum > 999999999) {
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
  onAddItem: (name: string, price: string, category: string, taxRate?: number) => void;
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
      onAddItem(item.name, String(item.unitPrice), item.category, item.taxRate);
    },
    [onAddItem],
  );

  const handleAddManualItem = useCallback(() => {
    const name = manualName.trim();
    if (!name || !manualPrice) return;

    const priceError = getManualPriceError(manualPrice);
    if (priceError !== null) {
      setManualPriceError(priceError);
      return;
    }

    setManualPriceError("");
    onAddItem(name, manualPrice, "other");
    setManualName("");
    setManualPrice("");
    onNewItemOpenChange(false);
  }, [manualName, manualPrice, onAddItem, onNewItemOpenChange]);

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
                      value={merchandiseSearch}
                      onChange={(e) => setMerchandiseSearch(e.target.value)}
                      placeholder="品目名で検索..."
                      className="flex-1 h-9"
                    />
                    <Select value={categoryFilter} onValueChange={setCategoryFilter}>
                      <SelectTrigger className="w-[120px] h-9">
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
                          <tr className={`border-b ${C.bgPage30} text-xs`}>
                            <th className="px-3 py-2 text-left font-medium">品目名</th>
                            <th className="px-3 py-2 text-left font-medium w-[70px]">区分</th>
                            <th className="px-3 py-2 text-right font-medium w-[90px]">単価</th>
                            <th className="px-3 py-2 text-right font-medium w-[60px]">税率</th>
                          </tr>
                        </thead>
                        <tbody>
                          {filteredMerchandise.map((item) => (
                            <tr
                              key={item.id}
                              onClick={() => handleSelectMerchandise(item)}
                              className={`border-b cursor-pointer ${C.hoverBgLight} transition-colors`}
                            >
                              <td className="px-3 py-2 text-sm font-medium">{item.name}</td>
                              <td className={`px-3 py-2 text-sm ${C.text50}`}>
                                {CATEGORY_LABELS[item.category as ItemCategory] ?? item.category}
                              </td>
                              <td className="px-3 py-2 text-sm text-right font-mono">
                                {formatCurrency(item.unitPrice)}
                              </td>
                              <td className={`px-3 py-2 text-sm text-right ${C.text50}`}>
                                {item.taxRate === DEFAULT_STANDARD_TAX_RATE ? "10%" : item.taxRate === DEFAULT_REDUCED_TAX_RATE ? "8%" : `${item.taxRate * 100}%`}
                              </td>
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
                    disabled={!manualName.trim() || !manualPrice}
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
          <TableHeader className={`sticky top-0 ${C.bgWhite} z-10 shadow-sm`}>
            <TableRow>
              <TableHead className="w-[100px]">区分</TableHead>
              <TableHead>項目名</TableHead>
              <TableHead className="text-right w-[90px]">単価</TableHead>
              <TableHead className="text-center w-[60px]">数量</TableHead>
              <TableHead className="text-center w-[90px]">割引</TableHead>
              <TableHead className="w-[100px] text-center">課税区分</TableHead>
              <TableHead className="text-center w-[70px]">税率</TableHead>
              <TableHead className="text-right w-[80px]">税額</TableHead>
              <TableHead className="text-center w-[60px]">保険</TableHead>
              <TableHead className="text-right w-[100px]">金額</TableHead>
              <TableHead className="w-[50px]" />
            </TableRow>
          </TableHeader>
          <TableBody>{itemRows}</TableBody>
        </Table>
      </CardContent>
      <div className={`p-4 ${C.bgPage} border-t flex justify-end gap-6 text-sm`}>
        <span>税抜小計: ¥{subtotal.toLocaleString()}</span>
        <span>消費税: ¥{taxTotal.toLocaleString()}</span>
        <span className="font-bold text-lg">
          合計: ¥{totalAmount.toLocaleString()}
        </span>
      </div>
    </Card>
  );
});
