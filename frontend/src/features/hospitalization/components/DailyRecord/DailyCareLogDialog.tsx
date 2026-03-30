// React/Framework
import { useState } from "react";

// External
import { format } from "date-fns";

// Internal
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Textarea } from "@/components/ui/textarea";
import { FormDialog } from "@/components/shared/FormDialog/FormDialog";

// Relative
import { H_STYLES } from "@/features/hospitalization/styles";

// Types
import type { CreateCareLogDTO } from "@/features/hospitalization/types";

type LogType = "food" | "excretion" | "medicine" | "other";

interface DailyCareLogDialogProps {
    open: boolean;
    onOpenChange: (open: boolean) => void;
    type: LogType;
    onSave: (data: CreateCareLogDTO) => void;
}

export function DailyCareLogDialog({ open, onOpenChange, type, onSave }: DailyCareLogDialogProps) {
    const getCurrentTime = () => format(new Date(), "HH:mm");

    const [form, setForm] = useState({
        value: "",
        notes: "",
        time: getCurrentTime()
    });
    const [prevOpen, setPrevOpen] = useState(false);

    if (open !== prevOpen) {
        setPrevOpen(open);
        if (open) {
            setForm({ value: "", notes: "", time: getCurrentTime() });
        }
    }

    const handleSave = () => {
        onSave({
            time: form.time,
            type: type,
            status: "completed",
            value: form.value,
            notes: form.notes,
            staff: "スタッフ"
        });
        onOpenChange(false);
    };

    const getTitle = () => {
        switch(type) {
            case 'food': return '食事記録';
            case 'excretion': return '排泄記録';
            default: return '活動・メモ';
        }
    };

    const getDescription = () => {
        switch(type) {
            case 'food': return '食事の内容や摂取量を記録してください。';
            case 'excretion': return '排泄の状態や詳細を記録してください。';
            default: return '活動内容や特記事項を記録してください。';
        }
    };

    const getPlaceholder = () => {
        switch(type) {
            case 'food': return '完食、1/2など';
            case 'excretion': return '良便、軟便など';
            default: return '内容';
        }
    };

    return (
        <FormDialog
            open={open}
            onClose={() => onOpenChange(false)}
            title={getTitle()}
            description={getDescription()}
            onSave={handleSave}
            saveLabel="記録"
        >
            <div className="space-y-4 py-4">
                <div className="space-y-2">
                    <Label>記録時刻</Label>
                    <Input type="time" value={form.time} onChange={e => setForm({...form, time: e.target.value})} className={H_STYLES.text.base} />
                </div>
                <div className="space-y-2">
                    <Label>内容・量</Label>
                    <Input
                        placeholder={getPlaceholder()}
                        value={form.value}
                        onChange={e => setForm({...form, value: e.target.value})}
                        className={H_STYLES.text.base}
                    />
                </div>
                <div className="space-y-2">
                    <Label>詳細メモ</Label>
                    <Textarea value={form.notes} onChange={e => setForm({...form, notes: e.target.value})} className={H_STYLES.text.base} />
                </div>
            </div>
        </FormDialog>
    );
}
