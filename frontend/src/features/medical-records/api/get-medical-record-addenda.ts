import { axios } from "@/lib/axios";
import type { MedicalRecordAddendumWithAuthorName } from "./types";

export const getMedicalRecordAddenda = async (
  medicalRecordId: string,
): Promise<MedicalRecordAddendumWithAuthorName[]> => {
  const { data } = await axios.get<MedicalRecordAddendumWithAuthorName[]>(
    `/v1/medical-records/${medicalRecordId}/addenda`,
  );
  return data ?? [];
};
