// React/Framework
import { ICON, C, LAYOUT } from "@/lib/design-tokens";
import { useState, useMemo, useCallback, memo, useTransition, useDeferredValue, useActionState, useEffect, useRef } from "react";
import { useParams, useNavigate, useLocation } from "react-router";

// External
import { Plus, Save, CreditCard, Printer, RotateCcw, EyeOff } from "lucide-react";
import { useQueryClient, useQuery } from "@tanstack/react-query";
import { toast } from "sonner";

// Internal
import { PageLayout } from "@/components/shared/PageLayout/PageLayout";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Badge } from "@/components/ui/badge";
import { Switch } from "@/components/ui/switch";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogTrigger, DialogFooter, DialogDescription } from "@/components/ui/dialog";
import { Separator } from "@/components/ui/separator";

// Shared Hooks
import { useAuth } from "@/hooks/use-auth";
import { usePermission } from "@/hooks/use-permission";

// Relative
import { LoadingFallback, ErrorFallback } from "@/components/shared/DataStates";
import { FormFieldError } from "@/components/shared/FormFieldError/FormFieldError";
import { useGetAccountingDetail } from "../api/get-accounting";
import { createAccounting } from "../api/create-accounting";
import { updateAccounting } from "../api/update-accounting";
import { updateBillingItem } from "../api/update-billing-item";
import { createBillingItem } from "../api/create-billing-item";
import { getUnbilledItems } from "../api/get-unbilled-items";
import { useGetRefunds } from "../api/get-refunds";
import { createRefund } from "../api/create-refund";
import { useGetAllMerchandiseItems } from "../api/get-merchandise-items";
import { useGetPet } from "@/hooks/use-pet";
import type { FrontendMerchandiseItem } from "../api/get-merchandise-items";
import { TaxTypeSelector } from "@/components/shared/TaxTypeSelector/TaxTypeSelector";
import { TaxRateSelector } from "@/components/shared/TaxRateSelector/TaxRateSelector";
import { queryKeys } from "@/lib/query-keys";
import { handleApiError } from "@/lib/handle-api-error";
import { calculateBillingTotals } from "@/lib/calculations";
import { AccountingDocument } from "../components/AccountingDocument";
import { cancelAccounting } from "../api/cancel-accounting";
import { ConfirmDialog } from "@/components/shared/ConfirmDialog/ConfirmDialog";
import { paths } from "@/config/paths";
import { NumberInput } from "@/components/shared/NumberInput/NumberInput";
import { DeleteIconButton } from "@/components/shared/DeleteIconButton/DeleteIconButton";
import { SubmitButton } from "@/components/shared/Form/SubmitButton";

// Types
import type { Accounting, AccountingItem, PaymentInfo, ItemCategory, PaymentMethod } from "../types";
import type { TaxType } from "@/types/generated/models";
import { ResourceAccounting } from "@/types/generated/models";

// ── 静的定数（rendering-hoist-jsx）──────────────────────────

const CATEGORY_LABELS: Record<ItemCategory, string> = {
  examination: "診察",
  test: "検査",
  procedure: "処置",
  surgery: "手術",
  medicine: "処方",
  food: "フード",
  goods: "物販",
  other: "その他",
};

// ── memo 化サブコンポーネント ──────────────────────────────

interface ItemListCardProps {
  items: AccountingItem[];
  subtotal: number;
  taxTotal: number;
  totalAmount: number;
  newItemOpen: boolean;
  onNewItemOpenChange: (open: boolean) => void;
  onAddItem: (name: string, price: string, category: string, taxRate?: number) => void;
  onDeleteItem: (id: string) => void;
  /** 既存会計の ID（billing_id）。未設定の場合は新規作成モード */
  accountingId?: string;
  onUpdateItemTax?: (itemId: string, taxType: TaxType, taxRate: number) => void;
  canEdit: boolean;
  canDelete: boolean;
}

const MERCHANDISE_CATEGORY_OPTIONS = [
  { value: "all", label: "すべて" },
  { value: "food", label: "フード" },
  { value: "goods", label: "物販" },
  { value: "other", label: "その他" },
];

// rendering-hoist-jsx: 静的 SelectItem リストをモジュールスコープに巻き上げ
const MERCHANDISE_CATEGORY_SELECT_ITEMS = MERCHANDISE_CATEGORY_OPTIONS.map((o) => (
  <SelectItem key={o.value} value={o.value}>{o.label}</SelectItem>
));

