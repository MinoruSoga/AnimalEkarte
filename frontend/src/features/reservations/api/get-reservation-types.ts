import { useQuery } from "@tanstack/react-query";
import { axios } from "@/lib/axios";
import { QUERY_STALE_TIMES, QUERY_GC_TIMES } from "@/lib/react-query";

// バックエンドの reservation_type_groups レスポンス型
interface ReservationTypeGroupSummary {
  id: number;
  name: string;
  color: string;
}

// バックエンドの reservation_types レスポンス型（group フィールド含む）
export interface ReservationTypeRaw {
  id: number;
  name: string;
  color: string;
  is_active: boolean;
  group_id: number | null;
  group: ReservationTypeGroupSummary | null;
}

// フロントエンド用にグループ化したデータ
export interface GroupedReservationTypes {
  label: string;
  types: ReservationTypeRaw[];
}

// GET /v1/masters/reservation-types — グループ情報を保持したまま取得
const fetchReservationTypesRaw = async (): Promise<ReservationTypeRaw[]> => {
  const { data } = await axios.get<ReservationTypeRaw[]>("/v1/masters/reservation-types");
  return data;
};

/**
 * 予約区分一覧をグループ情報付きで取得する。
 * BUG-341: ReservationFormFields でのグループ階層表示に使用。
 * useMasterItems("reservationType") の代わりに使い、group_id / group.name を保持する。
 */
export function useGetReservationTypesGrouped() {
  return useQuery({
    queryKey: ["masters", "reservationType", "grouped"],
    queryFn: fetchReservationTypesRaw,
    staleTime: QUERY_STALE_TIMES.STATIC,
    gcTime: QUERY_GC_TIMES.LONG,
    select: (data) => {
      // アクティブなものだけ対象
      const active = data.filter((t) => t.is_active);
      // グループ別に分類
      const map = new Map<string, GroupedReservationTypes>();
      for (const t of active) {
        const key = t.group_id != null ? String(t.group_id) : "__other__";
        const label = t.group?.name ?? "その他";
        if (!map.has(key)) map.set(key, { label, types: [] });
        map.get(key)!.types.push(t);
      }
      return [...map.values()];
    },
  });
}
