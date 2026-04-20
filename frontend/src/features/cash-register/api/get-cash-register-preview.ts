import { useQuery } from "@tanstack/react-query";
import { axios } from "@/lib/axios";
import type { ClosePreviewResult } from "@/types/generated/models";

export const getCashRegisterPreview = async (
  date: string,
  period: "am" | "pm",
): Promise<ClosePreviewResult> => {
  const { data } = await axios.get<ClosePreviewResult>("/v1/cash-register/preview", {
    params: { date, period },
  });
  return data;
};

export const useGetCashRegisterPreview = (date: string, period: "am" | "pm", enabled: boolean) =>
  useQuery({
    queryKey: ["cash-register-preview", date, period],
    queryFn: () => getCashRegisterPreview(date, period),
    enabled: enabled && !!date,
  });
