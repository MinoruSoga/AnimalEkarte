// React/Framework
import { useState } from "react";

// External
import { Send } from "lucide-react";
import { format } from "date-fns";

// Internal
import { Button } from "@/components/ui/button";
import { Textarea } from "@/components/ui/textarea";

// Internal
import { C } from "@/lib/design-tokens";

// Relative
import { H_STYLES } from "@/features/hospitalization/styles";

// Types
import type { CreateCareLogDTO } from "@/features/hospitalization/types";

interface DailyCareNoteFormProps {
    onSave: (data: CreateCareLogDTO) => void;
}

export function DailyCareNoteForm({ onSave }: DailyCareNoteFormProps) {
    const [note, setNote] = useState("");

    const handleSubmit = () => {
        if (!note.trim()) return;

        onSave({
            time: format(new Date(), "HH:mm"),
            type: "other",
            status: "completed",
            value: "経過記録",
            notes: note,
            staff: "スタッフ"
        });

        setNote("");
    };

    return (
        <div className={`flex flex-col ${H_STYLES.gap.default} ${H_STYLES.padding.box} bg-white border ${C.borderMedium} rounded-md mb-3 shadow-sm`}>
            <Textarea
                placeholder="経過記録や特記事項を入力..."
                value={note}
                onChange={(e) => setNote(e.target.value)}
                className={`min-h-[80px] border-0 focus-visible:ring-0 resize-none ${H_STYLES.padding.tight} ${C.textPlaceholder} ${H_STYLES.text.base}`}
            />
            <div className={`flex justify-between items-center border-t ${C.borderLight} pt-1.5 mt-0.5`}>
                <span className={`${H_STYLES.text.sm} font-medium ${C.text50}`}>{format(new Date(), "M/d HH:mm")}</span>
                <Button
                    size="sm"
                    onClick={handleSubmit}
                    disabled={!note.trim()}
                    variant="primary"
                    className={`gap-2 shadow-sm ${H_STYLES.button.action}`}
                >
                    <Send className={H_STYLES.button.icon} />
                    記録
                </Button>
            </div>
        </div>
    );
}
