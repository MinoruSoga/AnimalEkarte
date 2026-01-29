// React/Framework
import { useState, useMemo } from "react";
import { useParams, useNavigate, useLocation } from "react-router";

// External
import { Plus, Trash2, Save, CreditCard, Printer, FileText } from "lucide-react";
import { toast } from "sonner";

// Internal
import { PageLayout } from "@/components/shared/PageLayout";
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

// Relative
import { MOCK_ACCOUNTING_LIST } from "../api";
import { AccountingDocument } from "../components/AccountingDocument";

// Types
import type { Accounting, AccountingItem, PaymentInfo, ItemCategory, PaymentMethod } from "../types";

// Helper to create initial accounting data
function createInitialAccounting(
  id: string | undefined,
  locationState: { accountingItems?: AccountingItem[] } | null,
  locationSearch: string
): { accounting: Accounting | null; insurance: boolean; received: string; method: string } {
  // 新規作成
  if (!id) {
    const stateItems = locationState?.accountingItems || [];
    const searchParams = new URLSearchParams(locationSearch);
    const petId = searchParams.get('petId');

    return {
      accounting: {
        id: `acc_new_${Date.now()}`,
        ownerId: 'own_mock',
        ownerName: '新規 飼い主様',
        petId: petId || 'pet_mock',
        petName: '新規 ペットちゃん',
        petSpecies: '犬',
        status: 'waiting',
        scheduledDate: new Date().toISOString().split('T')[0],
        items: stateItems,
        payment: undefined
      },
      insurance: false,
      received: "",
      method: "cash"
    };
  }

  // 既存データ編集
  const data = MOCK_ACCOUNTING_LIST.find((a) => a.id === id);
  if (data) {
    return {
      accounting: { ...data },
      insurance: data.payment ? data.payment.insuranceAmount < 0 : false,
      received: data.payment ? data.payment.receivedAmount.toString() : "",
      method: data.payment ? data.payment.method : "cash"
    };
  }

  return { accounting: null, insurance: false, received: "", method: "cash" };
}

