// React/Framework
import React, { useState } from "react";

// External
import { Search, History, Plus } from "lucide-react";

// Internal
import { Button } from "../../../components/ui/button";
import { Input } from "../../../components/ui/input";
import { Card } from "../../../components/ui/card";
import { Badge } from "../../../components/ui/badge";
import { ScrollArea } from "../../../components/ui/scroll-area";

interface HistoryItem {
  id: number;
  date: string;
  author: string;
  type: string;
  title: string;
  content: string;
}

interface InterviewHistoryProps {
  historyItems: HistoryItem[];
}

export const InterviewHistory = React.memo(function InterviewHistory({
  historyItems,
}: InterviewHistoryProps) {
  const [searchTerm, setSearchTerm] = useState("");

  const filteredItems = historyItems.filter(item => 
    item.title.includes(searchTerm) || 
    item.content.includes(searchTerm) ||
    item.type.includes(searchTerm)
  );

  return (
    <Card className="flex-1 flex flex-col min-h-0 border border-[rgba(55,53,47,0.16)] shadow-sm bg-white rounded-md overflow-hidden">
      <div className="p-3 border-b border-[rgba(55,53,47,0.09)] bg-[#F7F6F3] flex items-center justify-between h-12 shrink-0">
        <div className="flex items-center gap-2">
          <History className="size-4 text-[#37352F]" />
          <h3 className="text-sm font-bold text-[#37352F]">過去のカルテ</h3>
        </div>
        <div className="flex items-center gap-2">
          <div className="relative">
            <Search className="absolute left-2.5 top-1/2 -translate-y-1/2 size-4 text-muted-foreground" />
            <Input
              placeholder="検索..."
              className="h-10 w-48 pl-9 text-sm bg-white border-[rgba(55,53,47,0.16)]"
              value={searchTerm}
              onChange={(e) => setSearchTerm(e.target.value)}
            />
          </div>
        </div>
      </div>

      <ScrollArea className="flex-1 p-0">
        <div className="divide-y divide-[rgba(55,53,47,0.09)]">
          {filteredItems.map((item) => (
            <div
              key={item.id}
              className="p-3 hover:bg-[#F7F6F3]/50 transition-colors group cursor-pointer"
            >
              <div className="flex items-start justify-between mb-1">
                <div className="flex items-center gap-2">
                  <span className="font-mono text-sm font-bold text-[#37352F]">
                    {item.date}
                  </span>
                  <Badge variant="secondary" className="text-sm px-2">
                    {item.type}
                  </Badge>
                </div>
                <span className="text-sm text-muted-foreground">
                  {item.author}
                </span>
              </div>
              <h4 className="text-sm font-bold text-[#37352F] mb-1">
                {item.title}
              </h4>
              <p className="text-sm text-[#37352F]/80 line-clamp-2 leading-snug">
                {item.content}
              </p>
              <div className="mt-2 hidden group-hover:flex justify-end">
                <Button
                  variant="ghost"
                  size="sm"
                  className="h-10 text-sm gap-1 text-blue-600 hover:text-blue-700 hover:bg-blue-50"
                >
                  <Plus className="size-4" />
                  引用
                </Button>
              </div>
            </div>
          ))}
          {filteredItems.length === 0 && (
            <div className="p-4 text-center text-sm text-muted-foreground">
              該当するカルテはありません
            </div>
          )}
        </div>
      </ScrollArea>
    </Card>
  );
});
