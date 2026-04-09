import { useMutation, useQueryClient } from "@tanstack/react-query";
import { axios } from "@/lib/axios";
import type { Staff, UpdateStaffRequest } from "./types";

export async function updateReservationStaff(
  clinicId: string,
  id: number,
  payload: UpdateStaffRequest
): Promise<Staff> {
  const { data } = await axios.put<Staff>(
    `/v1/clinics/${clinicId}/reservation-staffs/${id}`,
    payload
  );
  return data;
}

export function useUpdateReservationStaff(clinicId: string | null) {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: ({ id, payload }: { id: number; payload: UpdateStaffRequest }) =>
      updateReservationStaff(clinicId!, id, payload),
    onSuccess: () => {
      void queryClient.invalidateQueries({ queryKey: ["reservation-staffs", clinicId] });
    },
  });
}

export async function patchStaffStatus(
  clinicId: string,
  id: number,
  isActive: boolean
): Promise<Staff> {
  const { data } = await axios.patch<Staff>(
    `/v1/clinics/${clinicId}/reservation-staffs/${id}/status`,
    { is_active: isActive }
  );
  return data;
}

export function usePatchStaffStatus(clinicId: string | null) {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: ({ id, isActive }: { id: number; isActive: boolean }) =>
      patchStaffStatus(clinicId!, id, isActive),
    onSuccess: () => {
      void queryClient.invalidateQueries({ queryKey: ["reservation-staffs", clinicId] });
    },
  });
}

export async function patchStaffSortOrder(
  clinicId: string,
  id: number,
  sortOrder: number
): Promise<Staff> {
  const { data } = await axios.patch<Staff>(
    `/v1/clinics/${clinicId}/reservation-staffs/${id}/sort-order`,
    { sort_order: sortOrder }
  );
  return data;
}

export function usePatchStaffSortOrder(clinicId: string | null) {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: ({ id, sortOrder }: { id: number; sortOrder: number }) =>
      patchStaffSortOrder(clinicId!, id, sortOrder),
    onSuccess: () => {
      void queryClient.invalidateQueries({ queryKey: ["reservation-staffs", clinicId] });
    },
  });
}
