import React from "react";
import { Button } from "../../../components/ui/button";
import { Badge } from "../../../components/ui/badge";
import { CheckCircle } from "lucide-react";

interface ExaminationItem {
  id: number;
  name: string;
  result: string;
  unit: string;
  ref: string;
  status: string;
}

interface ExaminationGroupProps {
  group: {
    id: number;
    date: string;
    machine: string;
    items: ExaminationItem[];
  };
}

export const ExaminationGroup = React.memo(function ExaminationGroup({
  group,
}: ExaminationGroupProps) {
  return (
    <div className="flex flex-col gap-2">
      <div className="flex items-center justify-between w-full border-b border-[#37352F]/20 pb-2">
        <div className="flex items-center gap-4">
          <h3 className="text-base font-bold text-[#37352F] font-mono">
            {group.date}
          </h3>
          <Badge
            variant="secondary"
            className="bg-[#F7F6F3] text-[#37352F]/80 hover:bg-[#F7F6F3] font-normal px-2 py-0.5 text-sm h-10 border border-[rgba(55,53,47,0.16)]"
          >
            {group.machine}
          </Badge>
        </div>
        <Button
          variant="ghost"
          size="sm"
          className="h-10 text-sm text-[#37352F]/60 hover:text-[#37352F]"
        >
          詳細を表示
        </Button>
      </div>

      <div className="border border-[rgba(55,53,47,0.16)] rounded-lg bg-white overflow-hidden shadow-sm overflow-x-auto">
        <div className="min-w-[600px] grid grid-cols-[2fr_1.5fr_1.5fr_2fr_1.5fr] gap-0 border-b border-[rgba(55,53,47,0.16)] bg-[#F7F6F3] text-sm font-bold text-[#37352F]/80 h-12 items-center">
          <div className="p-2 border-r border-[rgba(55,53,47,0.16)] pl-3">
            項目名
          </div>
          <div className="p-2 border-r border-[rgba(55,53,47,0.16)] text-right">
            結果値
          </div>
          <div className="p-2 border-r border-[rgba(55,53,47,0.16)] text-center">
            単位
          </div>
          <div className="p-2 border-r border-[rgba(55,53,47,0.16)] text-center">
            基準値
          </div>
          <div className="p-2 text-center">判定</div>
        </div>
        {group.items.map((item, idx) => (
          <div
            key={item.id}
            className={`min-w-[600px] grid grid-cols-[2fr_1.5fr_1.5fr_2fr_1.5fr] gap-0 ${
              idx !== group.items.length - 1
                ? "border-b border-[rgba(55,53,47,0.16)]"
                : ""
            } bg-white text-sm text-[#37352F] items-center hover:bg-[#F7F6F3]/50 h-12 transition-colors`}
          >
            <div className="p-2 border-r border-[rgba(55,53,47,0.16)] pl-3 font-medium">
              {item.name}
            </div>
            <div
              className={`p-2 border-r border-[rgba(55,53,47,0.16)] text-right font-mono ${
                item.status === "high"
                  ? "text-red-600 font-bold"
                  : item.status === "low"
                  ? "text-blue-600 font-bold"
                  : ""
              }`}
            >
              {item.result}
            </div>
            <div className="p-2 border-r border-[rgba(55,53,47,0.16)] text-center text-[#37352F]/60 text-sm">
              {item.unit}
            </div>
            <div className="p-2 border-r border-[rgba(55,53,47,0.16)] text-center text-[#37352F]/60 text-sm">
              {item.ref}
            </div>
            <div className="p-2 flex justify-center items-center">
              {item.status === "high" && (
                <Badge
                  variant="destructive"
                  className="h-10 px-3 text-sm bg-red-500 hover:bg-red-600"
                >
                  HIGH
                </Badge>
              )}
              {item.status === "low" && (
                <Badge
                  variant="outline"
                  className="h-10 px-3 text-sm text-blue-600 border-blue-600 bg-blue-50"
                >
                  LOW
                </Badge>
              )}
              {item.status === "normal" && (
                <CheckCircle className="size-5 text-green-500/50" />
              )}
            </div>
          </div>
        ))}
      </div>
    </div>
  );
});
