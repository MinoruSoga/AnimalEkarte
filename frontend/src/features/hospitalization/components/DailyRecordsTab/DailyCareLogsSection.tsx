// React/Framework
import { useState, useCallback } from "react";

// External
import {
    UtensilsCrossed,
    Droplets,
    Pill,
    Stethoscope,
    MoreHorizontal,
    Plus,
} from "lucide-react";

// Internal
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Textarea } from "@/components/ui/textarea";
import {
    Select,
    SelectContent,
    SelectItem,
    SelectTrigger,
    SelectValue,
} from "@/components/ui/select";
import { FormDialog } from "@/components/shared/FormDialog/FormDialog";

// Types
import type { ApiCareLogRecord, CreateCareLogRecordRequest } from "../../api/daily-records-types";

type CareLogType = "food" | "excretion" | "medicine" | "treatment" | "other";

interface DailyCareLogsSectionProps {
    careLogs: ApiCareLogRecord[];
    onAddCareLog: (payload: CreateCareLogRecordRequest) => void;
    isPending: boolean;
}

interface CareLogFormState {
    time: string;
    type: CareLogType;
    value: string;
    notes: string;
}

const INITIAL_FORM: CareLogFormState = {
    time: "",
    type: "other",
    value: "",
    notes: "",
};

const CARE_LOG_TYPE_LABELS: Record<CareLogType, string> = {
    food: "食事",
    excretion: "排泄",
    medicine: "投薬",
    treatment: "処置",
    other: "その他",
};

function getCareLogIcon(type: string) {
    switch (type) {
        case "food":
            return <UtensilsCrossed className="h-3.5 w-3.5" />;
        case "excretion":
            return <Droplets className="h-3.5 w-3.5" />;
        case "medicine":
            return <Pill className="h-3.5 w-3.5" />;
        case "treatment":
            return <Stethoscope className="h-3.5 w-3.5" />;
        default:
            return <MoreHorizontal className="h-3.5 w-3.5" />;
    }
}

function getCareLogColor(type: string): string {
    switch (type) {
        case "food":
            return "bg-orange-50 border-orange-100 text-orange-700";
        case "excretion":
            return "bg-teal-50 border-teal-100 text-teal-700";
        case "medicine":
            return "bg-purple-50 border-purple-100 text-purple-700";
        case "treatment":
            return "bg-red-50 border-red-100 text-red-700";
        default:
            return "bg-gray-50 border-gray-100 text-gray-700";
    }
}

function getCurrentTime(): string {
    const now = new Date();
    return `${String(now.getHours()).padStart(2, "0")}:${String(now.getMinutes()).padStart(2, "0")}`;
}

export function DailyCareLogsSection({
    careLogs,
    onAddCareLog,
    isPending,
}: DailyCareLogsSectionProps) {
    const [isOpen, setIsOpen] = useState(false);
    const [form, setForm] = useState<CareLogFormState>(INITIAL_FORM);

    const handleOpen = useCallback(() => {
        setForm({ ...INITIAL_FORM, time: getCurrentTime() });
        setIsOpen(true);
    }, []);

    const handleChange = useCallback((e: React.ChangeEvent<HTMLInputElement | HTMLTextAreaElement>) => {
        const { name, value } = e.target;
        setForm((prev) => ({ ...prev, [name]: value }));
    }, []);

    const handleTypeChange = useCallback((value: string) => {
        setForm((prev) => ({ ...prev, type: value as CareLogType }));
    }, []);

    const handleSubmit = useCallback(() => {
        if (!form.time) return;

        const payload: CreateCareLogRecordRequest = {
            time: form.time,
            type: form.type,
        };
        if (form.value) payload.value = form.value;
        if (form.notes) payload.notes = form.notes;

        onAddCareLog(payload);
        setIsOpen(false);
    }, [form, onAddCareLog]);

    const sorted = [...careLogs].sort((a, b) => a.time.localeCompare(b.time));

    return (
        <div>
            <div className="flex items-center justify-between mb-2">
                <h4 className="flex items-center gap-1.5 text-sm font-bold text-[#37352F]">
                    <UtensilsCrossed className="size-4 text-orange-500" />
                    ケアログ
                </h4>
                <Button
                    variant="outline"
                    size="sm"
                    onClick={handleOpen}
                    className="h-7 gap-1 text-xs"
                >
                    <Plus className="h-3.5 w-3.5" />
                    追加
                </Button>
            </div>

            {sorted.length === 0 ? (
                <p className="text-xs text-[#37352F]/40 py-3 text-center bg-gray-50 rounded border border-dashed border-[rgba(55,53,47,0.16)]">
                    記録なし
                </p>
            ) : (
                <div className="space-y-1.5">
                    {sorted.map((log) => {
                        const colorClass = getCareLogColor(log.type);
                        return (
                            <div
                                key={log.id}
                                className={`text-xs rounded px-2.5 py-2 border ${colorClass}`}
                            >
                                <div className="flex items-center gap-2 mb-0.5">
                                    <span className="font-semibold">{log.time}</span>
                                    <span className="flex items-center gap-1">
                                        {getCareLogIcon(log.type)}
                                        {CARE_LOG_TYPE_LABELS[log.type as CareLogType] ?? log.type}
                                    </span>
                                    {log.value ? (
                                        <span className="text-[#37352F]/70">{log.value}</span>
                                    ) : null}
                                </div>
                                {log.notes ? (
                                    <p className="text-[#37352F]/60 mt-0.5">{log.notes}</p>
                                ) : null}
                            </div>
                        );
                    })}
                </div>
            )}

            <FormDialog
                open={isOpen}
                onClose={() => setIsOpen(false)}
                title="ケアログ追加"
                onSave={handleSubmit}
                saveLabel="追加"
                isPending={isPending}
                isSaveDisabled={!form.time}
                className="max-w-sm"
            >
                <div className="space-y-3 py-2">
                    <div>
                        <Label htmlFor="carelog-time" className="text-xs">
                            時刻 <span className="text-red-500">*</span>
                        </Label>
                        <Input
                            id="carelog-time"
                            name="time"
                            type="time"
                            value={form.time}
                            onChange={handleChange}
                            className="mt-1"
                        />
                    </div>
                    <div>
                        <Label className="text-xs">種別</Label>
                        <Select value={form.type} onValueChange={handleTypeChange}>
                            <SelectTrigger className="mt-1">
                                <SelectValue />
                            </SelectTrigger>
                            <SelectContent>
                                {(Object.keys(CARE_LOG_TYPE_LABELS) as CareLogType[]).map(
                                    (t) => (
                                        <SelectItem key={t} value={t}>
                                            {CARE_LOG_TYPE_LABELS[t]}
                                        </SelectItem>
                                    )
                                )}
                            </SelectContent>
                        </Select>
                    </div>
                    <div>
                        <Label htmlFor="carelog-value" className="text-xs">
                            値 / 量
                        </Label>
                        <Input
                            id="carelog-value"
                            name="value"
                            value={form.value}
                            onChange={handleChange}
                            placeholder="例: 100%, 通常便"
                            className="mt-1"
                        />
                    </div>
                    <div>
                        <Label htmlFor="carelog-notes" className="text-xs">
                            メモ
                        </Label>
                        <Textarea
                            id="carelog-notes"
                            name="notes"
                            value={form.notes}
                            onChange={handleChange}
                            placeholder="特記事項"
                            rows={2}
                            className="mt-1 resize-none"
                        />
                    </div>
                </div>
            </FormDialog>
        </div>
    );
}
