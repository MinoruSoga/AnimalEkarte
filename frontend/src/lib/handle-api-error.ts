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
export const CONFLICT_CODE_CAGE_NAME = "cage_name_conflict" as const;
export const CONFLICT_CODE_OCCUPATION_NAME = "occupation_name_conflict" as const;
export const CONFLICT_CODE_CONSULTATION_NAME = "consultation_name_conflict" as const;
export const CONFLICT_CODE_EXAM_TYPE_NAME = "exam_type_name_conflict" as const;
export const CONFLICT_CODE_PROCEDURE_NAME = "procedure_name_conflict" as const;
export const CONFLICT_CODE_VACCINE_NAME = "vaccine_name_conflict" as const;
export const CONFLICT_CODE_CHECKUP_TYPE_NAME = "checkup_type_name_conflict" as const;
export const CONFLICT_CODE_MEDICINE_NAME = "medicine_name_conflict" as const;

const KNOWN_CONFLICT_MESSAGES: Record<string, (name: string) => string> = {
  [CONFLICT_CODE_PERMISSION_GROUP_NAME]: (name) =>
    `権限グループ名『${name}』は既に使用されています`,
  [CONFLICT_CODE_ANIMAL_SPECIES_NAME]: (name) =>
    `動物種類『${name}』は既に使用されています`,
  [CONFLICT_CODE_SHIFT_TEMPLATE_NAME]: (name) =>
    `シフトテンプレート名『${name}』は既に使用されています`,
  [CONFLICT_CODE_LSTEP_AUTO_MANAGED_PREFIX]: (name) =>
    `自動管理プレフィックス『${name}』は既に使用されています`,
  [CONFLICT_CODE_CAGE_NAME]: (name) =>
    `ケージ『${name}』は既に使用されています`,
  [CONFLICT_CODE_OCCUPATION_NAME]: (name) =>
    `職種『${name}』は既に使用されています`,
  [CONFLICT_CODE_CONSULTATION_NAME]: (name) =>
    `診察『${name}』は既に使用されています`,
  [CONFLICT_CODE_EXAM_TYPE_NAME]: (name) =>
    `検査『${name}』は既に使用されています`,
  [CONFLICT_CODE_PROCEDURE_NAME]: (name) =>
    `処置『${name}』は既に使用されています`,
  [CONFLICT_CODE_VACCINE_NAME]: (name) =>
    `予防接種『${name}』は既に使用されています`,
  [CONFLICT_CODE_CHECKUP_TYPE_NAME]: (name) =>
    `定期健診『${name}』は既に使用されています`,
  [CONFLICT_CODE_MEDICINE_NAME]: (name) =>
    `薬剤『${name}』は既に使用されています`,
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

const ALREADY_EXISTS_RESOURCE_JA: Record<string, string> = {
  cage: "ケージ",
  occupation: "職種",
  insurance: "保険",
  diagnosis_type: "診断カテゴリ",
  diagnosis_name: "診断病名",
  trimming_course_type: "トリミングコース種別",
  trimming_course: "トリミングコース",
  trimming_option: "トリミングオプション",
  payment_method: "支払方法",
  animal_species: "動物種類",
  permission_group: "権限グループ",
  shift_template: "シフトテンプレート",
  medicine: "薬剤",
  consultation: "診察",
  procedure: "処置",
  vaccine: "ワクチン",
  checkup_type: "健診種別",
  hospitalization_plan: "入院プラン",
  chief_complaint_type: "主訴区分",
  reservation_type: "予約区分",
  merchandise: "物販",
  merchandise_item: "物販",
  campaign: "キャンペーン",
  lstep_condition_tag_mapping: "慢性疾患コード",
  clinic_holiday: "休診日",
};

/** マスタ重複の英語メッセージ（cage '' already exists 等）を日本語へ落とす（BUG-022）。 */
export function localizeAlreadyExistsMessage(serverMessage?: string): string | null {
  if (!serverMessage) {
    return null;
  }
  const match = serverMessage.match(/^(\w+) '(.*)' already exists$/i);
  if (match) {
    const resource = match[1].toLowerCase();
    const name = match[2].trim();
    const ja = ALREADY_EXISTS_RESOURCE_JA[resource];
    if (ja && name) return `${ja}『${name}』は既に使用されています`;
    // BUG-017: empty identifier must not become a type-only sentence
    // (consultation → 「診察は既に使用されています」 looks like the tab label).
  }
  if (/already exists/i.test(serverMessage)) {
    return "既に登録されています";
  }
  return null;
}

/** True when BE `error` already contains user-facing Japanese (hiragana/katakana/kanji). */
function isUserFacingJapanese(message: string | undefined): boolean {
  if (!message) {
    return false;
  }
  return /[\u3040-\u30ff\u3400-\u9fff]/.test(message);
}

/**
 * Extract a user-facing Japanese message from an API/unknown error without toasting.
 * 400/403/404 pass through Japanese `error` text only; English uses the status fallback.
 * Prefer BE Japanese `error` text for 409 business conflicts (reservation slot / no doctors).
 * Do not pass through English 409 messages; unknown-code English uses the reload fallback.
 */
export function extractApiErrorMessage(err: unknown, context = "操作"): string {
  if (axios.isAxiosError(err)) {
    const status = err.response?.status;
    const data = err.response?.data as ApiErrorBody | undefined;
    const serverMessage = data?.error;

    if (status === 400) {
      return (
        (isUserFacingJapanese(serverMessage) ? serverMessage : undefined) ??
        `${context}に失敗しました。入力内容を確認してください。`
      );
    }
    if (status === 401) {
      return "セッションが切れました。再度ログインしてください。";
    }
    if (status === 403) {
      return (
        (isUserFacingJapanese(serverMessage) ? serverMessage : undefined) ??
        `${context}の権限がありません。`
      );
    }
    if (status === 404) {
      return (
        (isUserFacingJapanese(serverMessage) ? serverMessage : undefined) ??
        `${context}対象が見つかりません。`
      );
    }
    if (status === 409) {
      return (
        localizeConflictMessage(data?.code, data?.params) ??
        localizeAlreadyExistsMessage(serverMessage) ??
        (isUserFacingJapanese(serverMessage) ? serverMessage : undefined) ??
        "他のユーザーによって更新されています。一度リロードしてください。"
      );
    }
    if (status !== undefined && status >= 500) {
      return "サーバーエラーが発生しました。しばらく経ってから再度お試しください。";
    }
    return `${context}に失敗しました。ネットワーク接続を確認してください。`;
  }
  if (err instanceof Error && err.message.trim()) {
    return err.message;
  }
  return `${context}中に予期しないエラーが発生しました。`;
}

/**
 * Centralized API error handler.
 * Extracts user-friendly messages from AxiosError and shows toast notifications.
 *
 * @param err - The caught error (unknown type)
 * @param context - Japanese context string for the operation (e.g. "保存", "削除")
 */
export function handleApiError(err: unknown, context: string): void {
  toast.error(extractApiErrorMessage(err, context));
  if (!axios.isAxiosError(err) && import.meta.env.DEV) {
    console.error("Non-Axios Error:", err);
  }
}
