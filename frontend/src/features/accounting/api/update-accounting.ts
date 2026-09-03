import { axios } from "@/lib/axios";
import { transformToAccounting } from "./transforms";
import type { Accounting } from "./transforms";
import type { BackendAccounting, UpdateAccountingRequest } from "./types";

export const updateAccounting = async (
  id: string,
  req: UpdateAccountingRequest,
): Promise<Accounting> => {
  const { data } = await axios.patch<BackendAccounting>(`/v1/accountings/${id}`, req);
  return transformToAccounting(data);
};
