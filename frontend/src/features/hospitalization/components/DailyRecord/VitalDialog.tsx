// React/Framework
import { useState } from "react";

// External
import { format } from "date-fns";

// Internal
import { Button } from "@/components/ui/button";
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogDescription, DialogFooter } from "@/components/ui/dialog";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Textarea } from "@/components/ui/textarea";

// Relative
import { H_STYLES } from "../../styles";

// Types
import type { CreateVitalDTO } from "../../types";

interface VitalDialogProps {
    open: boolean;
    onOpenChange: (open: boolean) => void;
    onSave: (data: CreateVitalDTO) => void;
}

export const VitalDialog = ({ open, onOpenChange, onSave }: VitalDialogProps) => {
    const getCurrentTime = () => format(new Date(), "HH:mm");

    const [form, setForm] = useState({
        temperature: "",
        heartRate: "",
        respirationRate: "",
        weight: "",
        notes: "",
        time: getCurrentTime()
    });
    const [prevOpen, setPrevOpen] = useState(false);

    if (open !== prevOpen) {
        setPrevOpen(open);
        if (open) {
            setForm({
                temperature: "",
                heartRate: "",
                respirationRate: "",
                weight: "",
                notes: "",
                time: getCurrentTime()
            });
        }
    }

    const handleSave = () => {
        onSave({
            time: form.time,
            temperature: form.temperature ? Number(form.temperature) : undefined,
            heartRate: form.heartRate ? Number(form.heartRate) : undefined,
            respirationRate: form.respirationRate ? Number(form.respirationRate) : undefined,
            weight: form.weight ? Number(form.weight) : undefined,
            notes: form.notes,
            staff: "担当医" // In a real app, get from auth context
        });
        onOpenChange(false);
    };

    return (
        <Dialog open={open} onOpenChange={onOpenChange}>
            <DialogContent>
                <DialogHeader>
                    <DialogTitle>バイタル記録</DialogTitle>
                    <DialogDescription>
                        患者のバイタルサインを記録してください。
                    </DialogDescription>
                </DialogHeader>
                <div className="grid grid-cols-2 gap-4 py-4">
                    <div className="col-span-2 space-y-2">
                        <Label>記録時刻</Label>
                        <Input type="time" value={form.time} onChange={e => setForm({...form, time: e.target.value})} className={H_STYLES.text.base} />
                    </div>
                    <div className="space-y-2">
                        <Label>体温 (℃)</Label>
                        <Input type="number" step="0.1" value={form.temperature} onChange={e => setForm({...form, temperature: e.target.value})} className={H_STYLES.text.base} />
                    </div>
                    <div className="space-y-2">
                        <Label>体重 (kg)</Label>
                        <Input type="number" step="0.01" value={form.weight} onChange={e => setForm({...form, weight: e.target.value})} className={H_STYLES.text.base} />
                    </div>
                    <div className="space-y-2">
                        <Label>心拍数 (/min)</Label>
                        <Input type="number" value={form.heartRate} onChange={e => setForm({...form, heartRate: e.target.value})} className={H_STYLES.text.base} />
                    </div>
                    <div className="space-y-2">
                        <Label>呼吸数 (/min)</Label>
                        <Input type="number" value={form.respirationRate} onChange={e => setForm({...form, respirationRate: e.target.value})} className={H_STYLES.text.base} />
                    </div>
                    <div className="col-span-2 space-y-2">
                        <Label>メモ</Label>
                        <Textarea value={form.notes} onChange={e => setForm({...form, notes: e.target.value})} className={H_STYLES.text.base} />
                    </div>
                </div>
                <DialogFooter>
                    <Button variant="outline" onClick={() => onOpenChange(false)} className={H_STYLES.button.action}>キャンセル</Button>
                    <Button onClick={handleSave} className={`bg-[#2EAADC] text-white ${H_STYLES.button.action}`}>記録</Button>
                </DialogFooter>
            </DialogContent>
        </Dialog>
    );
};
