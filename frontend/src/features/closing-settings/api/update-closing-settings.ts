import { useMutation, useQueryClient } from "@tanstack/react-query";
import { axios } from "@/lib/axios";
import { handleApiError } from "@/lib/handle-api-error";
import type { ClinicSettings } from "@/types/generated/models";

export interface UpdateClosingSettingsRequest {
  closing_am_pm_boundary?: string;
  closing_weekday_end?: string;
  closing_sunday_end?: string;
  closed_weekdays?: number[];
}

export const updateClosingSettings = async (
  data: UpdateClosingSettingsRequest,
): Promise<ClinicSettings> => {
  const { data: res } = await axios.patch<ClinicSettings>("/v1/closing-settings", data);
  return res;
};

export const useUpdateClosingSettings = () => {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: updateClosingSettings,
    onSuccess: () => qc.invalidateQueries({ queryKey: ["closing-settings"] }),
    onError: (error) => handleApiError(error, "休診設定の更新"),
  });
};
