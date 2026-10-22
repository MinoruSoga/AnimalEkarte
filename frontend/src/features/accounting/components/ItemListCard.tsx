import { memo, useCallback, useDeferredValue, useMemo, useState } from "react";
import { Plus } from "lucide-react";

import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "@/components/ui/dialog";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import { EmptyState } from "@/components/shared/DataStates";
import { DataTableRowButton } from "@/components/shared/DataTable/DataTableRowButton";
import { FormFieldError } from "@/components/shared/FormFieldError/FormFieldError";
import { MoneyInput } from "@/components/shared/SidePeek/MoneyInput";
import { C, ICON, LAYOUT } from "@/lib/design-tokens";
import { DEFAULT_STANDARD_TAX_RATE, DEFAULT_REDUCED_TAX_RATE } from "@/constants/tax";
import { formatCurrency } from "@/lib/format/number";
import { isPastJSTDate } from "@/lib/jst-date";
import { normalizeKana } from "@/lib/normalize-kana";
import { CATEGORY_LABELS } from "@/constants/item-category";
import type { TaxType } from "@/types/generated/models";

import { useGetAllMerchandiseItems } from "../api/get-merchandise-items";
import type { FrontendMerchandiseItem } from "../api/get-merchandise-items";
import type { AccountingItem, AddAccountingItemInput, ItemCategory } from "../types";

import { AccountingItemRow } from "./AccountingItemRow";
import { RequiredMark } from "@/components/shared/RequiredMark";

const MERCHANDISE_CATEGORY_OPTIONS = [
  { value: "all", label: "すべて" },
  { value: "food", label: "フード" },
  { value: "goods", label: "物販" },
  { value: "other", label: "その他" },
];

const MERCHANDISE_CATEGORY_SELECT_ITEMS = MERCHANDISE_CATEGORY_OPTIONS.map((o) => (
  <SelectItem key={o.value} value={o.value}>
    {o.label}
  </SelectItem>
));

/** EMR-246: 当日/過去のセクション見出し行がまたぐ明細テーブルの列数。 */
const ITEM_TABLE_COLUMN_COUNT = 11;

