// React/Framework
import { C, ICON } from "@/lib/design-tokens";
import { memo } from "react";

// External
import { CheckCircle } from "lucide-react";

// Internal
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";

interface ExamResult {
  id: number;
  name: string;
  result: string;
  unit: string;
  reference_value: string;
  status: string;
}

interface ExaminationGroupProps {
  group: {
    id: number;
    date: string;
    machine: string;
    items: ExamResult[];
  };
}

export const ExaminationGroup = memo(function ExaminationGroup({
  group,
}: ExaminationGroupProps) {
  return (
    <div className="flex flex-col gap-2">
      <div className={`flex items-center justify-between w-full border-b ${C.borderPrimary20} pb-2`}>
        <div className="flex items-center gap-4">
          <h3 className={`text-base font-bold ${C.text} font-mono`}>
            {group.date}
          </h3>
          <Badge
            variant="secondary"
            className={`${C.bgPage} ${C.text80} ${C.hoverBgPage} font-normal px-2 py-0.5 text-sm h-10 border ${C.borderMedium}`}
          >
            {group.machine}
          </Badge>
        </div>
        <Button
          variant="ghost"
          size="sm"
          className={`h-10 text-sm ${C.text60} ${C.hoverText}`}
        >
          詳細を表示
        </Button>
      </div>

      <div className={`border ${C.borderMedium} rounded-lg bg-white overflow-hidden shadow-sm overflow-x-auto`}>
        <div className={`min-w-[600px] grid grid-cols-[2fr_1.5fr_1.5fr_2fr_1.5fr] gap-0 border-b ${C.borderMedium} ${C.bgPage} text-sm font-bold ${C.text80} h-12 items-center`}>
          <div className={`p-2 border-r ${C.borderMedium} pl-3`}>
            項目名
          </div>
          <div className={`p-2 border-r ${C.borderMedium} text-right`}>
            結果値
          </div>
          <div className={`p-2 border-r ${C.borderMedium} text-center`}>
            単位
          </div>
          <div className={`p-2 border-r ${C.borderMedium} text-center`}>
            基準値
          </div>
          <div className="p-2 text-center">判定</div>
        </div>
        {group.items.map((item, idx) => (
          <div
            key={item.id}
            className={`min-w-[600px] grid grid-cols-[2fr_1.5fr_1.5fr_2fr_1.5fr] gap-0 ${
              idx !== group.items.length - 1
                ? `border-b ${C.borderMedium}`
                : ""
            } bg-white text-sm ${C.text} items-center ${C.hoverBgPageHalf} h-12 transition-colors`}
          >
            <div className={`p-2 border-r ${C.borderMedium} pl-3 font-medium`}>
              {item.name}
            </div>
            <div
              className={`p-2 border-r ${C.borderMedium} text-right font-mono ${
                item.status === "high"
                  ? `${C.danger} font-bold`
                  : item.status === "low"
                  ? `${C.accent} font-bold`
                  : ""
              }`}
            >
              {item.result}
            </div>
            <div className={`p-2 border-r ${C.borderMedium} text-center ${C.text60} text-sm`}>
              {item.unit}
            </div>
            <div className={`p-2 border-r ${C.borderMedium} text-center ${C.text60} text-sm`}>
              {item.reference_value}
            </div>
            <div className="p-2 flex justify-center items-center">
              {item.status === "high" ? (
                <Badge
                  variant="destructive"
                  className={`h-10 px-3 text-sm ${C.bgDanger} ${C.hoverBgDanger90}`}
                >
                  HIGH
                </Badge>
              ) : null}
              {item.status === "low" ? (
                <Badge
                  variant="outline"
                  className={`h-10 px-3 text-sm ${C.accent} ${C.borderAccent} ${C.bgAccent5}`}
                >
                  LOW
                </Badge>
              ) : null}
              {item.status === "normal" ? (
                <CheckCircle className={`${ICON.page} ${C.textStatusGreen} opacity-50`} />
              ) : null}
            </div>
          </div>
        ))}
      </div>
    </div>
  );
});
