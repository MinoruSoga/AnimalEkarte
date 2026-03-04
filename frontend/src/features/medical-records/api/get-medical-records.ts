import { useQuery } from "@tanstack/react-query";
import { axios } from "@/lib/axios";
import type { MedicalRecord } from "@/types";
import { transformMedicalRecord } from "./transforms";
import type { BackendMedicalRecord } from "./types";

export const getMedicalRecords = async (): Promise<MedicalRecord[]> => {
  const { data } = await axios.get<BackendMedicalRecord[]>("/v1/medical-records");
  return data.map(transformMedicalRecord);
};

export const useGetMedicalRecords = () => {
  return useQuery({
    queryKey: ["medical-records"],
    queryFn: getMedicalRecords,
  });
};
