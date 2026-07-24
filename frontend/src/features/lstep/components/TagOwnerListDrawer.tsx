import { memo, useState, useCallback, useTransition } from "react";
import { Link } from "react-router";
import { Download, Trash2 } from "lucide-react";
import {
  Sheet,
  SheetContent,
  SheetDescription,
  SheetHeader,
  SheetTitle,
} from "@/components/ui/sheet";
import { Button } from "@/components/ui/button";
import { C, STYLE, ICON } from "@/lib/design-tokens";
import { EmptyState } from "@/components/shared/DataStates";
import { handleApiError } from "@/lib/handle-api-error";
import { isAutoManagedTag } from "@/constants/lstep-auto-tag-prefixes";
import { paths } from "@/config/paths";
import { HISTORY_FETCH_LIMIT } from "@/config/fetch-limits";
import { useGetLstepTagOwners, fetchLstepTagOwnersCsv } from "../api/get-lstep-tag-owners";
import type { LstepTagOwner } from "../api/get-lstep-tag-owners";
import { BulkTagRemoveDialog } from "./BulkTagRemoveDialog";

interface TagOwnerListDrawerProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  tagName: string;
  ownerCount: number;
  canDelete?: boolean;
  canExportCsv?: boolean;
}

function formatDate(dateStr: string | null): string {
  if (!dateStr) return "—";
  return dateStr.slice(0, 10);
}

const LSTEP_CSV_EXPORT_LIMIT = 5_000;

function downloadCsv(blob: Blob, filename: string): void {
  const url = URL.createObjectURL(blob);
  const link = document.createElement("a");
  link.href = url;
  link.download = filename;
  link.click();
  URL.revokeObjectURL(url);
}

const TagOwnerListItem = memo(function TagOwnerListItem({
  owner,
}: {
  owner: LstepTagOwner;
}) {
  return (
    <li
      className={`flex items-center justify-between px-4 py-3 ${C.hoverBgPageHalf} transition-colors`}
    >
      <div className="flex flex-col gap-0.5 min-w-0">
        <span className={`text-sm font-medium ${C.text} truncate`}>
          {owner.owner_name}
        </span>
        <span className={`text-xs ${C.text50}`}>
          最終来院: {formatDate(owner.last_visit_date)}
        </span>
        {owner.reason ? (
          <span className={`text-xs ${C.text50} truncate`}>
            判定理由: {owner.reason}
          </span>
        ) : null}
      </div>
      <Link
        to={paths.owners.detail.getHref(owner.owner_id)}
        className={`inline-flex min-h-11 shrink-0 items-center ml-3 text-xs ${C.textBrand} hover:underline whitespace-nowrap`}
      >
        カルテを開く
      </Link>
    </li>
  );
});

export function TagOwnerListDrawer({
  open,
  onOpenChange,
  tagName,
  ownerCount,
  canDelete = false,
  canExportCsv = false,
}: TagOwnerListDrawerProps) {
  const [removeDialogOpen, setRemoveDialogOpen] = useState(false);
  const [csvLoading, startCsvTransition] = useTransition();

  const { data, isLoading } = useGetLstepTagOwners({
    tag: tagName,
    per_page: HISTORY_FETCH_LIMIT,
  });

  const owners = data?.owners ?? [];
  const isTruncated = typeof data?.total === "number" && data.total > owners.length;
  const isCsvOverLimit = Math.max(ownerCount, data?.total ?? 0) > LSTEP_CSV_EXPORT_LIMIT;
  const canBulkDelete = canDelete && !isAutoManagedTag(tagName) && !isTruncated;

  const handleExportCsv = useCallback(() => {
    if (ownerCount === 0 || !canExportCsv || isCsvOverLimit) return;
    startCsvTransition(async () => {
      try {
        const csv = await fetchLstepTagOwnersCsv(tagName);
        downloadCsv(csv, `lstep-tag-${tagName}-owners.csv`);
      } catch (error) {
        handleApiError(error, "CSV出力");
      }
    });
  }, [ownerCount, tagName, canExportCsv, isCsvOverLimit]);

  const handleRemoveClick = useCallback(() => {
    setRemoveDialogOpen(true);
  }, []);

  return (
    <>
      <Sheet open={open} onOpenChange={onOpenChange}>
        <SheetContent side="right" className="w-full max-w-full sm:max-w-[480px] flex flex-col p-0">
          <SheetHeader className="px-4 py-4 pr-16 border-b shrink-0">
            <SheetTitle className={`${C.text} text-base`}>
              タグ「{tagName}」の対象者一覧
            </SheetTitle>
            <SheetDescription className={C.text50}>
              {ownerCount}名
            </SheetDescription>
          </SheetHeader>

          {/* ツールバー */}
          <div className={`flex items-center gap-2 px-4 py-3 border-b ${C.borderLight} shrink-0`}>
            {canExportCsv ? (
              <Button
                variant="outline"
                className={`min-h-11 px-3 text-sm ${STYLE.btnOutline}`}
                onClick={handleExportCsv}
                disabled={isLoading || csvLoading || ownerCount === 0 || isCsvOverLimit}
              >
                <Download className={`mr-1.5 ${ICON.sm}`} />
                {csvLoading ? "取得中..." : "CSV"}
              </Button>
            ) : null}
            {canBulkDelete ? (
              <Button
                variant="outline"
                className={`min-h-11 px-3 text-sm ${C.danger} border ${C.borderDanger} ${C.hoverBgDanger5}`}
                onClick={handleRemoveClick}
                disabled={isLoading || owners.length === 0}
              >
                <Trash2 className={`mr-1.5 ${ICON.sm}`} />
                一括解除
              </Button>
            ) : null}
          </div>

          {isTruncated || isCsvOverLimit ? (
            <div
              className={`flex flex-col gap-1 border-b px-4 py-2 text-xs ${C.borderLight} ${C.text50}`}
              role="status"
            >
              {isTruncated ? <span>先頭100名を表示しています</span> : null}
              {isCsvOverLimit ? <span>CSV出力は5000名までです</span> : null}
            </div>
          ) : null}

          {/* オーナーリスト */}
          <div className="flex-1 overflow-y-auto">
            {isLoading ? (
              <div className={`py-12 text-center ${C.text50} text-sm`}>読み込み中...</div>
            ) : owners.length === 0 ? (
              <EmptyState message="対象者が見つかりません" />
            ) : (
              <ul className={`divide-y ${C.divideDivider}`}>
                {owners.map((owner) => (
                  <TagOwnerListItem key={owner.owner_id} owner={owner} />
                ))}
              </ul>
            )}
          </div>
        </SheetContent>
      </Sheet>

      {canBulkDelete ? (
        <BulkTagRemoveDialog
          open={removeDialogOpen}
          onOpenChange={setRemoveDialogOpen}
          tagName={tagName}
          ownerCount={owners.length}
          ownerIds={owners.map((o) => o.owner_id)}
        />
      ) : null}
    </>
  );
}
