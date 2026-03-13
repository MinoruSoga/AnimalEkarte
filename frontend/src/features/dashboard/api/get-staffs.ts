import { useQuery } from "@tanstack/react-query";
import { axios } from "@/lib/axios";

export interface BackendStaff {
  id: string;
  name: string;
  staff_role: string;
  is_active: boolean;
}

/** マスタのスタッフ一覧を取得する */
export function useStaffs() {
  return useQuery({
    queryKey: ["masters", "staffs"],
    queryFn: async () => {
      const { data } = await axios.get<BackendStaff[]>("/v1/masters/staffs");
      return data;
    },
    // スタッフ情報は頻繁に変わらないので30分キャッシュ
    staleTime: 30 * 60 * 1000,
  });
}

/** staffId → スタッフ名のMapを構築する */
export function buildStaffMap(staffs: BackendStaff[]): Map<string, string> {
  return new Map(staffs.map((s) => [s.id, s.name]));
}
