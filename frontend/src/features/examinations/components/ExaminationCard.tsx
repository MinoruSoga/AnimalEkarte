import { memo } from "react";
import { Calendar, FlaskConical, User } from "lucide-react";
import { Card, CardContent } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { BADGE, C, ICON } from "@/lib/design-tokens";
import type { ExaminationRecord } from "../api/transforms";

interface ExaminationCardProps {
  examination: ExaminationRecord;
  onClick?: () => void;
  className?: string;
}

function getStatusBadge(status: ExaminationRecord["status"]): string {
  switch (status) {
    case "確定":
      return BADGE.green;
    case "完了":
      return BADGE.blue;
    case "結果入力済み":
      return BADGE.blue;
    case "検査中":
      return BADGE.yellow;
    case "依頼中":
      return BADGE.gray;
    default:
      return BADGE.gray;
  }
}

export const ExaminationCard = memo(function ExaminationCard({
  examination,
  onClick,
  className,
}: ExaminationCardProps) {
  return (
    <Card
      className={`${C.bgWhite} border ${C.borderLight} shadow-none rounded-xs ${C.hoverBgPage} transition-colors ${onClick ? "cursor-pointer" : ""} ${className ?? ""}`}
      onClick={onClick}
    >
      <CardContent className="px-4 py-3">
        {/* Header row */}
        <div className="flex items-start justify-between gap-2">
          <div className="flex items-center gap-2 min-w-0">
            <FlaskConical className={`${ICON.action} shrink-0 ${C.text45}`} />
            <span className={`text-base font-medium ${C.text} truncate`}>
              {examination.testType}
            </span>
          </div>
          <Badge
            variant="outline"
            className={`text-xs px-1.5 h-5 font-normal border shrink-0 ${getStatusBadge(examination.status)}`}
          >
            {examination.status}
          </Badge>
        </div>

        {/* Meta row */}
        <div className={`flex items-center gap-4 mt-1.5 text-sm ${C.text60} flex-wrap`}>
          <span className="flex items-center gap-1">
            <Calendar className={`${ICON.xs} shrink-0`} />
            {examination.date}
          </span>
          {examination.doctor ? (
            <span className="flex items-center gap-1">
              <User className={`${ICON.xs} shrink-0`} />
              {examination.doctor}
            </span>
          ) : null}
        </div>

        {/* Result summary */}
        {examination.resultSummary ? (
          <p className={`mt-2 text-sm ${C.text70} line-clamp-2`}>
            {examination.resultSummary}
          </p>
        ) : null}
      </CardContent>
    </Card>
  );
});
