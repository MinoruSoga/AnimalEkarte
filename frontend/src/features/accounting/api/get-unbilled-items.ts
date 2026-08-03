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

/** BUG-013: typed unbilled warning (server payload is source/code/count/blocking only). */
export interface UnbilledWarning {
  source: string;
  code: string;
  count: number;
  blocking: boolean;
}

export interface UnbilledItemDetails {
  items: AccountingItem[];
  warnings: UnbilledWarning[];
}

interface BackendUnbilledDetails {
  items?: BackendAccountingItem[] | null;
  warnings?: UnbilledWarning[] | null;
}

function mapUnbilledItems(data: BackendAccountingItem[] | null | undefined): AccountingItem[] {
  return (data ?? []).map((item) => ({
    ...transformAccountingItem(item),
    id: getCandidateId(item),
  }));
}

/** Legacy raw-array getter — signature preserved for non-migrated callers. */
export const getUnbilledItems = async (petId: string): Promise<AccountingItem[]> => {
  const { data } = await axios.get<BackendAccountingItem[]>("/v1/billing-items/unbilled", {
    params: { pet_id: petId },
  });
  return mapUnbilledItems(data);
};

/**
 * BUG-013: additive details getter for new accounting consumer.
 * Returns billable candidates plus typed blocking warnings (no silent partial success).
 */
export const getUnbilledItemDetails = async (petId: string): Promise<UnbilledItemDetails> => {
  const { data } = await axios.get<BackendUnbilledDetails>("/v1/billing-items/unbilled-details", {
    params: { pet_id: petId },
  });
  return {
    items: mapUnbilledItems(data?.items),
    warnings: data?.warnings ?? [],
  };
};
