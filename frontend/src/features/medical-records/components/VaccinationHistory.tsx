import React, { useState } from "react";
import { Button } from "../../../components/ui/button";
import { Input } from "../../../components/ui/input";
import { Label } from "../../../components/ui/label";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "../../../components/ui/select";

interface HistoryItem {
  id: number;
  name: string;
  date: string;
  next: string;
}

interface VaccinationHistoryProps {
  historyItems: HistoryItem[];
}

export const VaccinationHistory = React.memo(function VaccinationHistory({
  historyItems,
}: VaccinationHistoryProps) {
  const [filterStartDate, setFilterStartDate] = useState("");
  const [filterEndDate, setFilterEndDate] = useState("");
  const [searchTerm, setSearchTerm] = useState("");
  const [sortOrder, setSortOrder] = useState("desc");

  const filteredItems = historyItems
    .filter((item) => {
      const matchesSearch = item.name.includes(searchTerm);
      // Simplify date filtering for mock
      return matchesSearch;
    })
    .sort((a, b) => {
      // Simplify sort for mock
      return sortOrder === "desc" ? 1 : -1;
    });

  return (
    <div className="col-span-6 flex flex-col gap-3">
      <h2 className="text-sm font-bold text-[#37352F]">予防接種履歴</h2>

      {/* Filters */}
      <div className="space-y-3 bg-white p-3 rounded-lg border border-[rgba(55,53,47,0.16)] shadow-sm">
        <div className="flex flex-col gap-1.5">
          <Label className="text-sm text-[#37352F]/60">実施日</Label>
          <div className="flex items-center gap-2">
            <Input
              type="date"
              value={filterStartDate}
              onChange={(e) => setFilterStartDate(e.target.value)}
              className="bg-white border-[rgba(55,53,47,0.16)] h-10 text-sm flex-1"
            />
            <span className="text-[#37352F] text-sm">〜</span>
            <Input
              type="date"
              value={filterEndDate}
              onChange={(e) => setFilterEndDate(e.target.value)}
              className="bg-white border-[rgba(55,53,47,0.16)] h-10 text-sm flex-1"
            />
          </div>
        </div>
        <div className="flex flex-col gap-1.5">
          <Label className="text-sm text-[#37352F]/60">検索単語</Label>
          <div className="flex gap-2">
            <Input
              value={searchTerm}
              onChange={(e) => setSearchTerm(e.target.value)}
              className="flex-1 bg-white border-[rgba(55,53,47,0.16)] h-10 text-sm"
              placeholder="検索..."
            />
            <Button
              variant="outline"
              className="h-10 bg-[#37352F] text-white hover:bg-[#37352F]/90 hover:text-white border-transparent text-sm shadow-sm px-3"
              onClick={() => setSearchTerm("")}
            >
              クリア
            </Button>
            <Select value={sortOrder} onValueChange={setSortOrder}>
              <SelectTrigger className="w-[80px] h-10 bg-white border-[rgba(55,53,47,0.16)] text-sm">
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="desc">降順</SelectItem>
                <SelectItem value="asc">昇順</SelectItem>
              </SelectContent>
            </Select>
          </div>
        </div>
      </div>

      {/* Table */}
      <div className="border border-[rgba(55,53,47,0.16)] rounded-lg bg-white overflow-hidden flex-1 flex flex-col shadow-sm">
        {/* Header */}
        <div className="flex items-center border-b border-[rgba(55,53,47,0.16)] bg-[#F7F6F3] text-sm font-bold text-[#37352F]/80 h-12 shrink-0">
          <div className="flex-1 px-3 text-center">予防接種名</div>
          <div className="w-[100px] px-2 text-center border-l border-[rgba(55,53,47,0.16)]">
            実施日
          </div>
          <div className="w-[100px] px-2 text-center border-l border-[rgba(55,53,47,0.16)]">
            次予定
          </div>
          <div className="w-[70px] px-2 text-center border-l border-[rgba(55,53,47,0.16)]">
            操作
          </div>
        </div>

        {/* Scrollable Rows */}
        <div className="flex-1 overflow-y-auto">
          {filteredItems.map((item) => (
            <div
              key={item.id}
              className="flex items-center border-b border-[rgba(55,53,47,0.16)] bg-white text-sm text-[#37352F] h-12 hover:bg-[#F7F6F3]/50 transition-colors"
            >
              <div className="flex-1 px-3 truncate font-medium">{item.name}</div>
              <div className="w-[100px] px-2 text-center border-l border-[rgba(55,53,47,0.16)] font-mono text-sm">
                {item.date}
              </div>
              <div className="w-[100px] px-2 text-center border-l border-[rgba(55,53,47,0.16)] font-mono text-sm">
                {item.next}
              </div>
              <div className="w-[70px] px-2 flex justify-center border-l border-[rgba(55,53,47,0.16)]">
                <Button
                  variant="outline"
                  size="sm"
                  className="h-10 w-[50px] text-sm bg-[#37352F] text-white hover:bg-[#37352F]/90 hover:text-white border-transparent px-0"
                >
                  複製
                </Button>
              </div>
            </div>
          ))}
        </div>
      </div>
    </div>
  );
});