const ItemListCard = memo(function ItemListCard({
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
  canEdit,
  canDelete,
}: ItemListCardProps) {
  const [categoryFilter, setCategoryFilter] = useState("all");
  const [merchandiseSearch, setMerchandiseSearch] = useState("");
  const deferredMerchandiseSearch = useDeferredValue(merchandiseSearch);
  // BUG-088: 手動入力モード
  const [addMode, setAddMode] = useState<"master" | "manual">("master");
  const [manualName, setManualName] = useState("");
  const [manualPrice, setManualPrice] = useState("");
  const [manualPriceError, setManualPriceError] = useState<string>("");

  // マスタデータ取得
  const { data: merchandiseItems = [] } = useGetAllMerchandiseItems();

  const filteredMerchandise = useMemo(() => {
    let result = merchandiseItems.filter((item) => item.isActive);
    if (categoryFilter !== "all") {
      result = result.filter((item) => item.category === categoryFilter);
    }
    if (deferredMerchandiseSearch) {
      const lower = deferredMerchandiseSearch.toLowerCase();
      result = result.filter((item) => item.name.toLowerCase().includes(lower));
    }
    return result;
  }, [merchandiseItems, categoryFilter, deferredMerchandiseSearch]);

  const handleSelectMerchandise = useCallback(
    (item: FrontendMerchandiseItem) => {
      onAddItem(item.name, String(item.unitPrice), item.category, item.taxRate);
    },
    [onAddItem],
  );

  const itemRows = useMemo(
    () =>
      items.map((item) => (
        <TableRow key={item.id} className="h-12">
          <TableCell>
            <Badge variant="outline" className="font-normal text-xs">
              {CATEGORY_LABELS[item.category as ItemCategory] ?? "その他"}
            </Badge>
          </TableCell>
          <TableCell className="font-medium">
            {item.name}
            {item.source === "medical_record" ? (
              <span className={`ml-2 text-[10px] ${C.accent} ${C.bgAccent5} px-1.5 py-0.5 rounded`}>
                カルテ連携
              </span>
            ) : null}
          </TableCell>
          <TableCell className="text-right">
            ¥{item.unitPrice.toLocaleString()}
          </TableCell>
          <TableCell className="text-center">
            <div className="flex items-center justify-center gap-2">
              {item.quantity}
            </div>
          </TableCell>
          <TableCell className="text-center">
            {accountingId !== undefined && onUpdateItemTax !== undefined && canEdit ? (
              <TaxTypeSelector
                value={item.taxType}
                onChange={(v) => onUpdateItemTax(item.id, v, item.taxRate)}
              />
            ) : (
              <span className={`text-sm ${C.text50}`}>
                {item.taxType === "excluded" ? "外税" : item.taxType === "included" ? "内税" : "非課税"}
              </span>
            )}
          </TableCell>
          <TableCell className="text-center">
            {accountingId !== undefined && onUpdateItemTax !== undefined && canEdit ? (
              <TaxRateSelector
                value={item.taxRate}
                onChange={(v) => onUpdateItemTax(item.id, item.taxType, v)}
              />
            ) : (
              <span className={`text-sm ${C.text50}`}>{Math.round(item.taxRate * 100)}%</span>
            )}
          </TableCell>
          <TableCell className="text-right font-mono text-sm">
            ¥{item.taxAmount.toLocaleString()}
          </TableCell>
          <TableCell className="text-center">
            {item.isInsuranceApplicable ? (
              <span className={`${C.textStatusGreen} font-bold text-xs`}>●</span>
            ) : (
              <span className={`${C.text20} text-xs`}>-</span>
            )}
          </TableCell>
          <TableCell className="text-right font-medium">
            ¥{(item.subtotal + item.taxAmount).toLocaleString()}
          </TableCell>
          <TableCell>
            {item.source === "manual" && canDelete ? (
              <DeleteIconButton onClick={() => onDeleteItem(item.id)} />
            ) : null}
          </TableCell>
        </TableRow>
      )),
    [items, accountingId, onDeleteItem, onUpdateItemTax, canEdit, canDelete],
  );

  return (
    <Card className="flex-1 flex flex-col overflow-hidden">
      <CardHeader className="flex flex-row items-center justify-between py-4 border-b shrink-0">
        <CardTitle className="text-base font-medium">明細一覧</CardTitle>
        {canEdit ? (
          <Dialog open={newItemOpen} onOpenChange={onNewItemOpenChange}>
          <DialogTrigger asChild>
            <Button variant="outline" size="sm" className="h-9">
              <Plus className={`mr-2 ${ICON.action}`} />
              物販・その他追加
            </Button>
          </DialogTrigger>
          <DialogContent className={`${LAYOUT.modal.md} max-h-[70vh] flex flex-col`}>
            <DialogHeader>
              <DialogTitle>物販・その他追加</DialogTitle>
              <DialogDescription>マスタから選択するか、手動入力で追加できます。</DialogDescription>
            </DialogHeader>
            {/* BUG-088: モード切替タブ */}
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
                              ¥{item.unitPrice.toLocaleString()}
                            </td>
                            <td className={`px-3 py-2 text-sm text-right ${C.text50}`}>
                              {item.taxRate === 0.1 ? "10%" : item.taxRate === 0.08 ? "8%" : `${item.taxRate * 100}%`}
                            </td>
                          </tr>
                        ))}
                      </tbody>
                    </table>
                  ) : (
                    <div className={`flex items-center justify-center h-full text-sm ${C.text50} py-8`}>
                      該当する品目がありません
                    </div>
                  )}
                </div>
              </>
            ) : (
              /* 手動入力フォーム */
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
                  onClick={() => {
                    if (!manualName.trim() || !manualPrice) return;
                    // BUG-072: 金額の範囲チェック（負の値・上限超過）
                    const priceNum = parseInt(manualPrice, 10);
                    if (isNaN(priceNum) || priceNum < 0) {
                      setManualPriceError("単価は0以上の整数で入力してください");
                      return;
                    }
                    if (priceNum > 999999999) {
                      setManualPriceError("単価は999,999,999円以下で入力してください");
                      return;
                    }
                    setManualPriceError("");
                    onAddItem(manualName.trim(), manualPrice, "other");
                    setManualName("");
                    setManualPrice("");
                    onNewItemOpenChange(false);
                  }}
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
              <TableHead className="w-[100px] text-center">課税区分</TableHead>
              <TableHead className="text-center w-[70px]">税率</TableHead>
              <TableHead className="text-right w-[80px]">税額</TableHead>
              <TableHead className="text-center w-[60px]">保険</TableHead>
              <TableHead className="text-right w-[100px]">金額</TableHead>
              <TableHead className="w-[50px]"></TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {itemRows}
          </TableBody>
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

// ── 保険カード ──────────────────────────────────────────

interface InsuranceCardProps {
  useInsurance: boolean;
  onUseInsuranceChange: (v: boolean) => void;
  insuranceRatio: string;
  onInsuranceRatioChange: (v: string) => void;
  insuranceAmount: number;
}

// rendering-hoist-jsx: 静的SelectItem定数をモジュールスコープに巻き上げ
const INSURANCE_RATIO_ITEMS = (
  <>
    <SelectItem value="0.5">50%</SelectItem>
    <SelectItem value="0.7">70%</SelectItem>
    <SelectItem value="0.9">90%</SelectItem>
    <SelectItem value="1.0">100%</SelectItem>
  </>
);

const InsuranceCard = memo(function InsuranceCard({
  useInsurance,
  onUseInsuranceChange,
  insuranceRatio,
  onInsuranceRatioChange,
  insuranceAmount,
}: InsuranceCardProps) {
  return (
    <Card>
      <CardHeader className={`py-3 px-4 ${C.bgPage} border-b`}>
        <div className="flex items-center justify-between">
          <CardTitle className="text-sm font-medium flex items-center gap-2">
            <CreditCard className={ICON.action} /> ペット保険（窓口精算）
          </CardTitle>
          <Switch checked={useInsurance} onCheckedChange={onUseInsuranceChange} />
        </div>
      </CardHeader>
      {useInsurance ? (
        <CardContent className="p-4 space-y-4">
          <div className="space-y-2">
            <Label className="text-xs">負担割合（保険会社が支払う割合）</Label>
            <Select value={insuranceRatio} onValueChange={onInsuranceRatioChange}>
              <SelectTrigger className="h-10">
                <SelectValue />
              </SelectTrigger>
              <SelectContent>{INSURANCE_RATIO_ITEMS}</SelectContent>
            </Select>
          </div>
          <div className={`flex justify-between items-center text-sm font-medium ${C.textStatusGreen} ${C.bgStatusGreen} p-2 rounded`}>
            <span>保険負担額（マイナス）</span>
            <span>{insuranceAmount.toLocaleString()} 円</span>
          </div>
        </CardContent>
      ) : null}
    </Card>
  );
});

// ── 決済カード ──────────────────────────────────────────

interface PaymentSplitDraft {
  method: PaymentMethod;
  amount: string;
  receivedAmount: string; // 現金のみ意味あり
}

const PAYMENT_METHOD_LABELS: Record<PaymentMethod, string> = {
  cash: "現金",
  credit_card: "カード",
  electronic_money: "電子マネー",
};

const PAYMENT_METHODS: PaymentMethod[] = ["cash", "credit_card", "electronic_money"];

interface PaymentCardProps {
  billingAmount: number;
  paymentSplits: PaymentSplitDraft[];
  onSplitsChange: (splits: PaymentSplitDraft[]) => void;
  isCompleted: boolean;
  canEdit: boolean;
  canCreate: boolean;
  isEditMode: boolean;
}

const PaymentCard = memo(function PaymentCard({
  billingAmount,
  paymentSplits,
  onSplitsChange,
  isCompleted,
  canEdit,
  canCreate,
  isEditMode,
}: PaymentCardProps) {
  const canSubmit = isEditMode ? canEdit : canCreate;

  const splitTotal = paymentSplits.reduce((sum, s) => sum + (parseInt(s.amount || "0", 10)), 0);
  const remaining = billingAmount - splitTotal;

  const isDisabled = remaining !== 0 || paymentSplits.some((s) => {
    if (!s.amount || parseInt(s.amount, 10) <= 0) return true;
    if (s.method === "cash") {
      const received = parseInt(s.receivedAmount || "0", 10);
      return received < parseInt(s.amount, 10);
    }
    return false;
  });

  const handleMethodChange = useCallback((idx: number, method: PaymentMethod) => {
    onSplitsChange(paymentSplits.map((s, i) => i === idx ? { ...s, method } : s));
  }, [paymentSplits, onSplitsChange]);

  const handleAmountChange = useCallback((idx: number, value: string) => {
    onSplitsChange(paymentSplits.map((s, i) => i === idx ? { ...s, amount: value } : s));
  }, [paymentSplits, onSplitsChange]);

  const handleReceivedChange = useCallback((idx: number, value: string) => {
    onSplitsChange(paymentSplits.map((s, i) => i === idx ? { ...s, receivedAmount: value } : s));
  }, [paymentSplits, onSplitsChange]);

  const handleRemoveSplit = useCallback((idx: number) => {
    onSplitsChange(paymentSplits.filter((_, i) => i !== idx));
  }, [paymentSplits, onSplitsChange]);

  const handleAddSplit = useCallback(() => {
    const rem = billingAmount - paymentSplits.reduce((sum, s) => sum + (parseInt(s.amount || "0", 10)), 0);
    onSplitsChange([...paymentSplits, { method: "cash", amount: rem > 0 ? rem.toString() : "", receivedAmount: "" }]);
  }, [paymentSplits, onSplitsChange, billingAmount]);

  return (
    <Card className="flex-1">
      <CardHeader className="py-3 px-4 border-b">
        <CardTitle className="text-base font-medium">決済情報</CardTitle>
      </CardHeader>
      <CardContent className="p-6 space-y-6">
        <div className="text-center space-y-1">
          <p className={`text-sm ${C.text50}`}>今回の請求金額</p>
          <p className={`text-4xl font-bold ${C.text}`}>
            ¥{billingAmount.toLocaleString()}
          </p>
        </div>

        <Separator />

        {canEdit ? (
          <div className="space-y-4">
            {paymentSplits.map((split, idx) => {
              const parsedAmount = parseInt(split.amount || "0", 10);
              const parsedReceived = parseInt(split.receivedAmount || "0", 10);
              const splitChange = split.method === "cash" ? parsedReceived - parsedAmount : 0;

              return (
                <div key={idx} className="border rounded-lg p-3 space-y-3">
                  <div className="flex items-center justify-between">
                    <Label className="text-xs font-medium">支払方法</Label>
                    {paymentSplits.length > 1 ? (
                      <DeleteIconButton
                        onClick={() => handleRemoveSplit(idx)}
                        aria-label="この支払いを削除"
                      />
                    ) : null}
                  </div>
                  <div className="grid grid-cols-3 gap-2">
                    {PAYMENT_METHODS.map((m) => (
                      <Button
                        key={m}
                        type="button"
                        variant={split.method === m ? "default" : "outline"}
                        onClick={() => handleMethodChange(idx, m)}
                        className="h-10 text-sm"
                      >
                        {PAYMENT_METHOD_LABELS[m]}
                      </Button>
                    ))}
                  </div>
                  <div className="space-y-1">
                    <Label className="text-xs">金額</Label>
                    <NumberInput
                      className="h-12 text-lg font-bold"
                      value={split.amount}
                      onChange={(v) => handleAmountChange(idx, v)}
                      suffix="円"
                      align="right"
                    />
                  </div>
                  {split.method === "cash" ? (
                    <>
                      <div className="space-y-1">
                        <Label className="text-xs">お預かり金額</Label>
                        <NumberInput
                          id={idx === 0 ? "receivedAmount" : undefined}
                          className="h-12 text-lg font-bold"
                          value={split.receivedAmount}
                          onChange={(v) => handleReceivedChange(idx, v)}
                          suffix="円"
                          align="right"
                        />
                        <div className="flex gap-2 justify-end">
                          <Button
                            type="button"
                            variant="outline"
                            size="sm"
                            onClick={() => handleReceivedChange(idx, parsedAmount.toString())}
                          >
                            丁度
                          </Button>
                          <Button
                            type="button"
                            variant="outline"
                            size="sm"
                            onClick={() => handleReceivedChange(idx, (Math.ceil(parsedAmount / 1000) * 1000).toString())}
                          >
                            千円単位
                          </Button>
                          <Button
                            type="button"
                            variant="outline"
                            size="sm"
                            onClick={() => handleReceivedChange(idx, (Math.ceil(parsedAmount / 10000) * 10000).toString())}
                          >
                            一万単位
                          </Button>
                        </div>
                      </div>
                      <div className={`${C.bgPrimary5} p-3 rounded-lg flex justify-between items-center`}>
                        <span className={`text-sm font-bold ${C.text60}`}>お釣り</span>
                        <span className={`text-xl font-bold ${splitChange < 0 ? C.danger : C.text}`}>
                          ¥{splitChange.toLocaleString()}
                        </span>
                      </div>
                    </>
                  ) : null}
                </div>
              );
            })}

            {remaining !== 0 ? (
              <p className={`text-xs text-right ${remaining < 0 ? C.danger : C.text50}`}>
                {remaining > 0
                  ? `残り ¥${remaining.toLocaleString()} 未入力`
                  : `¥${Math.abs(remaining).toLocaleString()} 超過`}
              </p>
            ) : null}

            <Button
              type="button"
              variant="outline"
              size="sm"
              onClick={handleAddSplit}
              className="w-full"
            >
              <Plus className={`mr-1 ${ICON.action}`} />
              支払方法を追加
            </Button>
          </div>
        ) : (
          <div className="space-y-3">
            {paymentSplits.map((split, idx) => (
              <div key={idx} className="flex justify-between items-center text-sm">
                <span className={C.text50}>{PAYMENT_METHOD_LABELS[split.method] ?? split.method}</span>
                <span className="font-medium">¥{parseInt(split.amount || "0", 10).toLocaleString()}</span>
              </div>
            ))}
          </div>
        )}

        {canSubmit ? (
          <SubmitButton
            className="w-full h-14 text-lg font-bold mt-4"
            size="lg"
            disabled={isDisabled}
            loadingText="処理中..."
          >
            <Save className={`mr-2 ${ICON.action}`} />
            {isCompleted ? "修正を保存する" : "会計を確定する"}
          </SubmitButton>
        ) : null}
      </CardContent>
    </Card>
  );
});

// ── 返金セクション ──────────────────────────────────────
interface RefundSectionProps {
  /** 会計 ID（バックエンドの billing_id に対応） */
  accountingId: string;
  totalAmount: number;
  isRefunding: boolean;
  onRefund: (amount: number, reason: string) => void;
  canEdit: boolean;
}

const RefundSection = memo(function RefundSection({
  accountingId,
  totalAmount,
  isRefunding,
  onRefund,
  canEdit,
}: RefundSectionProps) {
  const [refundDialogOpen, setRefundDialogOpen] = useState(false);
  const [refundAmount, setRefundAmount] = useState("");
  const [refundReason, setRefundReason] = useState("");
  const { data: refunds = [] } = useGetRefunds(accountingId);

  const totalRefunded = refunds.reduce((sum, r) => sum + r.amount, 0);
  const refundableAmount = totalAmount - totalRefunded;

  const handleSubmit = useCallback(() => {
    const amount = parseInt(refundAmount, 10);
    if (!amount || amount <= 0) return;
    onRefund(amount, refundReason);
    setRefundDialogOpen(false);
    setRefundAmount("");
    setRefundReason("");
  }, [refundAmount, refundReason, onRefund]);


  return (
    <Card>
      <CardHeader className={`py-3 px-4 ${C.bgSubtle} border-b`}>
        <div className="flex items-center justify-between">
          <CardTitle className="text-sm font-medium flex items-center gap-2">
            <RotateCcw className={`${ICON.action} ${C.textDiscount}`} />
            返金管理
            <span className={`text-xs font-normal ${C.text50}`}>
              残額 ¥{refundableAmount.toLocaleString()}
            </span>
            {totalRefunded > 0 ? (
              <span className={`text-xs font-normal ${C.textDiscount} ${C.bgDiscountLight} px-2 py-0.5 rounded`}>
                合計 ¥{totalRefunded.toLocaleString()} 返金済
              </span>
            ) : null}
          </CardTitle>
          {canEdit ? (
            <Dialog open={refundDialogOpen} onOpenChange={setRefundDialogOpen}>
            <DialogTrigger asChild>
              <Button
                variant="outline"
                size="sm"
                className="h-8 text-xs"
                disabled={refundableAmount <= 0}
              >
                <Plus className={`mr-1 ${ICON.action}`} />
                返金を登録
              </Button>
            </DialogTrigger>
            <DialogContent className="sm:max-w-sm">
              <DialogHeader>
                <DialogTitle>返金を登録</DialogTitle>
                <DialogDescription>返金金額と理由を入力してください。</DialogDescription>
              </DialogHeader>
              <div className="space-y-4 py-2">
                <div className="space-y-2">
                  <Label>返金金額（円）</Label>
                  <Input
                    type="number"
                    step={1}
                    min={1}
                    value={refundAmount}
                    onChange={(e) => setRefundAmount(e.target.value)}
                    placeholder="0"
                    className="h-10"
                  />
                </div>
                <div className="space-y-2">
                  <Label>返金理由（任意）</Label>
                  <Input
                    value={refundReason}
                    onChange={(e) => setRefundReason(e.target.value)}
                    placeholder="返金理由を入力..."
                    className="h-10"
                  />
                </div>
              </div>
              <DialogFooter>
                <Button variant="outline" onClick={() => setRefundDialogOpen(false)}>
                  キャンセル
                </Button>
                <Button
                  onClick={handleSubmit}
                  disabled={!refundAmount || parseInt(refundAmount, 10) <= 0 || isRefunding}
                >
                  {isRefunding ? "処理中..." : "登録する"}
                </Button>
              </DialogFooter>
            </DialogContent>
          </Dialog>
          ) : null}
        </div>
      </CardHeader>
      {refunds.length > 0 ? (
        <CardContent className="p-0">
          <table className="w-full text-sm">
            <thead>
              <tr className={`border-b ${C.bgPage30} text-xs`}>
                <th className="px-3 py-2 text-left font-medium">日時</th>
                <th className="px-3 py-2 text-left font-medium">処理者</th>
                <th className="px-3 py-2 text-right font-medium">金額</th>
                <th className="px-3 py-2 text-left font-medium">理由</th>
              </tr>
            </thead>
            <tbody>
              {refunds.map((r) => (
                <tr key={r.id} className="border-b last:border-0">
                  <td className={`px-3 py-2 font-mono text-xs ${C.text50}`}>
                    {new Date(r.refundedAt).toLocaleDateString("ja-JP")}
                  </td>
                  <td className={`px-3 py-2 text-xs ${C.text50}`}>
                    {r.refundedByName || "-"}
                  </td>
                  <td className={`px-3 py-2 text-right font-medium ${C.textDiscount}`}>
                    ¥{r.amount.toLocaleString()}
                  </td>
                  <td className={`px-3 py-2 ${C.text50} truncate max-w-[120px]`}>
                    {r.reason || "-"}
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </CardContent>
      ) : (
        <CardContent className={`p-4 text-center text-sm ${C.text50}`}>
          返金記録はありません
        </CardContent>
      )}
    </Card>
  );
});

// ── メインコンポーネント ──────────────────────────────────

interface AccountingDetailProps {
  invoiceRegistrationNumber?: string;
}

export const AccountingDetail = memo(function AccountingDetail({ invoiceRegistrationNumber }: AccountingDetailProps) {
  const { id } = useParams();
  const navigate = useNavigate();
  const location = useLocation();
  const queryClient = useQueryClient();
  const [, startTaxUpdateTransition] = useTransition();
  const [, startAddItemTransition] = useTransition();

  const locationState = location.state as { accountingItems?: AccountingItem[] } | null;

  // 既存データは API から取得
  const { data: fetchedAccounting, isLoading } = useGetAccountingDetail(id);

  // 新規作成時のペット情報取得（URL の petId から owner 情報を解決するため）
  const newPetId = useMemo(() => {
    if (id) return "";
    return new URLSearchParams(location.search).get("petId") ?? "";
  }, [id, location.search]);
  const { data: newPetData } = useGetPet(newPetId);

  // 新規作成の場合は location.state から items を引き継ぐ（派生データ）
  const baseAccounting = useMemo<Accounting | null>(() => {
    if (id) {
      return fetchedAccounting ?? null;
    }
    const stateItems = locationState?.accountingItems ?? [];
    return {
      id: "acc_new",
      ownerId: newPetData?.ownerId ?? "",
      ownerName: newPetData?.ownerName ?? "飼い主様",
      petId: newPetId,
      petName: newPetData?.name ?? "ペット",
      petSpecies: newPetData?.species ?? "犬",
      status: "waiting",
      scheduledDate: new Date().toISOString().split("T")[0],
      items: stateItems,
      payment: undefined,
      totalRefundedAmount: 0,
    };
  }, [id, fetchedAccounting, locationState, newPetId, newPetData]);

  // baseAccounting.items の安定参照 — useCallback deps に配列オブジェクトを渡さないための useMemo
  const baseItems = useMemo(() => baseAccounting?.items ?? [], [baseAccounting]);

  // 新規作成時: 未請求の治療明細をペットIDで取得してローカル明細を初期化
  const { data: unbilledItems } = useQuery({
    queryKey: ["unbilledItems", newPetId],
    queryFn: () => getUnbilledItems(newPetId),
    enabled: !id && !!newPetId,
    staleTime: 30_000,
  });

  // ユーザー操作による追加・削除を管理するローカル明細
  const [localItems, setLocalItems] = useState<AccountingItem[] | null>(null);

  // 未請求明細が取得できたらローカル明細を初期化（ユーザー未編集の場合のみ）
  useEffect(() => {
    if (unbilledItems && unbilledItems.length > 0 && localItems === null) {
      setLocalItems(unbilledItems);
    }
  }, [unbilledItems, localItems]);

  // 表示する明細: ローカル編集があればそちら優先
  const displayItems = useMemo(
    () => localItems ?? baseItems,
    [localItems, baseItems],
  );

  // 決済完了フラグ（API 更新後に画面に反映するためのローカル状態）
  const [completedPayment, setCompletedPayment] = useState<PaymentInfo | null>(null);

  // 実際に表示する accounting (payment は completedPayment 優先)
  const accounting: Accounting | null = useMemo(() => {
    if (!baseAccounting) return null;
    return {
      ...baseAccounting,
      items: displayItems,
      payment: completedPayment ?? baseAccounting.payment,
      status: completedPayment ? "completed" : baseAccounting.status,
    };
  }, [baseAccounting, displayItems, completedPayment]);

  // 保険・決済の状態（API データ取得後に同期）
  const [hasInsurance, setHasInsurance] = useState(() => (baseAccounting?.payment?.insuranceAmount ?? 0) < 0);
  const [insuranceRatio, setInsuranceRatio] = useState(() => baseAccounting?.payment?.insuranceRatio?.toString() ?? "0.5");

  const initPaymentSplits = (): PaymentSplitDraft[] => {
    const splits = baseAccounting?.paymentSplits;
    if (splits && splits.length > 0) {
      return splits.map((s) => ({
        method: s.method,
        amount: s.amount.toString(),
        receivedAmount: s.receivedAmount > 0 ? s.receivedAmount.toString() : "",
      }));
    }
    const payment = baseAccounting?.payment;
    if (payment) {
      return [{
        method: payment.method,
        amount: payment.billingAmount.toString(),
        receivedAmount: payment.receivedAmount > 0 ? payment.receivedAmount.toString() : "",
      }];
    }
    return [{ method: "cash", amount: "", receivedAmount: "" }];
  };

  const [paymentSplits, setPaymentSplits] = useState<PaymentSplitDraft[]>(initPaymentSplits);

  // 金額計算
  const calculation = useMemo(() => {
    if (!accounting) return null;

    const billingResult = calculateBillingTotals(
      accounting.items,
      0, // ownerDiscountRate (accounting detail supports specific line discounts instead)
      0, // globalDiscountAmount
      0.10, // taxRate
      hasInsurance ? parseFloat(insuranceRatio) : 0
    );

    return {
      subtotal: billingResult.subtotal,
      taxTotal: billingResult.tax,
      totalAmount: billingResult.total,
      insuranceAmount: billingResult.insuranceAmount,
      billingAmount: billingResult.billingAmount,
    };
  }, [accounting, hasInsurance, insuranceRatio]);

  // 追加アイテム用State
  const [newItemOpen, setNewItemOpen] = useState(false);
  const [isRefunding, startRefundTransition] = useTransition();

  // BUG-367: 明細兼領収書プレビュー State（旧 previewType 廃止）
  const [previewOpen, setPreviewOpen] = useState(false);

  // BUG-371: 精算済会計の修正確認 / キャンセル確認 モーダル状態
  const [editConfirmOpen, setEditConfirmOpen] = useState(false);
  const [cancelConfirmOpen, setCancelConfirmOpen] = useState(false);
  const [isCancelling, startCancelTransition] = useTransition();
  // editConfirmedRef: 確認モーダル OK 後の formAction 再実行用フラグ
  const editConfirmedRef = useRef(false);
  const formRef = useRef<HTMLFormElement>(null);

  interface FormState {
    success: boolean;
    timestamp: number;
  }

  /**
   * React 19 useActionState を使用した会計確定アクション
   */
  const [formState, formAction, _isPending] = useActionState(
    async (_prevState: FormState, _formData: FormData): Promise<FormState> => {
      if (!accounting || !calculation) return { success: false, timestamp: Date.now() };

      // BUG-371: 精算済会計の修正時は確認モーダル経由を強制
      // editConfirmed フラグが未設定の場合はモーダルを出して処理を中断
      if (accounting.status === "completed" && !editConfirmedRef.current) {
        setEditConfirmOpen(true);
        return { success: false, timestamp: Date.now() };
      }
      editConfirmedRef.current = false;

      // 決済 splits をリクエスト形式に変換
      const builtSplits = paymentSplits
        .filter((s) => s.amount && parseInt(s.amount, 10) > 0)
        .map((s) => {
          const amount = parseInt(s.amount, 10);
          const received = s.method === "cash" ? parseInt(s.receivedAmount || "0", 10) : 0;
          return {
            method: s.method as PaymentMethod,
            amount,
            received_amount: received,
            change_amount: s.method === "cash" ? Math.max(0, received - amount) : 0,
          };
        });

      // 代表支払方法 (cash > credit_card > electronic_money)
      const repMethod: PaymentMethod =
        builtSplits.some((s) => s.method === "cash") ? "cash" :
        builtSplits.some((s) => s.method === "credit_card") ? "credit_card" :
        "electronic_money";

      const cashSplit = builtSplits.find((s) => s.method === "cash");
      const totalReceived = cashSplit ? cashSplit.received_amount : calculation.billingAmount;
      const totalChange = cashSplit ? cashSplit.change_amount : 0;

      const paymentInfo: PaymentInfo = {
        subtotal: calculation.subtotal,
        taxTotal: calculation.taxTotal,
        totalAmount: calculation.totalAmount,
        insuranceAmount: calculation.insuranceAmount,
        discountAmount: 0,
        billingAmount: calculation.billingAmount,
        receivedAmount: totalReceived,
        changeAmount: totalChange,
        method: repMethod,
        insuranceRatio: hasInsurance ? parseFloat(insuranceRatio) : undefined,
      };

      try {
        if (!id) {
          // 新規会計: まず作成してから完了状態に更新
          const created = await createAccounting({
            pet_id: Number(accounting.petId),
            owner_id: Number(accounting.ownerId),
            scheduled_date: accounting.scheduledDate
              ? `${accounting.scheduledDate}T00:00:00Z`
              : new Date().toISOString(),
            subtotal: calculation.subtotal,
            tax_total: calculation.taxTotal,
            total_amount: calculation.totalAmount,
          });
          // BUG-ACCOUNTING-NEW-ITEMS-LOST: 新規作成後に明細を保存する
          for (const item of displayItems) {
            await createBillingItem({
              billing_id: Number(created.id),
              category: item.category,
              name: item.name,
              unit_price: item.unitPrice,
              quantity: item.quantity,
              tax_type: item.taxType,
              tax_rate: item.taxRate,
              is_insurance_applicable: item.isInsuranceApplicable,
              source: item.source,
            });
          }
          await updateAccounting(created.id, {
            status: "completed",
            subtotal: calculation.subtotal,
            tax_total: calculation.taxTotal,
            total_amount: calculation.totalAmount,
            insurance_ratio: hasInsurance ? parseFloat(insuranceRatio) : null,
            insurance_amount:
              calculation.insuranceAmount !== 0 ? calculation.insuranceAmount : null,
            billing_amount: calculation.billingAmount,
            received_amount: totalReceived,
            change_amount: totalChange,
            payment_method: repMethod,
            payment_splits: builtSplits,
            completed_at: new Date().toISOString(),
          });
          queryClient.invalidateQueries({ queryKey: ["accountings"] });
          toast.success("会計を登録・完了しました");
          navigate(paths.accounting.detail.getHref(created.id));
        } else {
          // 既存会計: 直接更新
          await updateAccounting(id, {
            status: "completed",
            subtotal: calculation.subtotal,
            tax_total: calculation.taxTotal,
            total_amount: calculation.totalAmount,
            insurance_ratio: hasInsurance ? parseFloat(insuranceRatio) : null,
            insurance_amount:
              calculation.insuranceAmount !== 0 ? calculation.insuranceAmount : null,
            billing_amount: calculation.billingAmount,
            received_amount: totalReceived,
            change_amount: totalChange,
            payment_method: repMethod,
            payment_splits: builtSplits,
            completed_at: new Date().toISOString(),
          });
          setCompletedPayment(paymentInfo);
          queryClient.invalidateQueries({ queryKey: ["accountings"] });
          queryClient.invalidateQueries({ queryKey: ["accounting", id] });
          queryClient.invalidateQueries({ queryKey: ["accounting-detail", id] });
          toast.success("会計を完了しました");
        }
        return { success: true, timestamp: Date.now() };
      } catch (error) {
        handleApiError(error, "会計の処理");
        return { success: false, timestamp: Date.now() };
      }
    },
    { success: false, timestamp: 0 }
  );

  // --- Focus Management (Accessibility) ---
  useEffect(() => {
    // If validation fails or amount is insufficient
    if (formState.success === false) {
      const element = document.getElementById("receivedAmount");
      if (element) {
        element.focus();
        element.scrollIntoView({ behavior: "smooth", block: "center" });
      }
    }
  }, [formState.success, formState.timestamp]);

  // clinic 情報（AccountingDocument に props 注入）
  const { user } = useAuth();
  const { canEdit, canCreate, canDelete } = usePermission("accounting");
  const canSubmit = id ? canEdit : canCreate;
  const clinicForDocument = useMemo(() => {
    const baseClinic = user?.clinic ?? null;
    if (!baseClinic) return null;
    return {
      ...baseClinic,
      invoiceRegistrationNumber,
    };
  // rerender-dependencies: user?.clinic（オブジェクト）の代わりに user（安定参照）を deps に使用
  }, [user, invoiceRegistrationNumber]);

  const handleAddItem = useCallback((name: string, price: string, category: string, taxRate?: number) => {
    const unitPrice = parseInt(price, 10);
    const qty = 1;
    const rate = taxRate ?? 0.1;
    const tempId = `manual_${crypto.randomUUID()}`;
    const newItem: AccountingItem = {
      id: tempId,
      category: category as ItemCategory,
      name,
      unitPrice,
      quantity: qty,
      taxType: "excluded" as TaxType,
      taxRate: rate,
      taxAmount: Math.round(unitPrice * qty * rate),
      subtotal: unitPrice * qty,
      isInsuranceApplicable: false,
      source: "manual",
    };

    // BUG-045: localItems が null (未編集) の場合は baseAccounting.items を seed として使用し、
    // 既存の治療明細を失わないようにする
    setLocalItems((prev) => [...(prev ?? baseItems), newItem]);
    setNewItemOpen(false);

    // 既存の会計 (id あり) の場合は POST API を呼び出してサーバーに永続化
    if (id) {
      startAddItemTransition(async () => {
        try {
          await createBillingItem({
            billing_id: Number(id),
            category,
            name,
            unit_price: unitPrice,
            quantity: qty,
            tax_type: "excluded",
            tax_rate: rate,
            is_insurance_applicable: false,
            source: "manual",
          });
          await queryClient.refetchQueries({ queryKey: queryKeys.accountings.detail(id) });
          setLocalItems(null);
          toast.success("明細を追加しました");
        } catch (error) {
          // 楽観的更新をロールバック
          setLocalItems((prev) => (prev ?? []).filter((i) => i.id !== tempId));
          handleApiError(error, "明細の追加");
        }
      });
    } else {
      // 新規会計（id なし）はローカル追加のみで API 呼び出しなし
      toast.success("明細を追加しました");
    }
  }, [id, queryClient, baseItems]);

  // BUG-045: 削除時も baseAccounting.items を seed として使用
  const handleDeleteItem = useCallback((itemId: string) => {
    setLocalItems((prev) => (prev ?? baseItems).filter((i) => i.id !== itemId));
  }, [baseItems]);

  const handleUpdateItemTax = useCallback(
    (itemId: string, taxType: TaxType, taxRate: number) => {
      if (!id) return;
      startTaxUpdateTransition(async () => {
        try {
          await updateBillingItem(itemId, { tax_type: taxType, tax_rate: taxRate });
          queryClient.invalidateQueries({ queryKey: queryKeys.accountings.detail(id) });
        } catch (error) {
          handleApiError(error, "税区分の更新");
        }
      });
    },
    [id, queryClient],
  );

  const handleRefund = useCallback(
    (amount: number, reason: string) => {
      if (!id) return;
      startRefundTransition(async () => {
        try {
          await createRefund(id, { amount, reason });
          queryClient.invalidateQueries({ queryKey: ["accounting-refunds", id] });
          queryClient.invalidateQueries({ queryKey: ["accountings"] });
          toast.success(`¥${amount.toLocaleString()} の返金を登録しました`);
      } catch (error) {
        handleApiError(error, "返金の登録");
      }
      });
    },
    [id, queryClient],
  );

  const handlePrint = useCallback(() => {
    setPreviewOpen(true);
  }, []);

  // BUG-371: 会計キャンセル（論理削除）実行
  const handleCancelConfirm = useCallback(() => {
    if (!id) return;
    startCancelTransition(async () => {
      try {
        await cancelAccounting(id);
        toast.success("会計をキャンセルしました");
        await queryClient.invalidateQueries({ queryKey: queryKeys.accountings.all() });
        navigate(paths.accounting.getHref());
      } catch (error) {
        handleApiError(error, "会計のキャンセル");
      } finally {
        setCancelConfirmOpen(false);
      }
    });
  }, [id, queryClient, navigate]);

  if (id && isLoading) return <LoadingFallback />;
  if (!accounting || !calculation) return <ErrorFallback message="データが見つかりません" />;

  return (
    <>
    <form ref={formRef} action={formAction}>
      <PageLayout
        className="print:hidden"
        title="会計精算"
        resource={ResourceAccounting}
        description={`受付No: ${accounting.id} | ${accounting.ownerName}様 - ${accounting.petName}ちゃん`}
        onBack={() => navigate(paths.accounting.getHref())}
        headerAction={
          accounting.status === "completed" ? (
            <div className="flex gap-2">
              <Button variant="outline" size="sm" onClick={handlePrint}>
                <Printer className={`mr-2 ${ICON.action}`} />
                明細兼領収書
              </Button>
              {canDelete ? (
                <Button
                  variant="outline"
                  size="sm"
                  onClick={() => setCancelConfirmOpen(true)}
                  className={C.danger}
                  disabled={isCancelling}
                >
                  会計をキャンセル
                </Button>
              ) : null}
            </div>
          ) : undefined
        }
      >
        {id && !canEdit ? (
          <div
            className={`flex items-center gap-2 px-4 py-2.5 rounded-md border mb-4 ${C.bgWarning50} ${C.borderWarning20} ${C.textWarning}`}
            role="status"
            aria-label="閲覧専用モード"
          >
            <EyeOff className={`shrink-0 h-4 w-4 ${C.textWarningIcon}`} aria-hidden="true" />
            <span className="text-sm font-medium">閲覧専用 — 編集権限がないため変更できません</span>
          </div>
        ) : null}
        <fieldset disabled={!canSubmit} className="border-0 p-0 m-0 min-w-0">
        <div className="flex flex-col lg:flex-row gap-6 h-[calc(100vh-140px)]">
          {/* 左カラム：明細リスト */}
          <div className="flex-1 flex flex-col gap-4 overflow-hidden">
            <ItemListCard
              items={accounting.items}
              subtotal={calculation.subtotal}
              taxTotal={calculation.taxTotal}
              totalAmount={calculation.totalAmount}
              newItemOpen={newItemOpen}
              onNewItemOpenChange={setNewItemOpen}
              onAddItem={handleAddItem}
              onDeleteItem={handleDeleteItem}
              accountingId={id}
              onUpdateItemTax={handleUpdateItemTax}
              canEdit={canEdit}
              canDelete={canDelete}
            />
          </div>

          {/* 右カラム：支払い情報 */}
          <div className="w-full lg:w-[400px] flex flex-col gap-4 overflow-y-auto">
            <InsuranceCard
              useInsurance={hasInsurance}
              onUseInsuranceChange={setHasInsurance}
              insuranceRatio={insuranceRatio}
              onInsuranceRatioChange={setInsuranceRatio}
              insuranceAmount={calculation.insuranceAmount}
            />

            <PaymentCard
              billingAmount={calculation.billingAmount}
              paymentSplits={paymentSplits}
              onSplitsChange={setPaymentSplits}
              isCompleted={accounting.status === "completed"}
              canEdit={canEdit}
              canCreate={canCreate}
              isEditMode={!!id}
            />

            {id && accounting.status === "completed" ? (
              <RefundSection
                accountingId={id}
                totalAmount={accounting.payment?.totalAmount ?? 0}
                isRefunding={isRefunding}
                onRefund={handleRefund}
                canEdit={canEdit}
              />
            ) : null}
          </div>
        </div>
        </fieldset>

        {/* Document Preview Modal */}
        <Dialog open={previewOpen} onOpenChange={setPreviewOpen}>
          <DialogContent className="max-w-4xl h-[90vh] overflow-hidden flex flex-col">
            <DialogHeader>
              <DialogTitle>明細兼領収書プレビュー</DialogTitle>
              <DialogDescription>印刷イメージを確認できます。</DialogDescription>
            </DialogHeader>
            <div className={`flex-1 ${C.bgActive} overflow-auto p-8 flex items-center justify-center`}>
              <div className="shadow-lg transform scale-100 origin-top">
                {accounting.payment ? (
                  <AccountingDocument
                    accounting={accounting}
                    paymentInfo={accounting.payment}
                    clinic={clinicForDocument}
                  />
                ) : null}
              </div>
            </div>
            <DialogFooter className="gap-2">
              <Button variant="outline" onClick={() => setPreviewOpen(false)}>
                閉じる
              </Button>
              <Button onClick={() => window.print()}>
                <Printer className={`mr-2 ${ICON.action}`} />
                印刷する
              </Button>
            </DialogFooter>
          </DialogContent>
        </Dialog>

        {/* BUG-371: 精算済修正の確認モーダル */}
        <ConfirmDialog
          open={editConfirmOpen}
          onClose={() => setEditConfirmOpen(false)}
          title="精算済みの会計を修正します"
          description="この操作は会計データに変更を加えます。よろしいですか?"
          confirmLabel="修正する"
          cancelLabel="キャンセル"
          onConfirm={() => {
            editConfirmedRef.current = true;
            setEditConfirmOpen(false);
            requestAnimationFrame(() => {
              formRef.current?.requestSubmit();
            });
          }}
        />

        {/* BUG-371: 会計キャンセル確認モーダル */}
        <ConfirmDialog
          open={cancelConfirmOpen}
          onClose={() => setCancelConfirmOpen(false)}
          title="この会計をキャンセルします"
          description="元に戻せません。キャンセルされた会計はステータスが「cancelled」になります。"
          confirmLabel="キャンセルする"
          cancelLabel="戻る"
          variant="destructive"
          isPending={isCancelling}
          onConfirm={handleCancelConfirm}
        />
      </PageLayout>
    </form>

      {/* Hidden Print Area */}
      {accounting.payment ? (
        <div className={`hidden print:block fixed inset-0 ${C.bgWhite} z-[9999] p-0 m-0 w-full h-full`}>
          <style type="text/css" media="print">
            {`
              @page { size: auto; margin: 0; }
              body { margin: 0; -webkit-print-color-adjust: exact; }
            `}
          </style>
          <div className="p-8">
            <AccountingDocument
              accounting={accounting}
              paymentInfo={accounting.payment}
              clinic={clinicForDocument}
            />
          </div>
        </div>
      ) : null}
    </>
  );
});
