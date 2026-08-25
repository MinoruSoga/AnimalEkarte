import { useQuery } from "@tanstack/react-query";
import { axios } from "@/lib/axios";
import { getStoredClinicId } from "@/lib/current-clinic";
import { queryKeys } from "@/lib/query-keys";
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
  group_id?: number | null;
  group?: ReservationTypeGroupSummary | null;
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

export interface ReservationAvailableTimeSlot {
  start_time: string;
  end_time: string;
}

interface ReservationStaffCourse {
  id: number;
  name?: string;
}

export interface ReservationStaff {
  id: number;
  name: string;
  is_active: boolean;
  /** Affirmative capability surface (TASK-021 Stage B). */
  capable_courses: ReservationStaffCourse[];
}

interface ReservationStaffWire {
  id: number;
  name: string;
  is_active: boolean;
  capable_courses?: ReservationStaffCourse[] | null;
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

export const getCurrentClinicId = (): string | null => {
  return getStoredClinicId();
};

const fetchReservationStaffs = async (clinicId: string): Promise<ReservationStaff[]> => {
  const { data } = await axios.get<ReservationStaffWire[]>(
    `/v1/clinics/${clinicId}/reservation-staffs`,
  );
  // Project only the positive contract so legacy wire fields cannot propagate.
  // Missing capabilities stay fail-closed rather than meaning "all capable".
  return data.map((staff) => ({
    id: staff.id,
    name: staff.name,
    is_active: staff.is_active,
    capable_courses: staff.capable_courses ?? [],
  }));
};

const fetchReservationAvailableTimes = async (
  reservationTypeId: string,
  date: string,
  staffId: string | null,
): Promise<ReservationAvailableTimeSlot[]> => {
  const { data } = await axios.get<ReservationAvailableTimeSlot[]>(
    "/v1/reservations/available-times",
    {
      params: {
        reservation_type_id: reservationTypeId,
        date,
        ...(staffId ? { staff_id: staffId } : {}),
      },
    },
  );
  return data;
};

/**
 * 予約区分一覧をグループ情報付きで取得する（共有フック）。
 * features/reservations と同一 query key を使用し React Query キャッシュを共有。
 * selectedTypeId を渡すと、その ID の無効区分だけを表示用に残す（BUG-015）。
 */
export function useGetReservationTypesGrouped(
  selectedTypeId?: string | number | null,
) {
  const selectedId =
    selectedTypeId === null || selectedTypeId === undefined || selectedTypeId === ""
      ? null
      : String(selectedTypeId);

  return useQuery({
    queryKey: queryKeys.masters.reservationTypesGrouped(),
    queryFn: fetchReservationTypesRaw,
    staleTime: QUERY_STALE_TIMES.STATIC,
    gcTime: QUERY_GC_TIMES.LONG,
    select: (data) => {
      const visible = data.filter(
        (t) => t.is_active || (selectedId !== null && String(t.id) === selectedId),
      );
      const map = new Map<string, GroupedReservationTypes>();
      for (const t of visible) {
        const groupId = t.group_id ?? t.group?.id ?? null;
        const key = groupId != null ? String(groupId) : "__other__";
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
    queryKey: queryKeys.shifts.onDutyStaffs(date!),
    queryFn: () => fetchOnDutyStaffs(date!),
    enabled: date !== null,
    staleTime: QUERY_STALE_TIMES.MEDIUM,
    gcTime: QUERY_GC_TIMES.SHORT,
  });
}

/**
 * 予約スタッフ一覧を取得する。
 * capable_courses は院内予約フォームの担当者候補を肯定形で絞り込む（TASK-021 Stage B）。
 * 欠落・未取得時は fail-closed（候補に載せない）。
 */
export function useGetReservationStaffs() {
  const clinicId = getCurrentClinicId();
  return useQuery({
    queryKey: queryKeys.clinics.reservationStaffs(clinicId!),
    queryFn: () => {
      if (clinicId === null) {
        return Promise.reject(new Error("clinic_id is required"));
      }
      return fetchReservationStaffs(clinicId);
    },
    enabled: clinicId !== null,
    staleTime: QUERY_STALE_TIMES.STATIC,
    gcTime: QUERY_GC_TIMES.LONG,
  });
}

/**
 * 院内予約フォーム用の空き枠一覧を取得する。
 * LIFF と同じ空き枠計算を使い、営業時間・スタッフシフト・既存予約・予約不可時間を反映する。
 */
export function useGetReservationAvailableTimes(
  reservationTypeId: string | null,
  date: string | null,
  staffId: string | null,
) {
  return useQuery({
    queryKey: queryKeys.reservations.availableTimes(reservationTypeId!, date!, staffId ?? undefined),
    queryFn: () => fetchReservationAvailableTimes(reservationTypeId!, date!, staffId),
    enabled: reservationTypeId !== null && date !== null,
    staleTime: QUERY_STALE_TIMES.REALTIME,
    gcTime: QUERY_GC_TIMES.SHORT,
    // BUG-015: inactive historical edits may 400; form keeps values and skips global toast.
    meta: { silentError: true },
  });
}
