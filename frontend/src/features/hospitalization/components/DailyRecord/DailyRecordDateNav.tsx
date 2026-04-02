// External
import { C, ICON } from "@/lib/design-tokens";
import { format } from "date-fns";
import { ja } from "date-fns/locale";
import { ChevronLeft, ChevronRight } from "lucide-react";

// Internal
import { Button } from "@/components/ui/button";

// Relative
import { H_STYLES } from "@/features/hospitalization/styles";

interface DailyRecordDateNavProps {
    date: Date;
    onPrev: () => void;
    onNext: () => void;
}

export function DailyRecordDateNav({ date, onPrev, onNext }: DailyRecordDateNavProps) {
    return (
        <div className={`flex items-center justify-between bg-white ${H_STYLES.padding.box} rounded-lg border ${C.borderMedium} shadow-sm`}>
            <Button variant="ghost" size="icon" onClick={onPrev} className="h-11 w-11">
                <ChevronLeft className={ICON.lg} />
            </Button>
            <div className={`font-bold ${C.text} ${H_STYLES.text.lg}`}>
                {format(date, "yyyy年 M月 d日 (EEE)", { locale: ja })}
            </div>
            <Button variant="ghost" size="icon" onClick={onNext} className="h-11 w-11">
                <ChevronRight className={ICON.lg} />
            </Button>
        </div>
    );
}
