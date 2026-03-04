import { useQuery } from "@tanstack/react-query";
import { axios } from "@/lib/axios";
import type { AccountingRecord } from "@/types";
import { transformAccounting } from "./transforms";
import type { BackendAccounting } from "./types";

export const getAccountings = async (): Promise<AccountingRecord[]> => {
  const { data } = await axios.get<BackendAccounting[]>("/v1/accountings");
  return data.map(transformAccounting);
};

export const useGetAccountings = () => {
  return useQuery({
    queryKey: ["accountings"],
    queryFn: getAccountings,
  });
};
