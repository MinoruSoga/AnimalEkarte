import { useMutation, useQueryClient } from "@tanstack/react-query";
import { axios } from "@/lib/axios";
import { handleApiError } from "@/lib/handle-api-error";
import { queryKeys } from "@/lib/query-keys";
import type { TrimmingCourseType as ModelTrimmingCourseType } from "@/types/generated/models";
import {
  useGetTrimmingCourseTypes,
  transformTrimmingCourseType,
  type TrimmingCourseType,
} from "@/hooks/use-trimming-course-types";

export { useGetTrimmingCourseTypes, type TrimmingCourseType };

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
// API functions
// ─────────────────────────────────────────────────

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

// ─────────────────────────────────────────────────
// Mutation hooks
// ─────────────────────────────────────────────────

export const useCreateTrimmingCourseType = () => {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: createTrimmingCourseType,
    onSuccess: () =>
      queryClient.invalidateQueries({
        queryKey: queryKeys.masters.category("trimming-course-types"),
      }),
    onError: (error) => handleApiError(error, "作成"),
  });
};

export const useUpdateTrimmingCourseType = () => {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: ({ id, req }: { id: string; req: UpdateTrimmingCourseTypeRequest }) =>
      updateTrimmingCourseType(id, req),
    onSuccess: () =>
      queryClient.invalidateQueries({
        queryKey: queryKeys.masters.category("trimming-course-types"),
      }),
    onError: (error) => handleApiError(error, "更新"),
  });
};

export const useDeleteTrimmingCourseType = () => {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: deleteTrimmingCourseType,
    onSuccess: () =>
      queryClient.invalidateQueries({
        queryKey: queryKeys.masters.category("trimming-course-types"),
      }),
    onError: (error) => handleApiError(error, "削除"),
  });
};
