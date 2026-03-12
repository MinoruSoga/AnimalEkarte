import axios from "axios";
import { toast } from "sonner";

/**
 * Centralized API error handler.
 * Extracts user-friendly messages from AxiosError and shows toast notifications.
 *
 * @param err - The caught error (unknown type)
 * @param context - Japanese context string for the operation (e.g. "保存", "削除")
 */
export function handleApiError(err: unknown, context: string): void {
  if (axios.isAxiosError(err)) {
    const status = err.response?.status;
    const serverMessage = err.response?.data?.message as string | undefined;

    if (status === 400) {
      toast.error(serverMessage ?? `入力値が不正です`);
    } else if (status === 401) {
      toast.error("認証が必要です。再度ログインしてください。");
    } else if (status === 403) {
      toast.error("この操作を実行する権限がありません。");
    } else if (status === 404) {
      toast.error(`${context}対象が見つかりません。`);
    } else if (status === 409) {
      toast.error(serverMessage ?? `${context}中に競合が発生しました。`);
    } else if (status !== undefined && status >= 500) {
      toast.error(`サーバーエラーが発生しました。しばらく経ってから再度お試しください。`);
    } else {
      toast.error(`${context}に失敗しました。ネットワーク接続を確認してください。`);
    }
    return;
  }

  // Non-Axios errors (unexpected)
  toast.error(`${context}中に予期しないエラーが発生しました。`);
}
