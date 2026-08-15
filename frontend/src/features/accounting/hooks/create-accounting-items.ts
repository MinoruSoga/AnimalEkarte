import { createBillingItem } from "../api/create-billing-item";
import type { CreateBillingItemRequest } from "../api/create-billing-item";
import type { AccountingItem } from "../types";

type CreateAccountingItemRequest = CreateBillingItemRequest & {
  other_reason?: string;
};

type CreateItem = (request: CreateAccountingItemRequest) => Promise<unknown>;

export async function createAccountingItemsSequentially(
  accountingId: string | number,
  items: ReadonlyArray<AccountingItem>,
  createItem: CreateItem = createBillingItem,
): Promise<void> {
  // 会計明細は副作用を伴う。失敗後の追加POSTを止めて部分作成範囲を限定するため、意図的に直列化する。
  for (const item of items) {
    await createItem({
      billing_id: Number(accountingId),
      category: item.category,
      name: item.name,
      unit_price: item.unitPrice,
      quantity: item.quantity,
      tax_type: item.taxType,
      tax_rate: item.taxRate,
      is_insurance_applicable: item.isInsuranceApplicable,
      source: item.source,
      ...(item.source === "manual" && item.category === "other" && item.otherReason !== undefined
        ? { other_reason: item.otherReason }
        : {}),
      merchandise_item_id: item.merchandiseItemId ? Number(item.merchandiseItemId) : undefined,
      vaccination_id: item.vaccinationId ? Number(item.vaccinationId) : undefined,
      treatment_id: item.treatmentId ? Number(item.treatmentId) : undefined,
      appointment_id: item.appointmentId ? Number(item.appointmentId) : undefined,
      trimming_course_id: item.trimmingCourseId ? Number(item.trimmingCourseId) : undefined,
      trimming_option_id: item.trimmingOptionId ? Number(item.trimmingOptionId) : undefined,
    });
  }
}
