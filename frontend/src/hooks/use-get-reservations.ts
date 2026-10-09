import { useQuery } from "@tanstack/react-query";
import { axios } from "@/lib/axios";
import { getStoredClinicId } from "@/lib/current-clinic";
import { useCurrentClinicId } from "@/hooks/use-current-clinic-id";
import { DAY_VIEW_FETCH_LIMIT } from "@/config/fetch-limits";
import { queryKeys } from "@/lib/query-keys";
import { QUERY_STALE_TIMES, QUERY_GC_TIMES } from "@/lib/react-query";
import { transformReservationWithIntent, type Reservation } from "./reservation-intent-response";
import type { Reservation as BackendReservation } from "@/types/generated/models";

interface ReservationsListResponse {
  data: BackendReservation[];
  total: number;
  page: number;
  limit: number;
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

function buildReservationParams(filters?: ReservationFilters): Record<string, string | number> {
  const params: Record<string, string | number> = { page: 1, limit: DAY_VIEW_FETCH_LIMIT };
  if (filters?.date) params.date = filters.date;
  // 期間レンジ指定時は表示期間の全予約を取得するため limit を引き上げる。
  // (BUG #82: limit=100 + start_time ASC で当日の予約が古い予約に押し出され予約管理に出ない問題の修正)
  if (filters?.startDate && filters?.endDate) {
    params.start_date = filters.startDate;
    params.end_date = filters.endDate;
    params.limit = 1000;
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
  // Codex PR606 #10: 複数ページにまたがる fetch は固定 clinic で発行する。
  // 発行時点の stored clinic をピン留めしないと、切替中に page2 以降が
  // 別クリニックへ飛んで混在データを返す（interceptor は発行毎に読む）。
  const pinnedClinicId = filters?.clinicId ?? getStoredClinicId();
  const requestConfig = {
    params,
    ...(pinnedClinicId ? { headers: { "X-Clinic-ID": pinnedClinicId } } : {}),
  };
  const { data: first } = await axios.get<ReservationsListResponse>(
    "/v1/reservations",
    requestConfig,
  );
  let rows = first.data;
  // EMR-308/Codex: レンジ取得は limit=1000 で打ち切られ、期限超過 tentative の
  // actionable list が total を超えると黙って欠落する。total まで全ページを取得する。
  let page = 1;
  while (rows.length < first.total) {
    page += 1;
    const { data: next } = await axios.get<ReservationsListResponse>("/v1/reservations", {
      ...requestConfig,
      params: { ...params, page },
    });
    if (next.data.length === 0) break; // total 不整合・並行削除への防御
    rows = rows.concat(next.data);
  }
  return rows.map(transformReservationWithIntent);
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
    queryFn: () => getReservations(filters),
    enabled: filters?.enabled ?? true,
    staleTime: QUERY_STALE_TIMES.REALTIME, // 予約一覧は高頻度変更
    gcTime: QUERY_GC_TIMES.SHORT,
  });
};
