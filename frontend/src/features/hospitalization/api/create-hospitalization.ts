import { useMutation, useQueryClient } from "@tanstack/react-query";
import { axios } from "@/lib/axios";
import type { Hospitalization } from "@/types";
import { transformHospitalization } from "./transforms";
import type {
  BackendHospitalization,
  CreateHospitalizationRequest,
} from "./types";

export const createHospitalization = async (
  req: CreateHospitalizationRequest
): Promise<Hospitalization> => {
  const { data } = await axios.post<BackendHospitalization>(
    "/v1/hospitalizations",
    {
      ...req,
      pet_id: Number(req.pet_id),
      owner_id: Number(req.owner_id),
      cage_id: req.cage_id ? Number(req.cage_id) : undefined,
    }
  );
  return transformHospitalization(data);
};

export const useCreateHospitalization = () => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: createHospitalization,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["hospitalizations"] });
    },
  });
};