const MANUAL_CATEGORY_SELECT_ITEMS = Object.entries(CATEGORY_LABELS).map(([value, label]) => (
  <SelectItem key={value} value={value}>
    {label}
  </SelectItem>
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
  /** EMR-229: 項目名編集 */
  onUpdateItemName?: (itemId: string, name: string) => void;
  /** EMR-230: 数量編集（手入力/社販明細） */
  onUpdateItemQuantity?: (itemId: string, quantity: number) => void;
  /** EMR-230: 金額（税抜小計）直接編集（手入力/社販明細） */
  onUpdateItemAmount?: (item: AccountingItem, amount: number) => void;
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
  onUpdateItemName,
  onUpdateItemQuantity,
  onUpdateItemAmount,
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
  // EMR-319: 必須未入力時に各フィールドへエラーを表示するための状態。
  const [manualNameError, setManualNameError] = useState("");
  const [manualCategoryError, setManualCategoryError] = useState("");
  const [manualOtherReasonError, setManualOtherReasonError] = useState("");

  const { data: merchandiseItems = [] } = useGetAllMerchandiseItems();

  const filteredMerchandise = useMemo(() => {
    let result = merchandiseItems.filter((item) => item.isActive);
    if (categoryFilter !== "all") {
      result = result.filter((item) => item.category === categoryFilter);
    }
    if (deferredMerchandiseSearch) {
      const normalizedTerm = normalizeKana(deferredMerchandiseSearch).toLowerCase();
      result = result.filter((item) =>
        normalizeKana(item.name).toLowerCase().includes(normalizedTerm),
      );
    }
    return result;
  }, [merchandiseItems, categoryFilter, deferredMerchandiseSearch]);

  const handleSelectMerchandise = useCallback(
    (item: FrontendMerchandiseItem) => {
      onAddItem({
        name: item.name,
        price: String(item.unitPrice),
        category: item.category,
        taxType: item.taxType,
        taxRate: item.taxRate,
        merchandiseItemId: item.id,
      });
    },
    [onAddItem],
  );

  const handleManualCategoryChange = useCallback((value: string) => {
    setManualCategory(value);
    setManualCategoryError("");
    if (value !== "other") {
      setManualOtherReason("");
      setManualOtherReasonError("");
    }
  }, []);

  const handleAddManualItem = useCallback(() => {
    const name = manualName.trim();
    const otherReason = manualOtherReason.trim();
    // EMR-319: 必須未入力を黙って return せず、各フィールドのエラー表示を立てる。
    const nameError = !name ? "品目名を入力してください" : "";
    const categoryError = !manualCategory ? "カテゴリを選択してください" : "";
    const otherReasonError =
      manualCategory === "other" && !otherReason ? "その他理由を入力してください" : "";
    // EMR-319: 必須エラーは全フィールド同時に表示する。単価も必須なので
    // 名称等の early return より先に評価しないと初回クリックで未表示のままになる。
    const priceError = getManualPriceError(manualPrice) ?? "";
    setManualNameError(nameError);
    setManualCategoryError(categoryError);
    setManualOtherReasonError(otherReasonError);
    setManualPriceError(priceError);
    if (nameError || categoryError || otherReasonError || priceError !== "") {
      return;
    }

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

  // EMR-246: 未請求候補は全期間集約のため、当日分と過去分を視覚的に区分する。
  // serviceDate なし（手動追加分など）は当日扱い。
  const { todayItems, pastItems } = useMemo(() => {
    const today: AccountingItem[] = [];
    const past: AccountingItem[] = [];
    for (const item of items) {
      if (item.serviceDate != null && isPastJSTDate(item.serviceDate)) {
        past.push(item);
      } else {
        today.push(item);
      }
    }
    return { todayItems: today, pastItems: past };
  }, [items]);

  const renderItemRows = useCallback(
    (groupItems: AccountingItem[]) =>
      groupItems.map((item) => (
        <AccountingItemRow
          key={item.id}
          item={item}
          accountingId={accountingId}
          canEdit={canEdit}
          canDelete={canDelete}
          onDeleteItem={onDeleteItem}
          onUpdateItemTax={onUpdateItemTax}
          onUpdateItemDiscount={onUpdateItemDiscount}
          onUpdateItemName={onUpdateItemName}
          onUpdateItemQuantity={onUpdateItemQuantity}
          onUpdateItemAmount={onUpdateItemAmount}
        />
      )),
    [
      accountingId,
      onDeleteItem,
      onUpdateItemTax,
      onUpdateItemDiscount,
      onUpdateItemName,
      onUpdateItemQuantity,
      onUpdateItemAmount,
      canEdit,
      canDelete,
    ],
  );

  const todayItemRows = useMemo(() => renderItemRows(todayItems), [renderItemRows, todayItems]);
  const pastItemRows = useMemo(() => renderItemRows(pastItems), [renderItemRows, pastItems]);
  const hasPastItems = pastItemRows.length > 0;

  return (
    <Card className="flex-1 flex flex-col overflow-hidden">
      <CardHeader className="flex flex-row items-center justify-between py-4 border-b shrink-0">
        <CardTitle className="text-base font-medium">明細一覧</CardTitle>
        {canEdit ? (
          <Dialog open={newItemOpen} onOpenChange={onNewItemOpenChange}>
            <DialogTrigger asChild>
              <Button type="button" variant="outline" size="sm">
                <Plus className={`mr-2 ${ICON.action}`} />
                物販・その他追加
              </Button>
            </DialogTrigger>
            <DialogContent className={`${LAYOUT.modal.md} max-h-[70vh] flex flex-col`}>
              <DialogHeader>
                <DialogTitle>物販・その他追加</DialogTitle>
                <DialogDescription>
                  マスタから選択するか、手動入力で追加できます。
                </DialogDescription>
              </DialogHeader>
              <div className="flex gap-0 border-b">
                <button
                  type="button"
                  onClick={() => setAddMode("master")}
                  className={`px-4 py-2 min-h-11 text-sm font-medium border-b-2 transition-colors ${
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
                  className={`px-4 py-2 min-h-11 text-sm font-medium border-b-2 transition-colors ${
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
                      <SelectContent>{MERCHANDISE_CATEGORY_SELECT_ITEMS}</SelectContent>
                    </Select>
                  </div>
                  <div className="flex-1 overflow-auto min-h-[200px] border rounded-md">
                    {filteredMerchandise.length > 0 ? (
                      <Table>
                        <TableHeader>
                          <TableRow className={`${C.bgPage} text-2xs font-semibold`}>
                            <TableHead>品目名</TableHead>
                            <TableHead className="w-[70px]">区分</TableHead>
                            <TableHead className="w-[90px] text-right">単価</TableHead>
                            <TableHead className="w-[60px] text-right">税率</TableHead>
                            <TableHead className="w-[80px]">操作</TableHead>
                          </TableRow>
                        </TableHeader>
                        <TableBody>
                          {filteredMerchandise.map((item) => (
                            <TableRow key={item.id} className={`${C.hoverBgLight}`}>
                              <TableCell className="font-medium">{item.name}</TableCell>
                              <TableCell className={C.text50}>
                                {CATEGORY_LABELS[item.category as ItemCategory] ?? item.category}
                              </TableCell>
                              <TableCell className="text-right font-mono">
                                {formatCurrency(item.unitPrice)}
                              </TableCell>
                              <TableCell className={`text-right ${C.text50}`}>
                                {item.taxRate === undefined
                                  ? "—"
                                  : item.taxRate === DEFAULT_STANDARD_TAX_RATE
                                    ? "10%"
                                    : item.taxRate === DEFAULT_REDUCED_TAX_RATE
                                      ? "8%"
                                      : `${item.taxRate * 100}%`}
                              </TableCell>
                              <TableCell>
                                <DataTableRowButton
                                  aria-label={`追加: ${item.name} (ID ${item.id})`}
                                  onClick={() => handleSelectMerchandise(item)}
                                >
                                  追加
                                </DataTableRowButton>
                              </TableCell>
                            </TableRow>
                          ))}
                        </TableBody>
                      </Table>
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
                    <Label htmlFor="manual-name">
                      品目名 <RequiredMark />
                    </Label>
                    <Input
                      id="manual-name"
                      autoFocus
                      value={manualName}
                      onChange={(e) => {
                        setManualName(e.target.value);
                        if (manualNameError) setManualNameError("");
                      }}
                      placeholder="例: 診察料（その他）"
                      aria-invalid={manualNameError ? true : undefined}
                      aria-describedby={manualNameError ? "manual-name-error" : undefined}
                    />
                    <FormFieldError id="manual-name-error" message={manualNameError} />
                  </div>
                  <div className="flex flex-col gap-1.5">
                    <Label htmlFor="manual-category">
                      カテゴリ <RequiredMark />
                    </Label>
                    <Select value={manualCategory} onValueChange={handleManualCategoryChange}>
                      <SelectTrigger
                        id="manual-category"
                        aria-label="カテゴリ"
                        aria-invalid={manualCategoryError ? true : undefined}
                        aria-describedby={manualCategoryError ? "manual-category-error" : undefined}
                      >
                        <SelectValue placeholder="カテゴリを選択" />
                      </SelectTrigger>
                      <SelectContent>{MANUAL_CATEGORY_SELECT_ITEMS}</SelectContent>
                    </Select>
                    <FormFieldError id="manual-category-error" message={manualCategoryError} />
                  </div>
                  {manualCategory === "other" ? (
                    <div className="flex flex-col gap-1.5">
                      <Label htmlFor="manual-other-reason">
                        その他理由 <RequiredMark />
                      </Label>
                      <Input
                        id="manual-other-reason"
                        value={manualOtherReason}
                        onChange={(e) => {
                          setManualOtherReason(e.target.value);
                          if (manualOtherReasonError) setManualOtherReasonError("");
                        }}
                        maxLength={MANUAL_OTHER_REASON_MAX_LENGTH}
                        placeholder="例: 締め時に分類を確認"
                        aria-invalid={manualOtherReasonError ? true : undefined}
                        aria-describedby={
                          manualOtherReasonError ? "manual-other-reason-error" : undefined
                        }
                      />
                      <FormFieldError
                        id="manual-other-reason-error"
                        message={manualOtherReasonError}
                      />
                    </div>
                  ) : null}
                  <div className="flex flex-col gap-1.5">
                    <Label htmlFor="manual-price">
                      単価（円）
                      <RequiredMark />
                    </Label>
                    <MoneyInput
                      variant="inline"
                      id="manual-price"
                      step={1}
                      min={0}
                      stringValue={manualPrice}
                      onStringChange={(v) => {
                        setManualPrice(v);
                        if (manualPriceError) setManualPriceError("");
                      }}
                      placeholder="例: 3000"
                      aria-invalid={manualPriceError ? true : undefined}
                      aria-describedby={manualPriceError ? "manual-price-error" : undefined}
                    />
                    <FormFieldError id="manual-price-error" message={manualPriceError} />
                  </div>
                  <Button type="button" className="w-full" onClick={handleAddManualItem}>
                    追加する
                  </Button>
                </div>
              )}
            </DialogContent>
          </Dialog>
        ) : null}
      </CardHeader>
      <p className={`px-4 py-2 text-sm ${C.text65}`}>
        横にスクロールしてすべての項目を確認できます
      </p>
      <CardContent
        role="region"
        aria-label="会計明細"
        tabIndex={0}
        className="p-0 overflow-auto flex-1 [&>div]:overflow-visible"
      >
        <Table>
          <TableHeader className={`sticky top-0 ${C.bgPage} z-10 shadow-level1`}>
            <TableRow>
              <TableHead className="w-[100px] whitespace-nowrap">区分</TableHead>
              <TableHead className="min-w-[260px] whitespace-nowrap">項目名</TableHead>
              <TableHead className="text-right w-[90px] whitespace-nowrap">単価</TableHead>
              <TableHead className="text-center w-[120px] whitespace-nowrap">数量</TableHead>
              <TableHead className="text-center w-[160px] whitespace-nowrap">割引</TableHead>
              <TableHead className="w-[110px] text-center whitespace-nowrap">課税区分</TableHead>
              <TableHead className="text-center w-[150px] whitespace-nowrap">税率</TableHead>
              <TableHead className="text-right w-[90px] whitespace-nowrap">税額</TableHead>
              <TableHead className="text-center w-[70px] whitespace-nowrap">保険</TableHead>
              <TableHead
                className={`sticky right-[80px] z-20 ${C.bgWhite} text-right min-w-[130px] w-[130px] whitespace-nowrap`}
              >
                金額
              </TableHead>
              <TableHead className={`sticky right-0 z-20 ${C.bgWhite} w-[80px]`} />
            </TableRow>
          </TableHeader>
          <TableBody>
            {hasPastItems ? (
              <TableRow className={C.bgPage}>
                <TableCell colSpan={ITEM_TABLE_COLUMN_COUNT}>
                  <span className={`text-2xs font-semibold ${C.text50}`}>当日分</span>
                </TableCell>
              </TableRow>
            ) : null}
            {todayItemRows}
            {hasPastItems ? (
              <TableRow className={C.bgPage}>
                <TableCell colSpan={ITEM_TABLE_COLUMN_COUNT}>
                  <span className={`text-2xs font-semibold ${C.text50}`}>過去の未請求分</span>
                </TableCell>
              </TableRow>
            ) : null}
            {pastItemRows}
          </TableBody>
        </Table>
      </CardContent>
      <div className={`p-4 ${C.bgPage} border-t flex flex-wrap justify-end gap-6 text-sm`}>
        <span>小計（明細合計）: {formatCurrency(subtotal)}</span>
        <span>消費税（内税分を含む）: {formatCurrency(taxTotal)}</span>
        <span className="font-semibold text-xl">合計: {formatCurrency(totalAmount)}</span>
      </div>
    </Card>
  );
});
