import { useMutation, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
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
  const { data } = await axios.patch<BackendHospitalization>(
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
    onSuccess: (_, { id }) => {
      queryClient.invalidateQueries({ queryKey: ["hospitalizations"] });
      queryClient.invalidateQueries({ queryKey: ["hospitalization", id] });
      queryClient.invalidateQueries({ queryKey: ["hospitalization", "raw", id] });
    },
    onError: (error: Error) => {
      toast.error(error.message || "入院記録の更新に失敗しました");
    },
  });
};
