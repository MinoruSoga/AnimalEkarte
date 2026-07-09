import { useQuery } from "@tanstack/react-query";
import { axios } from "@/lib/axios";
import { QUERY_STALE_TIMES, QUERY_GC_TIMES } from "@/lib/react-query";
import type { Vaccination } from "@/types/generated/models";

function formatDate(iso?: string): string {
  if (!iso) return "-";
  const d = new Date(iso);
  if (isNaN(d.getTime())) return "-";
  const yy = String(d.getFullYear()).slice(2);
  const m = String(d.getMonth() + 1);
  const day = String(d.getDate());
  return `${yy}/${m}/${day}`;
}

// 履歴 UI 向け transform。`medical-records/api/transforms.ts` の同名関数とは別物。
export function transformToHistoryItem(v: Vaccination) {
  return {
    id: v.id,
    name: v.vaccine?.name ?? `ワクチン(ID:${v.vaccine_id})`,
    date: formatDate(v.date),
    next: formatDate(v.next_date),
    vaccineId: v.vaccine_id ?? 0,
    lot1: v.lot1 ?? "",
    lot2: v.lot2 ?? "",
    lot3: v.lot3 ?? "",
    lot4: v.lot4 ?? "",
    nextDate: v.next_date ? v.next_date.split("T")[0] : "",
    remarks: v.remarks ?? "",
  };
}
export type PetVaccinationHistoryItem = ReturnType<typeof transformToHistoryItem>;

const getPetVaccinations = async (
  petId: string,
): Promise<PetVaccinationHistoryItem[]> => {
  const { data } = await axios.get<{ data: Vaccination[] }>("/v1/vaccinations", {
    params: { pet_id: Number(petId) },
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
    queryKey: ["vaccinations", "pet", petId],
    queryFn: () => getPetVaccinations(petId!),
    enabled: !!petId,
    staleTime: QUERY_STALE_TIMES.MEDIUM,
    gcTime: QUERY_GC_TIMES.STANDARD,
  });
};
