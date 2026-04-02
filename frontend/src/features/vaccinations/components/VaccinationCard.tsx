import { Calendar, Syringe, AlertTriangle } from "lucide-react";
import { Card, CardContent } from "@/components/ui/card";
import { C, ICON } from "@/lib/design-tokens";
import type { VaccinationRecord } from "@/types";

interface VaccinationCardProps {
  vaccination: VaccinationRecord;
  onClick?: () => void;
  className?: string;
}

function isOverdue(nextDate: string): boolean {
  if (!nextDate) return false;
  const next = new Date(nextDate);
  if (isNaN(next.getTime())) return false;
  return next < new Date();
}

export function VaccinationCard({
  vaccination,
  onClick,
  className,
}: VaccinationCardProps) {
  const overdue = isOverdue(vaccination.nextDate);

  return (
    <Card
      className={`bg-white border ${C.borderLight} shadow-none rounded-[4px] ${C.hoverBgPage} transition-colors ${onClick ? "cursor-pointer" : ""} ${className ?? ""}`}
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
            接種: {vaccination.date}
          </span>

          {/* 担当医 */}
          {/* VaccinationRecord に doctor フィールドは現状存在しないため
              ownerName を代替表示せず、フィールドが追加されたときに対応 */}
        </div>

        {/* 次回接種予定日 */}
        {vaccination.nextDate ? (
          <div
            className={`flex items-center gap-1.5 mt-1.5 text-sm ${
              overdue ? C.danger : C.text60
            }`}
          >
            {overdue ? (
              <AlertTriangle className={`${ICON.xs} shrink-0`} />
            ) : (
              <Calendar className={`${ICON.xs} shrink-0`} />
            )}
            <span>
              次回: {vaccination.nextDate}
              {overdue ? (
                <span className="ml-1.5 text-xs font-medium">（期限超過）</span>
              ) : null}
            </span>
          </div>
        ) : null}
      </CardContent>
    </Card>
  );
}
