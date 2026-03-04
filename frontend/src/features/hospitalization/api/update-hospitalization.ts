import { useMutation, useQueryClient } from "@tanstack/react-query";
import { axios } from "@/lib/axios";
import type { Hospitalization } from "@/types";
import { transformHospitalization } from "./transforms";
import type {
  BackendHospitalization,
  UpdateHospitalizationRequest,
} from "./types";

export const updateHospitalization = async (
  id: string,
  req: UpdateHospitalizationRequest
): Promise<Hospitalization> => {
  const { data } = await axios.put<BackendHospitalization>(
    `/v1/hospitalizations/${id}`,
    req
  );
  return transformHospitalization(data);
};

export const useUpdateHospitalization = () => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({
      id,
      req,
    }: {
      id: string;
      req: UpdateHospitalizationRequest;
    }) => updateHospitalization(id, req),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["hospitalizations"] });
    },
  });
};
