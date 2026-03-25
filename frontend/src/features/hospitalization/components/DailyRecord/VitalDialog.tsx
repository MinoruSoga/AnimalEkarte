// React/Framework
import { useState, useEffect } from "react";

// External
import { format } from "date-fns";

// Internal
import { Label } from "@/components/ui/label";
import { Textarea } from "@/components/ui/textarea";
import { FormDialog } from "@/components/shared/FormDialog/FormDialog";
import { NumberInput } from "@/components/shared/NumberInput/NumberInput";

// Relative
import { H_STYLES } from "../../styles";

// Types
import type { CreateVitalDTO } from "../../types";

interface VitalDialogProps {
    open: boolean;
    onOpenChange: (open: boolean) => void;
    onSave: (data: CreateVitalDTO) => void;
}

export function VitalDialog({ open, onOpenChange, onSave }: VitalDialogProps) {
    const getCurrentTime = () => format(new Date(), "HH:mm");

    const [form, setForm] = useState(() => ({
        temperature: "",
        heartRate: "",
        respirationRate: "",
        weight: "",
        notes: "",
        time: getCurrentTime()
    }));
    useEffect(() => {
        if (!open) return;
        // eslint-disable-next-line react-hooks/set-state-in-effect -- ダイアログ open 時にフォームをリセット
        setForm({
            temperature: "",
            heartRate: "",
            respirationRate: "",
            weight: "",
            notes: "",
            time: getCurrentTime()
        });
    }, [open]);

    const handleSave = () => {
        onSave({
            time: form.time,
            temperature: form.temperature ? Number(form.temperature) : undefined,
            heartRate: form.heartRate ? Number(form.heartRate) : undefined,
            respirationRate: form.respirationRate ? Number(form.respirationRate) : undefined,
            weight: form.weight ? Number(form.weight) : undefined,
            notes: form.notes,
            staff: "担当医"
        });
        onOpenChange(false);
    };

    return (
        <FormDialog
            open={open}
            onClose={() => onOpenChange(false)}
            title="バイタル記録"
            description="患者のバイタルサインを記録してください。"
            onSave={handleSave}
            saveLabel="記録"
        >
            <div className="grid grid-cols-2 gap-4 py-4">
                <div className="col-span-2 space-y-2">
                    <Label>記録時刻</Label>
                    <Input type="time" value={form.time} onChange={e => setForm(prev => ({...prev, time: e.target.value}))} className={H_STYLES.text.base} />
                </div>
                <div className="space-y-2">
                    <Label>体温 (℃)</Label>
                    <NumberInput step={0.1} value={form.temperature} onChange={v => setForm(prev => ({...prev, temperature: v}))} className={H_STYLES.text.base} suffix="℃" />
                </div>
                <div className="space-y-2">
                    <Label>体重 (kg)</Label>
                    <NumberInput step={0.01} value={form.weight} onChange={v => setForm(prev => ({...prev, weight: v}))} className={H_STYLES.text.base} suffix="kg" />
                </div>
                <div className="space-y-2">
                    <Label>心拍数 (/min)</Label>
                    <NumberInput value={form.heartRate} onChange={v => setForm(prev => ({...prev, heartRate: v}))} className={H_STYLES.text.base} suffix="/min" />
                </div>
                <div className="space-y-2">
                    <Label>呼吸数 (/min)</Label>
                    <NumberInput value={form.respirationRate} onChange={v => setForm(prev => ({...prev, respirationRate: v}))} className={H_STYLES.text.base} suffix="/min" />
                </div>
                <div className="col-span-2 space-y-2">
                    <Label>メモ</Label>
                    <Textarea value={form.notes} onChange={e => setForm(prev => ({...prev, notes: e.target.value}))} className={H_STYLES.text.base} />
                </div>
            </div>
        </FormDialog>
    );
}
