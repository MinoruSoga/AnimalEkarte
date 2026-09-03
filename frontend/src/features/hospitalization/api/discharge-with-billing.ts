import { axios } from "@/lib/axios";

export type DischargeWithBillingRequest = {
  discharge_date: string; // ISO8601
  create_accounting: boolean;
};

export type DischargeWithBillingResult = {
  hospitalization_id: number;
  accounting_id?: number;
  status: string;
};

export const dischargeWithBilling = async (
  id: string,
  req: DischargeWithBillingRequest,
): Promise<DischargeWithBillingResult> => {
  const { data } = await axios.post<DischargeWithBillingResult>(
    `/v1/hospitalizations/${id}/discharge-with-billing`,
    req,
  );
  return data;
};
