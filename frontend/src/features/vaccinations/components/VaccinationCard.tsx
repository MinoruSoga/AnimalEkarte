import { memo } from "react";
import { Calendar, Syringe, UserRound } from "lucide-react";
import { Card, CardContent } from "@/components/ui/card";
import { CheckupAlertBadge } from "@/components/shared/CheckupAlertBadge/CheckupAlertBadge";
import { formatDate } from "@/lib/format/date";
import { C, ICON } from "@/lib/design-tokens";
import type { VaccinationRecord } from "@/types";

interface VaccinationCardProps {
  vaccination: VaccinationRecord;
  /** EMR-428: 同じワクチンでより新しい接種がある等で警告を出すべきでないとき true（日付は表示する） */
  isReplaced?: boolean;
  onClick?: () => void;
  className?: string;
}

export const VaccinationCard = memo(function VaccinationCard({
  vaccination,
  isReplaced = false,
  onClick,
  className,
}: VaccinationCardProps) {
  return (
    <Card
      className={`${C.bgWhite} border ${C.borderLight} shadow-none rounded-xs ${C.hoverBgPage} transition-colors ${onClick ? "cursor-pointer" : ""} ${className ?? ""}`}
      onClick={onClick}
    >
      <CardContent className="px-4 py-3">
        {/* Header */}
        <div className="flex items-center gap-2 min-w-0">
          <Syringe className={`${ICON.action} shrink-0 ${C.text45}`} />
          <span className={`text-base font-medium ${C.text} truncate`}>
            {vaccination.vaccineName}
          </span>
        </div>

        {/* Meta */}
        <div className={`flex items-center gap-4 mt-1.5 text-sm ${C.text60} flex-wrap`}>
          {/* 接種日 */}
          <span className="flex items-center gap-1">
            <Calendar className={`${ICON.xs} shrink-0`} />
            接種: {formatDate(vaccination.date)}
          </span>

          {vaccination.doctor ? (
            <span className="flex items-center gap-1">
              <UserRound className={`${ICON.xs} shrink-0`} />
              担当医: {vaccination.doctor}
            </span>
          ) : null}
        </div>

        {/* 次回接種予定日 */}
        {vaccination.nextDate ? (
          <div className={`flex items-center gap-1.5 mt-1.5 text-sm ${C.text60}`}>
            <Calendar className={`${ICON.xs} shrink-0`} />
            <span>次回: {formatDate(vaccination.nextDate)}</span>
            {!isReplaced ? <CheckupAlertBadge nextDate={vaccination.nextDate} /> : null}
          </div>
        ) : null}
      </CardContent>
    </Card>
  );
});
