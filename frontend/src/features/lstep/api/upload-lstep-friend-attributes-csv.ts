import { useMutation, useQueryClient } from "@tanstack/react-query";
import { axios } from "@/lib/axios";
import type { LstepCsvImportItem } from "./get-lstep-csv-imports";

// POST /api/v1/clinics/:clinic_id/lstep/csv-imports/friend-attributes (multipart)
export function useUploadLstepFriendAttributesCsv() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: async (file: File) => {
      const clinicId = localStorage.getItem("auth_current_clinic:v1");
      const formData = new FormData();
      formData.append("file", file);
      const { data } = await axios.post<LstepCsvImportItem>(
        `/v1/clinics/${clinicId}/lstep/csv-imports/friend-attributes`,
        formData,
        { headers: { "Content-Type": "multipart/form-data" } }
      );
      return data;
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["lstep-csv-imports"] });
    },
  });
}
