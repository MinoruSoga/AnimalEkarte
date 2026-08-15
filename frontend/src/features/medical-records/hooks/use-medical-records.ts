import { useMemo } from "react";
import { useGetMedicalRecords } from "../api/get-medical-records";
import type { MedicalRecordFilters, MedicalRecordSortKey } from "../api/get-medical-records";
import type { ActiveFilter } from "@/components/shared/PropertyFilter/types";
import { toBackendMedicalRecordStatus } from "../api/transforms";

interface UseMedicalRecordsListParams {
  /** 確定済み検索語（useDeferredValue 適用後） */
  searchTerm: string;
  activeFilters: ActiveFilter[];
  /** #86: 拠点横断表示。2件以上のときのみ送信 */
  clinicIds?: string[];
  petId?: string;
  page: number;
  limit?: number;
  /** B-1 follow-up: 列ソート server 化 */
  sort?: MedicalRecordSortKey;
  order?: "asc" | "desc";
}

/**
 * BUG-B1: カルテ一覧の server-side pagination/search/filter。
 * PropertyFilter の ActiveFilter（診療日・ステータス・担当医ID・種ID）を BE query に変換して
 * `/v1/medical-records` を取得する。旧DB由来を含む全件（425,000件超）へページング可能。
 *
 * status/doctor/species は BE が単一値の完全一致のみ対応するため "is" 条件のみを送信対象とする
 * （is_not/is_empty は UI 側で選択不可に制限している。詳細は MedicalRecords.tsx 参照）。
 */
export function useMedicalRecordsList({
  searchTerm,
  activeFilters,
  clinicIds,
  petId,
  page,
  limit,
  sort,
  order,
}: UseMedicalRecordsListParams) {
  const filters = useMemo<MedicalRecordFilters>(() => {
    const dateFilter = activeFilters.find((f) => f.key === "date")?.value as
      | { from?: string; to?: string }
      | undefined;
    const statusFilter = activeFilters.find((f) => f.key === "status" && f.condition === "is");
    const doctorFilter = activeFilters.find((f) => f.key === "doctor" && f.condition === "is");
    const speciesFilter = activeFilters.find((f) => f.key === "species" && f.condition === "is");

    return {
      startDate: dateFilter?.from,
      endDate: dateFilter?.to,
      clinicIds,
      petId,
      search: searchTerm || undefined,
      status:
        typeof statusFilter?.value === "string"
          ? toBackendMedicalRecordStatus(statusFilter.value)
          : undefined,
      doctorId: typeof doctorFilter?.value === "string" ? doctorFilter.value : undefined,
      animalSpeciesId: typeof speciesFilter?.value === "string" ? speciesFilter.value : undefined,
      page,
      limit,
      sort,
      order,
    };
  }, [searchTerm, activeFilters, clinicIds, petId, page, limit, sort, order]);

  const { data, isLoading, isError } = useGetMedicalRecords(filters);

  return {
    records: data?.data ?? [],
    total: data?.total ?? 0,
    page: data?.page ?? page,
    limit: data?.limit ?? limit ?? 20,
    isLoading,
    isError,
  };
}
