import axios from "axios";
import { toast } from "sonner";

/** Stable BE domain conflict codes (BUG-023 / BUG-027 / BUG-026). */
export const CONFLICT_CODE_PERMISSION_GROUP_NAME =
  "permission_group_name_conflict" as const;
export const CONFLICT_CODE_ANIMAL_SPECIES_NAME =
  "animal_species_name_conflict" as const;
export const CONFLICT_CODE_SHIFT_TEMPLATE_NAME =
  "shift_template_name_conflict" as const;
export const CONFLICT_CODE_LSTEP_AUTO_MANAGED_PREFIX =
  "lstep_auto_managed_prefix_conflict" as const;

const KNOWN_CONFLICT_MESSAGES: Record<string, (name: string) => string> = {
  [CONFLICT_CODE_PERMISSION_GROUP_NAME]: (name) =>
    `権限グループ名『${name}』は既に使用されています`,
  [CONFLICT_CODE_ANIMAL_SPECIES_NAME]: (name) =>
    `動物種類『${name}』は既に使用されています`,
  [CONFLICT_CODE_SHIFT_TEMPLATE_NAME]: (name) =>
    `シフトテンプレート名『${name}』は既に使用されています`,
  [CONFLICT_CODE_LSTEP_AUTO_MANAGED_PREFIX]: (name) =>
    `自動管理プレフィックス『${name}』は既に使用されています`,
};

interface ApiErrorBody {
  error?: string;
  code?: string;
  params?: {
    name?: string;
  };
}

/**
 * Map a known 409 conflict code + safe params to a Japanese toast message.
 * Returns null for unknown codes, missing code, or empty/missing name
 * (caller keeps the existing 409 fallback — never show empty 『』).
 */
export function localizeConflictMessage(
  code: string | undefined,
  params: ApiErrorBody["params"] | undefined,
): string | null {
  if (!code) {
    return null;
  }
  const formatter = KNOWN_CONFLICT_MESSAGES[code];
  if (!formatter) {
    return null;
  }
  const name = params?.name?.trim();
  if (!name) {
    return null;
  }
  return formatter(name);
}

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
    const data = err.response?.data as ApiErrorBody | undefined;
    // バックエンドの RespondError 規約に合わせて data.error を取得
    const serverMessage = data?.error;

    if (status === 400) {
      toast.error(serverMessage ?? `${context}に失敗しました。入力内容を確認してください。`);
    } else if (status === 401) {
      // 401時は自動ログアウト・リダイレクトが行われるべきだが、ここでは通知のみ
      toast.error("セッションが切れました。再度ログインしてください。");
    } else if (status === 403) {
      // BUG-377: UI ゲートが漏れたケースのサイレント失敗を防ぐため、
      // 403 時は必ずサーバーメッセージまたは汎用メッセージをトースト表示する。
      toast.error(serverMessage ?? `${context}の権限がありません。`);
    } else if (status === 404) {
      toast.error(serverMessage ?? `${context}対象が見つかりません。`);
    } else if (status === 409) {
      // Prefer stable domain code → JA localization over raw English serverMessage
      // (BUG-023/027/026: internal id + empty string already-exists toasts).
      const localized = localizeConflictMessage(data?.code, data?.params);
      toast.error(
        localized ??
          serverMessage ??
          "他のユーザーによって更新されています。一度リロードしてください。",
      );
    } else if (status !== undefined && status >= 500) {
      toast.error(`サーバーエラーが発生しました。しばらく経ってから再度お試しください。`);
    } else {
      toast.error(`${context}に失敗しました。ネットワーク接続を確認してください。`);
    }
    return;
  }

  // Non-Axios errors (unexpected)
  if (import.meta.env.DEV) {
    console.error("Non-Axios Error:", err);
  }
  toast.error(`${context}中に予期しないエラーが発生しました。`);
}
