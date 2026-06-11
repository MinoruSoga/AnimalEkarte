import { useMutation } from "@tanstack/react-query";
import { toast } from "sonner";
import { axios } from "@/lib/axios";
import { handleApiError } from "@/lib/handle-api-error";

interface LstepBulkTagBody {
  owner_ids: string[];
  tag_name: string;
}

async function createLstepBulkTag(
  clinicId: string,
  body: LstepBulkTagBody
): Promise<void> {
  await axios.post(`/v1/clinics/${clinicId}/owners/lstep/bulk-tags`, body);
}

// POST /api/clinics/:clinic_id/owners/lstep/bulk-tags
export function useCreateLstepBulkTag() {
  const clinicId = localStorage.getItem("auth_current_clinic:v1") ?? null;

  return useMutation({
    mutationFn: (body: LstepBulkTagBody) => {
      if (clinicId === null) {
        return Promise.reject(new Error("clinic_id is not selected"));
      }
      return createLstepBulkTag(clinicId, body);
    },
    onSuccess: (_data, variables) => {
      toast.success(`${variables.owner_ids.length}件にタグを付与しました`);
    },
    onError: (error) => {
      handleApiError(error, "タグ付与");
    },
  });
}
