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
import { useGetLstepTagOwners, fetchAllLstepTagOwners } from "../api/get-lstep-tag-owners";
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

function buildOwnersCsv(
  owners: Array<{ owner_id: string; owner_name: string; last_visit_date: string | null }>,
  tagName: string
): string {
  const header = "owner_id,owner_name,tag_name,last_visit_date";
  const rows = owners.map(
    (o) =>
      `${o.owner_id},"${o.owner_name.replace(/"/g, '""')}","${tagName}",${o.last_visit_date ?? ""}`
  );
  return [header, ...rows].join("\n");
}

function downloadCsv(content: string, filename: string): void {
  const bom = "﻿";
  const blob = new Blob([bom + content], { type: "text/csv;charset=utf-8;" });
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
    per_page: 200,
  });

  const owners = data?.owners ?? [];
  const canBulkDelete = canDelete && !isAutoManagedTag(tagName);

  const handleExportCsv = useCallback(() => {
    if (ownerCount === 0 || !canExportCsv) return;
    startCsvTransition(async () => {
      try {
        const allOwners = await fetchAllLstepTagOwners(tagName, ownerCount);
        const csv = buildOwnersCsv(allOwners, tagName);
        downloadCsv(csv, `lstep-tag-${tagName}-owners.csv`);
      } catch (error) {
        handleApiError(error, "CSV出力");
      }
    });
  }, [ownerCount, tagName, canExportCsv]);

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
                disabled={isLoading || csvLoading || ownerCount === 0}
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
