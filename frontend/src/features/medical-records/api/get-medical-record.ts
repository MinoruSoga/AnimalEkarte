import { useQuery } from "@tanstack/react-query";
import { axios } from "@/lib/axios";
import { QUERY_STALE_TIMES, QUERY_GC_TIMES } from "@/lib/react-query";
import { queryKeys } from "@/lib/query-keys";
import { transformMedicalRecord } from "./transforms";
import type { MedicalRecord } from "./transforms";
import type { BackendMedicalRecord } from "./types";

const getMedicalRecord = async (id: string): Promise<MedicalRecord> => {
  const { data } = await axios.get<BackendMedicalRecord>(`/v1/medical-records/${id}`);
  return transformMedicalRecord(data);
};

export const useGetMedicalRecord = (id: string) => {
  return useQuery({
    queryKey: queryKeys.medicalRecords.detail(id),
    queryFn: () => getMedicalRecord(id),
    enabled: !!id,
    staleTime: QUERY_STALE_TIMES.MEDIUM,
    gcTime: QUERY_GC_TIMES.STANDARD,
  });
};
