import { axios } from "@/lib/axios";
import type { MedicalRecordAddendumResponse } from "@/types/generated/medicalrecord-responses";

export const getMedicalRecordAddenda = async (
  medicalRecordId: string,
): Promise<MedicalRecordAddendumResponse[]> => {
  const { data } = await axios.get<MedicalRecordAddendumResponse[]>(
    `/v1/medical-records/${medicalRecordId}/addenda`,
  );
  return data ?? [];
};
