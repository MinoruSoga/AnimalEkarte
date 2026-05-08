import { axios } from "@/lib/axios";
import type { MedicalRecordAddendum } from "@/types/generated/models";

export const getMedicalRecordAddenda = async (
  medicalRecordId: string,
): Promise<MedicalRecordAddendum[]> => {
  const { data } = await axios.get<MedicalRecordAddendum[]>(
    `/v1/medical-records/${medicalRecordId}/addenda`,
  );
  return data ?? [];
};
