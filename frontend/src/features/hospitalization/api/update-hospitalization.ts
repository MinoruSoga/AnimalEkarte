import { useMutation, useQueryClient } from "@tanstack/react-query";
import { axios } from "@/lib/axios";
import { handleApiError } from "@/lib/handle-api-error";
import { transformHospitalization } from "./transforms";
import type { Hospitalization } from "./transforms";
import type {
  BackendHospitalization,
  UpdateHospitalizationRequest,
} from "./types";

// バックエンドに送る実際のペイロード型（cage_id は uint64 互換の number）
type UpdateHospitalizationPayload = Omit<UpdateHospitalizationRequest, "cage_id"> & {
  cage_id?: number;
};

export const updateHospitalization = async (
  id: string,
  req: UpdateHospitalizationRequest
): Promise<Hospitalization> => {
  // cage_id は string として受け取るが、バックエンドは uint64（number）を期待するため変換する。
  // 空文字列はフィールドを省略する（ケージなし更新はバックエンドが未サポートのため送信しない）。
  const { cage_id, ...rest } = req;
  const payload: UpdateHospitalizationPayload = { ...rest };
  if (cage_id !== undefined && cage_id !== "") {
    payload.cage_id = Number(cage_id);
  }
  const { data } = await axios.patch<BackendHospitalization>(
    `/v1/hospitalizations/${id}`,
    payload
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
    onError: (error) => {
      handleApiError(error, "更新");
    },
  });
};
