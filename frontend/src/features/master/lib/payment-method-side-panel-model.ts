import type { PaymentMethod } from "../api/payment-method-master";

export interface PaymentMethodFormData {
  name: string;
  isActive: boolean;
}

export function paymentMethodToFormData(item: PaymentMethod | null): PaymentMethodFormData {
  return {
    name: item?.name ?? "",
    isActive: item?.isActive ?? true,
  };
}
