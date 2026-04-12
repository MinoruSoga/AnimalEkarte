import { useMutation, useQueryClient } from "@tanstack/react-query";
import { axios } from "@/lib/axios";
import type { VaccinationRecord } from "@/types";
import { handleApiError } from "@/lib/handle-api-error";
import { transformVaccination } from "./transforms";
import type { BackendVaccination, CreateVaccinationRequest } from "./types";

export const createVaccination = async (
  req: CreateVaccinationRequest
): Promise<VaccinationRecord> => {
  const { data } = await axios.post<BackendVaccination>(
    "/v1/vaccinations",
    req
  );
  return transformVaccination(data);
};

export const useCreateVaccination = () => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: createVaccination,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["vaccinations"] });
    },
    onError: (error) => handleApiError(error, "ワクチン接種登録"),
  });
};
