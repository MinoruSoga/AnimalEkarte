import React, { useDeferredValue, useState } from "react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { NotionDatePicker } from "@/components/shared/NotionDatePicker/NotionDatePicker";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { C } from "@/lib/design-tokens";

interface HistoryItem {
  id: number;
  name: string;
  date: string;
  next: string;
}

interface VaccinationHistoryProps {
  historyItems: HistoryItem[];
  isLoading?: boolean;
}

export const VaccinationHistory = React.memo(function VaccinationHistory({
  historyItems,
  isLoading = false,
}: VaccinationHistoryProps) {
  const [filterStartDate, setFilterStartDate] = useState("");
  const [filterEndDate, setFilterEndDate] = useState("");
  const [searchTerm, setSearchTerm] = useState("");
  const [sortOrder, setSortOrder] = useState("desc");
  const deferredSearch = useDeferredValue(searchTerm);

  const filteredItems = historyItems
    .filter((item) => {
      const matchesSearch = item.name.includes(deferredSearch);
      // Simplify date filtering for mock
      return matchesSearch;
    })
    .sort((_a, _b) => {
      // Simplify sort for mock
      return sortOrder === "desc" ? 1 : -1;
    });

  return (
    <div className="col-span-6 flex flex-col gap-3">
      <h2 className={`text-sm font-bold ${C.text}`}>予防接種履歴</h2>

      {/* Filters */}
      <div className="space-y-3 bg-white p-3 rounded-lg border border-[rgba(55,53,47,0.16)] shadow-sm">
        <div className="flex flex-col gap-1.5">
          <Label className={`text-sm ${C.text60}`}>実施日</Label>
          <div className="flex items-center gap-2">
            <NotionDatePicker
              value={filterStartDate}
              onChange={setFilterStartDate}
              placeholder="開始日"
              className="flex-1"
            />
            <span className={`${C.text} text-sm`}>〜</span>
            <NotionDatePicker
              value={filterEndDate}
              onChange={setFilterEndDate}
              placeholder="終了日"
              className="flex-1"
            />
          </div>
        </div>
        <div className="flex flex-col gap-1.5">
          <Label className={`text-sm ${C.text60}`}>検索単語</Label>
          <div className="flex gap-2">
            <Input
              value={searchTerm}
              onChange={(e) => setSearchTerm(e.target.value)}
              className="flex-1 bg-white border-[rgba(55,53,47,0.16)] h-10 text-sm"
              placeholder="検索..."
            />
            <Button
              variant="outline"
              className={`h-10 ${C.bgPrimary} text-white ${C.hoverBgPrimaryDark} hover:text-white border-transparent text-sm shadow-sm px-3`}
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
        <div className={`flex items-center border-b border-[rgba(55,53,47,0.16)] bg-[#F7F6F3] text-sm font-bold ${C.text80} h-12 shrink-0`}>
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
          {isLoading ? (
            <div className={`flex items-center justify-center h-24 text-sm ${C.text40}`}>
              読み込み中...
            </div>
          ) : filteredItems.length === 0 ? (
            <div className={`flex items-center justify-center h-24 text-sm ${C.text40}`}>
              接種記録がありません
            </div>
          ) : null}
          {!isLoading && filteredItems.map((item) => (
            <div
              key={item.id}
              className={`flex items-center border-b border-[rgba(55,53,47,0.16)] bg-white text-sm ${C.text} h-12 hover:bg-[#F7F6F3]/50 transition-colors`}
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
                  className={`h-10 w-[50px] text-sm ${C.bgPrimary} text-white ${C.hoverBgPrimaryDark} hover:text-white border-transparent px-0`}
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
