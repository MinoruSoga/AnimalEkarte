// React/Framework
import { memo, useDeferredValue, useState, useMemo } from "react";

// External
import { Search, History } from "lucide-react";
import { Link } from "react-router";

// Internal
import { Input } from "@/components/ui/input";
import { Badge } from "@/components/ui/badge";
import { ScrollArea } from "@/components/ui/scroll-area";
import { C, LAYOUT, ICON } from "@/lib/design-tokens";
import { EmptyState } from "@/components/shared/DataStates";
import { normalizedIncludes } from "@/lib/normalize-kana";
import { paths } from "@/config/paths";
import type { InterviewHistoryItem } from "../types";

interface InterviewHistoryProps {
  className?: string;
  historyItems: InterviewHistoryItem[];
}

export const InterviewHistory = memo(function InterviewHistory({
  className,
  historyItems,
}: InterviewHistoryProps) {
  const [searchTerm, setSearchTerm] = useState("");
  const deferredSearch = useDeferredValue(searchTerm);

  // js-cache-function-results: API 由来の filter 結果は useMemo でキャッシュ
  const filteredItems = useMemo(
    () =>
      historyItems.filter(
        (item) =>
          normalizedIncludes(item.title, deferredSearch) ||
          normalizedIncludes(item.content, deferredSearch) ||
          normalizedIncludes(item.type, deferredSearch),
      ),
    [historyItems, deferredSearch],
  );

  return (
    <div
      className={`flex flex-col border ${C.borderMedium} ${C.bgWhite} rounded-md overflow-hidden ${className ?? ""}`}
    >
      <div
        className={`p-3 border-b ${C.borderLight} ${C.bgPage} flex items-center justify-between min-h-12 shrink-0 gap-2`}
      >
        <div className="flex items-center gap-2 min-w-0">
          <History className={`${ICON.action} ${C.text}`} />
          <h3 className={`text-sm font-bold ${C.text}`}>問診抜粋</h3>
          <p className={`text-sm ${C.text60} truncate`}>全文は詳細で確認できます</p>
        </div>
        <div className="flex items-center gap-2">
          <div className="relative">
            <Search
              className={`absolute left-2.5 top-1/2 -translate-y-1/2 ${ICON.action} ${C.text60}`}
            />
            <label htmlFor="medical-record-history-search" className="sr-only">
              過去のカルテを検索
            </label>
            <Input
              id="medical-record-history-search"
              name="medicalRecordHistorySearch"
              placeholder="検索..."
              className={`${LAYOUT.touch.md} w-48 pl-9 text-sm ${C.bgWhite} ${C.borderMedium}`}
              value={searchTerm}
              onChange={(e) => setSearchTerm(e.target.value)}
            />
          </div>
        </div>
      </div>

      <ScrollArea className="flex-1 min-h-0">
        <div className={`divide-y ${C.borderDivider}`}>
          {filteredItems.map((item) => (
            <Link
              key={item.id}
              to={paths.medicalRecords.detail.getHref(item.id)}
              className={`block p-3 transition-colors ${C.hoverBgPageHalf}`}
            >
              <div className="flex items-start justify-between mb-1">
                <div className="flex items-center gap-2">
                  <span className={`font-mono text-sm font-bold ${C.text}`}>{item.date}</span>
                  <Badge variant="secondary" className="text-sm px-2">
                    {item.type}
                  </Badge>
                </div>
                <span className={`text-sm ${C.text60}`}>{item.author}</span>
              </div>
              <h4 className={`text-sm font-bold ${C.text} mb-1`}>{item.title}</h4>
              <p className={`text-sm ${C.text}/80 leading-snug whitespace-pre-wrap line-clamp-2`}>
                {item.content}
              </p>
            </Link>
          ))}
          {filteredItems.length === 0 ? (
            <EmptyState message="該当する抜粋はありません" />
          ) : null}
        </div>
      </ScrollArea>
    </div>
  );
});
