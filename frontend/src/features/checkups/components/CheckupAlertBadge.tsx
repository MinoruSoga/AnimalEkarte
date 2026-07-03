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
    // FE-refactor.md FD3: #F0D070/#7A5C00 have no matching entry in design-tokens.ts
    // PALETTE (checked against warningBg/warningText/noticeBg/noticeText — none match).
    // Left as raw hex to preserve the exact rendered color (behavior-preserving refactor);
    // aligning this with an existing warning token would change the visual and is out of
    // scope here.
    return <Badge className="bg-[#F0D070] text-[#7A5C00]">期限間近</Badge>;
  }
  return null;
}
