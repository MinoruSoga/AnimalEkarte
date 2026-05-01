import { useState, useCallback } from "react";
import { Link } from "react-router";
import { Download, Trash2 } from "lucide-react";
import {
  Sheet,
  SheetContent,
  SheetHeader,
  SheetTitle,
} from "@/components/ui/sheet";
import { Button } from "@/components/ui/button";
import { C, STYLE, ICON } from "@/lib/design-tokens";
import { axios } from "@/lib/axios";
import { handleApiError } from "@/lib/handle-api-error";
import { isAutoManagedTag } from "@/constants/lstep-auto-tag-prefixes";
import { useGetLstepTagOwners } from "../api/get-lstep-tag-owners";
import type { LstepTagOwner, LstepTagOwnersResponse } from "../api/get-lstep-tag-owners";
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

async function fetchAllOwnersForCsv(tagName: string, ownerCount: number): Promise<LstepTagOwner[]> {
  const clinicId = localStorage.getItem("auth_current_clinic:v1");
  const perPage = 100;
  const totalPages = Math.max(1, Math.ceil(ownerCount / perPage));
  const owners: LstepTagOwner[] = [];

  for (let page = 1; page <= totalPages; page += 1) {
    const { data } = await axios.get<LstepTagOwnersResponse>(
      `/v1/clinics/${clinicId}/lstep/owners`,
      { params: { tag: tagName, page, per_page: perPage } }
    );
    owners.push(...data.owners);
    if (owners.length >= data.total || data.owners.length === 0) {
      break;
    }
  }

  return owners;
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

export function TagOwnerListDrawer({
  open,
  onOpenChange,
  tagName,
  ownerCount,
  canDelete = false,
  canExportCsv = false,
}: TagOwnerListDrawerProps) {
  const [removeDialogOpen, setRemoveDialogOpen] = useState(false);
  const [csvLoading, setCsvLoading] = useState(false);

  const { data, isLoading } = useGetLstepTagOwners({
    tag: tagName,
    per_page: 200,
  });

  const owners = data?.owners ?? [];
  const canBulkDelete = canDelete && !isAutoManagedTag(tagName);

  const handleExportCsv = useCallback(async () => {
    if (ownerCount === 0 || !canExportCsv) return;
    setCsvLoading(true);
    try {
      const allOwners = await fetchAllOwnersForCsv(tagName, ownerCount);
      const csv = buildOwnersCsv(allOwners, tagName);
      downloadCsv(csv, `lstep-tag-${tagName}-owners.csv`);
    } catch (error) {
      handleApiError(error, "CSV出力");
    } finally {
      setCsvLoading(false);
    }
  }, [ownerCount, tagName, canExportCsv]);

  const handleRemoveClick = useCallback(() => {
    setRemoveDialogOpen(true);
  }, []);

  return (
    <>
      <Sheet open={open} onOpenChange={onOpenChange}>
        <SheetContent side="right" className="w-[480px] sm:max-w-[480px] flex flex-col p-0">
          <SheetHeader className="px-4 py-4 border-b shrink-0">
            <SheetTitle className={`${C.text} text-base`}>
              タグ「{tagName}」の対象者一覧
            </SheetTitle>
            <p className={`text-sm ${C.text50}`}>{ownerCount}名</p>
          </SheetHeader>

          {/* ツールバー */}
          <div className={`flex items-center gap-2 px-4 py-3 border-b ${C.borderLight} shrink-0`}>
            {canExportCsv ? (
              <Button
                variant="outline"
                className={`h-9 px-3 text-sm ${STYLE.btnOutline}`}
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
                className={`h-9 px-3 text-sm ${C.danger} border ${C.borderDanger} ${C.hoverBgDanger5}`}
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
              <div className={`py-12 text-center ${C.text40} text-sm`}>
                対象者が見つかりません
              </div>
            ) : (
              <ul className="divide-y divide-[rgba(55,53,47,0.09)]">
                {owners.map((owner) => (
                  <li
                    key={owner.owner_id}
                    className={`flex items-center justify-between px-4 py-3 ${C.hoverBgPageHalf} transition-colors`}
                  >
                    <div className="flex flex-col gap-0.5 min-w-0">
                      <span className={`text-sm font-medium ${C.text} truncate`}>
                        {owner.owner_name}
                      </span>
                      <span className={`text-xs ${C.text50}`}>
                        最終来院: {formatDate(owner.last_visit_date)}
                      </span>
                    </div>
                    <Link
                      to={`/owners/${owner.owner_id}`}
                      className={`shrink-0 ml-3 text-xs ${C.accent} hover:underline whitespace-nowrap`}
                    >
                      カルテを開く
                    </Link>
                  </li>
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
