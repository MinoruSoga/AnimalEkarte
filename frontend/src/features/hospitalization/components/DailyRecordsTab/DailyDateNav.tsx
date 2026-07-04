// React/Framework
import { C } from "@/lib/design-tokens";
import { useCallback, useMemo } from "react";

// Internal
import { CalendarNavToolbar } from "@/components/shared/CalendarNavToolbar";
import { formatJSTWallDate } from "@/lib/jst-date";

const WEEK_DAYS = ["日", "月", "火", "水", "木", "金", "土"];

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
    const canGoPrev = useMemo(
        () => selectedDate > admissionDate,
        [selectedDate, admissionDate]
    );
    const canGoNext = useMemo(
        () => selectedDate < dischargeDate,
        [selectedDate, dischargeDate]
    );

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

    // Format for display: YYYY-MM-DD -> YYYY年M月D日（曜日）
    const displayDate = useMemo(() => {
        const d = new Date(selectedDate + "T00:00:00");
        return `${d.getFullYear()}年${d.getMonth() + 1}月${d.getDate()}日（${WEEK_DAYS[d.getDay()]}）`;
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
