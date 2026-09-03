// React/Framework
import { memo, useActionState, useState, useCallback } from "react";

// External
import { Activity, Plus, Thermometer, Heart, Wind, Weight } from "lucide-react";

// Internal
import { C, ICON } from "@/lib/design-tokens";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { NumberInput } from "@/components/shared/NumberInput/NumberInput";
import { Dialog, DialogContent, DialogDescription, DialogHeader, DialogTitle, DialogFooter } from "@/components/ui/dialog";
import { Button } from "@/components/ui/button";
import { SubmitButton } from "@/components/shared/Form/SubmitButton";

// Types
import type { ApiVitalRecord, CreateVitalRecordRequest } from "../../api/daily-records-types";

interface DailyVitalsSectionProps {
    vitals: ApiVitalRecord[];
    onAddVital: (payload: CreateVitalRecordRequest) => Promise<void>;
    canCreate?: boolean;
}

interface VitalFormState {
    time: string;
    temperature: string;
    heart_rate: string;
    respiration_rate: string;
    weight: string;
    notes: string;
}

const INITIAL_FORM: VitalFormState = {
    time: "",
    temperature: "",
    heart_rate: "",
    respiration_rate: "",
    weight: "",
    notes: "",
};

interface VitalActionState {
    error: string | null;
}

const INITIAL_ACTION_STATE: VitalActionState = { error: null };

function getCurrentTime(): string {
    const now = new Date();
    return `${String(now.getHours()).padStart(2, "0")}:${String(now.getMinutes()).padStart(2, "0")}`;
}

