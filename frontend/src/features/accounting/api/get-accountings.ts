import { useQuery } from "@tanstack/react-query";
import { axios } from "@/lib/axios";
import type { Accounting } from "../types";
import { transformToAccounting } from "./transforms";
import type { BackendAccounting } from "./types";

interface AccountingsListResponse {
  data: BackendAccounting[];
  total: number;
  page: number;
  limit: number;
}

export const getAccountings = async (): Promise<Accounting[]> => {
  const { data } = await axios.get<AccountingsListResponse>("/v1/accountings");
  return data.data.map(transformToAccounting);
};

export const useGetAccountings = () => {
  return useQuery({
    queryKey: ["accountings"],
    queryFn: getAccountings,
  });
};
