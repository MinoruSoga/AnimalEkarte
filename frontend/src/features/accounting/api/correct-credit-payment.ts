import { axios } from "@/lib/axios";
import { transformToAccounting } from "./transforms";
import type { Accounting } from "./transforms";
import type { BackendAccounting, CorrectCreditPaymentRequest } from "./types";

// #189: 確定済み会計のクレジット（カード）金額を確定後に訂正する専用 API。
// 通常の updateAccounting (PATCH) とは別経路。理由必須・監査付きで BE が処理する。
//
// P2-12 (PR #186 review): 拠点横断で開いた会計から訂正する場合、対象は
// グローバル選択クリニックではなく会計自体のクリニック（clinicId）でなければならない。
// X-Clinic-ID ヘッダーを明示指定し、axios interceptor のグローバル値フォールバックを上書きする。
export const correctCreditPayment = async (
  id: string,
  clinicId: string,
  req: CorrectCreditPaymentRequest,
): Promise<Accounting> => {
  const { data } = await axios.post<BackendAccounting>(
    `/v1/accountings/${id}/credit-correction`,
    req,
    { headers: { "X-Clinic-ID": clinicId } },
  );
  return transformToAccounting(data);
};
