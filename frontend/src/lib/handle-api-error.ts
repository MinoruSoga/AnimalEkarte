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
    
    // バックエンドの RespondError 規約に合わせて data.error を取得
    const serverMessage = err.response?.data?.error as string | undefined;

    if (status === 400) {
      toast.error(serverMessage ?? `${context}に失敗しました。入力内容を確認してください。`);
    } else if (status === 401) {
      // 401時は自動ログアウト・リダイレクトが行われるべきだが、ここでは通知のみ
      toast.error("セッションが切れました。再度ログインしてください。");
    } else if (status === 403) {
      // 権限バッジ + ルートガードで既にUI制御済み。トーストは冗長なので抑制。
      return;
    } else if (status === 404) {
      toast.error(serverMessage ?? `${context}対象が見つかりません。`);
    } else if (status === 409) {
      toast.error(serverMessage ?? "他のユーザーによって更新されています。一度リロードしてください。");
    } else if (status !== undefined && status >= 500) {
      toast.error(`サーバーエラーが発生しました。しばらく経ってから再度お試しください。`);
    } else {
      toast.error(`${context}に失敗しました。ネットワーク接続を確認してください。`);
    }
    return;
  }

  // Non-Axios errors (unexpected)
  console.error("Non-Axios Error:", err);
  toast.error(`${context}中に予期しないエラーが発生しました。`);
}
