// React/Framework
import { ICON, C, LAYOUT } from "@/lib/design-tokens";
import { useCallback, useDeferredValue, useEffect, useMemo, useState } from "react";
import { useNavigate, useSearchParams } from "react-router";

// Components
import { UnpaidTab } from "../components/UnpaidTab";
import { UnifiedTabs, UnifiedTabsContent } from "@/components/shared/UnifiedTabs";
import { ClinicScopeFilter } from "@/components/shared/ClinicScopeFilter/ClinicScopeFilter";

// Hooks
import { useSortableData } from "@/hooks/use-sortable-data";
import { useClinicScope } from "@/hooks/use-clinic-scope";

// External
import { Plus, CreditCard, Info } from "lucide-react";

// Internal
import { PageLayout } from "@/components/shared/PageLayout/PageLayout";
import { PrimaryButton } from "@/components/shared/Form/PrimaryButton";
import { paths } from "@/config/paths";
import { LoadingFallback, ErrorFallback } from "@/components/shared/DataStates";

// Relative
import { AccountingListTable } from "../components/AccountingListTable";
import { calculateAccountingTotal } from "../components/accounting-list-table-model";
import { DailyAccountingTab } from "../components/DailyAccountingTab";
import { useGetAccountingsPage } from "../api/get-accountings";
import type { AccountingPageFilters } from "../api/get-accountings";
import { usePermission } from "@/hooks/use-permission";

// Types
import type { Accounting as AccountingType } from "../types";
import type { ActiveFilter } from "@/components/shared/PropertyFilter/types";
import { ResourceAccounting } from "@/types/generated/models";

const TABS = [
  { value: "list", label: "会計一覧" },
  { value: "daily", label: "当日会計" },
  { value: "unpaid", label: "未納者一覧" },
] as const;

const CLINIC_TOGGLE_RESET_PARAMS = ["page"] as const;

// BUG-411: サーバサイドページネーションのページサイズ。旧 usePagination() の既定値(20件)と
// backend parsePagination の既定 limit(20) を踏襲する。
const ACCOUNTING_PAGE_SIZE = 20;

// BUG-411: status / paymentMethod は backend 未対応のため client-only のまま。
// テキスト検索 (search) はサーバサイド化済み — ページ横断で飼主名・ペット名を探す。
const CLIENT_ONLY_FILTER_KEYS = ["status", "paymentMethod"];

// rerender-stable-empty-array: pageResult 未ロード時のフォールバックを毎レンダー新規配列にしない
// （useMemo の deps に使うため参照が安定している必要がある）
const EMPTY_ACCOUNTINGS: AccountingType[] = [];

