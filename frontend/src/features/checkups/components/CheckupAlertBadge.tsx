import { Badge } from "@/components/ui/badge";

interface CheckupAlertBadgeProps {
  nextDate: string | null | undefined;
}

function toISODate(d: Date): string {
  return d.toISOString().slice(0, 10);
}

export function CheckupAlertBadge({ nextDate }: CheckupAlertBadgeProps) {
  if (!nextDate) return null;
  const today = toISODate(new Date());
  if (nextDate < today) {
    return <Badge variant="destructive">期限切れ</Badge>;
  }
  const limit = new Date(today);
  limit.setDate(limit.getDate() + 30);
  if (nextDate <= toISODate(limit)) {
    return <Badge className="bg-[#F0D070] text-[#7A5C00]">期限間近</Badge>;
  }
  return null;
}
