import { useDeferredValue, useState, memo } from "react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { DatePicker } from "@/components/shared/DatePicker/DatePicker";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { C, STYLE } from "@/lib/design-tokens";
import { normalizedIncludes } from "@/lib/normalize-kana";

// rendering-hoist-jsx: 静的 SelectItem JSX をモジュール定数に巻き上げ
const SORT_ORDER_SELECT_ITEMS = (
  <>
    <SelectItem value="desc">降順</SelectItem>
    <SelectItem value="asc">昇順</SelectItem>
  </>
);

export interface VaccinationHistoryItem {
  id: number;
  name: string;
  date: string;
  next: string;
  vaccineId: number;
  lot1: string;
  lot2: string;
  lot3: string;
  lot4: string;
  nextDate: string;
  remarks: string;
}

interface VaccinationHistoryProps {
  historyItems: VaccinationHistoryItem[];
  isLoading?: boolean;
  onDuplicate?: (item: VaccinationHistoryItem) => void;
  canCreate?: boolean;
}

export const VaccinationHistory = memo(function VaccinationHistory({
  historyItems,
  isLoading = false,
  onDuplicate,
  canCreate = false,
}: VaccinationHistoryProps) {
  const [filterStartDate, setFilterStartDate] = useState("");
  const [filterEndDate, setFilterEndDate] = useState("");
  const [searchTerm, setSearchTerm] = useState("");
  const [sortOrder, setSortOrder] = useState("desc");
  const deferredSearch = useDeferredValue(searchTerm);

  const filteredItems = historyItems
    .filter((item) => {
      const matchesSearch = normalizedIncludes(item.name, deferredSearch);
      // Simplify date filtering for mock
      return matchesSearch;
    })
    .sort((_a, _b) => {
      // Simplify sort for mock
      return sortOrder === "desc" ? 1 : -1;
    });

  return (
    <div className="col-span-1 flex flex-col gap-3 lg:col-span-2">
      <h2 className={`text-base font-semibold ${C.text}`}>過去の接種履歴</h2>

      {/* Filters */}
      <div className={`space-y-3 ${C.bgWhite} p-3 rounded-lg border ${C.borderMedium}`}>
        <div className="flex flex-col gap-1.5">
          <Label className={`text-sm ${C.text60}`}>実施日</Label>
          <div className="flex items-center gap-2">
            <DatePicker
              value={filterStartDate}
              onChange={setFilterStartDate}
              placeholder="開始日"
              className="flex-1"
            />
            <span className={`${C.text} text-sm`}>〜</span>
            <DatePicker
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
              className={`flex-1 ${C.bgWhite} ${C.borderMedium} h-10 text-sm`}
              placeholder="検索..."
            />
            <Button
              variant="outline"
              className={`h-10 ${C.bgWhite} ${C.text} ${C.borderMedium} ${C.hoverBgPage} text-sm px-3`}
              onClick={() => setSearchTerm("")}
            >
              クリア
            </Button>
            <Select value={sortOrder} onValueChange={setSortOrder}>
              <SelectTrigger className={`w-[80px] h-10 ${C.bgWhite} ${C.borderMedium} text-sm`}>
                <SelectValue />
              </SelectTrigger>
              <SelectContent>{SORT_ORDER_SELECT_ITEMS}</SelectContent>
            </Select>
          </div>
        </div>
      </div>

      {/* Table */}
      <div
        className={`border ${C.borderMedium} rounded-lg ${C.bgWhite} overflow-hidden flex-1 flex flex-col`}
      >
        {/* Header — DESIGN.md ex-data-table-cell: canvas-soft 背景 + eyebrow 相当タイポグラフィ（STYLE.sectionLabel） */}
        <div
          className={`flex items-center border-b ${C.borderMedium} ${C.bgPage} ${STYLE.sectionLabel} h-12 shrink-0`}
        >
          <div className="flex-1 px-3 text-center">予防接種名</div>
          <div className={`w-[100px] px-2 text-center border-l ${C.borderMedium}`}>実施日</div>
          <div className={`w-[100px] px-2 text-center border-l ${C.borderMedium}`}>次予定</div>
          <div className={`w-[70px] px-2 text-center border-l ${C.borderMedium}`}>操作</div>
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
          {!isLoading
            ? filteredItems.map((item) => (
                <div
                  key={item.id}
                  className={`flex items-center border-b ${C.borderMedium} ${C.bgWhite} text-sm ${C.text} h-12 ${C.hoverBgPageHalf} transition-colors`}
                >
                  <div className="flex-1 px-3 truncate font-medium">{item.name}</div>
                  <div
                    className={`w-[100px] px-2 text-center border-l ${C.borderMedium} font-mono text-sm`}
                  >
                    {item.date}
                  </div>
                  <div
                    className={`w-[100px] px-2 text-center border-l ${C.borderMedium} font-mono text-sm`}
                  >
                    {item.next}
                  </div>
                  <div className={`w-[70px] px-2 flex justify-center border-l ${C.borderMedium}`}>
                    {canCreate ? (
                      <Button
                        size="sm"
                        className={`h-10 w-[50px] text-sm ${C.bgBrand} ${C.textOnBrand} ${C.hoverBgBrand} ${C.hoverTextOnBrand} rounded-full border-transparent px-0`}
                        onClick={() => onDuplicate?.(item)}
                      >
                        複製
                      </Button>
                    ) : null}
                  </div>
                </div>
              ))
            : null}
        </div>
      </div>
    </div>
  );
});
