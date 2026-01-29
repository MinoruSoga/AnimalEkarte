// React/Framework
import React from "react";

// External
import { FileText } from "lucide-react";

// Internal
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";

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
  return (
    <div className="flex flex-col gap-3">
      <div className="flex justify-end">
        <Button
          size="sm"
          className="bg-[#37352F] hover:bg-[#37352F]/90 text-white gap-2 h-10 text-sm shadow-sm border-transparent px-4"
        >
          <FileText className="size-4" />
          検査取り込み
        </Button>
      </div>

      {/* Filters */}
      <div className="flex flex-col md:flex-row items-end gap-4 bg-white p-4 rounded-lg border border-[rgba(55,53,47,0.16)] shadow-sm">
        <div className="flex flex-col gap-1.5 w-full md:w-[300px]">
          <Label className="text-sm font-medium text-[#37352F]/60">
            検査項目検索
          </Label>
          <Input
            value={searchTerm}
            onChange={(e) => onSearchChange(e.target.value)}
            className="bg-white border-[rgba(55,53,47,0.16)] h-10 text-sm"
            placeholder="WBC, Cre, etc..."
          />
        </div>

        <div className="flex flex-col gap-1.5 w-full md:w-[400px]">
          <Label className="text-sm font-medium text-[#37352F]/60">
            期間
          </Label>
          <div className="flex items-center gap-2">
            <Input
              type="date"
              value={dateStart}
              onChange={(e) => onDateStartChange(e.target.value)}
              className="bg-white border-[rgba(55,53,47,0.16)] h-10 flex-1 text-sm"
            />
            <span className="text-[#37352F] font-medium text-sm">〜</span>
            <Input
              type="date"
              value={dateEnd}
              onChange={(e) => onDateEndChange(e.target.value)}
              className="bg-white border-[rgba(55,53,47,0.16)] h-10 flex-1 text-sm"
            />
          </div>
        </div>
      </div>
    </div>
  );
});
