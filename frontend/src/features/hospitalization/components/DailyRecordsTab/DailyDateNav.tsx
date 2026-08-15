// React/Framework
import { C } from "@/lib/design-tokens";
import { useCallback, useMemo } from "react";

// Internal
import { CalendarNavToolbar } from "@/components/shared/CalendarNavToolbar";
import { formatJSTWallDate } from "@/lib/jst-date";
import { formatJapaneseDate } from "@/lib/format/date";

interface DailyDateNavProps {
    selectedDate: string; // YYYY-MM-DD
    admissionDate: string; // YYYY-MM-DD
    dischargeDate: string; // YYYY-MM-DD (today if not discharged)
    onDateChange: (date: string) => void;
}

function addDays(dateStr: string, days: number): string {
    const d = new Date(dateStr + "T00:00:00");
    d.setDate(d.getDate() + days);
    return formatJSTWallDate(d);
}

export function DailyDateNav({
    selectedDate,
    admissionDate,
    dischargeDate,
    onDateChange,
}: DailyDateNavProps) {
    const canGoPrev = selectedDate > admissionDate;
    const canGoNext = selectedDate < dischargeDate;

    const handlePrev = useCallback(() => {
        if (canGoPrev) {
            onDateChange(addDays(selectedDate, -1));
        }
    }, [canGoPrev, selectedDate, onDateChange]);

    const handleNext = useCallback(() => {
        if (canGoNext) {
            onDateChange(addDays(selectedDate, 1));
        }
    }, [canGoNext, selectedDate, onDateChange]);

    // FE4-8: 共有 formatJapaneseDate へ委譲（ローカル getter 由来のため出力同一）
    const displayDate = useMemo(() => {
        const d = new Date(selectedDate + "T00:00:00");
        return formatJapaneseDate(d);
    }, [selectedDate]);

    return (
        <CalendarNavToolbar
            layout="spread"
            size="sm"
            className="py-2 px-1"
            onPrev={handlePrev}
            onNext={handleNext}
            prevDisabled={!canGoPrev}
            nextDisabled={!canGoNext}
            prevAriaLabel="前日"
            nextAriaLabel="翌日"
            label={<span className={`text-sm font-semibold ${C.text}`}>{displayDate}</span>}
        />
    );
}
