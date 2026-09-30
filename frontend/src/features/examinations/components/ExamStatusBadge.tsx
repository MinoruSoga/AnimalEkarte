import { CheckCircle } from "lucide-react";

import { Badge } from "@/components/ui/badge";
import { BADGE, C, ICON } from "@/lib/design-tokens";

interface ExamStatusBadgeProps {
  status?: "normal" | "high" | "low";
  isAssessed?: boolean;
  /**
   * FE-RC-027: ExamPivotTable と ExamItemsTable の完全コピペ StatusBadge を統合。
   * pivot（compact）はセルを詰めるため normal/未判定以外の内訳を出さず、
   * items（既定）は判定列で normal=CheckCircle・未判定でもない未保存=「-」を明示する。
   * 挙動は各呼び出し元の既存表示を変えずに温存する。
   */
  compact?: boolean;
}

export function ExamStatusBadge({ status, isAssessed, compact = false }: ExamStatusBadgeProps) {
  const horizontalPadding = compact ? "px-2" : "px-3";

  if (status === "high") {
    return (
      <Badge
        variant="destructive"
        className={`h-8 ${horizontalPadding} text-xs ${C.bgDanger} ${C.hoverBgDanger90}`}
      >
        HIGH
      </Badge>
    );
  }

  if (status === "low") {
    return (
      <Badge variant="outline" className={`h-8 ${horizontalPadding} text-xs ${BADGE.blue}`}>
        LOW
      </Badge>
    );
  }

  if (isAssessed === false) {
    return (
      <Badge variant="outline" className={`h-8 ${horizontalPadding} text-xs ${BADGE.yellow}`}>
        未判定
        <span className="sr-only">（基準値未設定のため判定していない）</span>
      </Badge>
    );
  }

  if (compact) return null;

  if (status === "normal") {
    return (
      <CheckCircle
        role="img"
        aria-label="基準値内"
        className={`${ICON.action} ${C.textStatusGreen}`}
      />
    );
  }

  // 未判定（保存前）
  return <span className={`text-xs ${C.text60}`}>-</span>;
}
