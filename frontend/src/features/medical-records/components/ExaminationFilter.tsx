// React/Framework
import { C, ICON } from "@/lib/design-tokens";
import React from "react";

// External
import { FileText } from "lucide-react";

// Internal
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { NotionDatePicker } from "@/components/shared/NotionDatePicker/NotionDatePicker";
import { usePermission } from "@/features/auth";

interface ExaminationFilterProps {
  searchTerm: string;
  onSearchChange: (value: string) => void;
  dateStart: string;
  onDateStartChange: (value: string) => void;
  dateEnd: string;
  onDateEndChange: (value: string) => void;
}

export const ExaminationFilter = React.memo(function ExaminationFilter({
  searchTerm,
  onSearchChange,
  dateStart,
  onDateStartChange,
  dateEnd,
  onDateEndChange,
}: ExaminationFilterProps) {
  const { canCreate } = usePermission("medical-records");
  return (
    <div className="flex flex-col gap-3">
      <div className="flex justify-end">
        {canCreate ? (
          <Button
            size="sm"
            className={`${C.bgAccent} ${C.bgAccentHover} text-white gap-2 h-10 text-sm shadow-sm border-transparent px-4`}
          >
            <FileText className={ICON.action} />
            検査取り込み
          </Button>
        ) : null}
      </div>

      {/* Filters */}
      <div className={`flex flex-col md:flex-row items-end gap-4 bg-white p-4 rounded-lg border ${C.borderMedium} shadow-sm`}>
        <div className="flex flex-col gap-1.5 w-full md:w-[300px]">
          <Label className={`text-sm font-medium ${C.text60}`}>
            検査項目検索
          </Label>
          <Input
            value={searchTerm}
            onChange={(e) => onSearchChange(e.target.value)}
            className={`bg-white ${C.borderMedium} h-10 text-sm`}
            placeholder="WBC, Cre, etc..."
          />
        </div>

        <div className="flex flex-col gap-1.5 w-full md:w-[400px]">
          <Label className={`text-sm font-medium ${C.text60}`}>
            期間
          </Label>
          <div className="flex items-center gap-2">
            <NotionDatePicker
              value={dateStart}
              onChange={onDateStartChange}
              placeholder="開始日"
              className="flex-1"
            />
            <span className={`${C.text} font-medium text-sm`}>〜</span>
            <NotionDatePicker
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
