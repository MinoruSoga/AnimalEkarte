import { useMutation, useQueryClient } from "@tanstack/react-query";
import { axios } from "@/lib/axios";
import type { VaccinationRecord } from "@/types";
import { handleApiError } from "@/lib/handle-api-error";
import { transformVaccination } from "./transforms";
import type { BackendVaccination, UpdateVaccinationRequest } from "./types";

const updateVaccination = async (
  id: string,
  req: UpdateVaccinationRequest
): Promise<VaccinationRecord> => {
  const { data } = await axios.patch<BackendVaccination>(
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
    onError: (error) => handleApiError(error, "ワクチン接種更新"),
  });
};
