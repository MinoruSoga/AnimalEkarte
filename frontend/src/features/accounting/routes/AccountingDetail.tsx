// React/Framework
import { useState, useMemo, useCallback, memo, useTransition } from "react";
import { useParams, useNavigate, useLocation } from "react-router";

// External
import { Plus, Trash2, Save, CreditCard, Printer, FileText } from "lucide-react";
import { useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";

// Internal
import { PageLayout } from "@/components/shared/PageLayout/PageLayout";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Badge } from "@/components/ui/badge";
import { Switch } from "@/components/ui/switch";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
  DialogFooter,
  DialogDescription,
} from "@/components/ui/dialog";
import { Separator } from "@/components/ui/separator";

// Shared Hooks
import { useAuth } from "@/features/auth/hooks/use-auth";

// Relative
import { useGetAccountingDetail } from "../api/get-accounting";
import { updateAccounting } from "../api/update-accounting";
import { AccountingDocument } from "../components/AccountingDocument";
import { paths } from "@/config/paths";

// Types
import type { Accounting, AccountingItem, PaymentInfo, ItemCategory, PaymentMethod } from "../types";

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
  onAddItem: (name: string, price: string, category: string) => void;
  onDeleteItem: (id: string) => void;
}

const ItemListCard = memo(function ItemListCard({
  items,
  subtotal,
  taxTotal,
  totalAmount,
  newItemOpen,
  onNewItemOpenChange,
  onAddItem,
  onDeleteItem,
}: ItemListCardProps) {
  const [newItemName, setNewItemName] = useState("");
  const [newItemPrice, setNewItemPrice] = useState("");
  const [newItemCategory, setNewItemCategory] = useState("goods");

  const handleAdd = useCallback(() => {
    if (!newItemName || !newItemPrice) return;
    onAddItem(newItemName, newItemPrice, newItemCategory);
    setNewItemName("");
    setNewItemPrice("");
  }, [newItemName, newItemPrice, newItemCategory, onAddItem]);

  return (
    <Card className="flex-1 flex flex-col overflow-hidden">
      <CardHeader className="flex flex-row items-center justify-between py-4 border-b shrink-0">
        <CardTitle className="text-base font-medium">明細一覧</CardTitle>
        <Dialog open={newItemOpen} onOpenChange={onNewItemOpenChange}>
          <DialogTrigger asChild>
            <Button variant="outline" size="sm" className="h-9">
              <Plus className="mr-2 h-4 w-4" />
              物販・その他追加
            </Button>
          </DialogTrigger>
          <DialogContent>
            <DialogHeader>
              <DialogTitle>明細追加</DialogTitle>
              <DialogDescription>手動で明細項目を追加します。</DialogDescription>
            </DialogHeader>
            <div className="space-y-4 py-4">
              <div className="space-y-2">
                <Label>区分</Label>
                <Select value={newItemCategory} onValueChange={setNewItemCategory}>
                  <SelectTrigger>
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="food">療法食・フード</SelectItem>
                    <SelectItem value="goods">物販・ケア用品</SelectItem>
                    <SelectItem value="other">その他</SelectItem>
                  </SelectContent>
                </Select>
              </div>
              <div className="space-y-2">
                <Label>品目名</Label>
                <Input
                  value={newItemName}
                  onChange={(e) => setNewItemName(e.target.value)}
                  placeholder="例: ロイヤルカナン 3kg"
                />
              </div>
              <div className="space-y-2">
                <Label>単価 (税込)</Label>
                <Input
                  type="number"
                  value={newItemPrice}
                  onChange={(e) => setNewItemPrice(e.target.value)}
                  placeholder="0"
                />
              </div>
            </div>
            <DialogFooter>
              <Button onClick={handleAdd}>追加</Button>
            </DialogFooter>
          </DialogContent>
        </Dialog>
      </CardHeader>
      <CardContent className="p-0 overflow-auto flex-1">
        <Table>
          <TableHeader className="sticky top-0 bg-white z-10 shadow-sm">
            <TableRow>
              <TableHead className="w-[100px]">区分</TableHead>
              <TableHead>項目名</TableHead>
              <TableHead className="text-right w-[100px]">単価</TableHead>
              <TableHead className="text-center w-[80px]">数量</TableHead>
              <TableHead className="text-center w-[80px]">保険</TableHead>
              <TableHead className="text-right w-[120px]">金額</TableHead>
              <TableHead className="w-[50px]"></TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {items.map((item) => (
              <TableRow key={item.id} className="h-12">
                <TableCell>
                  <Badge variant="outline" className="font-normal text-xs">
                    {CATEGORY_LABELS[item.category as ItemCategory] ?? "その他"}
                  </Badge>
                </TableCell>
                <TableCell className="font-medium">
                  {item.name}
                  {item.source === "medical_record" ? (
                    <span className="ml-2 text-[10px] text-blue-500 bg-blue-50 px-1.5 py-0.5 rounded">
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
                  {item.isInsuranceApplicable ? (
                    <span className="text-green-600 font-bold text-xs">●</span>
                  ) : (
                    <span className="text-gray-300 text-xs">-</span>
                  )}
                </TableCell>
                <TableCell className="text-right font-medium">
                  ¥{(item.unitPrice * item.quantity).toLocaleString()}
                </TableCell>
                <TableCell>
                  {item.source === "manual" ? (
                    <Button
                      variant="ghost"
                      size="icon"
                      className="h-8 w-8 text-red-500 hover:text-red-700 hover:bg-red-50"
                      onClick={() => onDeleteItem(item.id)}
                    >
                      <Trash2 className="h-4 w-4" />
                    </Button>
                  ) : null}
                </TableCell>
              </TableRow>
            ))}
          </TableBody>
        </Table>
      </CardContent>
      <div className="p-4 bg-gray-50 border-t flex justify-end gap-6 text-sm">
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

const InsuranceCard = memo(function InsuranceCard({
  useInsurance,
  onUseInsuranceChange,
  insuranceRatio,
  onInsuranceRatioChange,
  insuranceAmount,
}: InsuranceCardProps) {
  return (
    <Card>
      <CardHeader className="py-3 px-4 bg-gray-50 border-b">
        <div className="flex items-center justify-between">
          <CardTitle className="text-sm font-medium flex items-center gap-2">
            <CreditCard className="h-4 w-4" /> ペット保険（窓口精算）
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
              <SelectContent>
                <SelectItem value="0.5">50%</SelectItem>
                <SelectItem value="0.7">70%</SelectItem>
                <SelectItem value="0.9">90%</SelectItem>
                <SelectItem value="1.0">100%</SelectItem>
              </SelectContent>
            </Select>
          </div>
          <div className="flex justify-between items-center text-sm font-medium text-green-700 bg-green-50 p-2 rounded">
            <span>保険負担額（マイナス）</span>
            <span>{insuranceAmount.toLocaleString()} 円</span>
          </div>
        </CardContent>
      ) : null}
    </Card>
  );
});

// ── 決済カード ──────────────────────────────────────────

interface PaymentCardProps {
  billingAmount: number;
  changeAmount: number;
  paymentMethod: string;
  onPaymentMethodChange: (v: string) => void;
  receivedAmount: string;
  onReceivedAmountChange: (v: string) => void;
  onComplete: () => void;
  isCompleted: boolean;
  isSaving: boolean;
}

const PaymentCard = memo(function PaymentCard({
  billingAmount,
  changeAmount,
  paymentMethod,
  onPaymentMethodChange,
  receivedAmount,
  onReceivedAmountChange,
  onComplete,
  isCompleted,
  isSaving,
}: PaymentCardProps) {
  return (
    <Card className="flex-1">
      <CardHeader className="py-3 px-4 border-b">
        <CardTitle className="text-base font-medium">決済情報</CardTitle>
      </CardHeader>
      <CardContent className="p-6 space-y-6">
        <div className="text-center space-y-1">
          <p className="text-sm text-gray-500">今回の請求金額</p>
          <p className="text-4xl font-bold text-[#37352F]">
            ¥{billingAmount.toLocaleString()}
          </p>
        </div>

        <Separator />

        <div className="space-y-4">
          <div className="space-y-2">
            <Label>支払方法</Label>
            <div className="grid grid-cols-3 gap-2">
              <Button
                type="button"
                variant={paymentMethod === "cash" ? "default" : "outline"}
                onClick={() => onPaymentMethodChange("cash")}
                className="h-12"
              >
                現金
              </Button>
              <Button
                type="button"
                variant={paymentMethod === "credit_card" ? "default" : "outline"}
                onClick={() => onPaymentMethodChange("credit_card")}
                className="h-12"
              >
                カード
              </Button>
              <Button
                type="button"
                variant={paymentMethod === "electronic_money" ? "default" : "outline"}
                onClick={() => onPaymentMethodChange("electronic_money")}
                className="h-12"
              >
                電子マネー
              </Button>
            </div>
          </div>

          <div className="space-y-2">
            <Label>お預かり金額</Label>
            <div className="relative">
              <Input
                type="number"
                className="h-14 text-xl font-bold text-right pr-10"
                placeholder="0"
                value={receivedAmount}
                onChange={(e) => onReceivedAmountChange(e.target.value)}
              />
              <span className="absolute right-3 top-4 text-gray-400">円</span>
            </div>
            <div className="flex gap-2 justify-end">
              <Button
                variant="outline"
                size="sm"
                onClick={() => onReceivedAmountChange(billingAmount.toString())}
              >
                丁度
              </Button>
              <Button
                variant="outline"
                size="sm"
                onClick={() =>
                  onReceivedAmountChange(
                    (Math.ceil(billingAmount / 1000) * 1000).toString(),
                  )
                }
              >
                千円単位
              </Button>
              <Button
                variant="outline"
                size="sm"
                onClick={() =>
                  onReceivedAmountChange(
                    (Math.ceil(billingAmount / 10000) * 10000).toString(),
                  )
                }
              >
                一万単位
              </Button>
            </div>
          </div>
        </div>

        <div className="bg-gray-100 p-4 rounded-lg flex justify-between items-center">
          <span className="font-bold text-gray-600">お釣り</span>
          <span
            className={`text-2xl font-bold ${changeAmount < 0 ? "text-red-500" : "text-gray-900"}`}
          >
            ¥{changeAmount.toLocaleString()}
          </span>
        </div>

        <Button
          className="w-full h-14 text-lg font-bold mt-4"
          size="lg"
          onClick={onComplete}
          disabled={
            changeAmount < 0 ||
            !receivedAmount ||
            isCompleted ||
            isSaving
          }
        >
          <Save className="mr-2 h-5 w-5" />
          {isCompleted ? "精算完了済み" : isSaving ? "処理中..." : "会計を確定する"}
        </Button>
      </CardContent>
    </Card>
  );
});

// ── メインコンポーネント ──────────────────────────────────

export function AccountingDetail() {
  const { id } = useParams();
  const navigate = useNavigate();
  const location = useLocation();
  const queryClient = useQueryClient();
  const [isSaving, startSaveTransition] = useTransition();

  const locationState = location.state as { accountingItems?: AccountingItem[] } | null;

  // 既存データは API から取得
  const { data: fetchedAccounting, isLoading } = useGetAccountingDetail(id);

  // 新規作成の場合は location.state から items を引き継ぐ（派生データ）
  const baseAccounting = useMemo<Accounting | null>(() => {
    if (id) {
      return fetchedAccounting ?? null;
    }
    const searchParams = new URLSearchParams(location.search);
    const petId = searchParams.get("petId");
    const stateItems = locationState?.accountingItems ?? [];
    return {
      id: "acc_new",
      ownerId: "",
      ownerName: "新規 飼い主様",
      petId: petId ?? "",
      petName: "新規 ペットちゃん",
      petSpecies: "犬",
      status: "waiting",
      scheduledDate: new Date().toISOString().split("T")[0],
      items: stateItems,
      payment: undefined,
    };
  }, [id, fetchedAccounting, locationState, location.search]);

  // ユーザー操作による追加・削除を管理するローカル明細
  const [localItems, setLocalItems] = useState<AccountingItem[] | null>(null);

  // 表示する明細: ローカル編集があればそちら優先
  const displayItems = useMemo(
    () => localItems ?? baseAccounting?.items ?? [],
    [localItems, baseAccounting?.items],
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
  const [receivedAmount, setReceivedAmount] = useState(() => baseAccounting?.payment?.receivedAmount?.toString() ?? "");
  const [paymentMethod, setPaymentMethod] = useState(() => baseAccounting?.payment?.method ?? "cash");

  // 追加アイテム用State
  const [newItemOpen, setNewItemOpen] = useState(false);

  // Document Preview State
  const [previewOpen, setPreviewOpen] = useState(false);
  const [previewType, setPreviewType] = useState<"receipt" | "statement">("receipt");

  // clinic 情報（AccountingDocument に props 注入）
  const { user } = useAuth();
  const clinic = user?.clinic ?? null;

  // 金額計算
  const calculation = useMemo(() => {
    if (!accounting) return null;

    let subtotal = 0;
    let taxTotal = 0;

    for (const item of accounting.items) {
      const price = item.unitPrice * item.quantity;
      const tax = Math.floor(price * item.taxRate);
      subtotal += price;
      taxTotal += tax;
    }

    const totalAmount = subtotal + taxTotal;

    let insuranceAmount = 0;
    if (hasInsurance) {
      const insuranceTargetTotal = accounting.items
        .filter((item) => item.isInsuranceApplicable)
        .reduce((sum, item) => sum + item.unitPrice * item.quantity, 0);

      const ratio = parseFloat(insuranceRatio);
      insuranceAmount = Math.floor(insuranceTargetTotal * ratio) * -1;
    }

    const billingAmount = totalAmount + insuranceAmount;
    const received = parseInt(receivedAmount || "0", 10);
    const changeAmount = received > billingAmount ? received - billingAmount : 0;

    return {
      subtotal,
      taxTotal,
      totalAmount,
      insuranceAmount,
      billingAmount,
      received,
      changeAmount,
    };
  }, [accounting, hasInsurance, insuranceRatio, receivedAmount]);

  const handleAddItem = useCallback((name: string, price: string, category: string) => {
    const newItem: AccountingItem = {
      id: `manual_${crypto.randomUUID()}`,
      category: category as ItemCategory,
      name,
      unitPrice: parseInt(price, 10),
      quantity: 1,
      taxRate: 0.1,
      isInsuranceApplicable: false,
      source: "manual",
    };

    setLocalItems((prev) => [...(prev ?? []), newItem]);
    setNewItemOpen(false);
    toast.success("明細を追加しました");
  }, []);

  const handleDeleteItem = useCallback((itemId: string) => {
    setLocalItems((prev) => (prev ?? []).filter((i) => i.id !== itemId));
  }, []);

  const handleComplete = useCallback(() => {
    if (!accounting || !calculation || !id) return;

    const paymentInfo: PaymentInfo = {
      subtotal: calculation.subtotal,
      taxTotal: calculation.taxTotal,
      totalAmount: calculation.totalAmount,
      insuranceAmount: calculation.insuranceAmount,
      discountAmount: 0,
      billingAmount: calculation.billingAmount,
      receivedAmount: calculation.received,
      changeAmount: calculation.changeAmount,
      method: paymentMethod as PaymentMethod,
      insuranceRatio: hasInsurance ? parseFloat(insuranceRatio) : undefined,
    };

    startSaveTransition(async () => {
      try {
        await updateAccounting(id, {
          status: "completed",
          subtotal: calculation.subtotal,
          tax_total: calculation.taxTotal,
          total_amount: calculation.totalAmount,
          insurance_ratio: hasInsurance ? parseFloat(insuranceRatio) : null,
          insurance_amount:
            calculation.insuranceAmount !== 0 ? calculation.insuranceAmount : null,
          billing_amount: calculation.billingAmount,
          received_amount: calculation.received,
          change_amount: calculation.changeAmount,
          payment_method: paymentMethod as PaymentMethod,
          completed_at: new Date().toISOString(),
        });
        setCompletedPayment(paymentInfo);
        queryClient.invalidateQueries({ queryKey: ["accountings"] });
        queryClient.invalidateQueries({ queryKey: ["accounting", id] });
        queryClient.invalidateQueries({ queryKey: ["accounting-detail", id] });
        toast.success("会計を完了しました");
      } catch {
        toast.error("会計の更新に失敗しました");
      }
    });
  }, [accounting, calculation, paymentMethod, hasInsurance, insuranceRatio, id, queryClient]);

  const handlePrint = useCallback((type: "receipt" | "statement") => {
    setPreviewType(type);
    setPreviewOpen(true);
  }, []);

  if (id && isLoading) return <div>読み込み中...</div>;
  if (!accounting || !calculation) return <div>データが見つかりません</div>;

  return (
    <>
      <PageLayout
        className="print:hidden"
        title="会計精算"
        description={`受付No: ${accounting.id} | ${accounting.ownerName}様 - ${accounting.petName}ちゃん`}
        onBack={() => navigate(paths.accounting.getHref())}
        headerAction={
          accounting.status === "completed" ? (
            <div className="flex gap-2">
              <Button variant="outline" size="sm" onClick={() => handlePrint("statement")}>
                <FileText className="mr-2 h-4 w-4" />
                診療明細書
              </Button>
              <Button variant="outline" size="sm" onClick={() => handlePrint("receipt")}>
                <Printer className="mr-2 h-4 w-4" />
                領収書発行
              </Button>
            </div>
          ) : undefined
        }
      >
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
              changeAmount={calculation.changeAmount}
              paymentMethod={paymentMethod}
              onPaymentMethodChange={(v) => setPaymentMethod(v as PaymentMethod)}
              receivedAmount={receivedAmount}
              onReceivedAmountChange={setReceivedAmount}
              onComplete={handleComplete}
              isCompleted={accounting.status === "completed"}
              isSaving={isSaving}
            />
          </div>
        </div>

        {/* Document Preview Modal */}
        <Dialog open={previewOpen} onOpenChange={setPreviewOpen}>
          <DialogContent className="max-w-4xl h-[90vh] overflow-hidden flex flex-col">
            <DialogHeader>
              <DialogTitle>
                {previewType === "receipt" ? "領収書プレビュー" : "診療明細書プレビュー"}
              </DialogTitle>
              <DialogDescription>印刷イメージを確認できます。</DialogDescription>
            </DialogHeader>
            <div className="flex-1 bg-gray-100 overflow-auto p-8 flex items-center justify-center">
              <div className="shadow-lg transform scale-100 origin-top">
                {accounting.payment ? (
                  <AccountingDocument
                    type={previewType}
                    accounting={accounting}
                    paymentInfo={accounting.payment}
                    clinic={clinic}
                  />
                ) : null}
              </div>
            </div>
            <DialogFooter className="gap-2">
              <Button variant="outline" onClick={() => setPreviewOpen(false)}>
                閉じる
              </Button>
              <Button onClick={() => window.print()}>
                <Printer className="mr-2 h-4 w-4" />
                印刷する
              </Button>
            </DialogFooter>
          </DialogContent>
        </Dialog>
      </PageLayout>

      {/* Hidden Print Area */}
      {accounting.payment ? (
        <div className="hidden print:block fixed inset-0 bg-white z-[9999] p-0 m-0 w-full h-full">
          <style type="text/css" media="print">
            {`
              @page { size: auto; margin: 0; }
              body { margin: 0; -webkit-print-color-adjust: exact; }
            `}
          </style>
          <div className="p-8">
            <AccountingDocument
              type={previewType}
              accounting={accounting}
              paymentInfo={accounting.payment}
              clinic={clinic}
            />
          </div>
        </div>
      ) : null}
    </>
  );
}
