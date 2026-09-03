import { isAxiosError } from "axios";

export interface EditableSubOwner {
  ownerId: number;
  name: string;
  nameKana: string;
  relationship: string;
}

export interface SaveState {
  kind: "idle" | "success" | "error";
  message: string;
}

interface ApiErrorResponse {
  error?: unknown;
}

export const INITIAL_SAVE_STATE: SaveState = {
  kind: "idle",
  message: "",
};

export const VERSION_CONFLICT_MESSAGE =
  "他の端末でペット情報が変更されました。再読み込みしてから、もう一度保存してください。";
export const SUB_OWNER_SEARCH_DEBOUNCE_MS = 300;

export function toEditableSubOwners(
  subOwners: ReadonlyArray<{
    owner_id: number;
    name: string;
    name_kana: string;
    relationship: string;
  }>,
): EditableSubOwner[] {
  return subOwners.map((subOwner) => ({
    ownerId: subOwner.owner_id,
    name: subOwner.name,
    nameKana: subOwner.name_kana,
    relationship: subOwner.relationship,
  }));
}

export function getSaveErrorMessage(error: unknown): string {
  if (!isAxiosError<ApiErrorResponse>(error)) {
    return "副飼主を保存できませんでした。時間をおいて再度お試しください。";
  }
  if (error.response?.status === 409) {
    return VERSION_CONFLICT_MESSAGE;
  }
  if (error.response?.status === 400) {
    const serverMessage = error.response.data?.error;
    return typeof serverMessage === "string" && serverMessage.trim() !== ""
      ? `副飼主を保存できませんでした。${serverMessage}`
      : "副飼主を保存できませんでした。入力内容を確認してください。";
  }
  return "副飼主を保存できませんでした。時間をおいて再度お試しください。";
}
