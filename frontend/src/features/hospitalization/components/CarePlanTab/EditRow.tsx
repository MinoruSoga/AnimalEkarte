// React/Framework
import { ICON, C } from "@/lib/design-tokens";
import { useState, useCallback } from "react";

// External
import { Loader2 } from "lucide-react";

// Internal
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Select, SelectContent, SelectTrigger, SelectValue } from "@/components/ui/select";

// Relative
import { TYPE_SELECT_ITEMS, TIMING_OPTIONS } from "./CarePlanBadges";
import { CarePlanRefSelect } from "./CarePlanRefSelect";

// Types
import type { CarePlanItem, CarePlanItemType, CarePlanTiming, UpdateCarePlanItemInput } from "../../api/care-plan-items";

/** type ごとに DDL(chk_care_plan_item_ref)が必須とするマスタ参照が要る種別 */
function requiresRef(type: CarePlanItemType): boolean {
    return type === "medicine" || type === "treatment" || type === "item";
}

/** refId を、現在の type に応じた正しい FK フィールドへ振り分ける */
function buildRefFields(type: CarePlanItemType, refId: string | null) {
    return {
        medicine_id: type === "medicine" ? refId : null,
        procedure_id: type === "treatment" ? refId : null,
        hospitalization_plan_id: type === "item" ? refId : null,
    };
}

/** item の既存 type に対応する FK 値を取り出す(type 未変更時の初期選択状態) */
function initialRefId(item: CarePlanItem): string | null {
    if (item.type === "medicine") return item.medicine_id ?? null;
    if (item.type === "treatment") return item.procedure_id ?? null;
    if (item.type === "item") return item.hospitalization_plan_id ?? null;
    return null;
}

interface EditRowProps {
    item: CarePlanItem;
    onSave: (input: UpdateCarePlanItemInput) => void;
    onCancel: () => void;
    isSaving: boolean;
}

export function EditRow({ item, onSave, onCancel, isSaving }: EditRowProps) {
    const [name, setName] = useState(item.name);
    const [type, setType] = useState<CarePlanItemType>(item.type);
    const [timing, setTiming] = useState<CarePlanTiming[]>(item.timing);
    const [refId, setRefId] = useState<string | null>(initialRefId(item));

    const needsRef = requiresRef(type);

    const handleTypeChange = useCallback((next: CarePlanItemType) => {
        setType(next);
        setRefId(null);
    }, []);

    const handleTimingToggle = useCallback((t: CarePlanTiming) => {
        setTiming((prev) =>
            prev.includes(t) ? prev.filter((x) => x !== t) : [...prev, t]
        );
    }, []);

    const handleSave = useCallback(() => {
        if (!name.trim()) return;
        if (needsRef && !refId) return;
        onSave({
            name: name.trim(),
            type,
            timing,
            ...buildRefFields(type, refId),
        });
    }, [name, type, timing, refId, needsRef, onSave]);

    const canSave = !isSaving && !!name.trim() && (!needsRef || !!refId);

    return (
        <div className={`flex flex-col gap-2 p-3 ${C.bgBrand5} rounded-lg border ${C.borderBrandLight}`}>
            <div className="flex gap-2 items-center">
                <Select value={type} onValueChange={(v) => handleTypeChange(v as CarePlanItemType)}>
                    <SelectTrigger className="w-28 h-8 text-xs">
                        <SelectValue />
                    </SelectTrigger>
                    <SelectContent>{TYPE_SELECT_ITEMS}</SelectContent>
                </Select>
                <Input
                    value={name}
                    onChange={(e) => setName(e.target.value)}
                    className="h-8 text-sm flex-1"
                    placeholder="名称"
                />
            </div>
            {needsRef ? (
                <CarePlanRefSelect type={type} value={refId} onChange={setRefId} />
            ) : null}
            <div className="flex items-center gap-3">
                <span className={`text-xs ${C.text50} shrink-0`}>タイミング:</span>
                <div className="flex gap-2">
                    {TIMING_OPTIONS.map((opt) => (
                        <label key={opt.value} className="flex items-center gap-1 cursor-pointer">
                            <input
                                type="checkbox"
                                checked={timing.includes(opt.value)}
                                onChange={() => handleTimingToggle(opt.value)}
                                className="rounded"
                            />
                            <span className="text-xs">{opt.label}</span>
                        </label>
                    ))}
                </div>
            </div>
            <div className="flex gap-2 justify-end">
                <Button
                    variant="ghost"
                    size="sm"
                    onClick={onCancel}
                    disabled={isSaving}
                    className="h-7 text-xs"
                >
                    キャンセル
                </Button>
                <Button
                    size="sm"
                    onClick={handleSave}
                    disabled={!canSave}
                    className="h-7 text-xs"
                >
                    {isSaving ? <Loader2 className={`${ICON.action} animate-spin mr-1`} /> : null}
                    保存
                </Button>
            </div>
        </div>
    );
}
