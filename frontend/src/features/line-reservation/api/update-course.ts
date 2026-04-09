import { useMutation, useQueryClient } from "@tanstack/react-query";
import { axios } from "@/lib/axios";
import type { ServiceType, UpdateServiceTypeRequest } from "./types";

export async function updateCourse(
  clinicId: string,
  id: number,
  payload: UpdateServiceTypeRequest
): Promise<ServiceType> {
  const { data } = await axios.put<ServiceType>(
    `/v1/clinics/${clinicId}/reservation-courses/${id}`,
    payload
  );
  return data;
}

export function useUpdateCourse(clinicId: string | null) {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: ({ id, payload }: { id: number; payload: UpdateServiceTypeRequest }) =>
      updateCourse(clinicId!, id, payload),
    onSuccess: () => {
      void queryClient.invalidateQueries({ queryKey: ["reservation-courses", clinicId] });
    },
  });
}

export async function patchCourseStatus(
  clinicId: string,
  id: number,
  isActive: boolean
): Promise<ServiceType> {
  const { data } = await axios.patch<ServiceType>(
    `/v1/clinics/${clinicId}/reservation-courses/${id}/status`,
    { is_active: isActive }
  );
  return data;
}

export function usePatchCourseStatus(clinicId: string | null) {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: ({ id, isActive }: { id: number; isActive: boolean }) =>
      patchCourseStatus(clinicId!, id, isActive),
    onSuccess: () => {
      void queryClient.invalidateQueries({ queryKey: ["reservation-courses", clinicId] });
    },
  });
}

export async function patchCourseSortOrder(
  clinicId: string,
  id: number,
  sortOrder: number
): Promise<ServiceType> {
  const { data } = await axios.patch<ServiceType>(
    `/v1/clinics/${clinicId}/reservation-courses/${id}/sort-order`,
    { sort_order: sortOrder }
  );
  return data;
}

export function usePatchCourseSortOrder(clinicId: string | null) {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: ({ id, sortOrder }: { id: number; sortOrder: number }) =>
      patchCourseSortOrder(clinicId!, id, sortOrder),
    onSuccess: () => {
      void queryClient.invalidateQueries({ queryKey: ["reservation-courses", clinicId] });
    },
  });
}
