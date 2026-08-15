// React/Framework
import { C, ICON, STYLE } from "@/lib/design-tokens";
import { memo } from "react";
import { Link, useLocation } from "react-router";

// External
import { CheckCircle } from "lucide-react";

// Internal
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { paths } from "@/config/paths";
import type { ExamGroup } from "../api/get-record-examinations";

interface ExaminationGroupProps {
  group: ExamGroup;
  petId?: string;
}

export const ExaminationGroup = memo(function ExaminationGroup({
  group,
  petId,
}: ExaminationGroupProps) {
  const location = useLocation();
  const historyLocation = `${location.pathname}${location.search}`;
  const pivotHref = petId
    ? `${paths.examinations.detail.getHref(group.id)}?${new URLSearchParams({
        petId,
        historyView: "pivot",
      }).toString()}`
    : undefined;

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
        {pivotHref ? (
          <Button
            asChild
            variant="ghost"
            size="sm"
            className={`h-10 text-sm ${C.text60} ${C.hoverText}`}
          >
            <Link
              to={pivotHref}
              state={{ from: historyLocation }}
              aria-label={`${group.date}の検歴を表示`}
            >
              検歴を表示
            </Link>
          </Button>
        ) : null}
      </div>

      <div className={`border ${C.borderMedium} rounded-lg ${C.bgWhite} overflow-hidden overflow-x-auto`}>
        <div className={`min-w-[600px] grid grid-cols-[2fr_1.5fr_1.5fr_2fr_1.5fr] gap-0 border-b ${C.borderMedium} ${C.bgPage} ${STYLE.sectionLabel} h-12 items-center`}>
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
            } ${C.bgWhite} text-sm ${C.text} items-center ${C.hoverBgPageHalf} h-12 transition-colors`}
          >
            <div className={`p-2 border-r ${C.borderMedium} pl-3 font-medium`}>
              {item.name}
            </div>
            <div
              className={`p-2 border-r ${C.borderMedium} text-right font-mono ${
                item.status === "high"
                  ? `${C.danger} font-bold`
                  : item.status === "low"
                  ? `${C.textStatusBlue} font-bold`
                  : ""
              }`}
            >
              {item.inspectionValue || item.result || "-"}
            </div>
            <div className={`p-2 border-r ${C.borderMedium} text-center ${C.text60} text-sm`}>
              {item.unit || "-"}
            </div>
            <div className={`p-2 border-r ${C.borderMedium} text-center ${C.text60} text-sm`}>
              {item.referenceValue || item.normalValue || "-"}
            </div>
            <div className="p-2 flex justify-center items-center">
              {item.status === "high" ? (
                <Badge
                  variant="destructive"
                  className={`h-10 px-3 text-sm ${C.bgDanger} ${C.hoverBgDanger90}`}
                >
                  HIGH
                </Badge>
              ) : item.status === "low" ? (
                <Badge
                  variant="outline"
                  className={`h-10 px-3 text-sm ${C.textStatusBlue} ${C.borderBlue400} ${C.bgStatusBlueLight}`}
                >
                  LOW
                </Badge>
              ) : item.isAssessed === false ? (
                <Badge
                  variant="outline"
                  className={`h-10 px-3 text-sm ${C.textWarning} ${C.borderWarning20} ${C.bgWarning50}`}
                >
                  未判定
                  <span className="sr-only">（基準値未設定のため判定していない）</span>
                </Badge>
              ) : item.status === "normal" ? (
                <CheckCircle
                  role="img"
                  aria-label="基準値内"
                  className={`${ICON.action} ${C.textStatusGreen} opacity-50`}
                />
              ) : null}
            </div>
          </div>
        ))}
      </div>
    </div>
  );
});
