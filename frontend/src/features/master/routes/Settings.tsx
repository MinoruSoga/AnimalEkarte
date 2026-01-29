// React/Framework
import { useState } from "react";

// External
import { Plus, Settings as SettingsIcon, Save, Trash2, Package } from "lucide-react";
import { toast } from "sonner";

// Internal
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { TableCell } from "@/components/ui/table";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { PageLayout } from "@/components/shared/PageLayout";
import { SearchFilterBar } from "@/components/shared/SearchFilterBar";
import { DataTable } from "@/components/shared/DataTable";
import { PrimaryButton } from "@/components/shared/Form";
import { StatusBadge } from "@/components/shared/StatusBadge";
import { DataTableRow } from "@/components/shared/DataTable";
import { RowActionButton } from "@/components/shared/RowActionButton";
import { getMasterStatusColor } from "@/utils/status-helpers";

// Relative
import { useMasterItems } from "@/hooks/use-master-items";

// Types
import type { MasterItem, InventoryItem } from "@/types";

interface SettingsPageProps {
    category?: string;
}

export const Settings = ({ category: propCategory }: SettingsPageProps) => {
  const category = propCategory || "examination"; 
  
  const [isEditing, setIsEditing] = useState(false);
  const [selectedItem, setSelectedItem] = useState<MasterItem | null>(null);
  const [searchTerm, setSearchTerm] = useState("");
  const [formData, setFormData] = useState<Partial<MasterItem>>({});
  
  const { data: filteredItems, add, update, remove } = useMasterItems(category, searchTerm);
  // Inventory feature is currently disabled/missing
  const inventory: InventoryItem[] = [];

  const getTabLabel = (tab: string) => {
    switch (tab) {
        case "examination": return "検査マスタ";
        case "vaccine": return "予防接種マスタ";
        case "medicine": return "薬剤マスタ";
        case "staff": return "スタッフマスタ";
        case "insurance": return "保険マスタ";
        case "consultation": return "診察マスタ";
        case "serviceType": return "予約区分マスタ";
        case "procedure": return "処置マスタ";
        case "hospitalization": return "入院マスタ";
        case "cage": return "ケージマスタ";
        default: return "マスタ";
    }
  }

  const getCategoryDefault = (tab: string) => {
     const map: Record<string, string> = {
         "examination": "検査",
         "vaccine": "予防",
         "medicine": "薬剤",
         "staff": "スタッフ",
         "insurance": "保険",
         "consultation": "診察",
         "serviceType": "診療内容",
         "procedure": "処置",
         "hospitalization": "入院",
         "cage": "ケージ",
     };
     return map[tab] || "";
  }

  const shouldShowPrice = (cat: string) => {
      return !["staff", "cage", "insurance"].includes(cat);
  }

  const shouldShowInventory = (_cat: string) => {
      return false; // Inventory management is disabled
      // return ["examination", "vaccine", "medicine", "procedure", "hospitalization"].includes(cat);
  }

  const getLabels = (cat: string) => {
      if (cat === "staff") return { code: "社員番号", name: "氏名", category: "職種" };
      if (cat === "cage") return { code: "コード", name: "ケージ名", category: "エリア" };
      if (cat === "insurance") return { code: "コード", name: "保険会社名", category: "種別" };
      if (cat === "medicine") return { code: "コード", name: "薬品名", category: "薬効分類" };
      return { code: "コード", name: "名称", category: "カテゴリ" };
  }

  const labels = getLabels(category);

  // Reset editing state when category changes (render-time state derivation)
  const [prevCategory, setPrevCategory] = useState(category);
  if (category !== prevCategory) {
    setPrevCategory(category);
    setIsEditing(false);
    setSelectedItem(null);
    setSearchTerm("");
    setFormData({});
  }

  const handleEdit = (item: MasterItem) => {
    setSelectedItem(item);
    setFormData({...item});
    setIsEditing(true);
  };

  const handleCreate = () => {
    setSelectedItem(null);
    setFormData({
        category: getCategoryDefault(category),
        status: "active",
        price: 0,
        inventoryId: undefined,
        defaultQuantity: 1
    });
    setIsEditing(true);
  };

  const handleCloseEdit = () => {
    setIsEditing(false);
    setSelectedItem(null);
    setFormData({});
  }

  const handleSave = () => {
      if (!formData.code || !formData.name) {
          toast.error("コードと名称は必須です");
          return;
      }
      
      try {
        if (selectedItem) {
            update(selectedItem.id, formData);
            toast.success("更新しました");
        } else {
            add(formData as Omit<MasterItem, "id">);
            toast.success("登録しました");
        }
        setIsEditing(false);
      } catch {
          toast.error("保存中にエラーが発生しました");
      }
  };

  const handleDelete = () => {
    if (selectedItem) {
        if (window.confirm("本当に削除しますか？")) {
            remove(selectedItem.id);
            setIsEditing(false);
            toast.success("削除しました");
        }
    }
  };

  const columns = [
    { header: labels.code, className: "w-[120px]" },
    { header: labels.name },
    { header: labels.category, className: "w-[100px]" },
    ...(shouldShowPrice(category) ? [{ header: "単価", className: "w-[100px]", align: "right" as const }] : []),
    ...(shouldShowInventory(category) ? [{ header: "在庫連携", className: "w-[150px]" }] : []),
    { header: "ステータス", className: "w-[100px]", align: "center" as const },
    { header: "操作", className: "w-[80px]", align: "right" as const },
  ];

  // Edit Form Component
  if (isEditing) {
      return (
        <PageLayout
          title={selectedItem ? `${getTabLabel(category)} 編集` : `${getTabLabel(category)} 新規登録`}
          onBack={handleCloseEdit}
          maxWidth="max-w-2xl"
          align="left"
        >
            <div className="bg-white p-6 rounded-lg border border-[rgba(55,53,47,0.16)] space-y-4 shadow-sm">
                <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                    <div className="space-y-1.5">
                        <Label className="text-sm text-[#37352F]/60">{labels.code} <span className="text-red-500">*</span></Label>
                        <Input 
                            className="h-10 text-sm bg-white text-[#37352F] border-[rgba(55,53,47,0.16)]" 
                            value={formData.code || ""}
                            onChange={(e) => setFormData({...formData, code: e.target.value})}
                            placeholder={`例: ${category === 'staff' ? 'ST-001' : 'EX-001'}`} 
                        />
                    </div>
                    <div className="space-y-1.5">
                        <Label className="text-sm text-[#37352F]/60">{labels.name} <span className="text-red-500">*</span></Label>
                        <Input 
                            className="h-10 text-sm bg-white text-[#37352F] border-[rgba(55,53,47,0.16)]" 
                            value={formData.name || ""}
                            onChange={(e) => setFormData({...formData, name: e.target.value})}
                            placeholder={`例: ${category === 'staff' ? '山田 太郎' : '血液検査'}`} 
                        />
                    </div>
                </div>

                <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                        <div className="space-y-1.5">
                        <Label className="text-sm text-[#37352F]/60">{labels.category}</Label>
                        <Input 
                            className="h-10 text-sm bg-white text-[#37352F] border-[rgba(55,53,47,0.16)]" 
                            value={formData.category || ""}
                            onChange={(e) => setFormData({...formData, category: e.target.value})}
                            placeholder={`例: ${category === 'staff' ? '獣医師' : '血液'}`} 
                        />
                    </div>
                    {shouldShowPrice(category) && (
                        <div className="space-y-1.5">
                            <Label className="text-sm text-[#37352F]/60">単価 (円)</Label>
                            <Input 
                                className="h-10 text-sm bg-white text-[#37352F] border-[rgba(55,53,47,0.16)]" 
                                type="number" 
                                value={formData.price || 0}
                                onChange={(e) => setFormData({...formData, price: Number(e.target.value)})}
                                placeholder="0" 
                            />
                        </div>
                    )}
                </div>

                {/* Inventory Link Section */}
                {shouldShowInventory(category) && (
                    <div className="pt-2 pb-2">
                        <div className="bg-gray-50/80 p-4 rounded-md border border-gray-100 space-y-3">
                            <div className="flex items-center gap-2 mb-2">
                                <Package className="w-4 h-4 text-gray-500" />
                                <Label className="text-sm font-medium text-[#37352F]">在庫連携設定</Label>
                            </div>
                            
                            <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
                                <div className="md:col-span-2 space-y-1.5">
                                    <Label className="text-xs text-[#37352F]/60">連携する在庫アイテム</Label>
                                    <Select 
                                        value={formData.inventoryId || "none"} 
                                        onValueChange={(val) => setFormData({...formData, inventoryId: val === "none" ? undefined : val})}
                                    >
                                        <SelectTrigger className="h-10 text-sm bg-white text-[#37352F] border-[rgba(55,53,47,0.16)]">
                                            <SelectValue placeholder="連携なし" />
                                        </SelectTrigger>
                                        <SelectContent>
                                            <SelectItem value="none" className="text-muted-foreground">連携なし</SelectItem>
                                            {inventory.map(inv => (
                                                <SelectItem key={inv.id} value={inv.id}>
                                                    {inv.name} <span className="text-xs text-muted-foreground ml-2">({inv.category})</span>
                                                </SelectItem>
                                            ))}
                                        </SelectContent>
                                    </Select>
                                </div>
                                <div className="space-y-1.5">
                                    <Label className="text-xs text-[#37352F]/60">消費数量</Label>
                                    <Input 
                                        className="h-10 text-sm bg-white text-[#37352F] border-[rgba(55,53,47,0.16)]" 
                                        type="number" 
                                        value={formData.defaultQuantity || 1}
                                        onChange={(e) => setFormData({...formData, defaultQuantity: Number(e.target.value)})}
                                        placeholder="1"
                                        min={1}
                                    />
                                </div>
                            </div>
                            <p className="text-xs text-muted-foreground mt-1">
                                ※ この処置が実施されると、選択された在庫アイテムから指定された数量が自動的に減算されます。
                            </p>
                        </div>
                    </div>
                )}

                <div className="space-y-1.5">
                    <Label className="text-sm text-[#37352F]/60">備考 / 詳細</Label>
                    <Input 
                        className="h-10 text-sm bg-white text-[#37352F] border-[rgba(55,53,47,0.16)]" 
                        value={formData.description || ""}
                        onChange={(e) => setFormData({...formData, description: e.target.value})}
                        placeholder="補足情報など" 
                    />
                </div>

                <div className="space-y-1.5">
                    <Label className="text-sm text-[#37352F]/60">ステータス</Label>
                    <div className="flex items-center gap-6 pt-2">
                            <label className="flex items-center gap-2 cursor-pointer text-sm text-[#37352F]">
                            <input 
                                type="radio" 
                                name="status" 
                                checked={formData.status !== "inactive"} 
                                onChange={() => setFormData({...formData, status: "active"})}
                                className="text-[#37352F] w-4 h-4" 
                            />
                            <span>有効</span>
                            </label>
                            <label className="flex items-center gap-2 cursor-pointer text-sm text-[#37352F]">
                            <input 
                                type="radio" 
                                name="status" 
                                checked={formData.status === "inactive"} 
                                onChange={() => setFormData({...formData, status: "inactive"})}
                                className="text-[#37352F] w-4 h-4" 
                            />
                            <span>無効</span>
                            </label>
                    </div>
                </div>

                <div className="flex justify-between pt-4 border-t border-[rgba(55,53,47,0.09)]">
                    {selectedItem ? (
                      <Button 
                        variant="ghost" 
                        size="sm" 
                        onClick={handleDelete}
                        className="h-10 text-sm text-[#E03E3E] hover:bg-[#E03E3E]/10 hover:text-[#E03E3E]"
                      >
                          <Trash2 className="mr-1.5 size-4" />
                          ���除
                      </Button>
                    ) : (
                      <div />
                    )}
                    <div className="flex gap-2">
                        <Button variant="outline" size="sm" onClick={handleCloseEdit} className="h-10 text-sm bg-white text-[#37352F] border-[rgba(55,53,47,0.16)]">キャンセル</Button>
                        <PrimaryButton onClick={handleSave} className="gap-2">
                            <Save className="size-4" />
                            保存
                        </PrimaryButton>
                    </div>
                </div>
            </div>
        </PageLayout>
      )
  }

  // List View
  return (
    <PageLayout
      title={getTabLabel(category)}
      icon={<SettingsIcon className="size-5 text-[#37352F]" />}
      headerAction={
        <PrimaryButton onClick={handleCreate}>
            <Plus className="mr-1.5 size-4" />
            新規登録
        </PrimaryButton>
      }
      maxWidth="max-w-full"
    >
      <div className="flex flex-col gap-4">
        <SearchFilterBar
            searchTerm={searchTerm}
            onSearchChange={setSearchTerm}
            placeholder="コード、名称、カテゴリで検索..."
            count={filteredItems.length}
        />

        <DataTable
            columns={columns}
            data={filteredItems}
            emptyMessage="データが見つかりません"
            renderRow={(item) => (
                <DataTableRow 
                    key={item.id} 
                    onClick={() => handleEdit(item)}
                >
                    <TableCell className="font-mono text-sm text-[#37352F]/80 py-2">{item.code}</TableCell>
                    <TableCell className="font-medium text-sm text-[#37352F] py-2">{item.name}</TableCell>
                    <TableCell className="text-sm text-[#37352F] py-2">{item.category}</TableCell>
                    <TableCell className="text-right font-mono text-sm text-[#37352F] py-2">
                        {shouldShowPrice(category) ? (item.price ? `¥${item.price.toLocaleString()}` : "-") : null}
                    </TableCell>
                    <TableCell className="text-sm text-[#37352F] py-2">
                        {shouldShowInventory(category) ? (
                            item.inventoryId ? (
                                <div className="flex items-center gap-1.5">
                                    <Package className="w-3.5 h-3.5 text-gray-400" />
                                    <span>{inventory.find(i => i.id === item.inventoryId)?.name || "不明な商品"}</span>
                                    <span className="text-xs text-gray-400">×{item.defaultQuantity || 1}</span>
                                </div>
                            ) : (
                                <span className="text-gray-300">-</span>
                            )
                        ) : null}
                    </TableCell>
                    <TableCell className="text-center py-2">
                        <StatusBadge colorClass={getMasterStatusColor(item.status)}>
                            {item.status === 'active' ? '有効' : '無効'}
                        </StatusBadge>
                    </TableCell>
                    <TableCell className="text-right py-2">
                        <RowActionButton onClick={() => handleEdit(item)} />
                    </TableCell>
                </DataTableRow>
            )}
        />
      </div>
    </PageLayout>
  );
}
