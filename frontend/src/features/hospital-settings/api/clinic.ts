import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { axios } from "@/lib/axios";
import type { BackendClinic, UpdateClinicRequest } from "./types";

export const getClinic = async (id: string): Promise<BackendClinic> => {
  const { data } = await axios.get<BackendClinic>(`/v1/clinics/${id}`);
  return data;
};

export const useGetClinic = (id: string) => {
  return useQuery({
    queryKey: ["clinic", id],
    queryFn: () => getClinic(id),
    enabled: !!id,
  });
};

export const updateClinic = async (id: string, req: UpdateClinicRequest): Promise<BackendClinic> => {
  const { data } = await axios.patch<BackendClinic>(`/v1/clinics/${id}`, req);
  return data;
};

export const useUpdateClinic = () => {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: ({ id, req }: { id: string; req: UpdateClinicRequest }) => updateClinic(id, req),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["clinic"] });
    },
  });
};

