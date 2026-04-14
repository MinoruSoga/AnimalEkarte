// BUG-371: 会計の論理削除（status=cancelled）
// 旧 DELETE /accountings/:id は廃止。本関数は POST /accountings/:id/cancel を呼び出す。
import { axios } from "@/lib/axios";

export const cancelAccounting = async (id: string): Promise<void> => {
  await axios.post(`/v1/accountings/${id}/cancel`);
};
