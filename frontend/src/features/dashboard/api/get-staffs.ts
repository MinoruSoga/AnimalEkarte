import { useQuery } from "@tanstack/react-query";
import { axios } from "@/lib/axios";
import { QUERY_STALE_TIMES } from "@/lib/react-query";

export interface BackendStaff {
  id: string;
  name: string;
  staff_role: string;
  is_active: boolean;
}

interface StaffListResponse {
  data: BackendStaff[];
  total: number;
  page: number;
  limit: number;
}

/** マスタのスタッフ一覧を取得する */
export function useGetStaffs() {
  return useQuery({
    queryKey: ["masters", "staffs"],
    queryFn: async () => {
      const { data } = await axios.get<StaffListResponse>("/v1/masters/staffs");
      return data.data;
    },
    // スタッフ情報は頻繁に変わらないので30分キャッシュ
    staleTime: QUERY_STALE_TIMES.STATIC,
  });
}

/** staffId → スタッフ名のMapを構築する */
export function buildStaffMap(staffs: BackendStaff[]): Map<string, string> {
  return new Map(staffs.map((s) => [s.id, s.name]));
}
