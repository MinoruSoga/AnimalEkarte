// React/Framework
import { useState } from "react";

// External
import { Send } from "lucide-react";
import { format } from "date-fns";

// Internal
import { Button } from "@/components/ui/button";
import { Textarea } from "@/components/ui/textarea";

// Relative
import { H_STYLES } from "../../styles";

// Types
import type { CreateCareLogDTO } from "../../types";

interface SimpleNoteFormProps {
    onSave: (data: CreateCareLogDTO) => void;
}

export const SimpleNoteForm = ({ onSave }: SimpleNoteFormProps) => {
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
        <div className={`flex flex-col ${H_STYLES.gap.default} ${H_STYLES.padding.box} bg-white border border-[rgba(55,53,47,0.16)] rounded-md mb-3 shadow-sm`}>
            <Textarea 
                placeholder="経過記録や特記事項を入力..." 
                value={note}
                onChange={(e) => setNote(e.target.value)}
                className={`min-h-[80px] border-0 focus-visible:ring-0 resize-none ${H_STYLES.padding.tight} placeholder:text-gray-400 ${H_STYLES.text.base}`}
            />
            <div className="flex justify-between items-center border-t border-gray-100 pt-1.5 mt-0.5">
                <span className={`${H_STYLES.text.sm} font-medium text-gray-500`}>{format(new Date(), "M/d HH:mm")}</span>
                <Button 
                    size="sm" 
                    onClick={handleSubmit}
                    disabled={!note.trim()}
                    className={`bg-[#2EAADC] hover:bg-[#2EAADC]/90 text-white gap-2 shadow-sm ${H_STYLES.button.action}`}
                >
                    <Send className={H_STYLES.button.icon} />
                    記録
                </Button>
            </div>
        </div>
    );
};