export function AccountingList() {
  const navigate = useNavigate();
  const [searchParams, setSearchParams] = useSearchParams();
  const { canCreate, canEdit } = usePermission("accounting");

  // #86 段階3: 拠点横断表示
  const {
    assignedClinics,
    selectedClinicIds,
    isMultiClinic,
    clinicNameById,
    handleToggleClinic,
  } = useClinicScope({ resetParamsOnToggle: CLINIC_TOGGLE_RESET_PARAMS });

  const tabParam = searchParams.get("tab");
  const activeTab = tabParam === "unpaid" ? "unpaid" : tabParam === "daily" ? "daily" : "list";
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

  // FE-144: URLクエリパラメータからページ番号を読み取る（BUG-411: サーバフェッチのキーそのものになる）
  const urlPage = Math.max(1, Number(searchParams.get("page") ?? 1) || 1);

  // 日付 + テキスト検索はサーバへ。ステータス・支払方法は backend 未対応のため client-only。
  const apiFilters = useMemo<AccountingPageFilters>(() => {
    const dateFilter = activeFilters.find((f) => f.key === "date")?.value as
      | { from?: string; to?: string }
      | undefined;
    return {
      startDate: dateFilter?.from,
      endDate: dateFilter?.to,
      search: deferredSearch.trim() || undefined,
      clinicIds: isMultiClinic ? selectedClinicIds : undefined,
      page: urlPage,
      limit: ACCOUNTING_PAGE_SIZE,
    };
  }, [activeFilters, deferredSearch, isMultiClinic, selectedClinicIds, urlPage]);

  const { data: pageResult, isLoading, isError } = useGetAccountingsPage(apiFilters);
  const accountings = pageResult?.data ?? EMPTY_ACCOUNTINGS;

  // js-cache-function-results: client-only フィルタ結果を useMemo でキャッシュ
  const filteredRecords = useMemo(() => {
    let result = accountings;

    // ActiveFilter からフィルタ適用（condition 対応）— status は client-only
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

    return result;
  }, [accountings, activeFilters]);

  const getSortValue = useCallback((item: AccountingType, key: string) => {
    if (key === "totalAmount") return calculateAccountingTotal(item);
    return String(item[key as keyof AccountingType] ?? "");
  }, []);

  const { activeSorts, setActiveSorts, toggleSort, directionFor, sortedData } =
    useSortableData(filteredRecords, { getSortValue });

  // client-only フィルタ中のみ、表示件数と Pagination total が乖離しうる。
  const hasPageScopedFilter = activeFilters.some((f) =>
    CLIENT_ONLY_FILTER_KEYS.includes(f.key),
  );

  // BUG-411: usePagination() のクライアントスライスを撤去し、サーバの total/page/limit をそのまま使う。
  // テキスト検索はサーバ total と一致。client-only フィルタ時のみページ内絞込（表示件数のみ乖離しうる）。
  const serverTotal = pageResult?.total ?? 0;
  const serverLimit = pageResult?.limit ?? ACCOUNTING_PAGE_SIZE;
  const serverPage = pageResult?.page ?? urlPage;
  const totalPages = Math.max(1, Math.ceil(serverTotal / serverLimit));
  const displayTotal = hasPageScopedFilter ? filteredRecords.length : serverTotal;
  const pagination = {
    paginatedData: sortedData,
    totalPages,
    totalCount: serverTotal,
    startIndex: serverTotal === 0 ? 0 : (serverPage - 1) * serverLimit + 1,
    endIndex: Math.min(serverPage * serverLimit, serverTotal),
    currentPage: serverPage,
  };

  // FE-144 / BUG-411: URLの page がサーバ total から導いた totalPages を超えている場合はクランプする
  // （日付フィルタ変更等で母集団が縮んだ場合に空ページへ迷い込むのを防ぐ。既存クランプ effect を踏襲）。
  useEffect(() => {
    if (isLoading) return;
    const clampedPage = Math.max(1, Math.min(urlPage, totalPages));
    if (clampedPage !== urlPage) {
      setSearchParams((prev) => {
        const next = new URLSearchParams(prev);
        if (clampedPage === 1) {
          next.delete("page");
        } else {
          next.set("page", String(clampedPage));
        }
        return next;
      }, { replace: true });
    }
  // eslint-disable-next-line react-hooks/exhaustive-deps -- setSearchParams は安定参照。urlPage/totalPages/isLoading の変化時のみ再評価する設計（FE-144踏襲）。
  }, [urlPage, totalPages, isLoading]);

  // FE-144: ページ変更時にURLクエリパラメータを更新
  const handlePageChange = useCallback((page: number) => {
    setSearchParams((prev) => {
      const next = new URLSearchParams(prev);
      if (page === 1) {
        next.delete("page");
      } else {
        next.set("page", String(page));
      }
      return next;
    }, { replace: true });
  }, [setSearchParams]);

  // BUG-411 react-reviewer指摘: 旧 usePagination() の resetKey（検索・フィルタ変更時に page を1へ戻す）
  // を撤去したままだと、ページ3閲覧中にフィルタ変更→表示件数とPaginationのtotalが食い違ったまま
  // page=3に留まる退行になる。検索・フィルタ変更時は常に page を1へ戻す（旧挙動を維持）。
  const handleSearchChange = useCallback((value: string) => {
    setSearchTerm(value);
    setSearchParams((prev) => {
      const next = new URLSearchParams(prev);
      next.delete("page");
      return next;
    }, { replace: true });
  }, [setSearchParams]);

  const handleFilterChange = useCallback((next: ActiveFilter[]) => {
    setActiveFilters(next);
    setSearchParams((prev) => {
      const params = new URLSearchParams(prev);
      params.delete("page");
      return params;
    }, { replace: true });
  }, [setSearchParams]);

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
      maxWidth={LAYOUT.pageContentMaxWidth.full}
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
          items={TABS}
          value={activeTab}
          onValueChange={handleTabChange}
        >
          <UnifiedTabsContent value="daily" className="mt-4">
            <DailyAccountingTab
              selectedClinicIds={isMultiClinic ? selectedClinicIds : undefined}
              clinicNameById={clinicNameById}
            />
          </UnifiedTabsContent>
          <UnifiedTabsContent value="unpaid" className="mt-4">
            <UnpaidTab />
          </UnifiedTabsContent>
          <UnifiedTabsContent value="list" className="mt-4">
            {isLoading ? <LoadingFallback /> : null}
            {isError ? <ErrorFallback /> : null}
            {!isLoading && !isError && hasPageScopedFilter ? (
              <div className={`flex items-start gap-2 rounded-md px-3 py-2 mb-3 text-sm ${C.textWarning}`}>
                <Info className={`${ICON.action} ${C.textWarningIcon} shrink-0 mt-0.5`} />
                <span>
                  ステータス・支払方法での絞り込みは現在表示中のページ内のみが対象です。
                  飼主名・ペット名の検索は全件を対象にサーバで行います。
                </span>
              </div>
            ) : null}
            {!isLoading && !isError ? (
              <AccountingListTable
                filteredCount={displayTotal}
                pagination={pagination}
                searchTerm={searchTerm}
                activeFilters={activeFilters}
                activeSorts={activeSorts}
                isFiltering={isFiltering}
                canEdit={canEdit}
                directionFor={directionFor}
                onSearchChange={handleSearchChange}
                onFilterChange={handleFilterChange}
                onSortChange={setActiveSorts}
                onToggleSort={toggleSort}
                onEdit={handleEdit}
                onMedicalRecordOpen={handleMedicalRecordOpen}
                onPageChange={handlePageChange}
                showClinicColumn={isMultiClinic}
                clinicNameById={clinicNameById}
              />
            ) : null}
          </UnifiedTabsContent>
        </UnifiedTabs>
      </div>
    </PageLayout>
  );
}
