import { axios } from "@/lib/axios";

export const deleteBillingItem = async (itemId: string): Promise<void> => {
  await axios.delete(`/v1/billing-items/${itemId}`);
};
