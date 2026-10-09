import { useCallback, useLayoutEffect, useMemo, useRef, useState } from "react";
import { LayoutGrid, List, Link2 } from "lucide-react";
import { toast } from "sonner";
import type {
  ActiveFilter,
  ActiveSort,
  FilterProperty,
} from "@/components/shared/PropertyFilter/types";
import { UnifiedTabs } from "@/components/shared/UnifiedTabs";
import { ToggleGroup, ToggleGroupItem } from "@/components/ui/toggle-group";
import { Button } from "@/components/ui/button";
import { PropertyFilter } from "@/components/shared/PropertyFilter/PropertyFilter";
import { Pagination } from "@/components/shared/Pagination/Pagination";
import { C, ICON, LAYOUT } from "@/lib/design-tokens";
import { resolveCageNameById } from "@/lib/cage-name";
import type { Hospitalization, MasterItem } from "@/types";
import { LoadingFallback, ErrorFallback } from "@/components/shared/DataStates";
import { Plus } from "lucide-react";
import { PageLayout } from "@/components/shared/PageLayout/PageLayout";
import { PrimaryButton } from "@/components/shared/Form/PrimaryButton";
import { ConfirmDialog } from "@/components/shared/ConfirmDialog/ConfirmDialog";
import { ResourceHospitalization, ResourceMasterHospitalization } from "@/types/generated/models";
import { usePermission } from "@/hooks/use-permission";
import { useMergeCages, useUnmergeCages } from "@/hooks/use-cage-merge";
import { useCageOccupancy } from "@/hooks/use-cage-occupancy";
import { HospitalizationBoard } from "../components/HospitalizationBoard";
import { HospitalizationListView } from "../components/HospitalizationListView";
import {
  HOSPITALIZATION_SORT_PROPERTIES,
  HOSPITALIZATION_TAB_ITEMS,
  isValidViewMode,
  type ServerPagePagination,
  type ViewMode,
} from "./hospitalization-list-model";

interface HospitalizationListContentProps {
  statusFilter: string;
  onStatusTabChange: (value: string) => void;
  filterProperties: FilterProperty[];
  activeFilters: ActiveFilter[];
  onFilterChange: (next: ActiveFilter[]) => void;
  searchTerm: string;
  onSearchChange: (value: string) => void;
  serverTotal: number;
  activeSorts: ActiveSort[];
  onSortChange: (sorts: ActiveSort[]) => void;
  viewMode: ViewMode;
  onViewModeChange: (mode: ViewMode) => void;
  cages: MasterItem[];
  boardHospitalizations: Hospitalization[];
  /** Codex PR606 #2: board の全件フェッチ状態（ページングと独立）。 */
  boardLoading: boolean;
  boardError: boolean;
  onNavigateToForm: (id?: string) => void;
  onMovePet: (hospitalizationId: string, targetCageId: string) => void;
  canCreate: boolean;
  canEdit: boolean;
  pagination: ServerPagePagination<Hospitalization>;
  onPageChange: (page: number) => void;
}

