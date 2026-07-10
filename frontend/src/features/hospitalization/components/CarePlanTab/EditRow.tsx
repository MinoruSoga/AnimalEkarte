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
import { TYPE_SELECT_ITEMS, TIMING_OPTIONS } from "./badges";

// Types
import type { CarePlanItem, CarePlanItemType, CarePlanTiming, UpdateCarePlanItemInput } from "../../api/care-plan-items";

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

    const handleTimingToggle = useCallback((t: CarePlanTiming) => {
        setTiming((prev) =>
            prev.includes(t) ? prev.filter((x) => x !== t) : [...prev, t]
        );
    }, []);

    const handleSave = useCallback(() => {
        if (!name.trim()) return;
        onSave({ name: name.trim(), type, timing });
    }, [name, type, timing, onSave]);

    return (
        <div className={`flex flex-col gap-2 p-3 ${C.bgBrand5} rounded-lg border ${C.borderBrandLight}`}>
            <div className="flex gap-2 items-center">
                <Select value={type} onValueChange={(v) => setType(v as CarePlanItemType)}>
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
                    disabled={isSaving || !name.trim()}
                    className="h-7 text-xs"
                >
                    {isSaving ? <Loader2 className={`${ICON.action} animate-spin mr-1`} /> : null}
                    保存
                </Button>
            </div>
        </div>
    );
}
