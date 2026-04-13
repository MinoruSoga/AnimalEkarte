import { useQuery } from "@tanstack/react-query";
import { axios } from "@/lib/axios";
import { QUERY_STALE_TIMES, QUERY_GC_TIMES } from "@/lib/react-query";

export interface BackendStaff {
  id: number;
  name: string;
  is_active: boolean;
}

/** マスタのスタッフ一覧を取得する */
export function useGetStaffs() {
  return useQuery({
    queryKey: ["reception", "staffs"],
    queryFn: async () => {
      const { data } = await axios.get<BackendStaff[]>("/v1/masters/staffs");
      return data;
    },
    // スタッフ情報は頻繁に変わらないので30分キャッシュ
    staleTime: QUERY_STALE_TIMES.STATIC,
    gcTime: QUERY_GC_TIMES.LONG,
  });
}

/** staffId → スタッフ名のMapを構築する */
export function buildStaffMap(staffs: BackendStaff[]): Map<string, string> {
  return new Map(staffs.map((s) => [String(s.id), s.name]));
}
