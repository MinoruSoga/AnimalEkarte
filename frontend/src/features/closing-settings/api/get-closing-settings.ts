import { useQuery } from "@tanstack/react-query";
import { axios } from "@/lib/axios";
import type { ClosingSettingsResponse } from "@/types/generated/models";

export const getClosingSettings = async (): Promise<ClosingSettingsResponse> => {
  const { data } = await axios.get<ClosingSettingsResponse>("/v1/closing-settings");
  return data;
};

export const useGetClosingSettings = () =>
  useQuery({
    queryKey: ["closing-settings"],
    queryFn: getClosingSettings,
  });
