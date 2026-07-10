import { useQuery } from "@tanstack/react-query";
import { axios } from "@/lib/axios";
import { QUERY_STALE_TIMES, QUERY_GC_TIMES } from "@/lib/react-query";
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

function transformImageGalleryGroup(g: { groupId: number; date: string; images: ImageGalleryItem[] }) {
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

const getMedicalRecordImages = async (
  medicalRecordId: string,
): Promise<ImageGalleryGroup[]> => {
  const { data } = await axios.get<MedicalRecordImage[]>(
    `/v1/medical-records/${medicalRecordId}/images`,
  );
  return groupImagesByDate(data ?? []);
};

export const useGetMedicalRecordImages = (medicalRecordId?: string) => {
  return useQuery({
    queryKey: ["record-images", medicalRecordId],
    queryFn: () => getMedicalRecordImages(medicalRecordId!),
    enabled: !!medicalRecordId,
    staleTime: QUERY_STALE_TIMES.MEDIUM,
    gcTime: QUERY_GC_TIMES.STANDARD,
  });
};
