import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { axios } from "@/lib/axios";
import { handleApiError } from "@/lib/handle-api-error";
import { QUERY_STALE_TIMES, QUERY_GC_TIMES } from "@/lib/react-query";
import type { TrimmingCourseType as ModelTrimmingCourseType } from "@/types/generated/models";

const TRIMMING_COURSE_TYPES_QUERY_KEY = ["masters", "trimming-course-types"] as const;

// ─────────────────────────────────────────────────
// Request types
// ─────────────────────────────────────────────────

type TrimmingCourseTypeRequestBase = Omit<
  ModelTrimmingCourseType,
  "id" | "clinic_id" | "created_at" | "updated_at"
>;

export type CreateTrimmingCourseTypeRequest = Pick<TrimmingCourseTypeRequestBase, "name"> &
  Partial<Omit<TrimmingCourseTypeRequestBase, "name">>;

export type UpdateTrimmingCourseTypeRequest = Partial<TrimmingCourseTypeRequestBase>;

// ─────────────────────────────────────────────────
// Transform
// ─────────────────────────────────────────────────

function transformTrimmingCourseType(data: ModelTrimmingCourseType) {
  return {
    id: String(data.id),
    name: data.name,
    isActive: data.is_active,
    sortOrder: data.sort_order,
  };
}

export type TrimmingCourseType = ReturnType<typeof transformTrimmingCourseType>;

// ─────────────────────────────────────────────────
// API functions
// ─────────────────────────────────────────────────

const getTrimmingCourseTypes = async (): Promise<TrimmingCourseType[]> => {
  const { data } = await axios.get<ModelTrimmingCourseType[]>("/v1/masters/trimming-course-types");
  return data.map(transformTrimmingCourseType);
};

const createTrimmingCourseType = async (
  req: CreateTrimmingCourseTypeRequest,
): Promise<TrimmingCourseType> => {
  const { data } = await axios.post<ModelTrimmingCourseType>(
    "/v1/masters/trimming-course-types",
    req,
  );
  return transformTrimmingCourseType(data);
};

const updateTrimmingCourseType = async (
  id: string,
  req: UpdateTrimmingCourseTypeRequest,
): Promise<TrimmingCourseType> => {
  const { data } = await axios.patch<ModelTrimmingCourseType>(
    `/v1/masters/trimming-course-types/${id}`,
    req,
  );
  return transformTrimmingCourseType(data);
};

const deleteTrimmingCourseType = async (id: string): Promise<void> => {
  await axios.delete(`/v1/masters/trimming-course-types/${id}`);
};

const reorderTrimmingCourseTypes = async (ids: number[]): Promise<void> => {
  await axios.patch("/v1/masters/trimming-course-types/reorder", { ids });
};

// ─────────────────────────────────────────────────
// Query hooks
// ─────────────────────────────────────────────────

export const useGetTrimmingCourseTypes = () =>
  useQuery({
    queryKey: TRIMMING_COURSE_TYPES_QUERY_KEY,
    queryFn: getTrimmingCourseTypes,
    staleTime: QUERY_STALE_TIMES.STATIC,
    gcTime: QUERY_GC_TIMES.LONG,
  });

export const useCreateTrimmingCourseType = () => {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: createTrimmingCourseType,
    onSuccess: () => queryClient.invalidateQueries({ queryKey: TRIMMING_COURSE_TYPES_QUERY_KEY }),
    onError: (error) => handleApiError(error, "作成"),
  });
};

export const useUpdateTrimmingCourseType = () => {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: ({ id, req }: { id: string; req: UpdateTrimmingCourseTypeRequest }) =>
      updateTrimmingCourseType(id, req),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: TRIMMING_COURSE_TYPES_QUERY_KEY }),
    onError: (error) => handleApiError(error, "更新"),
  });
};

export const useDeleteTrimmingCourseType = () => {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: deleteTrimmingCourseType,
    onSuccess: () => queryClient.invalidateQueries({ queryKey: TRIMMING_COURSE_TYPES_QUERY_KEY }),
    onError: (error) => handleApiError(error, "削除"),
  });
};

export const useReorderTrimmingCourseTypes = () => {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (ids: number[]) => reorderTrimmingCourseTypes(ids),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: TRIMMING_COURSE_TYPES_QUERY_KEY }),
    onError: (error) => handleApiError(error, "並び替え"),
  });
};
