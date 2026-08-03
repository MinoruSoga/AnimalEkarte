import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { axios } from "@/lib/axios";
import { handleApiError } from "@/lib/handle-api-error";
import { QUERY_STALE_TIMES, QUERY_GC_TIMES } from "@/lib/react-query";
import { queryKeys } from "@/lib/query-keys";
import type { InquiryTemplate as ModelInquiryTemplate } from "@/types/generated/models";

// ─────────────────────────────────────────────────
// Request types (derived from models.ts)
// ─────────────────────────────────────────────────

type InquiryTemplateRequestBase = Omit<
  ModelInquiryTemplate,
  "id" | "clinic_id" | "created_at" | "updated_at"
>;

export type CreateInquiryTemplateRequest = Pick<
  InquiryTemplateRequestBase,
  "category" | "title" | "content" | "is_active"
> &
  Partial<Omit<InquiryTemplateRequestBase, "category" | "title" | "content" | "is_active">>;

export type UpdateInquiryTemplateRequest = Partial<InquiryTemplateRequestBase>;

// ─────────────────────────────────────────────────
// Transform
// ─────────────────────────────────────────────────

function transformInquiryTemplate(data: ModelInquiryTemplate) {
  return {
    id: String(data.id ?? 0),
    clinicId: String(data.clinic_id ?? 0),
    category: data.category,
    title: data.title,
    content: data.content,
    isActive: data.is_active,
    sortOrder: data.sort_order,
    createdAt: data.created_at,
    updatedAt: data.updated_at,
  };
}

export type InquiryTemplate = ReturnType<typeof transformInquiryTemplate>;

// ─────────────────────────────────────────────────
// Query keys
// ─────────────────────────────────────────────────

// ─────────────────────────────────────────────────
// API functions
// ─────────────────────────────────────────────────

async function listInquiryTemplates(): Promise<InquiryTemplate[]> {
  const { data } = await axios.get<ModelInquiryTemplate[]>("/v1/masters/inquiry-templates");
  return data.map(transformInquiryTemplate);
}

async function createInquiryTemplate(
  req: CreateInquiryTemplateRequest
): Promise<InquiryTemplate> {
  const { data } = await axios.post<ModelInquiryTemplate>("/v1/masters/inquiry-templates", req);
  return transformInquiryTemplate(data);
}

async function updateInquiryTemplate(
  id: string,
  req: UpdateInquiryTemplateRequest
): Promise<InquiryTemplate> {
  const { data } = await axios.patch<ModelInquiryTemplate>(
    `/v1/masters/inquiry-templates/${id}`,
    req
  );
  return transformInquiryTemplate(data);
}

async function deleteInquiryTemplate(id: string): Promise<void> {
  await axios.delete(`/v1/masters/inquiry-templates/${id}`);
}

// ─────────────────────────────────────────────────
// TanStack Query hooks
// ─────────────────────────────────────────────────

export function useGetInquiryTemplates() {
  return useQuery({
    queryKey: queryKeys.masters.category("inquiry-templates"),
    queryFn: listInquiryTemplates,
    staleTime: QUERY_STALE_TIMES.STATIC,
    gcTime: QUERY_GC_TIMES.LONG,
  });
}

export function useCreateInquiryTemplate() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: createInquiryTemplate,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.masters.category("inquiry-templates") });
    },
    onError: (error) => handleApiError(error, "作成"),
  });
}

export function useUpdateInquiryTemplate() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: ({ id, req }: { id: string; req: UpdateInquiryTemplateRequest }) =>
      updateInquiryTemplate(id, req),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.masters.category("inquiry-templates") });
    },
    onError: (error) => handleApiError(error, "更新"),
  });
}

export function useDeleteInquiryTemplate() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: deleteInquiryTemplate,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.masters.category("inquiry-templates") });
    },
    onError: (error) => handleApiError(error, "削除"),
  });
}

