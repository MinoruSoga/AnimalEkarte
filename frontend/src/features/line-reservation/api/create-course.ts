import { useMutation, useQueryClient } from "@tanstack/react-query";
import { axios } from "@/lib/axios";
import type { ServiceType, CreateServiceTypeRequest } from "./types";

export async function createCourse(
  clinicId: string,
  payload: CreateServiceTypeRequest
): Promise<ServiceType> {
  const { data } = await axios.post<ServiceType>(
    `/v1/clinics/${clinicId}/reservation-courses`,
    payload
  );
  return data;
}

export function useCreateCourse(clinicId: string | null) {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (payload: CreateServiceTypeRequest) => createCourse(clinicId!, payload),
    onSuccess: () => {
      void queryClient.invalidateQueries({ queryKey: ["reservation-courses", clinicId] });
    },
  });
}
