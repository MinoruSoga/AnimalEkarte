import type { Accounting } from "./types";
import { getAccountings } from "./api/get-accountings";

export interface AccountingsLoaderData {
  accountings: Accounting[];
}

export const accountingsLoader = async (): Promise<AccountingsLoaderData> => {
  const accountings = await getAccountings();
  return { accountings };
};
