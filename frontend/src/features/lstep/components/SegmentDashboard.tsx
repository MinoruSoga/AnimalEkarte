import { useMemo } from "react";
import { Users } from "lucide-react";
import { Button } from "@/components/ui/button";
import { C, STYLE, ICON } from "@/lib/design-tokens";
import type { LstepTagSummaryItem } from "../api/get-lstep-tag-summary";
import { SEGMENT_GROUPS } from "../constants/segment-tags";

interface SegmentDashboardProps {
  summaryTags: LstepTagSummaryItem[];
  onViewOwners: (tagName: string, ownerCount: number) => void;
}

export function SegmentDashboard({ summaryTags, onViewOwners }: SegmentDashboardProps) {
  const countMap = useMemo(() => {
    const map = new Map<string, number>();
    for (const tag of summaryTags) {
      map.set(tag.tag_name, tag.owner_count);
    }
    return map;
  }, [summaryTags]);

  return (
    <div className="flex flex-col gap-4">
      {SEGMENT_GROUPS.map((group) => (
        <div key={group.title}>
          <p className={`text-xs font-semibold uppercase tracking-wide ${C.text50} mb-2`}>
            {group.title}
          </p>
          <div className={`grid ${group.gridCols} gap-2`}>
            {group.tags.map((seg) => {
              const count = countMap.get(seg.tagName) ?? 0;
              return (
                <div
                  key={seg.tagName}
                  className={`bg-white border ${C.borderLight} rounded-[4px] px-4 py-3 flex flex-col gap-1`}
                >
                  <p className={`text-xs font-medium ${C.text} truncate`}>{seg.label}</p>
                  <p className={`text-[10px] font-mono ${C.text40} truncate`}>{seg.tagName}</p>
                  <p className={`text-2xl font-bold ${C.text} mt-0.5`}>
                    {count.toLocaleString("ja-JP")}
                    <span className={`text-sm font-normal ${C.text50} ml-1`}>名</span>
                  </p>
                  <Button
                    variant="outline"
                    className={`h-8 px-2 text-xs mt-1 self-start ${STYLE.btnOutline}`}
                    onClick={() => onViewOwners(seg.tagName, count)}
                  >
                    <Users className={`mr-1 ${ICON.sm}`} />
                    対象者一覧
                  </Button>
                </div>
              );
            })}
          </div>
        </div>
      ))}
    </div>
  );
}
