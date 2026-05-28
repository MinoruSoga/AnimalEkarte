import type {
  CreatePaymentMethodRequest,
  UpdatePaymentMethodRequest,
} from "../api/payment-method-master";
import type { PaymentMethodFormData } from "../components/PaymentMethodSidePanelModel";

export function buildPaymentMethodCreateRequest(
  data: PaymentMethodFormData,
): CreatePaymentMethodRequest {
  return {
    name: data.name,
    is_active: data.isActive,
  };
}

export function buildPaymentMethodUpdateRequest(
  data: PaymentMethodFormData,
): UpdatePaymentMethodRequest {
  return buildPaymentMethodCreateRequest(data);
}
