// React/Framework
import { ICON, C } from "@/lib/design-tokens";
import { useCallback, useDeferredValue, useEffect, useMemo, useState } from "react";
import { useNavigate, useSearchParams } from "react-router";

// Components
import { UnpaidTab } from "../components/UnpaidTab";
import { UnifiedTabs, UnifiedTabsContent } from "@/components/shared/UnifiedTabs";
import { ClinicScopeFilter } from "@/components/shared/ClinicScopeFilter/ClinicScopeFilter";

// Hooks
import { useSortableData } from "@/hooks/use-sortable-data";
import { useAuth } from "@/hooks/use-auth";

// External
import { Plus, CreditCard } from "lucide-react";

// Internal
import { PageLayout } from "@/components/shared/PageLayout/PageLayout";
import { PrimaryButton } from "@/components/shared/Form/PrimaryButton";
import { paths } from "@/config/paths";
import { usePagination } from "@/hooks/use-pagination";
import { LoadingFallback, ErrorFallback } from "@/components/shared/DataStates";

// Relative
import { AccountingListTable } from "../components/AccountingListTable";
import { calculateAccountingTotal } from "../components/AccountingListTableModel";
import { DailyAccountingTab } from "../components/DailyAccountingTab";
import { useGetAccountings } from "../api/get-accountings";
import type { AccountingFilters } from "../api/get-accountings";
import { usePermission } from "@/hooks/use-permission";

// Types
import type { Accounting as AccountingType } from "../types";
import type { ActiveFilter } from "@/components/shared/NotionFilter/types";
import { ResourceAccounting } from "@/types/generated/models";

const TABS = [
  { value: "list", label: "会計一覧" },
  { value: "daily", label: "当日会計" },
  { value: "unpaid", label: "未納者一覧" },
] as const;

