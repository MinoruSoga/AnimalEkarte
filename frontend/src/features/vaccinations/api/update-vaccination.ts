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
    onSuccess: (_data, { id }) => {
      queryClient.invalidateQueries({ queryKey: ["vaccinations"] });
      // FE4-6 fix: detail クエリの実キーは ["vaccination", id]（単数形）。
      // list prefix invalidation はこれを包含しないため、更新後も詳細画面が stale のまま残っていた。
      queryClient.invalidateQueries({ queryKey: ["vaccination", id] });
    },
    onError: (error) => handleApiError(error, "ワクチン接種更新"),
  });
};
