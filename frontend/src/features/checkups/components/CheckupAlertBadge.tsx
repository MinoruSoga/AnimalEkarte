import { Badge } from "@/components/ui/badge";
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
    return <Badge className="bg-[#F0D070] text-[#7A5C00]">期限間近</Badge>;
  }
  return null;
}
