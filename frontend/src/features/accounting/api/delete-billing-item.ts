import { axios } from "@/lib/axios";

import type { DeleteBillingItemRequest } from "./types";

export const deleteBillingItem = async (
  itemId: string,
  req?: DeleteBillingItemRequest,
): Promise<void> => {
  // 締め後削除は body に post_close_reason が必要（BUG-021 / BUG-463）。
  // axios.delete の第2引数は config。data に body を載せる。
  await axios.delete(`/v1/billing-items/${itemId}`, req ? { data: req } : undefined);
};
