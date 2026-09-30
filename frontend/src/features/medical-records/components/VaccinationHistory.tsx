import { useDeferredValue, useMemo, useState, memo } from "react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { LoadingFallback } from "@/components/shared/DataStates";
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

/**
 * 表示用文字列を比較可能な YYYY-MM-DD へ正規化する。
 * 実データは use-pet-vaccinations formatDate 出力の "YY/M/D"（2桁年・パディング無し）、
 * その他は "YYYY-MM-DD" を想定。解釈不能なら "" を返す。
 */
function toComparableDate(display: string): string {
  if (/^\d{4}-\d{2}-\d{2}$/.test(display)) return display;
  const short = /^(\d{2})\/(\d{1,2})\/(\d{1,2})$/.exec(display);
  if (short) {
    return `20${short[1]}-${short[2].padStart(2, "0")}-${short[3].padStart(2, "0")}`;
  }
  return "";
}

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

  const filteredItems = useMemo(() => {
    const filtered = historyItems.filter((item) => {
      if (!normalizedIncludes(item.name, deferredSearch)) return false;
      const itemDate = toComparableDate(item.date);
      // 日付フィルタ指定時は解釈不能な表示日付を対象外にする
      if (filterStartDate && (itemDate === "" || itemDate < filterStartDate)) return false;
      if (filterEndDate && (itemDate === "" || itemDate > filterEndDate)) return false;
      return true;
    });
    // sort は安定ソート: 同日同士は渡された順を保つ
    return filtered.sort((a, b) => {
      const da = toComparableDate(a.date);
      const db = toComparableDate(b.date);
      if (da === db) return 0;
      const cmp = da < db ? -1 : 1;
      return sortOrder === "desc" ? -cmp : cmp;
    });
  }, [historyItems, deferredSearch, filterStartDate, filterEndDate, sortOrder]);

  return (
    <div className="col-span-1 flex flex-col gap-3 lg:col-span-2">
      <h2 className={`text-base font-semibold ${C.text}`}>過去の接種履歴</h2>

      {/* Filters */}
      <div className={`space-y-3 ${C.bgWhite} p-3 rounded-lg border ${C.borderMedium}`}>
        <div className="flex flex-col gap-1.5">
          <Label htmlFor="vacc-history-date-start" className={`text-sm ${C.text60}`}>
            実施日
          </Label>
          <div className="flex items-center gap-2">
            <DatePicker
              id="vacc-history-date-start"
              value={filterStartDate}
              onChange={setFilterStartDate}
              placeholder="開始日"
              className="flex-1"
            />
            <span className={`${C.text} text-sm`}>〜</span>
            <Label htmlFor="vacc-history-date-end" className="sr-only">
              実施日（終了）
            </Label>
            <DatePicker
              id="vacc-history-date-end"
              value={filterEndDate}
              onChange={setFilterEndDate}
              placeholder="終了日"
              className="flex-1"
            />
          </div>
        </div>
        <div className="flex flex-col gap-1.5">
          <Label htmlFor="vacc-history-search" className={`text-sm ${C.text60}`}>
            検索単語
          </Label>
          <div className="flex gap-2">
            <Input
              id="vacc-history-search"
              value={searchTerm}
              onChange={(e) => setSearchTerm(e.target.value)}
              className={`flex-1 ${C.bgWhite} ${C.borderMedium} text-sm`}
              placeholder="検索..."
            />
            <Button
              variant="outline"
              className={`${C.bgWhite} ${C.text} ${C.borderMedium} ${C.hoverBgPage} text-sm px-3`}
              onClick={() => setSearchTerm("")}
            >
              クリア
            </Button>
            <Select value={sortOrder} onValueChange={setSortOrder}>
              <SelectTrigger
                aria-label="並び順"
                className={`w-[80px] ${C.bgWhite} ${C.borderMedium} text-sm`}
              >
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
        <div className="flex-1 overflow-y-auto relative pb-20">
          {isLoading ? (
            <LoadingFallback />
          ) : filteredItems.length === 0 ? (
            <div
              role="status"
              className={`flex items-center justify-center h-24 text-sm ${C.text60}`}
            >
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
                        className={`w-[50px] text-sm ${C.bgActionPrimarySolid} ${C.textOnActionPrimary} ${C.hoverBgActionPrimarySolid} ${C.hoverTextOnActionPrimary} ${C.activeBgActionPrimarySolid} ${C.activeTextOnActionPrimary} rounded-md border-transparent px-0`}
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
