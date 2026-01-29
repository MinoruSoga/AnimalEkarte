// React/Framework
import { useState } from "react";

// External
import { Search } from "lucide-react";

// Internal
import { Button } from "@/components/ui/button";
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogDescription, DialogFooter } from "@/components/ui/dialog";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { Textarea } from "@/components/ui/textarea";
import { Badge } from "@/components/ui/badge";

// Shared
import { TreatmentSearchDialog } from "@/components/shared/TreatmentSearchDialog";
import type { TreatmentMasterItem } from "@/components/shared/TreatmentSearchDialog";

// Relative
import { H_STYLES } from "../../styles";

// Types
import type { CarePlanItem, CreateCarePlanDTO, UpdateCarePlanDTO } from "../../types";

interface CarePlanDialogProps {
    open: boolean;
    onOpenChange: (open: boolean) => void;
    editingPlan?: CarePlanItem;
    onCreate: (data: CreateCarePlanDTO) => void;
    onUpdate: (id: string, data: UpdateCarePlanDTO) => void;
}

const DEFAULT_FORM_STATE: Partial<CarePlanItem> = {
    type: "food",
    timing: ["morning", "night"],
    status: "active"
};

export const CarePlanDialog = ({
    open,
    onOpenChange,
    editingPlan,
    onCreate,
    onUpdate
}: CarePlanDialogProps) => {
    const [isSearchOpen, setIsSearchOpen] = useState(false);
    const [formData, setFormData] = useState<Partial<CarePlanItem>>(DEFAULT_FORM_STATE);
    const [prevOpen, setPrevOpen] = useState(false);

    // Reset or Populate form when dialog opens or editingPlan changes
    if (open !== prevOpen) {
        setPrevOpen(open);
        if (open) {
            if (editingPlan) {
                setFormData(editingPlan);
            } else {
                setFormData(DEFAULT_FORM_STATE);
            }
        }
    }

    const handleSave = () => {
        if (!formData.name || !formData.type) return;

        if (editingPlan) {
            onUpdate(editingPlan.id, formData);
        } else {
            onCreate(formData as CreateCarePlanDTO);
        }
        onOpenChange(false);
    };

    const handleSelectMaster = (item: TreatmentMasterItem) => {
        setFormData({
            ...formData,
            type: "treatment",
            name: item.name,
            masterId: item.code,
            unitPrice: item.unitPrice,
            category: item.category,
            description: item.category === "薬剤" ? "1錠" : "", 
        });
    };

    const toggleTiming = (time: string) => {
        const current = formData.timing || [];
        if (current.includes(time)) {
            setFormData({ ...formData, timing: current.filter(t => t !== time) });
        } else {
            setFormData({ ...formData, timing: [...current, time] });
        }
    };

    return (
        <Dialog open={open} onOpenChange={onOpenChange}>
            <DialogContent>
                <DialogHeader>
                    <DialogTitle>{editingPlan ? "プラン編集" : "新規プラン作成"}</DialogTitle>
                    <DialogDescription>
                        {editingPlan ? "既存のケアプランを編集します。" : "新しいケアプランを作成します。"}
                    </DialogDescription>
                </DialogHeader>
                <div className="space-y-4 py-4">
                    {/* Search Action */}
                    <div className="flex items-end gap-3 mb-2 p-3 bg-[#F7F6F3] rounded-md border border-[rgba(55,53,47,0.09)]">
                        <div className="flex-1">
                            <Label className={`${H_STYLES.text.sm} text-[#37352F]/60 mb-1 block`}>マスタから引用</Label>
                            <div className={`${H_STYLES.text.base} text-[#37352F]`}>
                                処置・検査・薬などをマスタから検索して入力できます
                            </div>
                        </div>
                        <Button 
                            variant="outline" 
                            className={`bg-white gap-2 text-[#37352F] ${H_STYLES.button.action}`}
                            onClick={() => setIsSearchOpen(true)}
                        >
                            <Search className={H_STYLES.button.icon} />
                            マスタ検索
                        </Button>
                    </div>

                    <div className="grid grid-cols-2 gap-4">
                        <div className="space-y-2">
                            <Label>種類</Label>
                            <Select 
                                value={formData.type} 
                                onValueChange={(val: "food" | "medicine" | "treatment" | "instruction" | "item") => setFormData({...formData, type: val})}
                            >
                                <SelectTrigger>
                                    <SelectValue />
                                </SelectTrigger>
                                <SelectContent>
                                    <SelectItem value="food">食事</SelectItem>
                                    <SelectItem value="medicine">投薬</SelectItem>
                                    <SelectItem value="treatment">処置・検査</SelectItem>
                                    <SelectItem value="instruction">指示・その他</SelectItem>
                                    <SelectItem value="item">持ち物</SelectItem>
                                </SelectContent>
                            </Select>
                        </div>
                        <div className="space-y-2">
                            <Label>名称</Label>
                            <Input 
                                placeholder="例: ロイヤルカナン消化器サポート" 
                                value={formData.name || ""}
                                onChange={(e) => setFormData({...formData, name: e.target.value})}
                            />
                        </div>
                    </div>

                    {formData.unitPrice !== undefined && (
                            <div className={`flex items-center gap-2 ${H_STYLES.text.base} text-[#37352F]/60 px-1`}>
                            <Badge variant="outline" className="font-mono bg-purple-50 text-purple-700 border-purple-200">
                                マスタ連動中
                            </Badge>
                            <span>単価: ¥{formData.unitPrice.toLocaleString()} / カテゴリ: {formData.category}</span>
                            </div>
                    )}

                    <div className="space-y-2">
                        <Label>詳細・指示量</Label>
                        <Input 
                            placeholder="例: 30g / 1錠 / 左前��" 
                            value={formData.description || ""}
                            onChange={(e) => setFormData({...formData, description: e.target.value})}
                        />
                    </div>

                    <div className="space-y-2">
                        <Label>タイミング</Label>
                        <div className="flex gap-2">
                            {["morning", "noon", "night"].map((time) => (
                                <div 
                                    key={time}
                                    onClick={() => toggleTiming(time)}
                                    className={`
                                        px-3 py-1.5 rounded-md ${H_STYLES.text.base} border cursor-pointer select-none transition-colors
                                        ${formData.timing?.includes(time) 
                                            ? "bg-[#2EAADC] text-white border-[#2EAADC]" 
                                            : "bg-white text-[#37352F] border-[rgba(55,53,47,0.16)] hover:bg-gray-50"}
                                    `}
                                >
                                    {time === 'morning' ? '朝' : time === 'noon' ? '昼' : '夜'}
                                </div>
                            ))}
                        </div>
                    </div>

                    <div className="space-y-2">
                        <Label>メモ・特記事項</Label>
                        <Textarea 
                            placeholder="例: ふやかして与える" 
                            value={formData.notes || ""}
                            onChange={(e) => setFormData({...formData, notes: e.target.value})}
                        />
                    </div>

                    <div className="space-y-2">
                        <Label>ステータス</Label>
                        <Select 
                            value={formData.status} 
                            onValueChange={(val: "active" | "completed" | "discontinued") => setFormData({...formData, status: val})}
                        >
                            <SelectTrigger>
                                <SelectValue />
                            </SelectTrigger>
                            <SelectContent>
                                <SelectItem value="active">実施中</SelectItem>
                                <SelectItem value="completed">完了</SelectItem>
                                <SelectItem value="discontinued">中止</SelectItem>
                            </SelectContent>
                        </Select>
                    </div>
                </div>
                <DialogFooter>
                    <Button variant="outline" onClick={() => onOpenChange(false)} className={H_STYLES.button.action}>キャンセル</Button>
                    <Button onClick={handleSave} className={`bg-[#2EAADC] text-white ${H_STYLES.button.action}`}>保存</Button>
                </DialogFooter>
            </DialogContent>

            <TreatmentSearchDialog 
                open={isSearchOpen} 
                onOpenChange={setIsSearchOpen} 
                onSelect={handleSelectMaster} 
            />
        </Dialog>
    );
};
