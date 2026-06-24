import { axios } from "@/lib/axios";
import { transformToAccounting } from "./transforms";
import type { Accounting } from "./transforms";
import type { BackendAccounting, CorrectCreditPaymentRequest } from "./types";

// #189: 確定済み会計のクレジット（カード）金額を確定後に訂正する専用 API。
// 通常の updateAccounting (PATCH) とは別経路。理由必須・監査付きで BE が処理する。
export const correctCreditPayment = async (
  id: string,
  req: CorrectCreditPaymentRequest,
): Promise<Accounting> => {
  const { data } = await axios.post<BackendAccounting>(
    `/v1/accountings/${id}/credit-correction`,
    req,
  );
  return transformToAccounting(data);
};
