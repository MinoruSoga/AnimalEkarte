import { useMutation, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { axios } from "@/lib/axios";
import { handleApiError } from "@/lib/handle-api-error";

export type LineSendType = "text" | "pdf_url" | "image_url";

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
  const clinicId = localStorage.getItem("auth_current_clinic:v1") ?? "";

  return useMutation({
    mutationFn: (body: LineSendRequest) =>
      sendLineMessage(clinicId, ownerId, body),
    onSuccess: () => {
      queryClient.invalidateQueries({
        queryKey: ["line-send-history", ownerId],
      });
      toast.success("LINEメッセージを送信しました");
    },
    onError: (error) => {
      handleApiError(error, "LINE送信");
    },
  });
}
