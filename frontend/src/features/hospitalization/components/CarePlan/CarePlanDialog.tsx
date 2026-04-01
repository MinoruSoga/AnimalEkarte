// React/Framework
import { memo, useEffect, useState } from "react";

// External
import { Search } from "lucide-react";

// Internal
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { Textarea } from "@/components/ui/textarea";
import { Badge } from "@/components/ui/badge";
import { FormDialog } from "@/components/shared/FormDialog/FormDialog";

// Shared
import { TreatmentSearchDialog } from "@/components/shared/TreatmentSearchDialog/TreatmentSearchDialog";
import type { TreatmentMasterItem } from "@/components/shared/TreatmentSearchDialog/TreatmentSearchDialog";

// Relative
import { H_STYLES } from "@/features/hospitalization/styles";

// Types
import type { CarePlanItem, CreateCarePlanDTO, UpdateCarePlanDTO } from "@/features/hospitalization/types";
import type { CarePlanTiming } from "@/types";
import { C } from "@/lib/design-tokens";

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

export const CarePlanDialog = memo(function CarePlanDialog({
    open,
    onOpenChange,
    editingPlan,
    onCreate,
    onUpdate
}: CarePlanDialogProps) {
    const [isSearchOpen, setIsSearchOpen] = useState(false);
    const [formData, setFormData] = useState<Partial<CarePlanItem>>(DEFAULT_FORM_STATE);

    // ダイアログ open 時にフォームを初期化
    useEffect(() => {
        if (open) {
            // eslint-disable-next-line react-hooks/set-state-in-effect -- ダイアログ open 時のフォームリセットパターン
            setFormData(editingPlan ?? DEFAULT_FORM_STATE);
        }
    }, [open, editingPlan]);

    const handleSave = () => {
        if (!formData.name || !formData.type) return;

        // type が "treatment" 以外の場合、マスタ由来フィールドを除外して送信
        const payload: Partial<CarePlanItem> = { ...formData };
        if (formData.type !== "treatment") {
            delete payload.unitPrice;
            delete payload.masterId;
            delete payload.category;
        }

        if (editingPlan) {
            onUpdate(editingPlan.id, payload);
        } else {
            onCreate(payload as CreateCarePlanDTO);
        }
        onOpenChange(false);
    };

    const handleSelectMaster = (item: TreatmentMasterItem) => {
        setFormData({
            ...formData,
            type: "treatment",
            name: item.name,
            masterId: item.id,
            unitPrice: item.unitPrice,
            category: item.category,
            description: item.category === "薬剤" ? "1錠" : "",
        });
    };

    const toggleTiming = (time: string) => {
        const current = formData.timing || [];
        const timingVal = time as CarePlanTiming;
        if (current.includes(timingVal)) {
            setFormData({ ...formData, timing: current.filter(t => t !== timingVal) });
        } else {
            setFormData({ ...formData, timing: [...current, timingVal] });
        }
    };

    return (
        <>
            <FormDialog
                open={open}
                onClose={() => onOpenChange(false)}
                title={editingPlan ? "プラン編集" : "新規プラン作成"}
                description={editingPlan ? "既存のケアプランを編集します。" : "新しいケアプランを作成します。"}
                onSave={handleSave}
            >
                <div className="space-y-4 py-4">
                    {/* Search Action */}
                    <div className="flex items-end gap-3 mb-2 p-3 bg-[#F7F6F3] rounded-md border border-[rgba(55,53,47,0.09)]">
                        <div className="flex-1">
                            <Label className={`${H_STYLES.text.sm} ${C.text60} mb-1 block`}>マスタから引用</Label>
                            <div className={`${H_STYLES.text.base} ${C.text}`}>
                                処置・検査・薬などをマスタから検索して入力できます
                            </div>
                        </div>
                        <Button
                            variant="outline"
                            className={`bg-white gap-2 ${C.text} ${H_STYLES.button.action}`}
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
                                onValueChange={(val: "food" | "medicine" | "treatment" | "instruction" | "item") =>
                                  setFormData((prev) => ({
                                    ...prev,
                                    type: val,
                                    unitPrice: undefined,
                                    masterId: undefined,
                                    category: undefined,
                                  }))
                                }
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
                                onChange={(e) => setFormData((prev) => ({...prev, name: e.target.value}))}
                            />
                        </div>
                    </div>

                    {formData.unitPrice !== undefined ? (
                        <div className={`flex items-center gap-2 ${H_STYLES.text.base} ${C.text60} px-1`}>
                            <Badge variant="outline" className="font-mono bg-purple-50 text-purple-700 border-purple-200">
                                マスタ連動中
                            </Badge>
                            <span>単価: ¥{formData.unitPrice.toLocaleString()} / カテゴリ: {formData.category}</span>
                        </div>
                    ) : null}

                    <div className="space-y-2">
                        <Label>詳細・指示量</Label>
                        <Input
                            placeholder="例: 30g / 1錠 / 左前腕"
                            value={formData.description || ""}
                            onChange={(e) => setFormData((prev) => ({...prev, description: e.target.value}))}
                        />
                    </div>

                    <div className="space-y-2">
                        <Label>タイミング</Label>
                        <div className="flex gap-2">
                            {(["morning", "noon", "night"] as CarePlanTiming[]).map((time) => (
                                <div
                                    key={time}
                                    onClick={() => toggleTiming(time)}
                                    className={`
                                        px-3 py-1.5 rounded-md ${H_STYLES.text.base} border cursor-pointer select-none transition-colors
                                        ${formData.timing?.includes(time)
                                            ? "bg-[#2EAADC] text-white border-[#2EAADC]"
                                            : "bg-white ${C.text} border-[rgba(55,53,47,0.16)] hover:bg-gray-50"}
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
                            onChange={(e) => setFormData((prev) => ({...prev, notes: e.target.value}))}
                        />
                    </div>

                    <div className="space-y-2">
                        <Label>ステータス</Label>
                        <Select
                            value={formData.status}
                            onValueChange={(val: "active" | "completed" | "discontinued") => setFormData((prev) => ({...prev, status: val}))}
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
            </FormDialog>

            <TreatmentSearchDialog
                open={isSearchOpen}
                onOpenChange={setIsSearchOpen}
                onSelect={handleSelectMaster}
            />
        </>
    );
});
