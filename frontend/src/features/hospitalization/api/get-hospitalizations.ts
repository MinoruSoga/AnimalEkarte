import { useQuery } from "@tanstack/react-query";
import { axios } from "@/lib/axios";
import { getStoredClinicId } from "@/lib/current-clinic";
import { queryKeys } from "@/lib/query-keys";
import { QUERY_STALE_TIMES, QUERY_GC_TIMES } from "@/lib/react-query";
import { useCurrentClinicId } from "@/hooks/use-current-clinic-id";
import {
  HOSPITALIZATION_LIST_DEFAULT_LIMIT,
  HOSPITALIZATION_LIST_DEFAULT_PAGE,
  type HospitalizationFilterStatus,
  type HospitalizationWireStatus,
  toHospitalizationWireStatus,
} from "../constants";
import { transformHospitalization } from "./transforms";
import type { Hospitalization } from "./transforms";
import type { BackendHospitalization } from "./types";

interface HospitalizationPaginatedResponse {
  data: BackendHospitalization[];
  total: number;
  page: number;
  limit: number;
}

interface HospitalizationSnapshotResponse {
  data: BackendHospitalization[];
  total: number;
  snapshot: boolean;
}

export interface HospitalizationFilters {
  petId?: string;
  startDate?: string; // YYYY-MM-DD（入院開始日の範囲）
  endDate?: string; // YYYY-MM-DD
  /** UI タブ値（active/reserved/discharged/all）。wire 変換は getHospitalizations 内で行う。 */
  statusFilter?: HospitalizationFilterStatus;
  page?: number;
  limit?: number;
}

/** サーバページング封筒を捨てない list 結果（BUG-009）。 */
export interface HospitalizationsResult {
  data: Hospitalization[];
  total: number;
  page: number;
  limit: number;
}

/**
 * Board view の全件フェッチは list limit(20) ではなく 200/page で回す。
 * ケージ occupancy の正確な集計のため全行が必要（PR606 #2）。
 */
const HOSPITALIZATION_SNAPSHOT_MAX_ROWS = 10000;

const buildHospitalizationFilterParams = (
  filters: HospitalizationFilters | undefined,
): Record<string, string | number> => {
  const params: Record<string, string | number> = {};

  if (filters?.petId) params.pet_id = filters.petId;
  if (filters?.startDate) params.start_date = filters.startDate;
  if (filters?.endDate) params.end_date = filters.endDate;

  if (filters?.statusFilter !== undefined) {
    const wireStatus: HospitalizationWireStatus | undefined = toHospitalizationWireStatus(
      filters.statusFilter,
    );
    if (wireStatus !== undefined) {
      params.status = wireStatus;
    }
  }
  return params;
};

const buildHospitalizationParams = (
  filters: HospitalizationFilters | undefined,
  page: number,
  limit: number,
): Record<string, string | number> => ({
  page,
  limit,
  ...buildHospitalizationFilterParams(filters),
});

/**
 * Codex PR606 #10: 複数ページをまたぐ fetch は固定 clinic で発行する。
 * インターセプタは X-Clinic-ID を明示しない限り発行時点の stored clinic を使うため、
 * fetch 中のクリニック切替で page2 以降が別クリニックに混入するのを防ぐ。
 */
const clinicScopeHeaders = (clinicId: string | null): Record<string, string> | undefined => {
  return clinicId ? { "X-Clinic-ID": clinicId } : undefined;
};

const fetchHospitalizationPage = async (
  params: Record<string, string | number>,
  headers?: Record<string, string>,
): Promise<HospitalizationPaginatedResponse> => {
  const { data } = await axios.get<HospitalizationPaginatedResponse>("/v1/hospitalizations", {
    params,
    ...(headers ? { headers } : {}),
  });
  return data;
};

