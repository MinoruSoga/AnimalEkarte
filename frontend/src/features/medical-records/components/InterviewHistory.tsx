// React/Framework
import { memo, useDeferredValue, useState } from "react";

// External
import { Search, History, Plus } from "lucide-react";

// Internal
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Badge } from "@/components/ui/badge";
import { ScrollArea } from "@/components/ui/scroll-area";
import { C, LAYOUT } from "@/lib/design-tokens";
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

  const filteredItems = historyItems.filter(item =>
    item.title.includes(deferredSearch) ||
    item.content.includes(deferredSearch) ||
    item.type.includes(deferredSearch)
  );

  return (
    <div className={`flex flex-col border ${C.borderMedium} bg-white rounded-md overflow-hidden ${className ?? ""}`}>
      <div className={`p-3 border-b ${C.borderLight} ${C.bgPage} flex items-center justify-between h-12 shrink-0`}>
        <div className="flex items-center gap-2">
          <History className={`size-4 ${C.text}`} />
          <h3 className={`text-sm font-bold ${C.text}`}>過去のカルテ</h3>
        </div>
        <div className="flex items-center gap-2">
          <div className="relative">
            <Search className={`absolute left-2.5 top-1/2 -translate-y-1/2 size-4 ${C.text60}`} />
            <Input
              placeholder="検索..."
              className={`${LAYOUT.touch.md} w-48 pl-9 text-sm bg-white ${C.borderMedium}`}
              value={searchTerm}
              onChange={(e) => setSearchTerm(e.target.value)}
            />
          </div>
        </div>
      </div>

      <ScrollArea className="flex-1">
        <div className={`divide-y ${C.borderDivider}`}>
          {filteredItems.map((item) => (
            <div
              key={item.id}
              className={`p-3 ${C.hoverBgPageHalf} transition-colors group cursor-pointer`}
            >
              <div className="flex items-start justify-between mb-1">
                <div className="flex items-center gap-2">
                  <span className={`font-mono text-sm font-bold ${C.text}`}>
                    {item.date}
                  </span>
                  <Badge variant="secondary" className="text-sm px-2">
                    {item.type}
                  </Badge>
                </div>
                <span className={`text-sm ${C.text60}`}>
                  {item.author}
                </span>
              </div>
              <h4 className={`text-sm font-bold ${C.text} mb-1`}>
                {item.title}
              </h4>
              <p className={`text-sm ${C.text}/80 line-clamp-2 leading-snug`}>
                {item.content}
              </p>
              <div className="mt-2 hidden group-hover:flex justify-end">
                <Button
                  variant="ghost"
                  size="sm"
                  className={`${LAYOUT.touch.md} text-sm gap-1 ${C.text60} ${C.hoverText} ${C.hoverBgPage}`}
                >
                  <Plus className="size-4" />
                  引用
                </Button>
              </div>
            </div>
          ))}
          {filteredItems.length === 0 ? (
            <div className={`p-4 text-center text-sm ${C.text60}`}>
              該当するカルテはありません
            </div>
          ) : null}
        </div>
      </ScrollArea>
    </div>
  );
});
