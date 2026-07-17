// React/Framework
import { useMutation, useQueryClient } from "@tanstack/react-query";

// Internal
import { axios } from "@/lib/axios";
import { handleApiError } from "@/lib/handle-api-error";
import { queryKeys } from "@/lib/query-keys";

// Relative
import type { MedicalRecordImage } from "@/types/generated/models";

// P2-15 (PR #186 review): 拠点横断で開いたカルテ（record.clinicId）の子リソースを操作する場合、
// グローバル選択クリニックではなくレコード自身の clinicId を X-Clinic-ID として送る必要がある。
// clinicId 省略時は axios インターセプタがグローバル選択値にフォールバックする（従来挙動を維持）。
function clinicHeader(clinicId?: string): Record<string, string> {
  return clinicId ? { "X-Clinic-ID": clinicId } : {};
}

// ── Upload ────────────────────────────────────────────────────────────

const uploadImage = async (
  medicalRecordId: string,
  file: File,
  clinicId?: string,
): Promise<MedicalRecordImage> => {
  const formData = new FormData();
  formData.append("file", file);
  const { data } = await axios.post<MedicalRecordImage>(
    `/v1/medical-records/${medicalRecordId}/images/upload`,
    formData,
    { headers: { "Content-Type": "multipart/form-data", ...clinicHeader(clinicId) } },
  );
  return data;
};

export const useCreateMedicalRecordImages = (medicalRecordId: string, clinicId?: string) => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (files: File[]) =>
      Promise.all(files.map((f) => uploadImage(medicalRecordId, f, clinicId))),
    onSuccess: () => {
      queryClient.invalidateQueries({
        queryKey: queryKeys.medicalRecords.images(medicalRecordId),
      });
    },
    onError: (error) => {
      handleApiError(error, "画像アップロード");
    },
  });
};

// ── Delete ────────────────────────────────────────────────────────────

const deleteImage = async (
  medicalRecordId: string,
  imageId: number,
  clinicId?: string,
): Promise<void> => {
  await axios.delete(
    `/v1/medical-records/${medicalRecordId}/images/${imageId}`,
    clinicId ? { headers: { "X-Clinic-ID": clinicId } } : undefined,
  );
};

export const useDeleteImage = (medicalRecordId: string, clinicId?: string) => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (imageId: number) => deleteImage(medicalRecordId, imageId, clinicId),
    onSuccess: () => {
      queryClient.invalidateQueries({
        queryKey: queryKeys.medicalRecords.images(medicalRecordId),
      });
    },
    onError: (error) => {
      handleApiError(error, "画像削除");
    },
  });
};
