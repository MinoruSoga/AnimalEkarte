import { useState } from "react";
import { format } from "date-fns";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Textarea } from "@/components/ui/textarea";
import { FormDialog } from "@/components/shared/FormDialog/FormDialog";
import { CreateCareLogDTO, Task } from "@/features/hospitalization/types";
import { DailyRecord } from "@/types";
import { H_STYLES } from "@/features/hospitalization/styles";
import { C } from "@/lib/design-tokens";

type CareLogType = DailyRecord["careLogs"][0]["type"];
const VALID_LOG_TYPES: CareLogType[] = ["food", "medicine", "treatment", "other", "excretion"];

interface TaskCompleteDialogProps {
    open: boolean;
    onOpenChange: (open: boolean) => void;
    task: Task | null;
    onConfirm: (log: CreateCareLogDTO) => void;
}

export function TaskCompleteDialog({ open, onOpenChange, task, onConfirm }: TaskCompleteDialogProps) {
    const getCurrentTime = () => format(new Date(), "HH:mm");

    const [form, setForm] = useState(() => ({
        notes: "",
        time: getCurrentTime()
    }));
    const [prevOpen, setPrevOpen] = useState(false);

    if (open !== prevOpen) {
        setPrevOpen(open);
        if (open) {
            setForm({ notes: "", time: getCurrentTime() });
        }
    }

    const handleConfirm = () => {
        if (!task) return;

        const taskType = task.type as CareLogType;
        const logType: CareLogType = VALID_LOG_TYPES.includes(taskType) ? taskType : "other";

        onConfirm({
            time: form.time,
            type: logType,
            status: "completed",
            value: "実施",
            notes: `${task.name} (${task.description}) 実施 ${form.notes ? `\n${form.notes}` : ""}`,
            staff: "担当医"
        });

        onOpenChange(false);
    };

    return (
        <FormDialog
            open={open}
            onClose={() => onOpenChange(false)}
            title="処置の実施記録"
            description="以下の処置を実施として記録します。"
            onSave={handleConfirm}
            saveLabel="実施記録を保存"
        >
            <div className="space-y-4 py-4">
                <div className={`${C.bgPage} p-3 rounded-md border ${C.borderLight}`}>
                    <div className={`font-bold ${C.text} ${H_STYLES.text.base}`}>{task?.name}</div>
                    <div className={`${C.text60} ${H_STYLES.text.sm}`}>{task?.description}</div>
                </div>

                <div className="space-y-2">
                    <Label htmlFor="task-time">実施時刻</Label>
                    <Input id="task-time" type="time" value={form.time} onChange={e => setForm(prev => ({...prev, time: e.target.value}))} className={H_STYLES.text.base} />
                </div>

                <div className="space-y-2">
                    <Label htmlFor="task-memo">実施メモ (任意)</Label>
                    <Textarea
                        id="task-memo"
                        placeholder="特記事項があれば入力..."
                        value={form.notes}
                        onChange={e => setForm(prev => ({...prev, notes: e.target.value}))}
                        className={H_STYLES.text.base}
                    />
                </div>
            </div>
        </FormDialog>
    );
}
