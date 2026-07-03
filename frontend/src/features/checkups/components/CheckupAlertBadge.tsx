import { Badge } from "@/components/ui/badge";
import { C } from "@/lib/design-tokens";
import { todayISODate, addDaysISO } from "../lib/today-iso";

interface CheckupAlertBadgeProps {
  nextDate: string | null | undefined;
}

export function CheckupAlertBadge({ nextDate }: CheckupAlertBadgeProps) {
  if (!nextDate) return null;
  const today = todayISODate();
  if (nextDate < today) {
    return <Badge variant="destructive">期限切れ</Badge>;
  }
  const limit = addDaysISO(today, 30);
  if (nextDate <= limit) {
    return <Badge className={`${C.bgCheckupDueSoon} ${C.textCheckupDueSoon}`}>期限間近</Badge>;
  }
  return null;
}
