// Internal
import { axios } from "@/lib/axios";
import { handleApiError } from "@/lib/handle-api-error";

export interface ChangePasswordInput {
  current_password: string;
  new_password: string;
}

/**
 * PUT /v1/users/me/password — 自分のパスワードを変更する（BUG-062 / FEAT-006）
 */
export const changeMyPassword = async (input: ChangePasswordInput): Promise<void> => {
  try {
    await axios.put("/v1/users/me/password", input);
  } catch (error) {
    handleApiError(error, "パスワード変更");
    throw error;
  }
};
