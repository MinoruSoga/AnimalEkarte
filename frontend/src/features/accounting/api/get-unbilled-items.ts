import { axios } from "@/lib/axios";
import type { BackendAccountingItem } from "./types";
import { transformAccountingItem } from "./transforms";
import type { AccountingItem } from "./transforms";

export const getUnbilledItems = async (petId: string): Promise<AccountingItem[]> => {
  const { data } = await axios.get<BackendAccountingItem[]>("/v1/billing-items/unbilled", {
    params: { pet_id: petId },
  });
  return (data ?? []).map((item) => ({
    ...transformAccountingItem(item),
    id: `treatment_${item.id}`,
  }));
};
