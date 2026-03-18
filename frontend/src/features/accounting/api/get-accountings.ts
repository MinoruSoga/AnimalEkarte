import { useQuery } from "@tanstack/react-query";
import { axios } from "@/lib/axios";
import type { Accounting, AccountingStatus } from "../types";
import { transformToAccounting } from "./transforms";
import type { BackendAccounting } from "./types";

interface AccountingsListResponse {
  data: BackendAccounting[];
  total: number;
  page: number;
  limit: number;
}

export const getAccountings = async (
  status?: AccountingStatus,
): Promise<Accounting[]> => {
  const params: Record<string, string> = {};
  if (status) params.status = status;
  const { data } = await axios.get<AccountingsListResponse>("/v1/accountings", { params });
  return data.data.map(transformToAccounting);
};

export const useGetAccountings = (status?: AccountingStatus) => {
  return useQuery({
    queryKey: ["accountings", { status }],
    queryFn: () => getAccountings(status),
  });
};
