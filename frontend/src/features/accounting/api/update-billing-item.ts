import { axios } from "@/lib/axios";
import type { UpdateBillingItemRequest, BackendAccountingItem } from "./types";
import { transformAccountingItem } from "./transforms";
import type { AccountingItem } from "./transforms";

export const updateBillingItem = async (
  itemId: string,
  req: UpdateBillingItemRequest,
): Promise<AccountingItem> => {
  const { data } = await axios.patch<BackendAccountingItem>(`/v1/billing-items/${itemId}`, req);
  return transformAccountingItem(data);
};
