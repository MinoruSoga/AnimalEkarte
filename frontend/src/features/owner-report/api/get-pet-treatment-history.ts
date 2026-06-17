import { useQuery } from "@tanstack/react-query";
import { axios } from "@/lib/axios";
import { QUERY_STALE_TIMES, QUERY_GC_TIMES } from "@/lib/react-query";

/** 治療履歴の絞り込み。#158: 投薬=medicine / 手術・処置=procedure / 治療=all。 */
export type TreatmentHistoryFilter = "medicine" | "procedure" | "all";

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
  const d = new Date(iso);
  if (isNaN(d.getTime())) return "-";
  const yy = String(d.getFullYear()).slice(2);
  const m = String(d.getMonth() + 1);
  const day = String(d.getDate());
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
    medicalRecordId: row.medical_record_id,
  };
}

const getPetTreatmentHistory = async (
  petId: string,
  filter: TreatmentHistoryFilter,
): Promise<PetTreatmentHistoryItem[]> => {
  const params: Record<string, string | number> = { limit: 100 };
  if (filter !== "all") params.item_type = filter;
  const { data } = await axios.get<PetTreatmentHistoryListResponse>(
    `/v1/pets/${petId}/treatment-history`,
    { params },
  );
  return (data.data ?? []).map(transformHistoryItem);
};

export const useGetPetTreatmentHistory = (
  petId: string | undefined,
  filter: TreatmentHistoryFilter,
) => {
  return useQuery({
    queryKey: ["pet-treatment-history", petId, filter],
    queryFn: () => getPetTreatmentHistory(petId!, filter),
    enabled: !!petId,
    staleTime: QUERY_STALE_TIMES.MEDIUM,
    gcTime: QUERY_GC_TIMES.STANDARD,
  });
};
