import { useState, useCallback, useMemo } from "react";
import { RefreshCw } from "lucide-react";
import { isAxiosError } from "axios";
import { useQueryClient } from "@tanstack/react-query";
import { PageLayout } from "@/components/shared/PageLayout/PageLayout";
import { ErrorFallback } from "@/components/shared/DataStates";
import { Button } from "@/components/ui/button";
import { C, ICON, STYLE, LAYOUT } from "@/lib/design-tokens";
import { queryKeys } from "@/lib/query-keys";
import { useAuth } from "@/hooks/use-auth";
import {
  ResourceLstepAnalytics,
  ResourceOwners,
} from "@/types/generated/models";
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
  const { hasPermission } = useAuth();
  const canViewAnalytics = hasPermission(ResourceLstepAnalytics, "view");
  const canDeleteOwnerTags = hasPermission(ResourceOwners, "delete");

  const { data, isLoading, isError, error } = useGetLstepTagSummary();

  const [drawerState, setDrawerState] = useState<DrawerState>({
    open: false,
    tagName: "",
    ownerCount: 0,
  });

  const tags = useMemo(() => data?.tags ?? [], [data]);

  const handleRefresh = useCallback(() => {
    queryClient.invalidateQueries({ queryKey: queryKeys.lstepTagSummary() });
  }, [queryClient]);

  const handleViewOwners = useCallback((tagName: string, ownerCount: number) => {
    setDrawerState({ open: true, tagName, ownerCount });
  }, []);

  const handleBulkRemove = useCallback((tagName: string, ownerCount: number) => {
    setDrawerState({ open: true, tagName, ownerCount });
  }, []);

  const handleDrawerOpenChange = useCallback((open: boolean) => {
    setDrawerState((prev) => ({ ...prev, open }));
  }, []);

  const relativeTime = data?.as_of ? formatRelativeMinutes(data.as_of) : null;
  const isForbidden = isAxiosError(error) && error.response?.status === 403;

  return (
    <PageLayout
      title="Lステップタグ管理"
      maxWidth={LAYOUT.pageContentMaxWidth.full}
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
        {/* タグサマリーテーブル */}
        {isError ? (
          <ErrorFallback
            message={
              isForbidden
                ? "アクセス権限がありません。Lステップ分析の閲覧権限を確認してください。"
                : "Lステップタグの取得に失敗しました"
            }
          />
        ) : (
          <TagSummaryTable
            tags={tags}
            isLoading={isLoading}
            onViewOwners={handleViewOwners}
            onBulkRemove={handleBulkRemove}
            canDelete={canDeleteOwnerTags}
          />
        )}
      </div>

      {/* 対象者一覧ドロワー */}
      <TagOwnerListDrawer
        open={drawerState.open}
        onOpenChange={handleDrawerOpenChange}
        tagName={drawerState.tagName}
        ownerCount={drawerState.ownerCount}
        canDelete={canDeleteOwnerTags}
        canExportCsv={canViewAnalytics}
      />
    </PageLayout>
  );
}
