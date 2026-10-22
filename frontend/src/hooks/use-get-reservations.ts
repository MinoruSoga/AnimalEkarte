import { useQuery } from "@tanstack/react-query";
import { axios } from "@/lib/axios";
import { getStoredClinicId } from "@/lib/current-clinic";
import { useCurrentClinicId } from "@/hooks/use-current-clinic-id";
import { queryKeys } from "@/lib/query-keys";
import { QUERY_STALE_TIMES, QUERY_GC_TIMES } from "@/lib/react-query";
import {
  transformReservationWithIntent,
  type Reservation,
  type ReservationApiResponse,
} from "./reservation-intent-response";

interface ReservationsListResponse {
  data: ReservationApiResponse[];
  total: number;
  snapshot: boolean;
}

export interface ReservationFilters {
  date?: string;
  /** 期間レンジ取得の開始日 (YYYY-MM-DD)。endDate とセットで使う（予約管理カレンダー） */
  startDate?: string;
  /** 期間レンジ取得の終了日 (YYYY-MM-DD、当日を含む)。startDate とセットで使う */
  endDate?: string;
  status?: string;
  source?: string;
  petId?: string;
  ownerId?: string;
  enabled?: boolean;
  /** #86: 拠点横断表示。複数医院IDをカンマ区切りで送信。未指定=現在の医院のみ */
  clinicIds?: string[];
  /** EMR-308/Codex: 拠点横断カルテ等、レコード自身の医院を操作対象にするときの
   *  X-Clinic-ID 上書き（P2-15 vitals の clinicHeaderConfig と同じ先例）。 */
  clinicId?: string;
}

function buildReservationParams(
  filters?: ReservationFilters,
): Record<string, string | number | boolean> {
  const params: Record<string, string | number | boolean> = { snapshot: true };
  if (filters?.date) params.date = filters.date;
  if (filters?.startDate && filters?.endDate) {
    params.start_date = filters.startDate;
    params.end_date = filters.endDate;
  }
  if (filters?.status) params.status = filters.status;
  if (filters?.source) params.source = filters.source;
  if (filters?.petId) params.pet_id = filters.petId;
  if (filters?.ownerId) params.owner_id = filters.ownerId;
  if (filters?.clinicIds && filters.clinicIds.length > 1)
    params.clinic_ids = filters.clinicIds.join(",");
  return params;
}

const getReservations = async (filters?: ReservationFilters): Promise<Reservation[]> => {
  const params = buildReservationParams(filters);
  // 発行時点の医院を固定し、同じ query key の医院スコープを維持する。
  const pinnedClinicId = filters?.clinicId ?? getStoredClinicId();
  const requestConfig = {
    params,
    ...(pinnedClinicId ? { headers: { "X-Clinic-ID": pinnedClinicId } } : {}),
  };
  const { data: first } = await axios.get<ReservationsListResponse>(
    "/v1/reservations",
    requestConfig,
  );
  // A paginated/partial response cannot safely represent the current calendar:
  // offset pages may observe inserts, deletes or rescheduling between requests.
  if (
    first.snapshot !== true ||
    !Number.isSafeInteger(first.total) ||
    first.total < 0 ||
    !Array.isArray(first.data) ||
    first.total !== first.data.length
  ) {
    throw new Error("予約一覧を完全に取得できませんでした。再読み込みしてください");
  }
  return first.data.map(transformReservationWithIntent);
};

/**
 * Shared hook for fetching reservations.
 * R-F2-S13: reservations feature から昇格。medical-records (use-medical-record-form) が
 * reservations feature を直接 import しないための cross-feature 共有。
 * queryKey prefix は reservations feature 側と同一を維持する。
 */
export const useGetReservations = (filters?: ReservationFilters) => {
  const currentClinicId = useCurrentClinicId();
  // Codex PR606 #10: in-flight 結果が別クリニックの key に書き込まれないよう
  // query key に clinic scope を含める（clinicIds 横断取得は行の clinicId で区別済み）。
  const scopeClinicId = filters?.clinicId ?? currentClinicId;
  return useQuery({
    queryKey: queryKeys.reservations.list({ ...filters, scopeClinicId }),
    queryFn: () => getReservations({ ...filters, clinicId: scopeClinicId ?? undefined }),
    enabled: filters?.enabled ?? true,
    staleTime: QUERY_STALE_TIMES.REALTIME, // 予約一覧は高頻度変更
    gcTime: QUERY_GC_TIMES.SHORT,
  });
};
