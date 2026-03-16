import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { axios } from "@/lib/axios";
import { QUERY_STALE_TIMES, QUERY_GC_TIMES } from "@/lib/react-query";
import type { InquiryTemplate as ModelInquiryTemplate } from "@/types/generated/models";

// ─────────────────────────────────────────────────
// Request types
// ─────────────────────────────────────────────────

export interface CreateInquiryTemplateRequest {
  category: string;
  title: string;
  content: string;
  is_active?: boolean;
  sort_order?: number;
}

export interface UpdateInquiryTemplateRequest {
  category?: string;
  title?: string;
  content?: string;
  is_active?: boolean;
  sort_order?: number;
}

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

const QUERY_KEY = ["masters", "inquiry-templates"] as const;

// ─────────────────────────────────────────────────
// API functions
// ─────────────────────────────────────────────────

export async function listInquiryTemplates(): Promise<InquiryTemplate[]> {
  const { data } = await axios.get<ModelInquiryTemplate[]>("/v1/masters/inquiry-templates");
  return data.map(transformInquiryTemplate);
}

export async function createInquiryTemplate(
  req: CreateInquiryTemplateRequest
): Promise<InquiryTemplate> {
  const { data } = await axios.post<ModelInquiryTemplate>("/v1/masters/inquiry-templates", req);
  return transformInquiryTemplate(data);
}

export async function updateInquiryTemplate(
  id: string,
  req: UpdateInquiryTemplateRequest
): Promise<InquiryTemplate> {
  const { data } = await axios.patch<ModelInquiryTemplate>(
    `/v1/masters/inquiry-templates/${id}`,
    req
  );
  return transformInquiryTemplate(data);
}

export async function deleteInquiryTemplate(id: string): Promise<void> {
  await axios.delete(`/v1/masters/inquiry-templates/${id}`);
}

// ─────────────────────────────────────────────────
// TanStack Query hooks
// ─────────────────────────────────────────────────

export function useGetInquiryTemplates() {
  return useQuery({
    queryKey: QUERY_KEY,
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
      queryClient.invalidateQueries({ queryKey: QUERY_KEY });
    },
  });
}

export function useUpdateInquiryTemplate() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: ({ id, req }: { id: string; req: UpdateInquiryTemplateRequest }) =>
      updateInquiryTemplate(id, req),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: QUERY_KEY });
    },
  });
}

export function useDeleteInquiryTemplate() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: deleteInquiryTemplate,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: QUERY_KEY });
    },
  });
}