const getHospitalizations = async (
  filters?: HospitalizationFilters,
  scopeClinicId: string | null = getStoredClinicId(),
): Promise<HospitalizationsResult> => {
  const page = filters?.page ?? HOSPITALIZATION_LIST_DEFAULT_PAGE;
  const limit = filters?.limit ?? HOSPITALIZATION_LIST_DEFAULT_LIMIT;
  const data = await fetchHospitalizationPage(
    buildHospitalizationParams(filters, page, limit),
    clinicScopeHeaders(scopeClinicId),
  );
  return {
    data: (data.data ?? []).map(transformHospitalization),
    total: data.total,
    page: data.page,
    limit: data.limit,
  };
};

/**
 * Board view 用に同一フィルタの全ページを取得する。
 * page/limit は無視して内部で 200/page の全ページ走査を行う（PR606 #2）。
 * 途中のクリニック切替混在を防ぐため X-Clinic-ID を 1回目の時点で固定する（#10）。
 */
const getAllHospitalizations = async (
  filters?: Omit<HospitalizationFilters, "page" | "limit">,
  scopeClinicId: string | null = getStoredClinicId(),
): Promise<Hospitalization[]> => {
  const headers = clinicScopeHeaders(scopeClinicId);
  // total 超過分を全ページ取得（getReservations と同契約）。空ページで早期打切り。
  const { data } = await axios.get<HospitalizationSnapshotResponse>("/v1/hospitalizations", {
    params: { ...buildHospitalizationFilterParams(filters), snapshot: true },
    ...(headers ? { headers } : {}),
  });
  if (data?.snapshot !== true) {
    throw new Error("hospitalization snapshot response is missing the snapshot marker");
  }
  if (!Number.isSafeInteger(data.total) || data.total < 0) {
    throw new Error("hospitalization snapshot total is invalid");
  }
  if (!Array.isArray(data.data)) {
    throw new Error("hospitalization snapshot rows are invalid");
  }
  if (data.data.length !== data.total) {
    throw new Error("hospitalization snapshot row count mismatch");
  }
  if (data.total > HOSPITALIZATION_SNAPSHOT_MAX_ROWS) {
    throw new Error("hospitalization snapshot exceeds the maximum row count");
  }
  const ids = new Set<number>();
  for (const row of data.data) {
    if (ids.has(row.id)) {
      throw new Error("hospitalization snapshot contains duplicate rows");
    }
    ids.add(row.id);
  }
  return data.data.map(transformHospitalization);
};

export const useGetHospitalizations = (
  filters?: HospitalizationFilters,
  options?: { enabled?: boolean },
) => {
  // Codex PR606 #10: 切替中の in-flight 結果が新クリニックの key に書き込まれないよう
  // query key に clinic scope を含める。
  const scopeClinicId = useCurrentClinicId();
  return useQuery({
    queryKey: queryKeys.hospitalizations.list({ ...filters, scopeClinicId }),
    queryFn: () => getHospitalizations(filters, scopeClinicId),
    enabled: options?.enabled ?? true,
    staleTime: QUERY_STALE_TIMES.MEDIUM,
    gcTime: QUERY_GC_TIMES.STANDARD,
  });
};

/** Board view 専用: フィルタ一致行を全ページ分返す（#2）。 */
export const useGetAllHospitalizations = (
  filters?: Omit<HospitalizationFilters, "page" | "limit">,
  options?: { enabled?: boolean },
) => {
  const scopeClinicId = useCurrentClinicId();
  return useQuery({
    queryKey: queryKeys.hospitalizations.list({ ...filters, allPages: true, scopeClinicId }),
    queryFn: () => getAllHospitalizations(filters, scopeClinicId),
    enabled: options?.enabled ?? true,
    staleTime: QUERY_STALE_TIMES.MEDIUM,
    gcTime: QUERY_GC_TIMES.STANDARD,
  });
};

/** テスト用に export（request params 契約の固定）。 */
export { getHospitalizations, getAllHospitalizations };
