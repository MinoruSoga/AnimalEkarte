import { useQuery } from "@tanstack/react-query";
import { axios } from "@/lib/axios";

export interface LstepCsvImportItem {
  id: string;
  clinic_id: string;
  csv_type: string;
  file_name: string;
  uploaded_by_user_id?: string;
  row_count: number;
  success_count: number;
  error_count: number;
  status: string;
  imported_at?: string;
  created_at: string;
}

// GET /api/v1/clinics/:clinic_id/lstep/csv-imports?limit=N
export function useGetLstepCsvImports(limit = 20) {
  return useQuery({
    queryKey: ["lstep-csv-imports", limit],
    queryFn: async () => {
      const clinicId = localStorage.getItem("auth_current_clinic:v1");
      const { data } = await axios.get<LstepCsvImportItem[]>(
        `/v1/clinics/${clinicId}/lstep/csv-imports?limit=${limit}`
      );
      return data;
    },
    staleTime: 60 * 1000, // 1分キャッシュ
  });
}
