import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { axios } from "@/lib/axios";
import { QUERY_STALE_TIMES, QUERY_GC_TIMES } from "@/lib/react-query";
import type { JobTitle as ModelJobTitle } from "@/types/generated/models";

// ─────────────────────────────────────────────────
// Types
// ─────────────────────────────────────────────────

export interface CreateJobTitleRequest {
  name: string;
  description?: string;
  is_active?: boolean;
}

export interface UpdateJobTitleRequest {
  name?: string;
  description?: string;
  is_active?: boolean;
}

// ─────────────────────────────────────────────────
// Transform
// ─────────────────────────────────────────────────

function transformJobTitle(data: ModelJobTitle) {
  return {
    id: String(data.id ?? 0),
    name: data.name,
    description: data.description,
    isActive: data.is_active,
    sortOrder: data.sort_order,
    createdAt: data.created_at,
    updatedAt: data.updated_at,
  };
}

export type JobTitle = ReturnType<typeof transformJobTitle>;

// ─────────────────────────────────────────────────
// API functions
// ─────────────────────────────────────────────────

const getAllJobTitles = async (): Promise<JobTitle[]> => {
  const { data } = await axios.get<ModelJobTitle[]>("/v1/masters/job-titles");
  return data.map(transformJobTitle);
};

const createJobTitle = async (req: CreateJobTitleRequest): Promise<JobTitle> => {
  const { data } = await axios.post<ModelJobTitle>("/v1/masters/job-titles", req);
  return transformJobTitle(data);
};

const updateJobTitle = async (id: string, req: UpdateJobTitleRequest): Promise<JobTitle> => {
  const { data } = await axios.patch<ModelJobTitle>(`/v1/masters/job-titles/${id}`, req);
  return transformJobTitle(data);
};

const deleteJobTitle = async (id: string): Promise<void> => {
  await axios.delete(`/v1/masters/job-titles/${id}`);
};

const reorderJobTitles = async (req: { ids: number[] }): Promise<void> => {
  await axios.patch("/v1/masters/job-titles/reorder", req);
};

// ─────────────────────────────────────────────────
// Query hooks
// ─────────────────────────────────────────────────

export const useGetAllJobTitles = () => {
  return useQuery({
    queryKey: ["masters", "job-titles"],
    queryFn: getAllJobTitles,
    staleTime: QUERY_STALE_TIMES.STATIC,
    gcTime: QUERY_GC_TIMES.LONG,
  });
};

export const useCreateJobTitle = () => {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: createJobTitle,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["masters", "job-titles"] });
    },
  });
};

export const useUpdateJobTitle = () => {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: ({ id, req }: { id: string; req: UpdateJobTitleRequest }) =>
      updateJobTitle(id, req),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["masters", "job-titles"] });
    },
  });
};

export const useDeleteJobTitle = () => {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: deleteJobTitle,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["masters", "job-titles"] });
    },
  });
};

export const useReorderJobTitles = () => {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: reorderJobTitles,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["masters", "job-titles"] });
    },
  });
};
