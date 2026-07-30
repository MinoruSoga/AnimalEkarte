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

/** SEC-CS-F08: 同時アップロード数の上限（無制限 Promise.all によるバースト防止） */
export const MEDICAL_RECORD_IMAGE_UPLOAD_CONCURRENCY = 3;

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

/**
 * 有界並列プールで items を処理する（SEC-CS-F08）。
 * JS は単一スレッドのため nextIndex のインクリメントは await 前に行い競合しない。
 */
export async function mapWithConcurrency<T, R>(
  items: readonly T[],
  concurrency: number,
  mapper: (item: T, index: number) => Promise<R>,
): Promise<R[]> {
  if (items.length === 0) {
    return [];
  }
  const limit = Math.max(1, Math.min(concurrency, items.length));
  const results: R[] = new Array(items.length);
  let nextIndex = 0;

  const workers = Array.from({ length: limit }, async () => {
    while (true) {
      const index = nextIndex;
      nextIndex += 1;
      if (index >= items.length) {
        return;
      }
      results[index] = await mapper(items[index], index);
    }
  });

  await Promise.all(workers);
  return results;
}

/** 診療画像の一括アップロード（有界並列）。テストから直接検証できるように export。 */
export const uploadMedicalRecordImages = (
  medicalRecordId: string,
  files: File[],
  clinicId?: string,
): Promise<MedicalRecordImage[]> =>
  mapWithConcurrency(files, MEDICAL_RECORD_IMAGE_UPLOAD_CONCURRENCY, (file) =>
    uploadImage(medicalRecordId, file, clinicId),
  );

export const useCreateMedicalRecordImages = (medicalRecordId: string, clinicId?: string) => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (files: File[]) => uploadMedicalRecordImages(medicalRecordId, files, clinicId),
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
