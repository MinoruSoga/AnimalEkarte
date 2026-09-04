import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { axios } from "@/lib/axios";
import { handleApiError } from "@/lib/handle-api-error";
import { QUERY_STALE_TIMES, QUERY_GC_TIMES } from "@/lib/react-query";
import { queryKeys } from "@/lib/query-keys";
import type { PaymentMethodMaster as ModelPaymentMethodMaster } from "@/types/generated/models";

// ─────────────────────────────────────────────────
// Request types (derived from models.ts)
// ─────────────────────────────────────────────────

type PaymentMethodRequestBase = Omit<
  ModelPaymentMethodMaster,
  "id" | "clinic_id" | "created_at" | "updated_at"
>;

export type CreatePaymentMethodRequest = Pick<PaymentMethodRequestBase, "name"> &
  Partial<Omit<PaymentMethodRequestBase, "name">>;

export type UpdatePaymentMethodRequest = Partial<PaymentMethodRequestBase>;

// ─────────────────────────────────────────────────
// Transform
// ─────────────────────────────────────────────────

function transformPaymentMethod(data: ModelPaymentMethodMaster) {
  return {
    id: String(data.id),
    name: data.name,
    isActive: data.is_active,
    displayOrder: data.display_order,
  };
}

export type PaymentMethod = ReturnType<typeof transformPaymentMethod>;

// ─────────────────────────────────────────────────
// API functions
// ─────────────────────────────────────────────────

const getPaymentMethods = async (): Promise<PaymentMethod[]> => {
  const { data } = await axios.get<ModelPaymentMethodMaster[]>("/v1/payment-methods");
  return data.map(transformPaymentMethod);
};

const createPaymentMethod = async (req: CreatePaymentMethodRequest): Promise<PaymentMethod> => {
  const { data } = await axios.post<ModelPaymentMethodMaster>("/v1/payment-methods", req);
  return transformPaymentMethod(data);
};

const updatePaymentMethod = async (
  id: string,
  req: UpdatePaymentMethodRequest,
): Promise<PaymentMethod> => {
  const { data } = await axios.patch<ModelPaymentMethodMaster>(`/v1/payment-methods/${id}`, req);
  return transformPaymentMethod(data);
};

const deletePaymentMethod = async (id: string): Promise<void> => {
  await axios.delete(`/v1/payment-methods/${id}`);
};

// ─────────────────────────────────────────────────
// Query hooks
// ─────────────────────────────────────────────────

export const useGetPaymentMethods = () =>
  useQuery({
    queryKey: queryKeys.masters.category("payment-methods"),
    queryFn: getPaymentMethods,
    staleTime: QUERY_STALE_TIMES.STATIC,
    gcTime: QUERY_GC_TIMES.LONG,
  });

export const useCreatePaymentMethod = () => {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: createPaymentMethod,
    onSuccess: () =>
      queryClient.invalidateQueries({ queryKey: queryKeys.masters.category("payment-methods") }),
    onError: (error) => handleApiError(error, "作成"),
  });
};

export const useUpdatePaymentMethod = () => {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: ({ id, req }: { id: string; req: UpdatePaymentMethodRequest }) =>
      updatePaymentMethod(id, req),
    onSuccess: () =>
      queryClient.invalidateQueries({ queryKey: queryKeys.masters.category("payment-methods") }),
    onError: (error) => handleApiError(error, "更新"),
  });
};

export const useDeletePaymentMethod = () => {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: deletePaymentMethod,
    onSuccess: () =>
      queryClient.invalidateQueries({ queryKey: queryKeys.masters.category("payment-methods") }),
    onError: (error) => handleApiError(error, "削除"),
  });
};
