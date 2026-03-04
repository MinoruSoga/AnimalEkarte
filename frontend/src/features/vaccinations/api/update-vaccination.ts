import { useMutation, useQueryClient } from "@tanstack/react-query";
import { axios } from "@/lib/axios";
import type { VaccinationRecord } from "@/types";
import { transformVaccination } from "./transforms";
import type { BackendVaccination, UpdateVaccinationRequest } from "./types";

export const updateVaccination = async (
  id: string,
  req: UpdateVaccinationRequest
): Promise<VaccinationRecord> => {
  const { data } = await axios.put<BackendVaccination>(
    `/v1/vaccinations/${id}`,
    req
  );
  return transformVaccination(data);
};

export const useUpdateVaccination = () => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({
      id,
      req,
    }: {
      id: string;
      req: UpdateVaccinationRequest;
    }) => updateVaccination(id, req),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["vaccinations"] });
    },
  });
};
