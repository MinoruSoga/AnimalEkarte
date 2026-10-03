import { useInfiniteQuery } from "@tanstack/react-query";
import { axios } from "@/lib/axios";
import { queryKeys } from "@/lib/query-keys";
import { QUERY_STALE_TIMES, QUERY_GC_TIMES } from "@/lib/react-query";
import { HISTORY_FETCH_LIMIT } from "@/config/fetch-limits";
import { toJSTWallDate } from "@/lib/jst-date";

/** 治療履歴の絞り込み。#158: 投薬=medicine / 手術・処置=procedure / 治療=all。 */
export type TreatmentHistoryFilter = "medicine" | "procedure" | "all";

/** #159 追加フィルタ。anesthesiaOnly=true で麻酔処置のみ / isSurgery=true で手術処置のみ。 */
export interface TreatmentHistoryOptions {
  anesthesiaOnly?: boolean;
  isSurgery?: boolean;
}

/** GET /v1/pets/:id/treatment-history の 1 行（バックエンド petTreatmentHistoryResponse に対応）。 */
export interface BackendPetTreatmentHistory {
  id: string;
  medical_record_id: string;
  /** 診療日 = medical_records.date 由来（treatments.created_at ではない）。 */
  date: string | null;
  item_type: string;
  content: string;
  memo: string;
  admin_route: string;
  quantity: number;
  unit_price: number;
  status: string;
  medicine_id?: string;
  medicine_name?: string;
  procedure_id?: string;
  procedure_name?: string;
  anesthesia?: string;
  is_surgery?: boolean;
}

interface PetTreatmentHistoryListResponse {
  data: BackendPetTreatmentHistory[];
  total: number;
  page: number;
  limit: number;
}

export interface PetTreatmentHistoryItem {
  id: string;
  /** 表示用 "YY/M/D"。日付不明は "-"。 */
  date: string;
  itemType: string;
  /** 表示名: 薬剤名 / 処置名 / なければ content。 */
  name: string;
  adminRoute: string;
  quantity: number;
  /** 麻酔種別の日本語ラベル（procedure のみ）。 */
  anesthesia?: string;
  /** 手術処置フラグ（procedure のみ）。 */
  isSurgery?: boolean;
  medicalRecordId: string;
}

const ANESTHESIA_LABEL: Record<string, string> = {
  none: "麻酔なし",
  local: "局所麻酔",
  sedation: "鎮静",
  general: "全身麻酔",
};

function formatDate(iso: string | null): string {
  if (!iso) return "-";
  const instant = new Date(iso);
  if (isNaN(instant.getTime())) return "-";
  // 絶対時刻を JST 壁日付に変換してから表示する（ローカル TZ 依存で日付がずれるのを防ぐ）。
  const jst = toJSTWallDate(instant);
  const yy = String(jst.getFullYear()).slice(2);
  const m = String(jst.getMonth() + 1);
  const day = String(jst.getDate());
  return `${yy}/${m}/${day}`;
}

export function transformHistoryItem(row: BackendPetTreatmentHistory): PetTreatmentHistoryItem {
  const name = row.medicine_name || row.procedure_name || row.content || "-";
  const anesthesia =
    row.anesthesia != null ? (ANESTHESIA_LABEL[row.anesthesia] ?? row.anesthesia) : undefined;
  return {
    id: row.id,
    date: formatDate(row.date),
    itemType: row.item_type,
    name,
    adminRoute: row.admin_route ?? "",
    quantity: row.quantity,
    anesthesia,
    isSurgery: row.is_surgery,
    medicalRecordId: row.medical_record_id,
  };
}

export interface PetTreatmentHistoryResult {
  items: PetTreatmentHistoryItem[];
  /**
   * SD-18: 取得上限(HISTORY_FETCH_LIMIT)により実件数より少ない可能性がある場合 true。
   * EMR-242: 追加ページ読み込み後も同じ判定を維持する（total > 全ページの累積 raw 行数）。
   */
  isTruncated: boolean;
}

/** 1 ページ分の取得結果。rawCount はレスポンス生行数。 */
interface PetTreatmentHistoryPage {
  items: PetTreatmentHistoryItem[];
  rawCount: number;
  total?: number;
}

const getPetTreatmentHistory = async (
  petId: string,
  filter: TreatmentHistoryFilter,
  options: TreatmentHistoryOptions = {},
  page: number,
): Promise<PetTreatmentHistoryPage> => {
  const params: Record<string, string | number | boolean> = {
    page,
    limit: HISTORY_FETCH_LIMIT,
  };
  if (filter !== "all") params.item_type = filter;
  if (options.anesthesiaOnly) params.anesthesia_only = true;
  if (options.isSurgery) params.is_surgery = true;
  const { data } = await axios.get<PetTreatmentHistoryListResponse>(
    `/v1/pets/${petId}/treatment-history`,
    { params },
  );
  const rawRows = data.data ?? [];
  return {
    items: rawRows.map(transformHistoryItem),
    rawCount: rawRows.length,
    total: data.total,
  };
};

/** 累積 raw 行数が total に届いていなければ true。 */
function historyHasMore(pages: ReadonlyArray<{ rawCount: number; total?: number }>): boolean {
  const fetched = pages.reduce((sum, page) => sum + page.rawCount, 0);
  const total = pages[pages.length - 1]?.total;
  return typeof total === "number" && total > fetched;
}

export const useGetPetTreatmentHistory = (
  petId: string | undefined,
  filter: TreatmentHistoryFilter,
  options: TreatmentHistoryOptions = {},
) => {
  return useInfiniteQuery({
    queryKey: queryKeys.petTreatmentHistory(petId!, filter, options),
    queryFn: ({ pageParam }) => getPetTreatmentHistory(petId!, filter, options, pageParam),
    initialPageParam: 1,
    getNextPageParam: (_lastPage, allPages) =>
      historyHasMore(allPages) ? allPages.length + 1 : undefined,
    // 消費側の既存 shape ({items, isTruncated}) を維持しつつ全ページを累積する。
    select: (data): PetTreatmentHistoryResult => ({
      items: data.pages.flatMap((page) => page.items),
      isTruncated: historyHasMore(data.pages),
    }),
    enabled: !!petId,
    staleTime: QUERY_STALE_TIMES.MEDIUM,
    gcTime: QUERY_GC_TIMES.STANDARD,
  });
};
