import { useState, useCallback, useMemo } from "react";
import { RefreshCw, Users } from "lucide-react";
import { useQueryClient } from "@tanstack/react-query";
import { PageLayout } from "@/components/shared/PageLayout/PageLayout";
import { Button } from "@/components/ui/button";
import { C, ICON, STYLE } from "@/lib/design-tokens";
import { usePermission } from "@/hooks/use-permission";
import { ResourceOwners } from "@/types/generated/models";
import { useGetLstepTagSummary } from "../api/get-lstep-tag-summary";
import { TagSummaryTable } from "../components/TagSummaryTable";
import { TagOwnerListDrawer } from "../components/TagOwnerListDrawer";

interface DrawerState {
  open: boolean;
  tagName: string;
  ownerCount: number;
}

function formatRelativeMinutes(asOf: string): string {
  const diffMs = Date.now() - new Date(asOf).getTime();
  const mins = Math.floor(diffMs / 60000);
  if (mins < 1) return "たった今";
  return `${mins}分前`;
}

export function LstepTagManagementPage() {
  const queryClient = useQueryClient();
  // usePermission("owners") の canEdit で一括解除の可否を判定
  const { canEdit } = usePermission(ResourceOwners);

  const { data, isLoading } = useGetLstepTagSummary();

  const [drawerState, setDrawerState] = useState<DrawerState>({
    open: false,
    tagName: "",
    ownerCount: 0,
  });

  const tags = useMemo(() => data?.tags ?? [], [data]);

  const handleRefresh = useCallback(() => {
    queryClient.invalidateQueries({ queryKey: ["lstep-tag-summary"] });
  }, [queryClient]);

  // 対象者一覧を開く（TagSummaryTable の「対象者一覧」ボタン）
  const handleViewOwners = useCallback((tagName: string, ownerCount: number) => {
    setDrawerState({ open: true, tagName, ownerCount });
  }, []);

  // 削除ボタン押下 → TagOwnerListDrawer を開く（canDelete=true でドロワー内に「一括解除」ボタン表示）
  const handleBulkRemove = useCallback((tagName: string, ownerCount: number) => {
    setDrawerState({ open: true, tagName, ownerCount });
  }, []);

  const handleDrawerOpenChange = useCallback((open: boolean) => {
    setDrawerState((prev) => ({ ...prev, open }));
  }, []);

  const relativeTime = data?.as_of ? formatRelativeMinutes(data.as_of) : null;

  return (
    <PageLayout
      title="Lステップタグ管理"
      maxWidth="max-w-full"
      headerAction={
        <div className="flex items-center gap-3">
          {relativeTime !== null ? (
            <span className={`text-sm ${C.text50}`}>
              最終更新: {relativeTime}
            </span>
          ) : null}
          <Button
            variant="outline"
            className={STYLE.btnOutline}
            onClick={handleRefresh}
            disabled={isLoading}
          >
            <RefreshCw className={`mr-1.5 ${ICON.sm} ${isLoading ? "animate-spin" : ""}`} />
            更新
          </Button>
        </div>
      }
    >
      <div className="flex flex-col gap-4 flex-1 min-h-0">
        {/* サマリーカード */}
        {data !== undefined ? (
          <div className={`bg-white border ${C.borderLight} rounded-[4px] px-5 py-4 flex items-center gap-3`}>
            <Users className={`${ICON.lg} ${C.textBrand}`} />
            <div>
              <p className={`text-2xl font-bold ${C.text}`}>
                {data.total_owners_with_lstep.toLocaleString("ja-JP")}名
              </p>
              <p className={`text-sm ${C.text50}`}>がLステップ連携済み</p>
            </div>
          </div>
        ) : null}

        {/* タグサマリーテーブル */}
        <TagSummaryTable
          tags={tags}
          isLoading={isLoading}
          onViewOwners={handleViewOwners}
          onBulkRemove={handleBulkRemove}
          canDelete={canEdit}
        />
      </div>

      {/* 対象者一覧ドロワー（一括解除ボタンはドロワー内に表示） */}
      <TagOwnerListDrawer
        open={drawerState.open}
        onOpenChange={handleDrawerOpenChange}
        tagName={drawerState.tagName}
        ownerCount={drawerState.ownerCount}
        canDelete={canEdit}
      />
    </PageLayout>
  );
}
