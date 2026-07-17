import { axios } from "@/lib/axios";

/** POST /v1/medical-records — checkupのサブリソース登録に必要なカルテを作成する。 */
export interface CreateMedicalRecordForCheckupInput {
  pet_id: string;
  owner_id: string;
  visit_date: string;
}

export const createMedicalRecordForCheckup = async (
  input: CreateMedicalRecordForCheckupInput,
): Promise<{ id: string }> => {
  const { data } = await axios.post<{ id: string }>("/v1/medical-records", input);
  return data;
};

/** POST /v1/medical-records/:medicalRecordId/checkups — 作成済みカルテに健診記録を登録する。 */
export interface CreateCheckupOnMedicalRecordInput {
  checkup_type_id: number;
  date: string;
  next_date?: string;
  doctor_id?: number;
  result?: string;
}

export const createCheckupOnMedicalRecord = async (
  medicalRecordId: string | number,
  input: CreateCheckupOnMedicalRecordInput,
): Promise<{ id: string }> => {
  const { data } = await axios.post<{ id: string }>(
    `/v1/medical-records/${medicalRecordId}/checkups`,
    input,
  );
  return data;
};
