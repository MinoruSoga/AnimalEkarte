import { useState, useEffect } from "react";
import { format } from "date-fns";
import { Button } from "../../../../components/ui/button";
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogDescription, DialogFooter } from "../../../../components/ui/dialog";
import { Input } from "../../../../components/ui/input";
import { Label } from "../../../../components/ui/label";
import { Textarea } from "../../../../components/ui/textarea";
import { CreateCareLogDTO, Task } from "../../types";
import { H_STYLES } from "../../styles";

interface TaskCompleteDialogProps {
    open: boolean;
    onOpenChange: (open: boolean) => void;
    task: Task | null;
    onConfirm: (log: CreateCareLogDTO) => void;
}

export const TaskCompleteDialog = ({ open, onOpenChange, task, onConfirm }: TaskCompleteDialogProps) => {
    const getCurrentTime = () => format(new Date(), "HH:mm");
    
    const [form, setForm] = useState({
        notes: "",
        time: getCurrentTime()
    });

    useEffect(() => {
        if (open) {
            setForm({ notes: "", time: getCurrentTime() });
        }
    }, [open]);

    const handleConfirm = () => {
        if (!task) return;
        
        const logType = (['food', 'medicine', 'treatment'].includes(task.type)) 
            ? task.type 
            : 'other';

        onConfirm({
            time: form.time,
            type: logType as any,
            status: "completed",
            value: "実施",
            notes: `${task.name} (${task.description}) 実施 ${form.notes ? `\n${form.notes}` : ""}`,
            staff: "担当医"
        });
        
        onOpenChange(false);
    };

    return (
        <Dialog open={open} onOpenChange={onOpenChange}>
            <DialogContent>
                <DialogHeader>
                    <DialogTitle>処置の実施記録</DialogTitle>
                    <DialogDescription>
                        以下の処置を実施として記録します。
                    </DialogDescription>
                </DialogHeader>
                <div className="space-y-4 py-4">
                    <div className="bg-[#F7F6F3] p-3 rounded-md border border-[rgba(55,53,47,0.09)]">
                        <div className={`font-bold text-[#37352F] ${H_STYLES.text.base}`}>{task?.name}</div>
                        <div className={`text-[#37352F]/60 ${H_STYLES.text.sm}`}>{task?.description}</div>
                    </div>

                    <div className="space-y-2">
                        <Label>実施時刻</Label>
                        <Input type="time" value={form.time} onChange={e => setForm({...form, time: e.target.value})} className={H_STYLES.text.base} />
                    </div>

                    <div className="space-y-2">
                        <Label>実施メモ (任意)</Label>
                        <Textarea 
                            placeholder="特記事項があれば入力..."
                            value={form.notes} 
                            onChange={e => setForm({...form, notes: e.target.value})} 
                            className={H_STYLES.text.base}
                        />
                    </div>
                </div>
                <DialogFooter>
                    <Button variant="outline" onClick={() => onOpenChange(false)} className={H_STYLES.button.action}>キャンセル</Button>
                    <Button onClick={handleConfirm} className={`bg-[#2EAADC] text-white ${H_STYLES.button.action}`}>実施記録を保存</Button>
                </DialogFooter>
            </DialogContent>
        </Dialog>
    );
};
