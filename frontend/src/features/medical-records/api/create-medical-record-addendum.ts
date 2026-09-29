import { axios } from "@/lib/axios";
import type { MedicalRecordAddendumResponse } from "@/types/generated/medicalrecord-responses";

export interface CreateMedicalRecordAddendumInput {
  after_text: string;
  reason: string;
}

export const createMedicalRecordAddendum = async (
  medicalRecordId: string,
  input: CreateMedicalRecordAddendumInput,
): Promise<MedicalRecordAddendumResponse> => {
  const { data } = await axios.post<MedicalRecordAddendumResponse>(
    `/v1/medical-records/${medicalRecordId}/addenda`,
    input,
  );
  return data;
};
