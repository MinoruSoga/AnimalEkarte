import { axios } from "@/lib/axios";
import type { UpdateBillingItemRequest, BackendAccountingItem } from "./types";

export const updateBillingItem = async (
  itemId: string,
  req: UpdateBillingItemRequest,
): Promise<BackendAccountingItem> => {
  const { data } = await axios.patch<BackendAccountingItem>(
    `/v1/billing-items/${itemId}`,
    req,
  );
  return data;
};

