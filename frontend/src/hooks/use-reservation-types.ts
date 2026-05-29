import { useQuery } from "@tanstack/react-query";
import { axios } from "@/lib/axios";
import { QUERY_STALE_TIMES, QUERY_GC_TIMES } from "@/lib/react-query";

// バックエンドの reservation_type_groups レスポンス型
interface ReservationTypeGroupSummary {
  id: number;
  name: string;
  color: string;
}

// バックエンドの reservation_types レスポンス型（group フィールド含む）
interface ReservationTypeRaw {
  id: number;
  name: string;
  color: string;
  is_active: boolean;
  duration_minutes: number;
  sort_order: number;
  is_internal: boolean;
  category: string;
  group_id: number | null;
  group: ReservationTypeGroupSummary | null;
}

// フロントエンド用にグループ化したデータ
interface GroupedReservationTypes {
  label: string;
  types: ReservationTypeRaw[];
}

// バックエンドの staffSummaryResponse 型（id + name のみ）
interface OnDutyStaff {
  id: number;
  name: string;
}

interface ReservationStaffExcludedCourse {
  id: number;
  name?: string;
}

export interface ReservationStaff {
  id: number;
  name: string;
  is_active: boolean;
  excluded_courses: ReservationStaffExcludedCourse[];
}

// GET /v1/masters/reservation-types
const fetchReservationTypesRaw = async (): Promise<ReservationTypeRaw[]> => {
  const { data } = await axios.get<ReservationTypeRaw[]>("/v1/masters/reservation-types");
  return data;
};

// GET /v1/shifts/on-duty-staffs?date=YYYY-MM-DD
const fetchOnDutyStaffs = async (date: string): Promise<OnDutyStaff[]> => {
  const { data } = await axios.get<OnDutyStaff[]>("/v1/shifts/on-duty-staffs", {
    params: { date },
  });
  return data;
};

export const getCurrentClinicId = (): string => {
  try {
    return localStorage.getItem("auth_current_clinic:v1") ?? "";
  } catch {
    return "";
  }
};

const fetchReservationStaffs = async (clinicId: string): Promise<ReservationStaff[]> => {
  const { data } = await axios.get<ReservationStaff[]>(
    `/v1/clinics/${clinicId}/reservation-staffs`,
  );
  return data;
};

/**
 * 予約区分一覧をグループ情報付きで取得する（共有フック）。
 * features/reservations と同一 query key を使用し React Query キャッシュを共有。
 */
export function useGetReservationTypesGrouped() {
  return useQuery({
    queryKey: ["masters", "reservationType", "grouped"],
    queryFn: fetchReservationTypesRaw,
    staleTime: QUERY_STALE_TIMES.STATIC,
    gcTime: QUERY_GC_TIMES.LONG,
    select: (data) => {
      const active = data.filter((t) => t.is_active);
      const map = new Map<string, GroupedReservationTypes>();
      for (const t of active) {
        const key = t.group_id != null ? String(t.group_id) : "__other__";
        const label = t.group?.name ?? "その他";
        if (!map.has(key)) map.set(key, { label, types: [] });
        const entry = map.get(key);
        if (entry) entry.types.push(t);
      }
      return [...map.values()];
    },
  });
}

/**
 * 指定日に出勤しているスタッフ一覧を取得する（共有フック）。
 * features/reservations と同一 query key を使用し React Query キャッシュを共有。
 */
export function useGetOnDutyStaffs(date: string | null) {
  return useQuery({
    queryKey: ["shifts", "on-duty-staffs", date],
    queryFn: () => fetchOnDutyStaffs(date!),
    enabled: date !== null,
    staleTime: QUERY_STALE_TIMES.MEDIUM,
    gcTime: QUERY_GC_TIMES.SHORT,
  });
}

/**
 * 予約スタッフ一覧を取得する。
 * excluded_courses は院内予約フォームの担当者候補から非対応コースを除外するために使う。
 */
export function useGetReservationStaffs() {
  const clinicId = getCurrentClinicId();
  return useQuery({
    queryKey: ["clinics", clinicId, "reservation-staffs"],
    queryFn: () => fetchReservationStaffs(clinicId),
    enabled: clinicId !== "",
    staleTime: QUERY_STALE_TIMES.STATIC,
    gcTime: QUERY_GC_TIMES.LONG,
  });
}
