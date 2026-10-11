import { useMutation, useQueryClient } from "@tanstack/react-query";
import { axios } from "@/lib/axios";
import { handleApiError } from "@/lib/handle-api-error";
import { queryKeys } from "@/lib/query-keys";
import { transformHospitalization } from "./transforms";
import type { Hospitalization } from "./transforms";
import type { BackendHospitalization, UpdateHospitalizationRequest } from "./types";

// バックエンドに送る実際のペイロード型（cage_id / doctor_id は uint64 互換の number）
type UpdateHospitalizationPayload = Omit<
  UpdateHospitalizationRequest,
  "cage_id" | "doctor_id" | "expected_cage_id"
> & {
  cage_id?: number;
  doctor_id?: number;
  expected_cage_id?: number;
};

export const updateHospitalization = async (
  id: string,
  req: UpdateHospitalizationRequest,
): Promise<Hospitalization> => {
  // cage_id / doctor_id は string として受け取るが、バックエンドは uint64（number）を期待するため変換する。
  // 空文字列は cage_id を省略する。ケージを外す（選択なし）更新は clear_cage_id: true で表し、
  // `rest` にそのまま含まれて BE へ渡る（EMR-316）。
  const { cage_id, doctor_id, expected_cage_id, ...rest } = req;
  const payload: UpdateHospitalizationPayload = { ...rest };
  if (cage_id !== undefined && cage_id !== "") {
    payload.cage_id = Number(cage_id);
  }
  if (expected_cage_id !== undefined && expected_cage_id !== "") {
    payload.expected_cage_id = Number(expected_cage_id);
  }
  if (doctor_id !== undefined && doctor_id !== "") {
    const parsed = Number(doctor_id);
    if (Number.isFinite(parsed) && parsed > 0) {
      payload.doctor_id = parsed;
    }
  }
  const { data } = await axios.patch<BackendHospitalization>(`/v1/hospitalizations/${id}`, payload);
  return transformHospitalization(data);
};

export const useUpdateHospitalization = () => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({ id, req }: { id: string; req: UpdateHospitalizationRequest }) =>
      updateHospitalization(id, req),
    onSuccess: (_, { id }) => {
      queryClient.invalidateQueries({ queryKey: queryKeys.hospitalizations.all() });
      queryClient.invalidateQueries({ queryKey: queryKeys.hospitalizations.detail(id) });
      queryClient.invalidateQueries({ queryKey: queryKeys.hospitalizations.detailRaw(id) });
    },
    onError: (error) => {
      handleApiError(error, "更新");
    },
  });
};
