import { useQuery } from "@tanstack/react-query";
import { axios } from "@/lib/axios";
import { QUERY_STALE_TIMES, QUERY_GC_TIMES } from "@/lib/react-query";
import type { Vaccination } from "@/types/generated/models";

export interface PetVaccinationHistoryItem {
  id: number;
  name: string;
  date: string;
  next: string;
  // 複製に使う生データ
  vaccineId: number;
  lot1: string;
  lot2: string;
  lot3: string;
  lot4: string;
  nextDate: string;
  remarks: string;
}

function formatDate(iso?: string): string {
  if (!iso) return "-";
  const d = new Date(iso);
  if (isNaN(d.getTime())) return "-";
  const yy = String(d.getFullYear()).slice(2);
  const m = String(d.getMonth() + 1);
  const day = String(d.getDate());
  return `${yy}/${m}/${day}`;
}

function transformToHistoryItem(v: Vaccination): PetVaccinationHistoryItem {
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

const getPetVaccinations = async (
  petId: string,
): Promise<PetVaccinationHistoryItem[]> => {
  const { data } = await axios.get<{ data: Vaccination[] }>("/v1/vaccinations", {
    params: { pet_id: Number(petId) },
  });
  return (data.data ?? []).map(transformToHistoryItem);
};

export const useGetPetVaccinations = (petId?: string) => {
  return useQuery({
    queryKey: ["vaccinations", "pet", petId],
    queryFn: () => getPetVaccinations(petId!),
    enabled: !!petId,
    staleTime: QUERY_STALE_TIMES.MEDIUM,
    gcTime: QUERY_GC_TIMES.STANDARD,
  });
};
