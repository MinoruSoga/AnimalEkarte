import { useMutation } from "@tanstack/react-query";
import { toast } from "sonner";
import { axios } from "@/lib/axios";
import { handleApiError } from "@/lib/handle-api-error";

export interface LineLinkTokenResult {
  token: string;
  expiresAt: string;
  liffUrl: string;
}

interface GenerateLineLinkTokenApiResponse {
  token: string;
  expires_at: string;
  liff_url?: string;
}

async function generateLineLinkToken(ownerId: string): Promise<LineLinkTokenResult> {
  const { data } = await axios.post<GenerateLineLinkTokenApiResponse>(
    `/v1/owners/${ownerId}/line/link-token`,
  );
  return {
    token: data.token,
    expiresAt: data.expires_at,
    liffUrl: data.liff_url ?? "",
  };
}

export function useGenerateLineLinkToken(ownerId: string) {
  return useMutation({
    mutationFn: () => generateLineLinkToken(ownerId),
    onSuccess: (result) => {
      if (!result.liffUrl) {
        // LIFF ID 未設定のクリニックでは BE が liff_url を空で返す（SD-14）。
        // トークン自体は発行済みだが URL を提示できないため、無音失敗を避けて明示する。
        toast.error("LIFF IDが未設定のため連携用URLを発行できません。設定を確認してください。");
      }
    },
    onError: (error) => {
      handleApiError(error, "LINE連携用URL発行");
    },
  });
}
