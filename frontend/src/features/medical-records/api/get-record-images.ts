import { useQuery } from "@tanstack/react-query";
import { axios } from "@/lib/axios";
import { QUERY_STALE_TIMES, QUERY_GC_TIMES } from "@/lib/react-query";
import type { RecordImage } from "@/types/generated/models";

export interface ImageGalleryItem {
  id: number;
  name: string;
  src: string | null;
  label: string;
}

export interface ImageGalleryGroup {
  id: number;
  date: string;
  images: ImageGalleryItem[];
}

function formatGroupDate(iso: string): string {
  if (!iso) return "-";
  // Trim to "YYYY/MM/DD HH:MM:SS" format
  const d = iso.slice(0, 19).replace("T", " ");
  return d.replace(/-/g, "/");
}

function groupImagesByDate(images: RecordImage[]): ImageGalleryGroup[] {
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
    groupMap.get(dateKey)!.images.push({
      id: img.id,
      name: img.file_name || img.description || `画像${img.id}`,
      src: img.image_url || null,
      label: img.description || img.file_name || `画像${img.id}`,
    });
  });

  return Array.from(groupMap.values()).map((g) => ({
    id: g.groupId,
    date: g.date,
    images: g.images,
  }));
}

const getRecordImages = async (
  medicalRecordId: string,
): Promise<ImageGalleryGroup[]> => {
  const { data } = await axios.get<RecordImage[]>(
    `/v1/medical-records/${medicalRecordId}/images`,
  );
  return groupImagesByDate(data ?? []);
};

export const useGetRecordImages = (medicalRecordId?: string) => {
  return useQuery({
    queryKey: ["record-images", medicalRecordId],
    queryFn: () => getRecordImages(medicalRecordId!),
    enabled: !!medicalRecordId,
    staleTime: QUERY_STALE_TIMES.MEDIUM,
    gcTime: QUERY_GC_TIMES.STANDARD,
  });
};
