import { keepPreviousData, useQuery } from "@tanstack/react-query";
import { axios } from "@/lib/axios";
import { QUERY_STALE_TIMES, QUERY_GC_TIMES } from "@/lib/react-query";
import { queryKeys } from "@/lib/query-keys";
import { transformMedicalRecord } from "@/lib/transforms/medical-record";
import type { MedicalRecordResponse } from "@/types/generated/medicalrecord-responses";
import type { MedicalRecordFilters, MedicalRecordsResult } from "@/hooks/use-medical-records";

/**
 * EMR-245: カルテ一覧ページ専用の取得。共有版 useGetMedicalRecords
 * （@/hooks/use-medical-records、owner-report / partner-record-link も利用）
 * に表示列フィルタ（ownerName / petName / chiefComplaint）を加えた拡張。
 * パラメータ直列化は共有版と同じ規則 — 共有版に変更が入ったら合わせること。
 */

export interface MedicalRecordsPageFilters extends MedicalRecordFilters {
  /** 飼主名の部分一致（BE: owner_name、name/name_kana 畳込み ILIKE） */
  ownerName?: string;
  /** ペット名の部分一致（BE: pet_name、name/name_kana 畳込み ILIKE） */
  petName?: string;
  /** 主訴の部分一致（BE: chief_complaint、inquiries.chief_complaint ILIKE） */
  chiefComplaint?: string;
}

interface MedicalRecordsListResponse {
  data: MedicalRecordResponse[];
  total: number;
  page: number;
  limit: number;
}

const DEFAULT_PAGE = 1;
const DEFAULT_LIMIT = 20;

async function getMedicalRecordsPage(
  filters?: MedicalRecordsPageFilters,
): Promise<MedicalRecordsResult> {
  const params: Record<string, string | number> = {
    page: filters?.page ?? DEFAULT_PAGE,
    limit: filters?.limit ?? DEFAULT_LIMIT,
  };
  if (filters?.startDate) params.start_date = filters.startDate;
  if (filters?.endDate) params.end_date = filters.endDate;
  if (filters?.petId) params.pet_id = filters.petId;
  if (filters?.ownerId) params.owner_id = filters.ownerId;
  if (filters?.clinicIds?.length) {
    params.clinic_ids = filters.clinicIds.join(",");
  }
  if (filters?.search) params.search = filters.search;
  if (filters?.status) params.status = filters.status;
  if (filters?.doctorId) params.doctor_id = filters.doctorId;
  if (filters?.animalSpeciesId) params.animal_species_id = filters.animalSpeciesId;
  if (filters?.medicineId) params.medicine_id = filters.medicineId;
  if (filters?.procedureId) params.procedure_id = filters.procedureId;
  if (filters?.consultationId) params.consultation_id = filters.consultationId;
  if (filters?.inventoryId) params.inventory_id = filters.inventoryId;
  // EMR-245: 表示列フィルタは横断 search とは独立したクエリパラメータで送る
  if (filters?.ownerName) params.owner_name = filters.ownerName;
  if (filters?.petName) params.pet_name = filters.petName;
  if (filters?.chiefComplaint) params.chief_complaint = filters.chiefComplaint;
  if (filters?.sort) params.sort = filters.sort;
  if (filters?.order) params.order = filters.order;

  const { data } = await axios.get<MedicalRecordsListResponse>("/v1/medical-records", {
    params,
  });
  return {
    data: data.data.map(transformMedicalRecord),
    total: data.total,
    page: data.page,
    limit: data.limit,
  };
}

export function useGetMedicalRecordsPage(
  filters?: MedicalRecordsPageFilters,
  options?: { preservePreviousData?: boolean; enabled?: boolean },
) {
  return useQuery({
    queryKey: queryKeys.medicalRecords.list(filters),
    queryFn: () => getMedicalRecordsPage(filters),
    placeholderData: options?.preservePreviousData ? keepPreviousData : undefined,
    enabled: options?.enabled ?? true,
    staleTime: QUERY_STALE_TIMES.MEDIUM,
    gcTime: QUERY_GC_TIMES.STANDARD,
  });
}
