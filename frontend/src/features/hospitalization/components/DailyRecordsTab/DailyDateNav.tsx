// React/Framework
import { C, ICON } from "@/lib/design-tokens";
import { useCallback, useMemo } from "react";

// External
import { ChevronLeft, ChevronRight } from "lucide-react";

// Internal
import { Button } from "@/components/ui/button";

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
    return d.toISOString().split("T")[0];
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
        <div className="flex items-center justify-between py-2 px-1">
            <Button
                variant="ghost"
                size="sm"
                onClick={handlePrev}
                disabled={!canGoPrev}
                className="h-8 w-8 p-0"
                aria-label="前日"
            >
                <ChevronLeft className={ICON.action} />
            </Button>
            <span className={`text-sm font-semibold ${C.text}`}>{displayDate}</span>
            <Button
                variant="ghost"
                size="sm"
                onClick={handleNext}
                disabled={!canGoNext}
                className="h-8 w-8 p-0"
                aria-label="翌日"
            >
                <ChevronRight className={ICON.action} />
            </Button>
        </div>
    );
}
