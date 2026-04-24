import { axios } from "@/lib/axios";
import { transformToAccounting } from "./transforms";
import type { Accounting } from "./transforms";
import type { BackendAccounting, CreateAccountingRequest } from "./types";

export const createAccounting = async (
  req: CreateAccountingRequest,
): Promise<Accounting> => {
  const { data } = await axios.post<BackendAccounting>(
    "/v1/accountings",
    req,
  );
  return transformToAccounting(data);
};
