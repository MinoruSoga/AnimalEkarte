import { axios } from "@/lib/axios";
import type { AccountingItem } from "../types";
import type { BackendAccountingItem } from "./types";
import { transformAccountingItem } from "./transforms";

export interface CreateBillingItemRequest {
  billing_id: number;
  category: string;
  name: string;
  unit_price: number;
  quantity: number;
  tax_type: string;
  tax_rate: number;
  is_insurance_applicable: boolean;
  source: string;
}

export const createBillingItem = async (
  req: CreateBillingItemRequest,
): Promise<AccountingItem> => {
  const { data } = await axios.post<BackendAccountingItem>("/v1/billing-items", req);
  return transformAccountingItem(data);
};
