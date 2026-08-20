import { axios } from "@/lib/axios";
import type { BackendAccountingItem } from "./types";
import { transformAccountingItem } from "./transforms";
import type { AccountingItem } from "./transforms";

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
  merchandise_item_id?: number;
  vaccination_id?: number;
  exam_id?: number;
  treatment_id?: number;
  appointment_id?: number;
  trimming_course_id?: number;
  trimming_option_id?: number;
  /** #115 / BUG-021: レジ締め済み期間の明細追加理由（BE が締め判定時に必須） */
  post_close_reason?: string;
}

export const createBillingItem = async (
  req: CreateBillingItemRequest,
): Promise<AccountingItem> => {
  const { data } = await axios.post<BackendAccountingItem>("/v1/billing-items", req);
  return transformAccountingItem(data);
};
