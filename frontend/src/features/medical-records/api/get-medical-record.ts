import { useQuery } from "@tanstack/react-query";
import { axios } from "@/lib/axios";
import type { MedicalRecord } from "@/types";
import { transformMedicalRecord } from "./transforms";
import type { BackendMedicalRecord } from "./types";

interface MedicalRecordListResponse {
  data: BackendMedicalRecord[];
  total: number;
  page: number;
  limit: number;
}

export const getMedicalRecord = async (id: string): Promise<MedicalRecord> => {
  const { data } = await axios.get<BackendMedicalRecord>(
    `/v1/medical-records/${id}`
  );
  return transformMedicalRecord(data);
};

export const useGetMedicalRecord = (id: string) => {
  return useQuery({
    queryKey: ["medical-record", id],
    queryFn: () => getMedicalRecord(id),
    enabled: !!id,
  });
};

