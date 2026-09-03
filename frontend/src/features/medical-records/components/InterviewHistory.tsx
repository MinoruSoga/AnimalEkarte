// React/Framework
import { memo, useDeferredValue, useState, useCallback, useMemo } from "react";

// External
import { Search, History, Plus, ChevronDown, ChevronUp } from "lucide-react";

// Internal
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Badge } from "@/components/ui/badge";
import { ScrollArea } from "@/components/ui/scroll-area";
import { C, LAYOUT, ICON } from "@/lib/design-tokens";
import { EmptyState } from "@/components/shared/DataStates";
import { normalizedIncludes } from "@/lib/normalize-kana";
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
  // BUG-MEDI-011: 展開中のカルテIDを管理
  const [expandedId, setExpandedId] = useState<string | null>(null);

  const toggleExpand = useCallback((id: string) => {
    setExpandedId((prev) => (prev === id ? null : id));
  }, []);

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
        className={`p-3 border-b ${C.borderLight} ${C.bgPage} flex items-center justify-between h-12 shrink-0`}
      >
        <div className="flex items-center gap-2">
          <History className={`${ICON.action} ${C.text}`} />
          <h3 className={`text-sm font-bold ${C.text}`}>過去のカルテ</h3>
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
          {filteredItems.map((item) => {
            const isExpanded = expandedId === item.id;
            return (
              <div
                key={item.id}
                role="button"
                tabIndex={0}
                aria-expanded={isExpanded}
                className={`p-3 transition-colors group cursor-pointer ${isExpanded ? C.bgPage30 : C.hoverBgPageHalf}`}
                onClick={() => toggleExpand(item.id)}
                onKeyDown={(e) => {
                  if (e.target !== e.currentTarget) return;
                  if (e.key === "Enter" || e.key === " ") {
                    e.preventDefault();
                    toggleExpand(item.id);
                  }
                }}
              >
                <div className="flex items-start justify-between mb-1">
                  <div className="flex items-center gap-2">
                    <span className={`font-mono text-sm font-bold ${C.text}`}>{item.date}</span>
                    <Badge variant="secondary" className="text-sm px-2">
                      {item.type}
                    </Badge>
                  </div>
                  <div className="flex items-center gap-1">
                    <span className={`text-sm ${C.text60}`}>{item.author}</span>
                    {/* BUG-MEDI-011: 展開/折りたたみトグル */}
                    {isExpanded ? (
                      <ChevronUp className={`${ICON.action} ${C.text60}`} />
                    ) : (
                      <ChevronDown className={`${ICON.action} ${C.text60}`} />
                    )}
                  </div>
                </div>
                <h4 className={`text-sm font-bold ${C.text} mb-1`}>{item.title}</h4>
                {/* BUG-MEDI-011: 展開時は全文、折りたたみ時は2行に制限 */}
                <p
                  className={`text-sm ${C.text}/80 leading-snug whitespace-pre-wrap ${isExpanded ? "" : "line-clamp-2"}`}
                >
                  {item.content}
                </p>
                {/* BUG-MEDI-011: 展開時は「折りたたむ」ボタンを表示、未展開時はホバーで「引用」ボタン */}
                <div
                  className={`mt-2 flex justify-end gap-1 ${isExpanded ? "" : "hidden group-hover:flex"}`}
                >
                  <Button
                    variant="ghost"
                    size="sm"
                    className={`${LAYOUT.touch.md} text-sm gap-1 ${C.text60} ${C.hoverText} ${C.hoverBgPage}`}
                    onClick={(e) => {
                      e.stopPropagation();
                    }}
                  >
                    <Plus className={ICON.action} />
                    引用
                  </Button>
                  {isExpanded ? (
                    <Button
                      variant="ghost"
                      size="sm"
                      className={`${LAYOUT.touch.md} text-sm gap-1 ${C.text60} ${C.hoverText} ${C.hoverBgPage}`}
                      onClick={(e) => {
                        e.stopPropagation();
                        toggleExpand(item.id);
                      }}
                    >
                      <ChevronUp className={ICON.action} />
                      折りたたむ
                    </Button>
                  ) : null}
                </div>
              </div>
            );
          })}
          {filteredItems.length === 0 ? <EmptyState message="該当するカルテはありません" /> : null}
        </div>
      </ScrollArea>
    </div>
  );
});
