import { useState, useCallback, useMemo } from "react";
import { Search, Users, Trash2 } from "lucide-react";
import { Input } from "@/components/ui/input";
import { Button } from "@/components/ui/button";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import { C, STYLE, ICON } from "@/lib/design-tokens";
import { isAutoManagedTag } from "@/constants/lstep-auto-tag-prefixes";
import type { LstepTagSummaryItem } from "../api/get-lstep-tag-summary";

type TagCategory = "all" | "auto" | "manual";

interface TagSummaryTableProps {
  tags: LstepTagSummaryItem[];
  isLoading: boolean;
  onViewOwners: (tagName: string, ownerCount: number) => void;
  onBulkRemove: (tagName: string, ownerCount: number) => void;
  canDelete?: boolean;
}

export function TagSummaryTable({
  tags,
  isLoading,
  onViewOwners,
  onBulkRemove,
  canDelete = false,
}: TagSummaryTableProps) {
  const [categoryFilter, setCategoryFilter] = useState<TagCategory>("all");
  const [searchTerm, setSearchTerm] = useState("");

  const handleCategoryChange = useCallback((value: string) => {
    setCategoryFilter(value as TagCategory);
  }, []);

  const handleSearchChange = useCallback(
    (e: React.ChangeEvent<HTMLInputElement>) => {
      setSearchTerm(e.target.value);
    },
    []
  );

  const filteredTags = useMemo(() => {
    let result = tags;

    if (categoryFilter === "auto") {
      result = result.filter((t) => isAutoManagedTag(t.tag_name));
    } else if (categoryFilter === "manual") {
      result = result.filter((t) => !isAutoManagedTag(t.tag_name));
    }

    if (searchTerm) {
      const lower = searchTerm.toLowerCase();
      result = result.filter((t) => t.tag_name.toLowerCase().includes(lower));
    }

    return result;
  }, [tags, categoryFilter, searchTerm]);

  return (
    <div className="flex flex-col gap-3">
      {/* フィルタバー */}
      <div className={`bg-white border ${C.borderLight} rounded-[4px] p-3 flex flex-wrap gap-3 items-center`}>
        <div className="relative min-w-[200px]">
          <Search className={STYLE.searchIcon} />
          <Input
            className={`${STYLE.searchInput} pl-8`}
            placeholder="タグ名を検索..."
            value={searchTerm}
            onChange={handleSearchChange}
          />
        </div>

        <Select value={categoryFilter} onValueChange={handleCategoryChange}>
          <SelectTrigger className={`h-11 w-[160px] ${C.borderMedium} ${C.text} bg-white`}>
            <SelectValue />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="all">すべて</SelectItem>
            <SelectItem value="auto">自動タグ</SelectItem>
            <SelectItem value="manual">手動タグ</SelectItem>
          </SelectContent>
        </Select>

        <span className={`text-sm ${C.text50} ml-auto`}>
          {filteredTags.length}件
        </span>
      </div>

      {/* テーブル */}
      <div className={STYLE.tableContainer}>
        <Table>
          <TableHeader>
            <TableRow className={STYLE.tableHeaderRow}>
              <TableHead className={`${STYLE.tableHeaderCell} px-4`}>タグ名</TableHead>
              <TableHead className={`${STYLE.tableHeaderCell} px-4 w-24 text-right`}>飼い主数</TableHead>
              <TableHead className={`${STYLE.tableHeaderCell} px-4 w-28`}>種別</TableHead>
              <TableHead className={`${STYLE.tableHeaderCell} px-4 w-40 text-right`}>操作</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {isLoading ? (
              <TableRow>
                <TableCell colSpan={4} className={STYLE.tableEmpty}>
                  読み込み中...
                </TableCell>
              </TableRow>
            ) : filteredTags.length === 0 ? (
              <TableRow>
                <TableCell colSpan={4} className={STYLE.tableEmpty}>
                  タグが見つかりません
                </TableCell>
              </TableRow>
            ) : (
              filteredTags.map((tag) => {
                const isAuto = isAutoManagedTag(tag.tag_name);
                return (
                  <TableRow key={tag.tag_name} className={STYLE.tableRow}>
                    <TableCell className={`${STYLE.tableCell} px-4 font-mono`}>
                      {tag.tag_name}
                    </TableCell>
                    <TableCell className={`${STYLE.tableCell} px-4 text-right font-mono`}>
                      {tag.owner_count}
                    </TableCell>
                    <TableCell className={`${STYLE.tableCell} px-4`}>
                      {isAuto ? (
                        <span className={`inline-flex items-center px-2 py-0.5 rounded text-xs font-medium border ${C.bgStatusGray} ${C.textStatusGray} ${C.borderMuted}`}>
                          自動
                        </span>
                      ) : (
                        <span className={`inline-flex items-center px-2 py-0.5 rounded text-xs font-medium border ${C.bgAccentLight} ${C.textAccentDark} ${C.borderAccentBadge}`}>
                          手動
                        </span>
                      )}
                    </TableCell>
                    <TableCell className={`${STYLE.tableCell} px-4`}>
                      <div className="flex items-center justify-end gap-2">
                        <Button
                          variant="outline"
                          className={`h-9 px-3 text-sm ${STYLE.btnOutline}`}
                          onClick={() => onViewOwners(tag.tag_name, tag.owner_count)}
                        >
                          <Users className={`mr-1.5 ${ICON.sm}`} />
                          対象者一覧
                        </Button>
                        {!isAuto && canDelete ? (
                          <Button
                            variant="outline"
                            className={`h-9 px-3 text-sm ${C.danger} border ${C.borderDanger} ${C.hoverBgDanger5}`}
                            onClick={() => onBulkRemove(tag.tag_name, tag.owner_count)}
                          >
                            <Trash2 className={`mr-1.5 ${ICON.sm}`} />
                            削除
                          </Button>
                        ) : null}
                      </div>
                    </TableCell>
                  </TableRow>
                );
              })
            )}
          </TableBody>
        </Table>
      </div>
    </div>
  );
}
