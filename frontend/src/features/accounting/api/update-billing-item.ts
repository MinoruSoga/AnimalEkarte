import { axios } from "@/lib/axios";
import type { AccountingItem } from "../types";
import type { UpdateBillingItemRequest, BackendAccountingItem } from "./types";
import { transformAccountingItem } from "./transforms";

export const updateBillingItem = async (
  itemId: string,
  req: UpdateBillingItemRequest,
): Promise<AccountingItem> => {
  const { data } = await axios.patch<BackendAccountingItem>(
    `/v1/billing-items/${itemId}`,
    req,
  );
  return transformAccountingItem(data);
};