export const AccountingDetail = () => {
  const { id } = useParams();
  const navigate = useNavigate();
  const location = useLocation();

  // Compute initial values
  const locationState = location.state as { accountingItems?: AccountingItem[] } | null;
  const initialValues = useMemo(
    () => createInitialAccounting(id, locationState, location.search),
    [id, locationState, location.search]
  );

  const [accounting, setAccounting] = useState<Accounting | null>(initialValues.accounting);

  // 保険設定
  const [useInsurance, setUseInsurance] = useState(initialValues.insurance);
  const [insuranceRatio, setInsuranceRatio] = useState<string>("0.5"); // 50%, 70%

  // 支払い入力
  const [receivedAmount, setReceivedAmount] = useState<string>(initialValues.received);
  const [paymentMethod, setPaymentMethod] = useState<string>(initialValues.method);

  // 追加アイテム用State
  const [newItemOpen, setNewItemOpen] = useState(false);
  const [newItemName, setNewItemName] = useState("");
  const [newItemPrice, setNewItemPrice] = useState("");
  const [newItemCategory, setNewItemCategory] = useState("goods");

  // Document Preview State
  const [previewOpen, setPreviewOpen] = useState(false);
  const [previewType, setPreviewType] = useState<"receipt" | "statement">("receipt");

  // 金額計算
  const calculation = useMemo(() => {
    if (!accounting) return null;

    let subtotal = 0;
    let taxTotal = 0;

    accounting.items.forEach(item => {
      const price = item.unitPrice * item.quantity;
      const tax = Math.floor(price * item.taxRate);
      subtotal += price;
      taxTotal += tax;
    });

    const totalAmount = subtotal + taxTotal;
    
    // 保険計算
    let insuranceAmount = 0;
    if (useInsurance) {
      const insuranceTargetTotal = accounting.items
        .filter(item => item.isInsuranceApplicable)
        .reduce((sum, item) => sum + (item.unitPrice * item.quantity), 0);
      
      const ratio = parseFloat(insuranceRatio);
      insuranceAmount = Math.floor(insuranceTargetTotal * ratio) * -1; // マイナス値
    }

    const billingAmount = totalAmount + insuranceAmount; // 他の値引きがあればここに追加
    const received = parseInt(receivedAmount || "0", 10);
    const changeAmount = received > billingAmount ? received - billingAmount : 0;

    return {
      subtotal,
      taxTotal,
      totalAmount,
      insuranceAmount,
      billingAmount,
      received,
      changeAmount
    };
  }, [accounting, useInsurance, insuranceRatio, receivedAmount]);

  const handleAddItem = () => {
    if (!accounting || !newItemName || !newItemPrice) return;

    const newItem: AccountingItem = {
      id: `manual_${Date.now()}`,
      category: newItemCategory as ItemCategory,
      name: newItemName,
      unitPrice: parseInt(newItemPrice, 10),
      quantity: 1,
      taxRate: 0.1, // デフォルト10%
      isInsuranceApplicable: false, // 物販は基本的に対象外とする
      source: "manual"
    };

    setAccounting({
      ...accounting,
      items: [...accounting.items, newItem]
    });

    setNewItemName("");
    setNewItemPrice("");
    setNewItemOpen(false);
    toast.success("明細を追加しました");
  };

  const handleDeleteItem = (itemId: string) => {
    if (!accounting) return;
    setAccounting({
      ...accounting,
      items: accounting.items.filter(i => i.id !== itemId)
    });
  };

  const handleComplete = () => {
    if (!accounting || !calculation) return;

    // 保存処理（モックなのでコンソール出力のみ）
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
      insuranceRatio: useInsurance ? parseFloat(insuranceRatio) : undefined
    };

    setAccounting(prev => prev ? ({ ...prev, status: "completed", payment: paymentInfo }) : null);
    toast.success("会計を完了しました");
    // navigate("/accounting"); // 完了後も画面に留まって領収書発行を促す
  };

  const handlePrint = (type: "receipt" | "statement") => {
    setPreviewType(type);
    setPreviewOpen(true);
  };

  if (!accounting || !calculation) return <div>Loading...</div>;

  return (
    <>
    <PageLayout 
      className="print:hidden"
      title="会計精算" 
      description={`受付No: ${accounting.id} | ${accounting.ownerName}様 - ${accounting.petName}ちゃん`}
      onBack={() => navigate("/accounting")}
      headerAction={accounting.status === "completed" ? (
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
      ) : undefined}
    >
      <div className="flex flex-col lg:flex-row gap-6 h-[calc(100vh-140px)]">
        {/* 左カラム：明細リスト */}
        <div className="flex-1 flex flex-col gap-4 overflow-hidden">
          <Card className="flex-1 flex flex-col overflow-hidden">
            <CardHeader className="flex flex-row items-center justify-between py-4 border-b shrink-0">
              <CardTitle className="text-base font-medium">明細一覧</CardTitle>
              <Dialog open={newItemOpen} onOpenChange={setNewItemOpen}>
                <DialogTrigger asChild>
                  <Button variant="outline" size="sm" className="h-9">
                    <Plus className="mr-2 h-4 w-4" />
                    物販・その他追加
                  </Button>
                </DialogTrigger>
                <DialogContent>
                  <DialogHeader>
                    <DialogTitle>明細追加</DialogTitle>
                    <DialogDescription>
                        手動で明細項目を追加します。
                    </DialogDescription>
                  </DialogHeader>
                  <div className="space-y-4 py-4">
                    <div className="space-y-2">
                      <Label>区分</Label>
                      <Select value={newItemCategory} onValueChange={setNewItemCategory}>
                        <SelectTrigger><SelectValue /></SelectTrigger>
                        <SelectContent>
                          <SelectItem value="food">療法食・フード</SelectItem>
                          <SelectItem value="goods">物販・ケア用品</SelectItem>
                          <SelectItem value="other">その他</SelectItem>
                        </SelectContent>
                      </Select>
                    </div>
                    <div className="space-y-2">
                      <Label>品目名</Label>
                      <Input value={newItemName} onChange={e => setNewItemName(e.target.value)} placeholder="例: ロイヤルカナン 3kg" />
                    </div>
                    <div className="space-y-2">
                      <Label>単価 (税込)</Label>
                      <Input type="number" value={newItemPrice} onChange={e => setNewItemPrice(e.target.value)} placeholder="0" />
                    </div>
                  </div>
                  <DialogFooter>
                    <Button onClick={handleAddItem}>追加</Button>
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
                  {accounting.items.map((item) => (
                    <TableRow key={item.id} className="h-12">
                      <TableCell>
                        <Badge variant="outline" className="font-normal text-xs">
                          {item.category === 'examination' ? '診察' :
                           item.category === 'medicine' ? '処方' :
                           item.category === 'test' ? '検査' :
                           item.category === 'food' ? 'フード' : 'その他'}
                        </Badge>
                      </TableCell>
                      <TableCell className="font-medium">
                        {item.name}
                        {item.source === 'medical_record' && (
                          <span className="ml-2 text-[10px] text-blue-500 bg-blue-50 px-1.5 py-0.5 rounded">カルテ連携</span>
                        )}
                      </TableCell>
                      <TableCell className="text-right">¥{item.unitPrice.toLocaleString()}</TableCell>
                      <TableCell className="text-center">
                        <div className="flex items-center justify-center gap-2">
                           {/* 数量変更は今回省略 */}
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
                        {item.source === 'manual' && (
                          <Button 
                            variant="ghost" 
                            size="icon" 
                            className="h-8 w-8 text-red-500 hover:text-red-700 hover:bg-red-50"
                            onClick={() => handleDeleteItem(item.id)}
                          >
                            <Trash2 className="h-4 w-4" />
                          </Button>
                        )}
                      </TableCell>
                    </TableRow>
                  ))}
                </TableBody>
              </Table>
            </CardContent>
            {/* 明細フッター：小計 */}
            <div className="p-4 bg-gray-50 border-t flex justify-end gap-6 text-sm">
                <span>税抜小計: ¥{calculation.subtotal.toLocaleString()}</span>
                <span>消費税: ¥{calculation.taxTotal.toLocaleString()}</span>
                <span className="font-bold text-lg">合計: ¥{calculation.totalAmount.toLocaleString()}</span>
            </div>
          </Card>
        </div>

        {/* 右カラム：支払い情報 */}
        <div className="w-full lg:w-[400px] flex flex-col gap-4 overflow-y-auto">
          {/* 保険適用 */}
          <Card>
            <CardHeader className="py-3 px-4 bg-gray-50 border-b">
              <div className="flex items-center justify-between">
                <CardTitle className="text-sm font-medium flex items-center gap-2">
                  <CreditCard className="h-4 w-4" /> ペット保険（窓口精算）
                </CardTitle>
                <Switch checked={useInsurance} onCheckedChange={setUseInsurance} />
              </div>
            </CardHeader>
            {useInsurance && (
              <CardContent className="p-4 space-y-4">
                 <div className="space-y-2">
                    <Label className="text-xs">負担割合（保険会社が支払う割合）</Label>
                    <Select value={insuranceRatio} onValueChange={setInsuranceRatio}>
                        <SelectTrigger className="h-10"><SelectValue /></SelectTrigger>
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
                    <span>{calculation.insuranceAmount.toLocaleString()} 円</span>
                 </div>
              </CardContent>
            )}
          </Card>

          {/* 支払い・お釣り */}
          <Card className="flex-1">
             <CardHeader className="py-3 px-4 border-b">
                <CardTitle className="text-base font-medium">決済情報</CardTitle>
             </CardHeader>
             <CardContent className="p-6 space-y-6">
                
                {/* 請求額表示 */}
                <div className="text-center space-y-1">
                    <p className="text-sm text-gray-500">今回の請求金額</p>
                    <p className="text-4xl font-bold text-[#37352F]">
                        ¥{calculation.billingAmount.toLocaleString()}
                    </p>
                </div>

                <Separator />

                <div className="space-y-4">
                    <div className="space-y-2">
                        <Label>支払方法</Label>
                        <div className="grid grid-cols-3 gap-2">
                            <Button 
                                type="button" 
                                variant={paymentMethod === 'cash' ? 'default' : 'outline'}
                                onClick={() => setPaymentMethod('cash')}
                                className="h-12"
                            >
                                現金
                            </Button>
                            <Button 
                                type="button" 
                                variant={paymentMethod === 'credit_card' ? 'default' : 'outline'}
                                onClick={() => setPaymentMethod('credit_card')}
                                className="h-12"
                            >
                                カード
                            </Button>
                            <Button 
                                type="button" 
                                variant={paymentMethod === 'electronic_money' ? 'default' : 'outline'}
                                onClick={() => setPaymentMethod('electronic_money')}
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
                                onChange={(e) => setReceivedAmount(e.target.value)}
                            />
                            <span className="absolute right-3 top-4 text-gray-400">円</span>
                        </div>
                        <div className="flex gap-2 justify-end">
                            <Button variant="outline" size="sm" onClick={() => setReceivedAmount(calculation.billingAmount.toString())}>
                                丁度
                            </Button>
                            <Button variant="outline" size="sm" onClick={() => setReceivedAmount((Math.ceil(calculation.billingAmount / 1000) * 1000).toString())}>
                                千円単位
                            </Button>
                            <Button variant="outline" size="sm" onClick={() => setReceivedAmount((Math.ceil(calculation.billingAmount / 10000) * 10000).toString())}>
                                一万単位
                            </Button>
                        </div>
                    </div>
                </div>

                <div className="bg-gray-100 p-4 rounded-lg flex justify-between items-center">
                    <span className="font-bold text-gray-600">お釣り</span>
                    <span className={`text-2xl font-bold ${calculation.changeAmount < 0 ? 'text-red-500' : 'text-gray-900'}`}>
                        ¥{calculation.changeAmount.toLocaleString()}
                    </span>
                </div>

                <Button 
                    className="w-full h-14 text-lg font-bold mt-4" 
                    size="lg"
                    onClick={handleComplete}
                    disabled={calculation.changeAmount < 0 || !receivedAmount || accounting.status === "completed"}
                >
                    <Save className="mr-2 h-5 w-5" />
                    {accounting.status === "completed" ? "精算完了済み" : "会計を確定する"}
                </Button>

             </CardContent>
          </Card>
        </div>
      </div>

      {/* Document Preview Modal */}
      <Dialog open={previewOpen} onOpenChange={setPreviewOpen}>
        <DialogContent className="max-w-4xl h-[90vh] overflow-hidden flex flex-col">
            <DialogHeader>
                <DialogTitle>{previewType === "receipt" ? "領収書プレビュー" : "診療明細書プレビュー"}</DialogTitle>
                <DialogDescription>
                    印刷イメージを確認できます。
                </DialogDescription>
            </DialogHeader>
            <div className="flex-1 bg-gray-100 overflow-auto p-8 flex items-center justify-center">
                <div className="shadow-lg transform scale-100 origin-top">
                    {accounting && accounting.payment && (
                        <AccountingDocument 
                            type={previewType} 
                            accounting={accounting} 
                            paymentInfo={accounting.payment} 
                        />
                    )}
                </div>
            </div>
            <DialogFooter className="gap-2">
                <Button variant="outline" onClick={() => setPreviewOpen(false)}>閉じる</Button>
                <Button onClick={() => {
                    window.print();
                    // setPreviewOpen(false); // 印刷ダイアログが開いている間は閉じない方が自然かも
                }}>
                    <Printer className="mr-2 h-4 w-4" />
                    印刷する
                </Button>
            </DialogFooter>
        </DialogContent>
      </Dialog>
    </PageLayout>

    {/* Hidden Print Area */}
    {accounting && accounting.payment && (
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
                />
            </div>
        </div>
    )}
    </>
  );
}