function HospitalizationListContent({
  statusFilter,
  onStatusTabChange,
  filterProperties,
  activeFilters,
  onFilterChange,
  searchTerm,
  onSearchChange,
  serverTotal,
  activeSorts,
  onSortChange,
  viewMode,
  onViewModeChange,
  cages,
  boardHospitalizations,
  boardLoading,
  boardError,
  onNavigateToForm,
  onMovePet,
  canCreate,
  canEdit,
  pagination,
  onPageChange,
}: HospitalizationListContentProps) {
  // EMR-306: ケージ連結モード。merge/unmerge は master-hospitalization:edit 権限が必要。
  // unowned な HospitalizationList.tsx を経由しないよう、状態・mutation・権限 wiring は
  // この owned component 内に閉じ込める（board には props で渡すだけ）。
  const { canEdit: canMergeCages } = usePermission(ResourceMasterHospitalization);
  const [mergeMode, setMergeMode] = useState(false);
  const [pendingMerge, setPendingMerge] = useState<{
    draggedCageId: string;
    targetCageId: string;
  } | null>(null);
  const [pendingUnmergeRepId, setPendingUnmergeRepId] = useState<string | null>(null);
  const mergeMutation = useMergeCages();
  const unmergeMutation = useUnmergeCages();

  // 臨床安全境界: 権限は UI 表示だけでなく mutation 直前にも再検査する
  // （master/CageSettings.tsx の canEditRef と同型）。
  const canMergeRef = useRef(canMergeCages === true);
  useLayoutEffect(() => {
    canMergeRef.current = canMergeCages === true;
  }, [canMergeCages]);

  const boardMergeMode = canMergeCages === true && mergeMode;

  // EMR-306 (Grok review MEDIUM-2): merge-mode の占有判定は表示タブ/ページに依存させない。
  // admitted/reserved の cageId 全集合は共有フック useCageOccupancy が取得する
  // （advisory。最終可否は BE）。boardMergeMode 中のみ有効化し、それ以外では
  // 余分な API 呼び出しを発生させない。
  const { occupiedCageIds } = useCageOccupancy({ enabled: boardMergeMode });
  const hasMergeOccupant = useMemo(() => {
    // 非 merge-mode・占有クエリ不明（読込中/エラー/契約外応答）は undefined —
    // board 側は表示中データの占有判定へフォールバックする（advisory であり、
    // 最終可否は BE が持つ）。
    if (!boardMergeMode || occupiedCageIds === undefined) return undefined;
    const occupied = occupiedCageIds;
    return (cageId: string) => occupied.has(cageId);
  }, [boardMergeMode, occupiedCageIds]);

  const handleRequestMerge = useCallback((draggedCageId: string, targetCageId: string) => {
    setPendingMerge({ draggedCageId, targetCageId });
  }, []);

  const handleRequestUnmerge = useCallback((representativeId: string) => {
    setPendingUnmergeRepId(representativeId);
  }, []);

  const cageNameById = useCallback((id: string) => resolveCageNameById(cages, id), [cages]);

  const handleMergeConfirm = useCallback(() => {
    if (canMergeRef.current !== true || pendingMerge === null) return;
    const memberCageId = Number(pendingMerge.draggedCageId);
    if (!Number.isFinite(memberCageId)) return;
    // 連結は「drop 先の target cage が代表」「drop された cage が member」になる。
    mergeMutation.mutate(
      { representativeId: pendingMerge.targetCageId, memberCageIds: [memberCageId] },
      {
        onSuccess: () => {
          setPendingMerge(null);
          toast.success("ケージを連結しました");
        },
      },
    );
  }, [pendingMerge, mergeMutation]);

  const handleUnmergeConfirm = useCallback(() => {
    if (canMergeRef.current !== true || pendingUnmergeRepId === null) return;
    unmergeMutation.mutate(pendingUnmergeRepId, {
      onSuccess: () => {
        setPendingUnmergeRepId(null);
        toast.success("ケージの連結を解除しました");
      },
    });
  }, [pendingUnmergeRepId, unmergeMutation]);

  return (
    <>
      <div className="flex flex-col gap-4">
        <UnifiedTabs
          items={HOSPITALIZATION_TAB_ITEMS}
          value={statusFilter}
          onValueChange={onStatusTabChange}
          className="w-full"
        />

        <div className="flex items-center gap-4">
          <div className="flex-1">
            <PropertyFilter
              properties={filterProperties}
              activeFilters={activeFilters}
              onFilterChange={onFilterChange}
              searchTerm={searchTerm}
              onSearchChange={onSearchChange}
              searchPlaceholder="飼主名、ペット名、入院No..."
              count={serverTotal}
              sortProperties={HOSPITALIZATION_SORT_PROPERTIES}
              activeSorts={activeSorts}
              onSortChange={onSortChange}
            />
          </div>
          {canMergeCages && viewMode === "board" ? (
            <Button
              type="button"
              variant={mergeMode ? "default" : "outline"}
              size="sm"
              aria-pressed={mergeMode}
              aria-label="ケージ連結モード"
              className="h-11 shrink-0 gap-1.5"
              onClick={() => setMergeMode((prev) => !prev)}
            >
              <Link2 className={ICON.action} />
              連結
            </Button>
          ) : null}
          <div
            className={`${C.bgWhite} rounded-sm border ${C.borderMedium} p-1 h-11 flex items-center`}
          >
            <ToggleGroup
              type="single"
              value={viewMode}
              onValueChange={(v) => v && isValidViewMode(v) && onViewModeChange(v)}
            >
              <ToggleGroupItem
                value="board"
                size="sm"
                aria-label="Board View"
                className="-my-1 h-11 min-w-11"
              >
                <LayoutGrid className={ICON.action} />
              </ToggleGroupItem>
              <ToggleGroupItem
                value="list"
                size="sm"
                aria-label="List View"
                className="-my-1 h-11 min-w-11"
              >
                <List className={ICON.action} />
              </ToggleGroupItem>
            </ToggleGroup>
          </div>
        </div>

        {viewMode === "board" ? (
          // Codex PR606 #2: board は全件フェッチ中の部分ページを描画しない。
          // ケージ占有・未割当の欠落が臨床配置図で見えなくなるのを防ぐ。
          boardLoading ? (
            <LoadingFallback />
          ) : boardError ? (
            <ErrorFallback />
          ) : (
            <HospitalizationBoard
              cages={cages}
              hospitalizations={boardHospitalizations}
              onNavigateToForm={onNavigateToForm}
              onMovePet={onMovePet}
              canCreate={canCreate}
              canEdit={canEdit}
              mergeMode={boardMergeMode}
              canMerge={canMergeCages === true}
              hasMergeOccupant={hasMergeOccupant}
              onRequestMerge={handleRequestMerge}
              onRequestUnmerge={handleRequestUnmerge}
            />
          )
        ) : (
          <>
            <HospitalizationListView
              hospitalizations={pagination.paginatedData}
              onNavigate={onNavigateToForm}
              canEdit={canEdit}
            />
            {pagination.totalPages > 1 ? (
              <Pagination
                currentPage={pagination.currentPage}
                totalPages={pagination.totalPages}
                totalCount={pagination.totalCount}
                startIndex={pagination.startIndex}
                endIndex={pagination.endIndex}
                onPageChange={onPageChange}
                onPrev={() => onPageChange(pagination.currentPage - 1)}
                onNext={() => onPageChange(pagination.currentPage + 1)}
              />
            ) : null}
          </>
        )}
      </div>

      <ConfirmDialog
        open={pendingMerge !== null}
        onClose={() => setPendingMerge(null)}
        title="ケージを連結しますか？"
        description={
          pendingMerge !== null
            ? `「${cageNameById(pendingMerge.draggedCageId)}」を「${cageNameById(pendingMerge.targetCageId)}」に連結します。${cageNameById(pendingMerge.targetCageId)}が代表ケージになります。`
            : ""
        }
        confirmLabel="連結する"
        isPending={mergeMutation.isPending}
        onConfirm={handleMergeConfirm}
      />
      <ConfirmDialog
        open={pendingUnmergeRepId !== null}
        onClose={() => setPendingUnmergeRepId(null)}
        title="ケージの連結を解除しますか？"
        description={
          pendingUnmergeRepId !== null
            ? `「${cageNameById(pendingUnmergeRepId)}」の連結を解除します。構成ケージはそれぞれ独立したケージに戻ります。`
            : ""
        }
        confirmLabel="解除"
        variant="destructive"
        isPending={unmergeMutation.isPending}
        onConfirm={handleUnmergeConfirm}
      />
    </>
  );
}

interface HospitalizationListPageViewProps extends HospitalizationListContentProps {
  canCreateHeader: boolean;
  onCreate: () => void;
}

export function HospitalizationListPageView({
  canCreateHeader,
  onCreate,
  ...contentProps
}: HospitalizationListPageViewProps) {
  return (
    <PageLayout
      title="入院・ホテル管理"
      resource={ResourceHospitalization}
      headerAction={
        canCreateHeader ? (
          <PrimaryButton onClick={onCreate}>
            <Plus className={`mr-1.5 ${ICON.action}`} />
            新規入院登録
          </PrimaryButton>
        ) : null
      }
      maxWidth={LAYOUT.pageContentMaxWidth.full}
    >
      <HospitalizationListContent {...contentProps} />
    </PageLayout>
  );
}
