import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { C } from "@/lib/design-tokens";

import type { UnpaidGroupBy } from "../lib/unpaid-tab-model";

interface UnpaidTabFiltersProps {
  groupBy: UnpaidGroupBy;
  startDate: string;
  endDate: string;
  monthParam: string;
  onStartDateChange: (next: string) => void;
  onEndDateChange: (next: string) => void;
  onMonthChange: (next: string) => void;
  onGroupByChange: (next: UnpaidGroupBy) => void;
}

export function UnpaidTabFilters({
  groupBy,
  startDate,
  endDate,
  monthParam,
  onStartDateChange,
  onEndDateChange,
  onMonthChange,
  onGroupByChange,
}: UnpaidTabFiltersProps) {
  return (
    <div className="flex items-end gap-4 flex-wrap">
      {groupBy !== "monthly" ? (
        <>
          <div className="space-y-1.5">
            <Label htmlFor="startDate" className={`text-sm ${C.text60}`}>
              開始日
            </Label>
            <Input
              id="startDate"
              type="date"
              value={startDate}
              onChange={(e) => onStartDateChange(e.target.value)}
              className="h-9 text-sm"
            />
          </div>
          <div className="space-y-1.5">
            <Label htmlFor="endDate" className={`text-sm ${C.text60}`}>
              終了日
            </Label>
            <Input
              id="endDate"
              type="date"
              value={endDate}
              onChange={(e) => onEndDateChange(e.target.value)}
              className="h-9 text-sm"
            />
          </div>
        </>
      ) : (
        <div className="space-y-1.5">
          <Label htmlFor="monthPicker" className={`text-sm ${C.text60}`}>
            対象月
          </Label>
          <Input
            id="monthPicker"
            type="month"
            value={monthParam}
            onChange={(e) => onMonthChange(e.target.value)}
            className="h-9 text-sm"
          />
        </div>
      )}
      <div className="flex gap-2">
        <Button
          type="button"
          variant={groupBy === "owner" ? "default" : "outline"}
          size="sm"
          onClick={() => onGroupByChange("owner")}
        >
          飼主単位
        </Button>
        <Button
          type="button"
          variant={groupBy === "billing" ? "default" : "outline"}
          size="sm"
          onClick={() => onGroupByChange("billing")}
        >
          会計単位
        </Button>
        <Button
          type="button"
          variant={groupBy === "monthly" ? "default" : "outline"}
          size="sm"
          onClick={() => onGroupByChange("monthly")}
        >
          月次繰越
        </Button>
      </div>
    </div>
  );
}
