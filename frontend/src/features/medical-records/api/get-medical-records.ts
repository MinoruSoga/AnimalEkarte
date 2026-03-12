import { useQuery } from "@tanstack/react-query";
import { axios } from "@/lib/axios";
import type { MedicalRecord } from "@/types";
import { transformMedicalRecord } from "./transforms";
import type { BackendMedicalRecord } from "./types";

interface MedicalRecordsListResponse {
  data: BackendMedicalRecord[];
  total: number;
  page: number;
  limit: number;
}

export const getMedicalRecords = async (): Promise<MedicalRecord[]> => {
  const { data } = await axios.get<MedicalRecordsListResponse>("/v1/medical-records");
  return data.data.map(transformMedicalRecord);
};

export const useGetMedicalRecords = () => {
  return useQuery({
    queryKey: ["medical-records"],
    queryFn: getMedicalRecords,
  });
};
