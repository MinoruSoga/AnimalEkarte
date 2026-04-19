// React/Framework
import { useMutation, useQueryClient } from "@tanstack/react-query";

// Internal
import { axios } from "@/lib/axios";
import { handleApiError } from "@/lib/handle-api-error";

// Relative
import type { MedicalRecordImage } from "@/types/generated/models";

// ── Upload ────────────────────────────────────────────────────────────

const uploadImage = async (
  medicalRecordId: string,
  file: File,
): Promise<MedicalRecordImage> => {
  const formData = new FormData();
  formData.append("file", file);
  const { data } = await axios.post<MedicalRecordImage>(
    `/v1/medical-records/${medicalRecordId}/images/upload`,
    formData,
    { headers: { "Content-Type": "multipart/form-data" } },
  );
  return data;
};

export const useCreateMedicalRecordImages = (medicalRecordId: string) => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (files: File[]) =>
      Promise.all(files.map((f) => uploadImage(medicalRecordId, f))),
    onSuccess: () => {
      queryClient.invalidateQueries({
        queryKey: ["record-images", medicalRecordId],
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
): Promise<void> => {
  await axios.delete(
    `/v1/medical-records/${medicalRecordId}/images/${imageId}`,
  );
};

export const useDeleteImage = (medicalRecordId: string) => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (imageId: number) => deleteImage(medicalRecordId, imageId),
    onSuccess: () => {
      queryClient.invalidateQueries({
        queryKey: ["record-images", medicalRecordId],
      });
    },
    onError: (error) => {
      handleApiError(error, "画像削除");
    },
  });
};
