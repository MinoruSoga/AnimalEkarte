import { HISTORY_FETCH_LIMIT } from "@/config/fetch-limits";
import { useGetMedicalRecords } from "@/hooks/use-medical-records";
import { useGetReservations } from "@/hooks/use-get-reservations";
import { usePermission } from "@/hooks/use-permission";
import { useGetPetVaccinations } from "@/hooks/use-pet-vaccinations";
import { useGetPetCheckupResults } from "./use-pet-checkup-results";
import { formatJSTDate, todayJSTISO } from "@/lib/jst-date";
import {
  ResourceCheckups,
  ResourceExaminations,
  ResourceReservations,
  ResourceTrimming,
  ResourceVaccinations,
} from "@/types/generated/models";

import { useGetPetExaminations } from "../api/get-pet-examinations";
import { useGetPetTreatmentHistory } from "../api/get-pet-treatment-history";
import { useGetPetTrimmingHistory } from "../api/get-pet-trimming-history";

function addDaysISO(date: string, days: number): string {
  const instant = new Date(`${date}T00:00:00+09:00`);
  instant.setUTCDate(instant.getUTCDate() + days);
  return formatJSTDate(instant);
}

function useClinicalPermissions() {
  return {
    examination: usePermission(ResourceExaminations),
    vaccination: usePermission(ResourceVaccinations),
    checkup: usePermission(ResourceCheckups),
    trimming: usePermission(ResourceTrimming),
    reservation: usePermission(ResourceReservations),
  };
}

/**
 * EMR-242: 種類別履歴の手動追加読み込み制御。
 * hasMore の判定は各 api hook（useInfiniteQuery）側で total > 累積 raw 行数により行う。
 */
export interface HistoryLoadMoreControl {
  /** 残ページがある場合 true。 */
  hasMore: boolean;
  /** 次ページ取得中。 */
  isLoadingMore: boolean;
  /** ユーザー操作で次ページを読み込む。初回マウント時の自動取得は行わない。 */
  onLoadMore: () => void;
}

/** テストで素のオブジェクトに差し替えられた query でも動くよう optional フィールドで受ける。 */
interface LoadMoreQuerySource {
  hasNextPage?: boolean;
  isFetchingNextPage?: boolean;
  fetchNextPage?: () => Promise<unknown>;
}

function loadMoreControl(query: LoadMoreQuerySource): HistoryLoadMoreControl {
  return {
    hasMore: query.hasNextPage === true,
    isLoadingMore: query.isFetchingNextPage === true,
    onLoadMore: () => {
      void query.fetchNextPage?.();
    },
  };
}

function useClinicalQueries(
  petId: string,
  today: string,
  permissions: ReturnType<typeof useClinicalPermissions>,
) {
  const medicalRecordsQuery = useGetMedicalRecords({
    petId,
    status: "finalized",
    page: 1,
    limit: HISTORY_FETCH_LIMIT,
    sort: "date",
    order: "desc",
  });
  const examinationsQuery = useGetPetExaminations(
    permissions.examination.canView ? petId : undefined,
  );
  const vaccinationsQuery = useGetPetVaccinations(
    permissions.vaccination.canView ? petId : undefined,
  );
  const checkupsQuery = useGetPetCheckupResults(permissions.checkup.canView ? petId : undefined);
  const treatmentsQuery = useGetPetTreatmentHistory(petId, "all");
  const trimmingQuery = useGetPetTrimmingHistory(permissions.trimming.canView ? petId : undefined);
  const reservationsQuery = useGetReservations({
    startDate: today,
    endDate: addDaysISO(today, 365),
    petId,
    enabled: permissions.reservation.canView,
  });

  return {
    medicalRecordsQuery,
    examinationsQuery,
    vaccinationsQuery,
    checkupsQuery,
    treatmentsQuery,
    trimmingQuery,
    reservationsQuery,
  };
}

export function useOwnerClinicalBriefingData(petId: string) {
  const permissions = useClinicalPermissions();
  const today = todayJSTISO();
  const queries = useClinicalQueries(petId, today, permissions);

  return {
    permissions,
    today,
    ...queries,
    // EMR-242: 追加読み込み対象は検査・治療履歴・トリミングの3系統のみ
    // （診療行 medical-records と予防接種行は対象外）。
    historyLoadMore: {
      examinations: loadMoreControl(queries.examinationsQuery),
      treatments: loadMoreControl(queries.treatmentsQuery),
      trimming: loadMoreControl(queries.trimmingQuery),
    },
  };
}

export type OwnerClinicalBriefingData = ReturnType<typeof useOwnerClinicalBriefingData>;