export const DailyVitalsSection = memo(function DailyVitalsSection({ vitals, onAddVital, canCreate = false }: DailyVitalsSectionProps) {
    const [isOpen, setIsOpen] = useState(false);
    const [form, setForm] = useState<VitalFormState>(INITIAL_FORM);

    const handleOpen = useCallback(() => {
        setForm({ ...INITIAL_FORM, time: getCurrentTime() });
        setIsOpen(true);
    }, []);

    const handleClose = useCallback(() => {
        setIsOpen(false);
    }, []);

    const handleChange = useCallback((e: React.ChangeEvent<HTMLInputElement>) => {
        const { name, value } = e.target;
        setForm((prev) => ({ ...prev, [name]: value }));
    }, []);

    const [state, formAction] = useActionState(
        async (_prevState: VitalActionState, _formData: FormData): Promise<VitalActionState> => {
            if (!form.time) return { error: "時刻は必須です" };

            const payload: CreateVitalRecordRequest = {
                time: form.time.length === 5 ? `${form.time}:00` : form.time,
            };
            if (form.temperature) payload.temperature = parseFloat(form.temperature);
            if (form.heart_rate) payload.heart_rate = parseInt(form.heart_rate, 10);
            if (form.respiration_rate) payload.respiration_rate = parseInt(form.respiration_rate, 10);
            if (form.weight) payload.weight = parseFloat(form.weight);
            if (form.notes) payload.notes = form.notes;

            await onAddVital(payload);
            setIsOpen(false);
            return { error: null };
        },
        INITIAL_ACTION_STATE,
    );

    const sorted = [...vitals].sort((a, b) => a.time.localeCompare(b.time));

    return (
        <div>
            <div className="flex items-center justify-between mb-2">
                <h4 className={`flex items-center gap-1.5 text-sm font-bold ${C.text}`}>
                    <Activity className={`${ICON.action} ${C.textMedicalBlue}`} />
                    バイタル
                </h4>
                {canCreate ? (
                    <Button
                        variant="outline"
                        size="sm"
                        onClick={handleOpen}
                        className="h-7 gap-1 text-xs"
                    >
                        <Plus className={ICON.action} />
                        追加
                    </Button>
                ) : null}
            </div>

            {sorted.length === 0 ? (
                <p className={`text-xs ${C.text40} py-3 text-center ${C.bgPage} rounded border border-dashed ${C.borderMedium}`}>
                    記録なし
                </p>
            ) : (
                <div className="space-y-1.5">
                    {sorted.map((v) => (
                        <div
                            key={v.id}
                            className={`text-xs ${C.bgBrandLight} rounded px-2.5 py-2 border ${C.borderBrandLight}`}
                        >
                            <div className="flex items-center gap-2 mb-1">
                                <span className={`font-semibold ${C.textBrandDark}`}>{v.time}</span>
                            </div>
                            <div className={`flex flex-wrap gap-x-3 gap-y-0.5 ${C.text70}`}>
                                {v.temperature !== undefined && v.temperature !== null ? (
                                    <span className="flex items-center gap-1">
                                        <Thermometer className={ICON.action} />
                                        {v.temperature}℃
                                    </span>
                                ) : null}
                                {v.heart_rate !== undefined && v.heart_rate !== null ? (
                                    <span className="flex items-center gap-1">
                                        <Heart className={ICON.action} />
                                        {v.heart_rate}bpm
                                    </span>
                                ) : null}
                                {v.respiration_rate !== undefined && v.respiration_rate !== null ? (
                                    <span className="flex items-center gap-1">
                                        <Wind className={ICON.action} />
                                        {v.respiration_rate}回/分
                                    </span>
                                ) : null}
                                {v.weight !== undefined && v.weight !== null ? (
                                    <span className="flex items-center gap-1">
                                        <Weight className={ICON.action} />
                                        {v.weight}kg
                                    </span>
                                ) : null}
                            </div>
                            {v.notes ? (
                                <p className={`mt-1 ${C.text60}`}>{v.notes}</p>
                            ) : null}
                        </div>
                    ))}
                </div>
            )}

            <Dialog open={isOpen} onOpenChange={setIsOpen}>
                <DialogContent className="max-w-sm">
                    <DialogHeader>
                        <DialogTitle>バイタル追加</DialogTitle>
                        <DialogDescription>
                            時刻、体温、心拍数、呼吸数、体重などのバイタルを記録します。
                        </DialogDescription>
                    </DialogHeader>
                    <form action={formAction} className="contents">
                        {state.error ? (
                            <p role="alert" className={`text-xs ${C.textNotionRed}`}>
                                {state.error}
                            </p>
                        ) : null}
                        <div className="space-y-3 py-2">
                            <div>
                                <Label htmlFor="vital-time" className="text-xs">
                                    時刻 <span className={C.textRequired}>*</span>
                                </Label>
                                <Input
                                    id="vital-time"
                                    name="time"
                                    type="time"
                                    value={form.time}
                                    onChange={handleChange}
                                    className="mt-1"
                                />
                            </div>
                            <div className="grid w-full grid-cols-1 gap-2 sm:grid-cols-2">
                                <div>
                                    <Label htmlFor="vital-temp" className="text-xs">
                                        体温 (℃)
                                    </Label>
                                    <NumberInput
                                        id="vital-temp"
                                        step={0.1}
                                        value={form.temperature}
                                        onChange={(v) => setForm((prev) => ({ ...prev, temperature: v }))}
                                        placeholder="38.5"
                                        suffix="℃"
                                        className="mt-1"
                                    />
                                </div>
                                <div>
                                    <Label htmlFor="vital-hr" className="text-xs">
                                        心拍数 (bpm)
                                    </Label>
                                    <NumberInput
                                        id="vital-hr"
                                        value={form.heart_rate}
                                        onChange={(v) => setForm((prev) => ({ ...prev, heart_rate: v }))}
                                        placeholder="80"
                                        suffix="/min"
                                        className="mt-1"
                                    />
                                </div>
                                <div>
                                    <Label htmlFor="vital-rr" className="text-xs">
                                        呼吸数 (回/分)
                                    </Label>
                                    <NumberInput
                                        id="vital-rr"
                                        value={form.respiration_rate}
                                        onChange={(v) => setForm((prev) => ({ ...prev, respiration_rate: v }))}
                                        placeholder="20"
                                        suffix="/min"
                                        className="mt-1"
                                    />
                                </div>
                                <div>
                                    <Label htmlFor="vital-weight" className="text-xs">
                                        体重 (kg)
                                    </Label>
                                    <NumberInput
                                        id="vital-weight"
                                        step={0.01}
                                        value={form.weight}
                                        onChange={(v) => setForm((prev) => ({ ...prev, weight: v }))}
                                        placeholder="5.2"
                                        suffix="kg"
                                        className="mt-1"
                                    />
                                </div>
                            </div>
                            <div>
                                <Label htmlFor="vital-notes" className="text-xs">
                                    メモ
                                </Label>
                                <Input
                                    id="vital-notes"
                                    name="notes"
                                    value={form.notes}
                                    onChange={handleChange}
                                    placeholder="特記事項"
                                    className="mt-1"
                                />
                            </div>
                        </div>
                        <DialogFooter>
                            <Button type="button" variant="outline" onClick={handleClose} size="sm">
                                キャンセル
                            </Button>
                            <SubmitButton size="sm" disabled={!form.time}>
                                保存
                            </SubmitButton>
                        </DialogFooter>
                    </form>
                </DialogContent>
            </Dialog>
        </div>
    );
});
