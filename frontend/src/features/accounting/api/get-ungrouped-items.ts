import { useQuery } from "@tanstack/react-query";
import { axios } from "@/lib/axios";
import { queryKeys } from "@/lib/query-keys";
import { QUERY_STALE_TIMES } from "@/lib/react-query";

export interface UngroupedSameDaySummary {
  medicalRecordCount: number;
  trimmingCount: number;
  hasUngrouped: boolean;
}

interface BackendUngroupedSameDay {
  medical_record_count: number;
  trimming_count: number;
  has_ungrouped: boolean;
}

const getUngroupedSameDay = async (
  petId: string,
  date: string,
): Promise<UngroupedSameDaySummary> => {
  const { data } = await axios.get<BackendUngroupedSameDay>(
    "/v1/billing-items/ungrouped-same-day",
    {
      params: { pet_id: petId, date },
    },
  );
  return {
    medicalRecordCount: data.medical_record_count ?? 0,
    trimmingCount: data.trimming_count ?? 0,
    hasUngrouped: data.has_ungrouped ?? false,
  };
};

/** #77: 同日同ペットの未会計対象化項目(診察/トリミング)件数を取得する取り残し警告用フック。 */
export const useGetUngroupedSameDay = (petId: string, date: string, enabled: boolean) =>
  useQuery({
    queryKey: queryKeys.accounting.ungroupedItems(petId, date),
    queryFn: () => getUngroupedSameDay(petId, date),
    enabled: enabled && !!petId,
    staleTime: QUERY_STALE_TIMES.SHORT,
  });
