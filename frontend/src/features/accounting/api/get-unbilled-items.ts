import { axios } from "@/lib/axios";
import type { BackendAccountingItem } from "./types";
import { transformAccountingItem } from "./transforms";
import type { AccountingItem } from "./transforms";

function getCandidateId(item: BackendAccountingItem): string {
  if (item.vaccination_id != null) {
    return `vaccination_${item.vaccination_id}`;
  }
  if (item.treatment_id != null) {
    return `treatment_${item.treatment_id}`;
  }
  return `${item.source ?? "medical_record"}_${item.id}`;
}

export const getUnbilledItems = async (petId: string): Promise<AccountingItem[]> => {
  const { data } = await axios.get<BackendAccountingItem[]>("/v1/billing-items/unbilled", {
    params: { pet_id: petId },
  });
  return (data ?? []).map((item) => ({
    ...transformAccountingItem(item),
    id: getCandidateId(item),
  }));
};
