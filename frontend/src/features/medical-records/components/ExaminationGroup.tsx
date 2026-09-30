// React/Framework
import { C, ICON, STYLE } from "@/lib/design-tokens";
import { memo, useEffect, useRef } from "react";
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
  highlighted?: boolean;
}

export const ExaminationGroup = memo(function ExaminationGroup({
  group,
  petId,
  highlighted = false,
}: ExaminationGroupProps) {
  const location = useLocation();
  const highlightRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    if (!highlighted) return;
    highlightRef.current?.scrollIntoView({ block: "start" });
  }, [highlighted]);
  const historyLocation = `${location.pathname}${location.search}`;
  const pivotHref = petId
    ? `${paths.examinations.detail.getHref(group.id)}?${new URLSearchParams({
        petId,
        historyView: "pivot",
      }).toString()}`
    : undefined;

  return (
    <div
      ref={highlightRef}
      id={`exam-group-${group.id}`}
      aria-current={highlighted ? "true" : undefined}
      className="flex flex-col gap-2"
    >
      <div
        className={`flex items-center justify-between w-full border-b ${C.borderPrimary20} pb-2`}
      >
        <div className="flex items-center gap-4">
          <h3 className={`text-base font-bold ${C.text} font-mono`}>{group.date}</h3>
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
            className={`h-11 text-sm ${C.text60} ${C.hoverText}`}
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

      <div
        className={`border ${C.borderMedium} rounded-lg ${C.bgWhite} overflow-hidden overflow-x-auto`}
      >
        {/* EMR-227: div グリッドの疑似表をネイティブ table 化し、列見出し/行見出しの semantics を持たせる */}
        <table className="min-w-[600px] w-full table-fixed border-collapse">
          <colgroup>
            {/* 2fr:1.5fr:1.5fr:2fr:1.5fr の比率を維持 */}
            <col className="w-[23.5%]" />
            <col className="w-[17.6%]" />
            <col className="w-[17.6%]" />
            <col className="w-[23.5%]" />
            <col className="w-[17.6%]" />
          </colgroup>
          <thead>
            <tr className={`border-b ${C.borderMedium} ${C.bgPage} h-12`}>
              <th
                scope="col"
                className={`${STYLE.tableHeaderCell} border-r ${C.borderMedium} text-left`}
              >
                項目名
              </th>
              <th
                scope="col"
                className={`${STYLE.tableHeaderCell} border-r ${C.borderMedium} text-right`}
              >
                結果値
              </th>
              <th
                scope="col"
                className={`${STYLE.tableHeaderCell} border-r ${C.borderMedium} text-center`}
              >
                単位
              </th>
              <th
                scope="col"
                className={`${STYLE.tableHeaderCell} border-r ${C.borderMedium} text-center`}
              >
                基準値
              </th>
              <th scope="col" className={`${STYLE.tableHeaderCell} text-center`}>
                判定
              </th>
            </tr>
          </thead>
          <tbody>
            {group.items.map((item, idx) => (
              <tr
                key={item.id}
                className={`${
                  idx !== group.items.length - 1 ? `border-b ${C.borderMedium}` : ""
                } ${C.bgWhite} ${C.hoverBgPageHalf} h-12 transition-colors`}
              >
                <th
                  scope="row"
                  className={`${STYLE.tableHeaderCell} border-r ${C.borderMedium} text-left`}
                >
                  {item.name}
                </th>
                <td
                  className={`${STYLE.tableCellMono} border-r ${C.borderMedium} text-right ${
                    item.status === "high"
                      ? `${C.danger} font-semibold`
                      : item.status === "low"
                        ? `${C.textStatusBlue} font-semibold`
                        : ""
                  }`}
                >
                  {item.inspectionValue || item.result || "-"}
                </td>
                <td className={`${STYLE.tableCellMuted} border-r ${C.borderMedium} text-center`}>
                  {item.unit || "-"}
                </td>
                <td className={`${STYLE.tableCellMuted} border-r ${C.borderMedium} text-center`}>
                  {item.referenceValue || item.normalValue || "-"}
                </td>
                <td className={`${STYLE.tableCell} text-center`}>
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
                      className={`${ICON.action} ${C.textStatusGreen} opacity-50 inline-block`}
                    />
                  ) : null}
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </div>
  );
});
