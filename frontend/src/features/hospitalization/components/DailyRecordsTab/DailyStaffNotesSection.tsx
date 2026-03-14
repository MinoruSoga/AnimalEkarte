// React/Framework
import { useState, useCallback } from "react";

// External
import { MessageSquare, Plus } from "lucide-react";

// Internal
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Textarea } from "@/components/ui/textarea";
import {
    Dialog,
    DialogContent,
    DialogHeader,
    DialogTitle,
    DialogFooter,
} from "@/components/ui/dialog";

// Types
import type { ApiStaffNoteRecord, CreateStaffNoteRecordRequest } from "../../api/daily-records-types";

interface DailyStaffNotesSectionProps {
    staffNotes: ApiStaffNoteRecord[];
    onAddStaffNote: (payload: CreateStaffNoteRecordRequest) => void;
    isPending: boolean;
}

interface StaffNoteFormState {
    time: string;
    content: string;
}

const INITIAL_FORM: StaffNoteFormState = {
    time: "",
    content: "",
};

function getCurrentTime(): string {
    const now = new Date();
    return `${String(now.getHours()).padStart(2, "0")}:${String(now.getMinutes()).padStart(2, "0")}`;
}

export function DailyStaffNotesSection({
    staffNotes,
    onAddStaffNote,
    isPending,
}: DailyStaffNotesSectionProps) {
    const [isOpen, setIsOpen] = useState(false);
    const [form, setForm] = useState<StaffNoteFormState>(INITIAL_FORM);

    const handleOpen = useCallback(() => {
        setForm({ ...INITIAL_FORM, time: getCurrentTime() });
        setIsOpen(true);
    }, []);

    const handleClose = useCallback(() => {
        setIsOpen(false);
    }, []);

    const handleChange = useCallback(
        (e: React.ChangeEvent<HTMLInputElement | HTMLTextAreaElement>) => {
            const { name, value } = e.target;
            setForm((prev) => ({ ...prev, [name]: value }));
        },
        []
    );

    const handleSubmit = useCallback(() => {
        if (!form.time || !form.content.trim()) return;

        onAddStaffNote({
            time: form.time,
            content: form.content.trim(),
        });
        setIsOpen(false);
    }, [form, onAddStaffNote]);

    const sorted = [...staffNotes].sort((a, b) => a.time.localeCompare(b.time));

    return (
        <div>
            <div className="flex items-center justify-between mb-2">
                <h4 className="flex items-center gap-1.5 text-sm font-bold text-[#37352F]">
                    <MessageSquare className="h-4 w-4 text-green-500" />
                    スタッフメモ
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
                    {sorted.map((note) => (
                        <div
                            key={note.id}
                            className="text-xs bg-green-50 rounded px-2.5 py-2 border border-green-100"
                        >
                            <span className="font-semibold text-green-700 mr-2">{note.time}</span>
                            <span className="text-[#37352F]/80">{note.content}</span>
                        </div>
                    ))}
                </div>
            )}

            <Dialog open={isOpen} onOpenChange={setIsOpen}>
                <DialogContent className="max-w-sm">
                    <DialogHeader>
                        <DialogTitle>スタッフメモ追加</DialogTitle>
                    </DialogHeader>
                    <div className="space-y-3 py-2">
                        <div>
                            <Label htmlFor="note-time" className="text-xs">
                                時刻 <span className="text-red-500">*</span>
                            </Label>
                            <Input
                                id="note-time"
                                name="time"
                                type="time"
                                value={form.time}
                                onChange={handleChange}
                                className="mt-1"
                            />
                        </div>
                        <div>
                            <Label htmlFor="note-content" className="text-xs">
                                内容 <span className="text-red-500">*</span>
                            </Label>
                            <Textarea
                                id="note-content"
                                name="content"
                                value={form.content}
                                onChange={handleChange}
                                placeholder="メモ内容を入力"
                                rows={3}
                                className="mt-1 resize-none"
                            />
                        </div>
                    </div>
                    <DialogFooter>
                        <Button variant="outline" onClick={handleClose} size="sm">
                            キャンセル
                        </Button>
                        <Button
                            onClick={handleSubmit}
                            disabled={!form.time || !form.content.trim() || isPending}
                            size="sm"
                        >
                            保存
                        </Button>
                    </DialogFooter>
                </DialogContent>
            </Dialog>
        </div>
    );
}