export function AccountingList() {
  const navigate = useNavigate();
  const [searchParams, setSearchParams] = useSearchParams();
  const { canCreate, canEdit } = usePermission("accounting");

  // #86 段階3: 拠点横断表示
  const { user, currentClinicId } = useAuth();
  const assignedClinics = useMemo(() => user?.clinics ?? [], [user?.clinics]);
  const clinicsParam = searchParams.get("clinics");
  const selectedClinicIds = useMemo(
    () =>
      clinicsParam
        ? clinicsParam.split(",").filter(Boolean)
        : currentClinicId
          ? [currentClinicId]
          : [],
    [clinicsParam, currentClinicId],
  );
  const handleToggleClinic = useCallback((clinicId: string) => {
    setSearchParams((prev) => {
      const next = new URLSearchParams(prev);
      const current =
        next.get("clinics")?.split(",").filter(Boolean) ??
        (currentClinicId ? [currentClinicId] : []);
      const updated = current.includes(clinicId)
        ? current.filter((id) => id !== clinicId)
        : [...current, clinicId];
      if (updated.length === 0) return prev;
      if (updated.length === 1 && updated[0] === currentClinicId) {
        next.delete("clinics");
      } else {
        next.set("clinics", updated.join(","));
      }
      return next;
    });
  }, [setSearchParams, currentClinicId]);

  const tabParam = searchParams.get("tab");
  const activeTab = tabParam === "unpaid" ? "unpaid" : tabParam === "daily" ? "daily" : "list";
  const tabItems = TABS;
  const handleTabChange = useCallback((tab: string) => {
    setSearchParams((prev) => {
      const next = new URLSearchParams(prev);
      if (tab === "unpaid") {
        next.set("tab", "unpaid");
        next.delete("page");
        next.delete("daily_date");
      } else if (tab === "daily") {
        next.set("tab", "daily");
        next.delete("page");
        next.delete("group_by");
        next.delete("reference_date");
      } else {
        next.delete("tab");
        next.delete("page");
        next.delete("group_by");
        next.delete("reference_date");
        next.delete("daily_date");
      }
      return next;
    }, { replace: true });
  }, [setSearchParams]);

  const [searchTerm, setSearchTerm] = useState("");
  const [activeFilters, setActiveFilters] = useState<ActiveFilter[]>([]);
  const deferredSearch = useDeferredValue(searchTerm);
  const isFiltering = searchTerm !== deferredSearch;

  // activeFilters から日付フィルタのみを抽出してAPIに渡す
  // ステータスフィルタは is_not/is_empty/is_not_empty 条件があるためクライアントサイドのまま維持
  const apiFilters = useMemo<AccountingFilters>(() => {
    const dateFilter = activeFilters.find((f) => f.key === "date")?.value as
      | { from?: string; to?: string }
      | undefined;
    return {
      startDate: dateFilter?.from,
      endDate: dateFilter?.to,
      clinicIds: selectedClinicIds.length > 1 ? selectedClinicIds : undefined,
    };
  }, [activeFilters, selectedClinicIds]);

  const { data: accountings = [], isLoading, isError } = useGetAccountings(apiFilters);

  // js-cache-function-results: フィルタ結果を useMemo でキャッシュ
  const filteredRecords = useMemo(() => {
    let result = accountings;

    // ActiveFilter からフィルタ適用（condition 対応）
    const statusFilter = activeFilters.find((f) => f.key === "status");
    if (statusFilter && typeof statusFilter.value === "string") {
      result = result.filter((r) => {
        switch (statusFilter.condition) {
          case "is":
            return r.status === statusFilter.value;
          case "is_not":
            return r.status !== statusFilter.value;
          case "is_empty":
            return !r.status;
          case "is_not_empty":
            return !!r.status;
          default:
            return r.status === statusFilter.value;
        }
      });
    }

    // 支払方法フィルタ（クライアントサイド）
    const paymentMethodFilter = activeFilters.find((f) => f.key === "paymentMethod");
    if (paymentMethodFilter && typeof paymentMethodFilter.value === "string") {
      result = result.filter((r) => {
        const method = r.payment?.method ?? "";
        switch (paymentMethodFilter.condition) {
          case "is":           return method === paymentMethodFilter.value;
          case "is_not":       return method !== paymentMethodFilter.value;
          case "is_empty":     return !method;
          case "is_not_empty": return !!method;
          default:             return method === paymentMethodFilter.value;
        }
      });
    }

    // テキスト検索
    if (deferredSearch) {
      const lowerTerm = deferredSearch.toLowerCase();
      result = result.filter(
        (r) =>
          r.ownerName.toLowerCase().includes(lowerTerm) ||
          r.petName.toLowerCase().includes(lowerTerm),
      );
    }

    return result;
  }, [accountings, activeFilters, deferredSearch]);

  const getSortValue = useCallback((item: AccountingType, key: string) => {
    if (key === "totalAmount") return calculateAccountingTotal(item);
    return String(item[key as keyof AccountingType] ?? "");
  }, []);

  const { activeSorts, setActiveSorts, toggleSort, directionFor, sortedData } =
    useSortableData(filteredRecords, { getSortValue });

  const pagination = usePagination(sortedData, {
    pageSize: 20,
    resetKey: [deferredSearch, JSON.stringify(activeFilters)].join("|"),
  });

  // FE-144: URLクエリパラメータからページ番号を読み取る
  const urlPage = Number(searchParams.get("page") ?? 1);

  // FE-144: URLのページ番号とローカル状態を同期（URLが変わったときのみ）
  // rerender-dependencies: pagination（オブジェクト）を destructure し primitive を deps に使用
  const { totalPages, currentPage, goToPage } = pagination;
  useEffect(() => {
    const clampedPage = Math.max(1, Math.min(urlPage, totalPages));
    if (clampedPage !== currentPage) {
      goToPage(clampedPage);
    }
  // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [urlPage, totalPages]);

  // FE-144: ページ変更時にURLクエリパラメータを更新
  const handlePageChange = useCallback((page: number) => {
    goToPage(page);
    setSearchParams((prev) => {
      const next = new URLSearchParams(prev);
      if (page === 1) {
        next.delete("page");
      } else {
        next.set("page", String(page));
      }
      return next;
    }, { replace: true });
  }, [goToPage, setSearchParams]);

  const handleCreate = useCallback(() => {
    navigate(paths.accounting.selectPet.getHref());
  }, [navigate]);

  const handleEdit = useCallback((id: string) => {
    navigate(paths.accounting.detail.getHref(id));
  }, [navigate]);

  const handleMedicalRecordOpen = useCallback((medicalRecordId: string) => {
    navigate(paths.medicalRecords.detail.getHref(medicalRecordId));
  }, [navigate]);

  return (
    <PageLayout
      title="会計管理"
      resource={ResourceAccounting}
      icon={<CreditCard className={`${ICON.page} ${C.text}`} />}
      headerAction={
        activeTab === "list" && canCreate ? (
          <PrimaryButton onClick={handleCreate}>
            <Plus className={`mr-1.5 ${ICON.action}`} />
            新規会計登録
          </PrimaryButton>
        ) : null
      }
      maxWidth="max-w-full"
    >
      <div className="flex flex-col gap-4">
        {assignedClinics.length > 1 ? (
          <ClinicScopeFilter
            clinics={assignedClinics}
            selectedIds={selectedClinicIds}
            onToggle={handleToggleClinic}
          />
        ) : null}
        <UnifiedTabs
          items={tabItems}
          value={activeTab}
          onValueChange={handleTabChange}
        >
          <UnifiedTabsContent value="daily" className="mt-4">
            <DailyAccountingTab
              selectedClinicIds={selectedClinicIds.length > 1 ? selectedClinicIds : undefined}
            />
          </UnifiedTabsContent>
          <UnifiedTabsContent value="unpaid" className="mt-4">
            <UnpaidTab />
          </UnifiedTabsContent>
          <UnifiedTabsContent value="list" className="mt-4">
            {isLoading ? <LoadingFallback /> : null}
            {isError ? <ErrorFallback /> : null}
            {!isLoading && !isError ? (
              <AccountingListTable
                filteredCount={filteredRecords.length}
                pagination={pagination}
                searchTerm={searchTerm}
                activeFilters={activeFilters}
                activeSorts={activeSorts}
                isFiltering={isFiltering}
                canEdit={canEdit}
                directionFor={directionFor}
                onSearchChange={setSearchTerm}
                onFilterChange={setActiveFilters}
                onSortChange={setActiveSorts}
                onToggleSort={toggleSort}
                onEdit={handleEdit}
                onMedicalRecordOpen={handleMedicalRecordOpen}
                onPageChange={handlePageChange}
              />
            ) : null}
          </UnifiedTabsContent>
        </UnifiedTabs>
      </div>
    </PageLayout>
  );
}
