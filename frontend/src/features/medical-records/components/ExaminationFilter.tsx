// React/Framework
import { C, ICON } from "@/lib/design-tokens";
import { memo, useId } from "react";

// External
import { FileText } from "lucide-react";

// Internal
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { DatePicker } from "@/components/shared/DatePicker/DatePicker";
import { usePermission } from "@/hooks/use-permission";

interface ExaminationFilterProps {
  searchTerm: string;
  onSearchChange: (value: string) => void;
  dateStart: string;
  onDateStartChange: (value: string) => void;
  dateEnd: string;
  onDateEndChange: (value: string) => void;
  onImport?: () => void;
}

export const ExaminationFilter = memo(function ExaminationFilter({
  searchTerm,
  onSearchChange,
  dateStart,
  onDateStartChange,
  dateEnd,
  onDateEndChange,
  onImport,
}: ExaminationFilterProps) {
  const { canCreate } = usePermission("medical-records");
  const dateRangeId = useId();
  const startDateId = `${dateRangeId}-start-date`;
  const endDateId = `${dateRangeId}-end-date`;
  return (
    <div className="flex flex-col gap-3">
      <div className="flex justify-end">
        {canCreate ? (
          <Button
            type="button"
            size="sm"
            onClick={onImport}
            className={`${C.bgActionPrimarySolid} ${C.textOnActionPrimary} ${C.hoverBgActionPrimarySolid} ${C.hoverTextOnActionPrimary} ${C.activeBgActionPrimarySolid} ${C.activeTextOnActionPrimary} gap-2 h-11 text-sm shadow-none rounded-full border-transparent px-4`}
          >
            <FileText className={ICON.action} />
            検査取り込み
          </Button>
        ) : null}
      </div>

      {/* Filters */}
      <div
        className={`flex flex-col md:flex-row items-end gap-4 ${C.bgWhite} p-4 rounded-lg border ${C.borderMedium}`}
      >
        <div className="flex flex-col gap-1.5 w-full md:w-[300px]">
          <Label htmlFor="examination-filter-search" className={`text-sm font-medium ${C.text60}`}>
            検査項目検索
          </Label>
          <Input
            id="examination-filter-search"
            value={searchTerm}
            onChange={(e) => onSearchChange(e.target.value)}
            className={`${C.bgWhite} ${C.borderMedium} text-sm`}
            placeholder="WBC, Cre, etc..."
          />
        </div>

        <div className="flex flex-col gap-1.5 w-full md:w-[400px]">
          {/* 期間は2入力のグループ名のため span + 各ピッカーへ sr-only Label */}
          <span className={`text-sm font-medium ${C.text60}`}>期間</span>
          <div className="flex items-center gap-2">
            <Label htmlFor={startDateId} className="sr-only">
              開始日
            </Label>
            <DatePicker
              id={startDateId}
              value={dateStart}
              onChange={onDateStartChange}
              placeholder="開始日"
              className="flex-1"
            />
            <span className={`${C.text} font-medium text-sm`}>〜</span>
            <Label htmlFor={endDateId} className="sr-only">
              終了日
            </Label>
            <DatePicker
              id={endDateId}
              value={dateEnd}
              onChange={onDateEndChange}
              placeholder="終了日"
              className="flex-1"
            />
          </div>
        </div>
      </div>
    </div>
  );
});
