import { useQuery } from "@tanstack/react-query";
import { axios } from "@/lib/axios";
import type { BackendInsurance } from "./types";

export const getInsurances = async (): Promise<BackendInsurance[]> => {
  const { data } = await axios.get<BackendInsurance[]>("/v1/masters/insurances");
  return data;
};

export const useGetInsurances = () => {
  return useQuery({
    queryKey: ["masters", "insurances"],
    queryFn: getInsurances,
    staleTime: 30 * 60 * 1000, // 静的マスタデータは30分キャッシュ
  });
};
