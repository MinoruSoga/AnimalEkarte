import { useState } from "react";
import { C, ICON, LAYOUT } from "@/lib/design-tokens";
import { Button } from "@/components/ui/button";
import { Send } from "lucide-react";
import { CHECKUP_SYNC_OWNER_LIMIT } from "@/constants/lstep-checkup-sync";
import { usePermission } from "@/hooks/use-permission";
import { useGetCheckupSyncPreview } from "../api/get-checkup-sync-preview";
import { useCreateCheckupSync } from "../api/create-checkup-sync";
import type { CheckupSyncResult } from "../api/create-checkup-sync";
import type { CheckupSyncParams, CheckupType } from "../api/get-checkup-sync-preview";
import { CheckupSyncFilterForm } from "../components/CheckupSyncFilterForm";
import { CheckupSyncPreviewTable } from "../components/CheckupSyncPreviewTable";
import { CheckupSyncConfirmDialog } from "../components/CheckupSyncConfirmDialog";
import { todayJSTISO } from "@/lib/jst-date";
import { PageLayout } from "@/components/shared/PageLayout/PageLayout";
import { ResourceHospitalSettings } from "@/types/generated/models";

function buildDefaultTagName(checkupType: CheckupType): string {
  return `checkup_done_${checkupType}_${todayJSTISO().slice(0, 7)}`;
}

export function CheckupSyncPage() {
  const [searchParams, setSearchParams] = useState<CheckupSyncParams | null>(null);
  const [selectedOwnerIds, setSelectedOwnerIds] = useState<Set<string>>(new Set());
  const [confirmDialogOpen, setConfirmDialogOpen] = useState(false);
  const [tagName, setTagName] = useState("");
  const [syncResult, setSyncResult] = useState<CheckupSyncResult | null>(null);

  const { canCreate } = usePermission(ResourceHospitalSettings);
  const { data: previewData, isFetching } = useGetCheckupSyncPreview(searchParams);
  const { mutate: createCheckupSync, isPending } = useCreateCheckupSync();

  function handleSearch(params: CheckupSyncParams) {
    setSearchParams(params);
    setSelectedOwnerIds(new Set());
    setSyncResult(null);
  }

  function handleOpenConfirm() {
    if (!searchParams) return;
    setTagName(buildDefaultTagName(searchParams.checkup_type));
    setConfirmDialogOpen(true);
  }

  function handleConfirm() {
    if (!searchParams || selectedOwnerIds.size > CHECKUP_SYNC_OWNER_LIMIT) return;
    createCheckupSync(
      {
        checkup_type: searchParams.checkup_type,
        owner_ids: Array.from(selectedOwnerIds),
        tag_name: tagName,
      },
      {
        onSuccess: (data) => {
          setSyncResult(data);
          setConfirmDialogOpen(false);
          setSelectedOwnerIds(new Set());
        },
      }
    );
  }

  return (
    <PageLayout
      title="健診リマインダー抽出"
      description="Lステップタグを一括付与して健診リマインダーをLINE送信します。"
      maxWidth={LAYOUT.pageContentMaxWidth.formNarrow}
      resource={ResourceHospitalSettings}
    >
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
              eligibleCount={previewData.eligible_count}
              lineLinkedCount={previewData.line_linked_count}
              optOutCount={previewData.opt_out_count}
              noLivingPetCount={previewData.no_living_pet_count}
              totalCount={previewData.total_count}
            />
          ) : null}

          {/* 実行結果 */}
          {syncResult !== null ? (
            <div className={`rounded-lg border ${C.borderLight} p-4 ${C.bgPage30}`}>
              <p className={`text-sm font-medium ${C.text}`}>Lステップタグの一括付与が完了しました</p>
              <p className={`mt-1 text-sm ${C.text60}`}>
                成功: <span className={`font-semibold ${C.text}`}>{syncResult.success_count}件</span>
                {syncResult.skip_count > 0 ? (
                  <>
                    {'　'}スキップ: <span className={`font-semibold ${C.text50}`}>{syncResult.skip_count}件</span>
                  </>
                ) : null}
                {syncResult.failed_count > 0 ? (
                  <>
                    {'　'}失敗: <span className={`font-semibold ${C.textNotice}`}>{syncResult.failed_count}件</span>
                  </>
                ) : null}
              </p>
            </div>
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
                    を選択中（最大{CHECKUP_SYNC_OWNER_LIMIT}名）
                  </>
                ) : (
                  "LINE連携済みの対象者を選択してください"
                )}
              </p>
              <Button
                type="button"
                onClick={handleOpenConfirm}
                disabled={
                  selectedOwnerIds.size === 0 ||
                  selectedOwnerIds.size > CHECKUP_SYNC_OWNER_LIMIT
                }
                className={`${C.bgBrand} ${C.hoverBgBrand} ${C.hoverTextOnBrand} ${C.textOnBrand} h-11 px-4 text-base rounded-full transition-colors shadow-none border-transparent`}
              >
                <Send className={ICON.sm} />
                タグを一括付与する
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
    </PageLayout>
  );
}
