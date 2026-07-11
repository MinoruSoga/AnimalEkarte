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

// Types
import type { CarePlanItemType, CarePlanTiming } from "../../api/care-plan-items";

const INITIAL_TIMING: CarePlanTiming[] = ["morning"];

interface AddFormProps {
    onSubmit: (type: CarePlanItemType, name: string, timing: CarePlanTiming[]) => void;
    isSubmitting: boolean;
}

export function AddForm({ onSubmit, isSubmitting }: AddFormProps) {
    const [type, setType] = useState<CarePlanItemType>("instruction");
    const [name, setName] = useState("");
    const [timing, setTiming] = useState<CarePlanTiming[]>(INITIAL_TIMING);

    const handleTimingToggle = useCallback((t: CarePlanTiming) => {
        setTiming((prev) =>
            prev.includes(t) ? prev.filter((x) => x !== t) : [...prev, t]
        );
    }, []);

    const handleSubmit = useCallback(() => {
        if (!name.trim()) return;
        onSubmit(type, name.trim(), timing);
        setName("");
        setType("instruction");
        setTiming(INITIAL_TIMING);
    }, [name, type, timing, onSubmit]);

    return (
        <div className={`border-t ${C.borderLight} pt-3 mt-2`}>
            <p className={`text-xs font-medium ${C.text60} mb-2`}>新しいケアプラン項目を追加</p>
            <div className="flex flex-col gap-2">
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
                        placeholder="名称を入力"
                        onKeyDown={(e) => {
                            if (e.key === "Enter") handleSubmit();
                        }}
                    />
                </div>
                <div className="flex items-center justify-between">
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
                    <Button
                        size="sm"
                        onClick={handleSubmit}
                        disabled={isSubmitting || !name.trim()}
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
