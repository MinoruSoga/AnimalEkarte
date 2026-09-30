import { useMutation } from "@tanstack/react-query";

import { axios } from "@/lib/axios";

export interface ChangeMyPasswordInput {
  current_password: string;
  new_password: string;
}

/**
 * 自分のパスワード変更（PUT /v1/users/me/password）。
 * 共有レイヤのダイアログから HTTP を切り離し、副作用を hooks 層へ委譲する。
 * エラーのステータス別ハンドリング（401=現在のパスワード誤り等）は呼び出し側の責務。
 */
export function useChangeMyPassword() {
  return useMutation({
    mutationFn: async (input: ChangeMyPasswordInput) => {
      await axios.put("/v1/users/me/password", input);
    },
  });
}
