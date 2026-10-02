import { useMemo } from "react";
import { useGetMedicalRecordsPage } from "../api/get-medical-records-page";
import type { MedicalRecordsPageFilters } from "../api/get-medical-records-page";
import type { MedicalRecordSortKey } from "../api/get-medical-records";
import type { ActiveFilter } from "@/components/shared/PropertyFilter/types";
import { toBackendMedicalRecordStatus } from "../api/transforms";

interface UseMedicalRecordsListParams {
  /** 確定済み検索語（useDeferredValue 適用後） */
  searchTerm: string;
  activeFilters: ActiveFilter[];
  /** #86: 拠点横断表示。明示選択の非空 clinicIds は直列化し、未指定・current既定scopeは送信しない */
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
 * PropertyFilter の ActiveFilter（診療日・ステータス・担当医ID・種ID・表示列テキスト）を
 * BE query に変換して `/v1/medical-records` を取得する。旧DB由来を含む全件（425,000件超）へ
 * ページング可能。
 *
 * status/doctor/species は BE が単一値の完全一致のみ対応するため "is" 条件のみを送信対象とする
 * （is_not/is_empty は UI 側で選択不可に制限している。詳細は MedicalRecords.tsx 参照）。
 *
 * EMR-245: 表示列フィルタ owner_name / pet_name / chief_complaint は
 * condition "contains" 固定（BE の ILIKE 部分一致）で、横断検索 search とは
 * 別パラメータとして直列化される。
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
  const filters = useMemo<MedicalRecordsPageFilters>(() => {
    const dateFilter = activeFilters.find((f) => f.key === "date")?.value as
      { from?: string; to?: string } | undefined;
    const statusFilter = activeFilters.find((f) => f.key === "status" && f.condition === "is");
    const doctorFilter = activeFilters.find((f) => f.key === "doctor" && f.condition === "is");
    const speciesFilter = activeFilters.find((f) => f.key === "species" && f.condition === "is");
    // EMR-245: 表示列フィルタ（type:"text" → condition:"contains" 固定）を
    // 独立クエリパラメータへ。横断検索の search とは別パラメータで送る。
    const ownerNameFilter = activeFilters.find(
      (f) => f.key === "owner_name" && f.condition === "contains",
    );
    const petNameFilter = activeFilters.find(
      (f) => f.key === "pet_name" && f.condition === "contains",
    );
    const chiefComplaintFilter = activeFilters.find(
      (f) => f.key === "chief_complaint" && f.condition === "contains",
    );
    const medicineFilter = activeFilters.find((f) => f.key === "medicine" && f.condition === "is");
    const procedureFilter = activeFilters.find(
      (f) => f.key === "procedure" && f.condition === "is",
    );
    const consultationFilter = activeFilters.find(
      (f) => f.key === "consultation" && f.condition === "is",
    );
    const inventoryFilter = activeFilters.find(
      (f) => f.key === "inventory" && f.condition === "is",
    );

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
      ownerName: typeof ownerNameFilter?.value === "string" ? ownerNameFilter.value : undefined,
      petName: typeof petNameFilter?.value === "string" ? petNameFilter.value : undefined,
      chiefComplaint:
        typeof chiefComplaintFilter?.value === "string" ? chiefComplaintFilter.value : undefined,
      medicineId: typeof medicineFilter?.value === "string" ? medicineFilter.value : undefined,
      procedureId: typeof procedureFilter?.value === "string" ? procedureFilter.value : undefined,
      consultationId:
        typeof consultationFilter?.value === "string" ? consultationFilter.value : undefined,
      inventoryId: typeof inventoryFilter?.value === "string" ? inventoryFilter.value : undefined,
      page,
      limit,
      sort,
      order,
    };
  }, [searchTerm, activeFilters, clinicIds, petId, page, limit, sort, order]);

  const { data, isLoading, isError, isPlaceholderData } = useGetMedicalRecordsPage(filters, {
    preservePreviousData: true,
  });

  return {
    records: data?.data ?? [],
    total: data?.total ?? 0,
    page: data?.page ?? page,
    limit: data?.limit ?? limit ?? 20,
    isLoading,
    isError,
    isPlaceholderData,
  };
}
