// React/Framework
import { ICON, C } from "@/lib/design-tokens";
import { useState, useCallback } from "react";

// External
import { Plus, Loader2 } from "lucide-react";

// Internal
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Select, SelectContent, SelectTrigger, SelectValue } from "@/components/ui/select";

// Relative
import { TYPE_SELECT_ITEMS, TIMING_OPTIONS } from "./CarePlanBadges";
import { CarePlanRefSelect } from "./CarePlanRefSelect";

// Types
import type { CarePlanItemType, CarePlanTiming, CreateCarePlanItemInput } from "../../api/care-plan-items";

const INITIAL_TIMING: CarePlanTiming[] = ["morning"];

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

interface AddFormProps {
    onSubmit: (input: CreateCarePlanItemInput) => void;
    isSubmitting: boolean;
}

export function AddForm({ onSubmit, isSubmitting }: AddFormProps) {
    const [type, setType] = useState<CarePlanItemType>("instruction");
    const [name, setName] = useState("");
    const [timing, setTiming] = useState<CarePlanTiming[]>(INITIAL_TIMING);
    const [refId, setRefId] = useState<string | null>(null);

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

    const handleSubmit = useCallback(() => {
        if (!name.trim()) return;
        if (needsRef && !refId) return;
        onSubmit({
            type,
            name: name.trim(),
            timing,
            ...buildRefFields(type, refId),
        });
        setName("");
        setType("instruction");
        setTiming(INITIAL_TIMING);
        setRefId(null);
    }, [name, type, timing, refId, needsRef, onSubmit]);

    const canSubmit = !isSubmitting && !!name.trim() && (!needsRef || !!refId);

    return (
        <div className={`border-t ${C.borderLight} pt-3 mt-2`}>
            <p className={`text-xs font-medium ${C.text60} mb-2`}>新しいケアプラン項目を追加</p>
            <div className="flex flex-col gap-2">
                <div className="flex gap-2 items-center">
                    <Select value={type} onValueChange={(v) => handleTypeChange(v as CarePlanItemType)}>
                        <SelectTrigger className="w-28 h-11 text-xs" aria-label="ケアプラン項目種別">
                            <SelectValue />
                        </SelectTrigger>
                        <SelectContent>{TYPE_SELECT_ITEMS}</SelectContent>
                    </Select>
                    <Input
                        value={name}
                        onChange={(e) => setName(e.target.value)}
                        aria-label="ケアプラン項目名"
                        className="h-11 text-sm flex-1"
                        placeholder="名称を入力"
                        onKeyDown={(e) => {
                            if (e.key === "Enter") handleSubmit();
                        }}
                    />
                </div>
                {needsRef ? (
                    <CarePlanRefSelect type={type} value={refId} onChange={setRefId} />
                ) : null}
                <div className="flex items-center justify-between">
                    <div className="flex items-center gap-3">
                        <span className={`text-xs ${C.text50} shrink-0`}>タイミング:</span>
                        <div className="flex gap-2">
                            {TIMING_OPTIONS.map((opt) => (
                                <label key={opt.value} className="flex min-h-11 min-w-11 items-center gap-1 cursor-pointer">
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
                    <Button
                        size="sm"
                        onClick={handleSubmit}
                        disabled={!canSubmit}
                        className="h-8 text-xs gap-1"
                    >
                        {isSubmitting ? (
                            <Loader2 className={`${ICON.action} animate-spin`} />
                        ) : (
                            <Plus className={ICON.action} />
                        )}
                        追加
                    </Button>
                </div>
            </div>
        </div>
    );
}
