import { ICON, C, LAYOUT } from "@/lib/design-tokens";
import { LoadingFallback } from "@/components/shared/DataStates";
import { useState, useMemo, useDeferredValue, useCallback, useTransition } from "react";
import { useNavigate } from "react-router";
import { usePermission } from "@/hooks/use-permission";
import { useModalState } from "@/hooks/use-modal-state";
import { usePagination } from "@/hooks/use-pagination";
import { Plus, FileText } from "lucide-react";
import { PageLayout } from "@/components/shared/PageLayout/PageLayout";
import { PrimaryButton } from "@/components/shared/Form/PrimaryButton";
import { ConfirmDialog } from "@/components/shared/ConfirmDialog/ConfirmDialog";
import { paths } from "@/config/paths";
import { useGetEstimates } from "../api/get-estimates";
import { useDeleteEstimate } from "../api/delete-estimate";
import type { ActiveFilter, ActiveSort } from "@/components/shared/PropertyFilter/types";
import { ResourceEstimates } from "@/types/generated/models";
import { EstimateListContent } from "./EstimateListPanels";
import { filterAndSortEstimates } from "./estimate-list-model";

export function EstimateList() {
  const navigate = useNavigate();
  const { canCreate, canEdit, canDelete } = usePermission("estimates");
  const [searchTerm, setSearchTerm] = useState("");
  const [activeFilters, setActiveFilters] = useState<ActiveFilter[]>([]);
  const [activeSorts, setActiveSorts] = useState<ActiveSort[]>([]);
  const deleteModal = useModalState<string>();

  const deferredSearch = useDeferredValue(searchTerm);

  const { data: result, isLoading, isError } = useGetEstimates();
  const { mutate: deleteEstimate } = useDeleteEstimate();
  const [isDeletePending, startDeleteTransition] = useTransition();

  const filtered = useMemo(
    () => filterAndSortEstimates(result?.data ?? [], activeFilters, deferredSearch, activeSorts),
    [result?.data, activeFilters, deferredSearch, activeSorts],
  );

  const pagination = usePagination(filtered, {
    resetKey: deferredSearch,
  });

  const deleteItemId = deleteModal.item;
  const closeDeleteModal = deleteModal.close;
  const handleDeleteConfirm = useCallback(() => {
    if (deleteItemId == null) return;
    startDeleteTransition(() => {
      deleteEstimate(deleteItemId, {
        onSuccess: () => closeDeleteModal(),
      });
    });
  }, [deleteItemId, closeDeleteModal, deleteEstimate]);

  const handleSortChange = useCallback((sorts: ActiveSort[]) => {
    setActiveSorts(sorts);
  }, []);

  const openDeleteModal = deleteModal.open;
  const handleOpenDetail = useCallback((id: string) => {
    navigate(paths.estimates.detail.getHref(id));
  }, [navigate]);
  const handleOpenEdit = useCallback((id: string) => {
    navigate(paths.estimates.edit.getHref(id));
  }, [navigate]);

  if (isLoading) {
    return <LoadingFallback />;
  }
  if (isError) {
    return <div className={`p-4 ${C.danger}`}>データの取得に失敗しました</div>;
  }

  return (
    <PageLayout
      title="見積書管理"
      resource={ResourceEstimates}
      icon={<FileText className={`${ICON.page} ${C.text}`} />}
      headerAction={
        canCreate ? (
          <PrimaryButton colorVariant="primary" onClick={() => navigate(paths.estimates.new.getHref())}>
            <Plus className={`mr-1.5 ${ICON.action}`} />
            新規見積書登録
          </PrimaryButton>
        ) : null
      }
      maxWidth={LAYOUT.pageContentMaxWidth.full}
    >
      <EstimateListContent
        searchTerm={searchTerm}
        onSearchChange={setSearchTerm}
        activeFilters={activeFilters}
        onFilterChange={setActiveFilters}
        activeSorts={activeSorts}
        onSortChange={handleSortChange}
        filteredCount={filtered.length}
        pagination={pagination}
        canEdit={canEdit}
        canDelete={canDelete}
        onOpenDetail={handleOpenDetail}
        onOpenEdit={handleOpenEdit}
        onDelete={openDeleteModal}
      />

      <ConfirmDialog
        open={deleteModal.isOpen}
        onClose={deleteModal.close}
        onConfirm={handleDeleteConfirm}
        title="見積書を削除しますか?"
        description="この操作は取り消せません。"
        confirmLabel="削除"
        variant="destructive"
        isPending={isDeletePending}
      />
    </PageLayout>
  );
}
