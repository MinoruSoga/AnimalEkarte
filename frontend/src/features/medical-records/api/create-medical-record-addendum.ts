import { axios } from "@/lib/axios";
import type { MedicalRecordAddendum } from "@/types/generated/models";

export interface CreateMedicalRecordAddendumInput {
  after_text: string;
  reason: string;
}

export const createMedicalRecordAddendum = async (
  medicalRecordId: string,
  input: CreateMedicalRecordAddendumInput,
): Promise<MedicalRecordAddendum> => {
  const { data } = await axios.post<MedicalRecordAddendum>(
    `/v1/medical-records/${medicalRecordId}/addenda`,
    input,
  );
  return data;
};
