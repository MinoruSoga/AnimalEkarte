import type {
  CreatePaymentMethodRequest,
  PaymentMethod,
  UpdatePaymentMethodRequest,
} from "../api/payment-method-master";
import type { PaymentMethodFormData } from "../lib/payment-method-side-panel-model";

/**
 * BUG-029: reject duplicate names on the client so save never looks successful
 * when no new row is created.
 */
export function validatePaymentMethodForm(
  data: PaymentMethodFormData,
  options: {
    existing: PaymentMethod[] | undefined;
    editingId: string | null;
  },
): string | null {
  const name = data.name.trim();
  if (!name) {
    return "名称は必須です";
  }
  const conflict = (options.existing ?? []).some(
    (item) =>
      item.name.trim() === name && (options.editingId === null || item.id !== options.editingId),
  );
  if (conflict) {
    return `支払方法名「${name}」は既に使用されています`;
  }
  return null;
}

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
