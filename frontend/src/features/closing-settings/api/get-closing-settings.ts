import { useQuery } from "@tanstack/react-query";
import { axios } from "@/lib/axios";
import { queryKeys } from "@/lib/query-keys";
import type { ClinicSettings, ClosingSpecialPeriod } from "@/types/generated/models";

interface ClosingSettingsResponse {
  settings: ClinicSettings;
  special_periods: ClosingSpecialPeriod[];
}

const getClosingSettings = async (): Promise<ClosingSettingsResponse> => {
  const { data } = await axios.get<ClosingSettingsResponse>("/v1/closing-settings");
  return data;
};

export const useGetClosingSettings = () =>
  useQuery({
    queryKey: queryKeys.closingSettings.all(),
    queryFn: getClosingSettings,
  });
