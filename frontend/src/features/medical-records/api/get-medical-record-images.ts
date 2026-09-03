import { useQuery } from "@tanstack/react-query";
import { axios } from "@/lib/axios";
import { QUERY_STALE_TIMES, QUERY_GC_TIMES } from "@/lib/react-query";
import { queryKeys } from "@/lib/query-keys";
import type { MedicalRecordImage } from "@/types/generated/models";

function transformImageGalleryItem(img: MedicalRecordImage) {
  return {
    id: img.id,
    name: img.file_name || img.description || `画像${img.id}`,
    src: img.image_url || null,
    label: img.description || img.file_name || `画像${img.id}`,
    mimeType: img.mime_type || undefined,
  };
}
type ImageGalleryItem = ReturnType<typeof transformImageGalleryItem>;

function transformImageGalleryGroup(g: {
  groupId: number;
  date: string;
  images: ImageGalleryItem[];
}) {
  return {
    id: g.groupId,
    date: g.date,
    images: g.images,
  };
}
export type ImageGalleryGroup = ReturnType<typeof transformImageGalleryGroup>;

function formatGroupDate(iso: string): string {
  if (!iso) return "-";
  // Trim to "YYYY/MM/DD HH:MM:SS" format
  const d = iso.slice(0, 19).replace("T", " ");
  return d.replace(/-/g, "/");
}

function groupImagesByDate(images: MedicalRecordImage[]): ImageGalleryGroup[] {
  const groupMap = new Map<string, { groupId: number; date: string; images: ImageGalleryItem[] }>();

  images.forEach((img) => {
    const dateKey = img.created_at ? img.created_at.slice(0, 10) : "unknown";
    if (!groupMap.has(dateKey)) {
      groupMap.set(dateKey, {
        groupId: img.id,
        date: formatGroupDate(img.created_at),
        images: [],
      });
    }
    groupMap.get(dateKey)!.images.push(transformImageGalleryItem(img));
  });

  return Array.from(groupMap.values()).map(transformImageGalleryGroup);
}

// P2-15 (PR #186 review): 拠点横断で開いたカルテ（record.clinicId）の子リソースを操作する場合、
// グローバル選択クリニックではなくレコード自身の clinicId を X-Clinic-ID として送る必要がある。
// clinicId 省略時は axios インターセプタがグローバル選択値にフォールバックする（従来挙動を維持）。
function clinicHeaderConfig(clinicId?: string) {
  return clinicId ? { headers: { "X-Clinic-ID": clinicId } } : undefined;
}

const getMedicalRecordImages = async (
  medicalRecordId: string,
  clinicId?: string,
): Promise<ImageGalleryGroup[]> => {
  const { data } = await axios.get<MedicalRecordImage[]>(
    `/v1/medical-records/${medicalRecordId}/images`,
    clinicHeaderConfig(clinicId),
  );
  return groupImagesByDate(data ?? []);
};

export const useGetMedicalRecordImages = (medicalRecordId?: string, clinicId?: string) => {
  return useQuery({
    queryKey: queryKeys.medicalRecords.images(medicalRecordId!, clinicId),
    queryFn: () => getMedicalRecordImages(medicalRecordId!, clinicId),
    enabled: !!medicalRecordId,
    staleTime: QUERY_STALE_TIMES.MEDIUM,
    gcTime: QUERY_GC_TIMES.STANDARD,
  });
};
