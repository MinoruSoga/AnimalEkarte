// React/Framework
import { useDeferredValue, useMemo, useState } from "react";

// External
import { ClipboardCheck } from "lucide-react";

// Internal
import { PageLayout } from "@/components/shared/PageLayout/PageLayout";
import { formatDate } from "@/utils/format/date";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import { useGetCheckups } from "../api/get-checkups";

export function CheckupsList() {
  const [searchTerm, setSearchTerm] = useState("");
  const deferredSearch = useDeferredValue(searchTerm);

  const { data: checkups = [], isLoading, error } = useGetCheckups();

  const filtered = useMemo(() => {
    if (!deferredSearch) return checkups;
    const q = deferredSearch.toLowerCase();
    return checkups.filter(
      (c) =>
        c.petName.toLowerCase().includes(q) ||
        c.ownerName.toLowerCase().includes(q) ||
        c.checkupTypeName.toLowerCase().includes(q) ||
        c.result.toLowerCase().includes(q),
    );
  }, [checkups, deferredSearch]);

  return (
    <PageLayout title="定期健診">
      <div className="space-y-4">
        {/* 検索バー */}
        <div className="flex items-center gap-2">
          <input
            type="text"
            placeholder="ペット名・飼主名・種別で検索..."
            value={searchTerm}
            onChange={(e) => setSearchTerm(e.target.value)}
            className="h-9 w-72 rounded-md border border-input bg-background px-3 py-1 text-sm shadow-sm placeholder:text-muted-foreground focus:outline-none focus:ring-1 focus:ring-ring"
          />
          <span className="text-sm text-muted-foreground">{filtered.length}件</span>
        </div>

        {/* テーブル */}
        <div className="rounded-lg border bg-card shadow-sm overflow-hidden">
          <Table>
            <TableHeader>
              <TableRow className="bg-muted/50 hover:bg-muted/50">
                <TableHead className="whitespace-nowrap">飼主名</TableHead>
                <TableHead className="whitespace-nowrap">ペット名</TableHead>
                <TableHead className="whitespace-nowrap">健診種別</TableHead>
                <TableHead className="whitespace-nowrap">実施日</TableHead>
                <TableHead className="whitespace-nowrap">次回予定</TableHead>
                <TableHead className="whitespace-nowrap">結果・所見</TableHead>
                <TableHead className="whitespace-nowrap">担当医</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {isLoading ? (
                <TableRow>
                  <TableCell colSpan={7} className="h-24 text-center text-sm text-muted-foreground">
                    読み込み中...
                  </TableCell>
                </TableRow>
              ) : error ? (
                <TableRow>
                  <TableCell colSpan={7} className="h-24 text-center text-sm text-destructive">
                    データの取得に失敗しました
                  </TableCell>
                </TableRow>
              ) : filtered.length === 0 ? (
                <TableRow>
                  <TableCell colSpan={7} className="h-24 text-center">
                    <div className="flex flex-col items-center gap-2 text-muted-foreground">
                      <ClipboardCheck className="size-8 opacity-40" />
                      <span className="text-sm">定期健診の記録がありません</span>
                    </div>
                  </TableCell>
                </TableRow>
              ) : (
                filtered.map((c) => (
                  <TableRow key={c.id} className="hover:bg-muted/30">
                    <TableCell className="whitespace-nowrap text-sm">{c.ownerName || "-"}</TableCell>
                    <TableCell className="whitespace-nowrap text-sm font-medium">{c.petName || "-"}</TableCell>
                    <TableCell className="whitespace-nowrap text-sm">{c.checkupTypeName || "-"}</TableCell>
                    <TableCell className="whitespace-nowrap font-mono text-sm">{formatDate(c.date)}</TableCell>
                    <TableCell className="whitespace-nowrap font-mono text-sm">
                      {c.nextDate ? formatDate(c.nextDate) : "-"}
                    </TableCell>
                    <TableCell className="max-w-xs truncate text-sm">{c.result || "-"}</TableCell>
                    <TableCell className="whitespace-nowrap text-sm">{c.doctorName || "-"}</TableCell>
                  </TableRow>
                ))
              )}
            </TableBody>
          </Table>
        </div>
      </div>
    </PageLayout>
  );
}
