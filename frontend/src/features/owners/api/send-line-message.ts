import { useMutation, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { axios } from "@/lib/axios";
import { getStoredClinicId } from "@/lib/current-clinic";
import { handleApiError } from "@/lib/handle-api-error";
import { queryKeys } from "@/lib/query-keys";

export type LineSendType = "text" | "pdf_url" | "image_url";

export function isLineSendType(value: string): value is LineSendType {
  return value === "text" || value === "pdf_url" || value === "image_url";
}

export interface LineSendRequest {
  message_type: LineSendType;
  text?: string;
  file_id?: number | null;
  file_name?: string;
  purpose?: string;
}

async function sendLineMessage(
  clinicId: string,
  ownerId: string,
  body: LineSendRequest
): Promise<void> {
  await axios.post(
    `/v1/clinics/${clinicId}/owners/${ownerId}/line/send`,
    body
  );
}

export function useSendLineMessage(ownerId: string) {
  const queryClient = useQueryClient();
  const clinicId = getStoredClinicId();

  return useMutation({
    mutationFn: (body: LineSendRequest) => {
      if (clinicId === null) {
        return Promise.reject(new Error("clinic_id is not selected"));
      }
      return sendLineMessage(clinicId, ownerId, body);
    },
    onSuccess: () => {
      queryClient.invalidateQueries({
        queryKey: queryKeys.lineSendHistory(ownerId),
      });
      toast.success("LINEメッセージを送信しました");
    },
    onError: (error) => {
      handleApiError(error, "LINE送信");
    },
  });
}
