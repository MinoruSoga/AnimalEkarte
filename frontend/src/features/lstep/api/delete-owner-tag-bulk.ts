import { useState } from "react";
import { useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { axios } from "@/lib/axios";
import { requireStoredClinicId } from "@/lib/current-clinic";
import { queryKeys } from "@/lib/query-keys";

interface BulkRemoveProgress {
  done: number;
  total: number;
  isRunning: boolean;
}

interface UseBulkRemoveTagResult {
  bulkRemove: (tagName: string, ownerIds: string[]) => Promise<void>;
  progress: BulkRemoveProgress;
}

// DELETE /api/clinics/:clinic_id/owners/:owner_id/lstep/tags/:tag_name (逐次呼び出し)
// useMutation を使わず直接 axios を呼ぶ（進捗管理のため）
export function useBulkRemoveTag(): UseBulkRemoveTagResult {
  const queryClient = useQueryClient();
  const [progress, setProgress] = useState<BulkRemoveProgress>({
    done: 0,
    total: 0,
    isRunning: false,
  });

  const bulkRemove = async (tagName: string, ownerIds: string[]): Promise<void> => {
    const clinicId = requireStoredClinicId();
    setProgress({ done: 0, total: ownerIds.length, isRunning: true });

    for (const ownerId of ownerIds) {
      await axios.delete(`/v1/clinics/${clinicId}/owners/${ownerId}/lstep/tags/${tagName}`);
      setProgress((prev) => ({ ...prev, done: prev.done + 1 }));
    }

    setProgress((prev) => ({ ...prev, isRunning: false }));
    queryClient.invalidateQueries({ queryKey: queryKeys.lstepTagSummary() });
    toast.success(`タグ「${tagName}」を${ownerIds.length}名から解除しました`);
  };

  return { bulkRemove, progress };
}
