import { useQuery } from "@tanstack/react-query";
import { axios } from "@/lib/axios";
import { HISTORY_FETCH_LIMIT } from "@/config/fetch-limits";
import { queryKeys } from "@/lib/query-keys";
import { QUERY_STALE_TIMES, QUERY_GC_TIMES } from "@/lib/react-query";
import { formatJSTDate, toJSTWallDate } from "@/lib/jst-date";
import type { Vaccination } from "@/types/generated/models";

/**
 * SD-19: 治療履歴（get-pet-treatment-history.ts）と同様、絶対時刻を JST 壁日付に
 * 変換してから表示する。以前はブラウザローカル TZ の getter で整形しており、
 * JST 以外のローカル TZ で閲覧すると日付が 1 日ずれ得た。
 */
function formatDate(iso?: string): string {
  if (!iso) return "-";
  const d = new Date(iso);
  if (isNaN(d.getTime())) return "-";
  const jst = toJSTWallDate(d);
  const yy = String(jst.getFullYear()).slice(2);
  const m = String(jst.getMonth() + 1);
  const day = String(jst.getDate());
  return `${yy}/${m}/${day}`;
}

/** 判定・比較用 YYYY-MM-DD。UTC の split/slice ではなく JST 壁日付に揃える。 */
function toNextDateISO(iso?: string | null): string {
  if (!iso) return "";
  const d = new Date(iso);
  if (isNaN(d.getTime())) return "";
  return formatJSTDate(d);
}

// 履歴 UI 向け transform。`medical-records/api/transforms.ts` の同名関数とは別物。
function transformToHistoryItem(v: Vaccination) {
  return {
    id: v.id,
    name: v.vaccine?.name ?? `ワクチン(ID:${v.vaccine_id})`,
    medicalRecordId: v.medical_record_id != null ? String(v.medical_record_id) : undefined,
    price: v.vaccine?.price ?? null,
    date: formatDate(v.date),
    next: formatDate(v.next_date),
    vaccineId: v.vaccine_id ?? 0,
    lot1: v.lot1 ?? "",
    lot2: v.lot2 ?? "",
    lot3: v.lot3 ?? "",
    lot4: v.lot4 ?? "",
    nextDate: toNextDateISO(v.next_date),
    remarks: v.remarks ?? "",
  };
}
export type PetVaccinationHistoryItem = ReturnType<typeof transformToHistoryItem>;

const getPetVaccinations = async (
  petId: string,
): Promise<PetVaccinationHistoryItem[]> => {
  // BUG-007: always pass page/limit + pet_id. Default BE limit=20 page-window
  // hid newer rows behind future-dated seed data when clients omitted limit.
  const { data } = await axios.get<{ data: Vaccination[] }>("/v1/vaccinations", {
    params: { pet_id: petId, page: 1, limit: HISTORY_FETCH_LIMIT },
  });
  return (data.data ?? []).map(transformToHistoryItem);
};

/**
 * Shared hook for fetching a pet's vaccination history.
 * queryKey ["vaccinations", "pet", petId] は use-vaccinations.ts の
 * useCreateVaccination の invalidateQueries(["vaccinations"]) と prefix 一致し、
 * mutation 成功時にこの query も無効化される。
 */
export const useGetPetVaccinations = (petId?: string) => {
  return useQuery({
    queryKey: queryKeys.vaccinations.byPet(petId!),
    queryFn: () => getPetVaccinations(petId!),
    enabled: !!petId,
    staleTime: QUERY_STALE_TIMES.MEDIUM,
    gcTime: QUERY_GC_TIMES.STANDARD,
  });
};
