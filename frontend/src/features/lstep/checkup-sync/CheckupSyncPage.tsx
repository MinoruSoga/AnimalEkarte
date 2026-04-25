import { useState } from "react";
import { C, STYLE, ICON } from "@/lib/design-tokens";
import { Button } from "@/components/ui/button";
import { Send } from "lucide-react";
import { usePermission } from "@/hooks/use-permission";
import { useGetCheckupSyncPreview } from "../api/get-checkup-sync-preview";
import { useCreateCheckupSync } from "../api/create-checkup-sync";
import type { CheckupSyncParams, CheckupType } from "../api/get-checkup-sync-preview";
import { CheckupSyncFilterForm } from "./CheckupSyncFilterForm";
import { CheckupSyncPreviewTable } from "./CheckupSyncPreviewTable";
import { CheckupSyncConfirmDialog } from "./CheckupSyncConfirmDialog";

function buildDefaultTagName(checkupType: CheckupType): string {
  const now = new Date();
  const yyyy = now.getFullYear();
  const mm = String(now.getMonth() + 1).padStart(2, "0");
  return `checkup_done_${checkupType}_${yyyy}-${mm}`;
}

export function CheckupSyncPage() {
  const [searchParams, setSearchParams] = useState<CheckupSyncParams | null>(null);
  const [selectedOwnerIds, setSelectedOwnerIds] = useState<Set<string>>(new Set());
  const [confirmDialogOpen, setConfirmDialogOpen] = useState(false);
  const [tagName, setTagName] = useState("");

  const { canCreate } = usePermission("hospital-settings");
  const { data: previewData, isFetching } = useGetCheckupSyncPreview(searchParams);
  const { mutate: createCheckupSync, isPending } = useCreateCheckupSync();

  function handleSearch(params: CheckupSyncParams) {
    setSearchParams(params);
    setSelectedOwnerIds(new Set());
  }

  function handleOpenConfirm() {
    if (!searchParams) return;
    setTagName(buildDefaultTagName(searchParams.checkup_type));
    setConfirmDialogOpen(true);
  }

  function handleConfirm() {
    if (!searchParams) return;
    createCheckupSync(
      {
        checkup_type: searchParams.checkup_type,
        owner_ids: Array.from(selectedOwnerIds),
        tag_name: tagName,
      },
      {
        onSuccess: () => {
          setConfirmDialogOpen(false);
          setSelectedOwnerIds(new Set());
        },
      }
    );
  }

  return (
    <div className={`${STYLE.pageContent} max-w-5xl mx-auto`}>
      {/* ページヘッダー */}
      <div className="mb-6">
        <h1 className={`text-2xl font-bold ${C.text} leading-tight`}>
          健診リマインダー抽出
        </h1>
        <p className={`mt-1.5 text-sm ${C.text60}`}>
          検診の種類・動物種・最終来院日の条件で対象者を絞り込み、LステップタグでリマインダーをまとめてLINE送信します。
        </p>
      </div>

      {/* 抽出条件フォーム */}
      <CheckupSyncFilterForm onSearch={handleSearch} isLoading={isFetching} />

      {/* プレビューテーブル */}
      {searchParams !== null ? (
        <div className="mt-6 space-y-4">
          {isFetching ? (
            <div className={`text-center py-10 ${C.text50} text-sm`}>
              対象者を検索中...
            </div>
          ) : null}

          {!isFetching && previewData ? (
            <CheckupSyncPreviewTable
              owners={previewData.owners}
              selectedIds={selectedOwnerIds}
              onSelectionChange={setSelectedOwnerIds}
              lineLinkedCount={previewData.line_linked_count}
              totalCount={previewData.total_count}
            />
          ) : null}

          {/* 送信アクション */}
          {!isFetching && previewData && canCreate ? (
            <div className={`flex items-center justify-between py-3 border-t ${C.borderLight}`}>
              <p className={`text-sm ${C.text60}`}>
                {selectedOwnerIds.size > 0 ? (
                  <>
                    <span className={`font-semibold ${C.text}`}>
                      {selectedOwnerIds.size}名
                    </span>
                    を選択中
                  </>
                ) : (
                  "LINE連携済みの対象者を選択してください"
                )}
              </p>
              <Button
                type="button"
                onClick={handleOpenConfirm}
                disabled={selectedOwnerIds.size === 0}
                className={STYLE.btnPrimary}
              >
                <Send className={ICON.sm} />
                一括送信する
              </Button>
            </div>
          ) : null}
        </div>
      ) : null}

      {/* 確認ダイアログ */}
      {searchParams !== null ? (
        <CheckupSyncConfirmDialog
          open={confirmDialogOpen}
          onOpenChange={setConfirmDialogOpen}
          checkupType={searchParams.checkup_type}
          selectedCount={selectedOwnerIds.size}
          tagName={tagName}
          onTagNameChange={setTagName}
          onConfirm={handleConfirm}
          isPending={isPending}
        />
      ) : null}
    </div>
  );
}
