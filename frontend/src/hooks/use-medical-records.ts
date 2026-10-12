import { keepPreviousData, useQuery } from "@tanstack/react-query";
import { axios } from "@/lib/axios";
import { formatDate } from "@/lib/format/date";
import { QUERY_STALE_TIMES, QUERY_GC_TIMES } from "@/lib/react-query";
import { queryKeys } from "@/lib/query-keys";
import { transformMedicalRecord, type MedicalRecord } from "@/lib/transforms/medical-record";
import type { MedicalRecordResponse } from "@/types/generated/medicalrecord-responses";

/**
 * EMR-356: 体重単位の wire literal union。BE `model.BodyWeightUnit` の既存定数
 * BodyWeightUnitKg="Kg" / BodyWeightUnitG="g" に ground する。
 * hooks → features の依存は禁止・generated/models は hooks から restricted のため
 * 同形の literal union を局所宣言する（生成 union と構造的に同一）。
 */
type WireBodyWeightUnit = "Kg" | "g";

/**
 * EMR-356: weight_measurements の wire 投影。生成型 medicalrecord-responses.ts は
 * 手編集しないため、追加フィールドはローカルの additive 拡張として型付けする
 * （`make codegen` 反映後も構造的に同一）。
 */
interface WireWeightMeasurement {
  id: number;
  value: number;
  unit: WireBodyWeightUnit;
  recorded_at: string;
}

type MedicalRecordResponseWithWeights = MedicalRecordResponse & {
  /** [] = 測定なし（権威ある空）、キー欠落 = 投影未提供の旧 API。 */
  weight_measurements?: WireWeightMeasurement[];
};

/**
 * FE-RC-015: list / history query の実体。
 * wire→UI 変換は @/lib/transforms/medical-record が正本（followup3）。
 */

export type MedicalRecordSortKey = "date" | "owner_name" | "pet_name" | "status";

export interface MedicalRecordFilters {
  startDate?: string;
  endDate?: string;
  petId?: string;
  ownerId?: string;
  clinicIds?: string[];
  search?: string;
  status?: string;
  doctorId?: string;
  animalSpeciesId?: string;
  /** 薬剤マスタID完全一致フィルタ（treatments.medicine_id）。名前 ILIKE とは別経路 */
  medicineId?: string;
  /** 処置マスタID完全一致フィルタ（treatments.procedure_id） */
  procedureId?: string;
  /** 診察マスタID完全一致フィルタ（treatments.consultation_id） */
  consultationId?: string;
  /** 在庫品ID完全一致フィルタ（treatments.inventory_id、その他項目） */
  inventoryId?: string;
  page?: number;
  limit?: number;
  sort?: MedicalRecordSortKey;
  order?: "asc" | "desc";
}

export type MedicalRecordsResult = {
  data: MedicalRecord[];
  total: number;
  page: number;
  limit: number;
};

interface MedicalRecordsListResponse {
  data: MedicalRecordResponseWithWeights[];
  total: number;
  page: number;
  limit: number;
}

const DEFAULT_PAGE = 1;
const DEFAULT_LIMIT = 20;

export async function getMedicalRecords(
  filters?: MedicalRecordFilters,
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

export function useGetMedicalRecords(
  filters?: MedicalRecordFilters,
  options?: { preservePreviousData?: boolean; enabled?: boolean },
) {
  return useQuery({
    queryKey: queryKeys.medicalRecords.list(filters),
    queryFn: () => getMedicalRecords(filters),
    placeholderData: options?.preservePreviousData ? keepPreviousData : undefined,
    enabled: options?.enabled ?? true,
    staleTime: QUERY_STALE_TIMES.MEDIUM,
    gcTime: QUERY_GC_TIMES.STANDARD,
  });
}

/** EMR-182: 前回複写ペイロード（feature InterviewHistoryCopySource と同形）。 */
interface MedicalRecordInterviewHistoryCopySource {
  /** EMR-219: 治療明細複写の起点となる元カルテ id。明細行は適用時に lazy fetch する。 */
  recordId?: string;
  chiefComplaint?: string;
  treatmentPolicy?: string;
  chiefComplaintTypeId?: number;
}

/**
 * EMR-356: 行に相関する体重測定の表示投影（feature InterviewWeightMeasurement と同形）。
 * 正本は vital_records。measuredAt は測定時点の ISO instant。
 */
export interface MedicalRecordWeightMeasurement {
  id: string;
  value: number;
  unit: WireBodyWeightUnit;
  measuredAt: string;
}

/** FEAT-003: 問診履歴 UI 行（feature InterviewHistoryItem と同形）。 */
export interface MedicalRecordInterviewHistoryItem {
  id: string;
  date: string;
  author: string;
  type: string;
  title: string;
  content: string;
  /** EMR-182: 1項目でも複写可能な値を持つ行にのみ設定される。 */
  copySource?: MedicalRecordInterviewHistoryCopySource;
  /** NO32: 統合タイムラインの時系列ソート用 raw date（feature 側の同形フィールドと対応）。 */
  sortDate?: string;
  /** EMR-356: undefined=投影未提供（旧 API）、[]=測定なし。複写 payload には含めない。 */
  weightMeasurements?: MedicalRecordWeightMeasurement[];
}

/**
 * EMR-182: 前回複写ペイロード。複写可能な値が1つも無いときは undefined。
 * EMR-219: recordId は明細複写の起点となるため、id を持つレコードでは問診値が無くても保持する。
 */
function toCopySource(
  record: MedicalRecordResponse,
): MedicalRecordInterviewHistoryCopySource | undefined {
  const source: MedicalRecordInterviewHistoryCopySource = {};
  if (record.id != null) source.recordId = String(record.id);
  const inquiry = record.inquiry;
  if (inquiry) {
    if (inquiry.chief_complaint) source.chiefComplaint = inquiry.chief_complaint;
    if (inquiry.notes) source.treatmentPolicy = inquiry.notes;
    if (inquiry.chief_complaint_type_id != null) {
      source.chiefComplaintTypeId = inquiry.chief_complaint_type_id;
    }
  }
  return Object.keys(source).length > 0 ? source : undefined;
}

export function transformToHistoryItem(
  record: MedicalRecordResponseWithWeights,
): MedicalRecordInterviewHistoryItem {
  const chiefComplaint = record.inquiry?.chief_complaint ?? "";
  const content = chiefComplaint || "（記録なし）";
  const item: MedicalRecordInterviewHistoryItem = {
    id: String(record.id ?? 0),
    date: formatDate(record.date),
    author: record.doctor?.name ?? "-",
    type: record.status === "finalized" ? "確定済" : "作成中",
    title: chiefComplaint || record.record_no,
    content,
    copySource: toCopySource(record),
    // NO32: 統合タイムラインのソートキー（表示用 date とは別に raw date を保持）。
    sortDate: record.date,
  };
  // EMR-356: キーが存在する場合のみ投影する（undefined=旧 API の未提供と
  // 「測定なし」を区別）。他カルテの最新体重で代理表示はしない。
  if (record.weight_measurements !== undefined) {
    item.weightMeasurements = record.weight_measurements.map((m) => ({
      id: String(m.id),
      value: m.value,
      unit: m.unit,
      measuredAt: m.recorded_at,
    }));
  }
  return item;
}

/**
 * FEAT-003: ペットの問診履歴（InterviewHistoryItem[] 互換）。
 * EMR-356: record 自身の clinicId を任意で受け、クエリキー・X-Clinic-ID・
 * 行フィルタに反映する（拠点横断で別医院の履歴キャッシュ/行を混入させない）。
 * 取得エラーは isError で返す（「記録なし」と区別するため、空配列にはしない）。
 */
export function useGetPetMedicalHistory(
  petId: string | undefined,
  excludeRecordId?: string,
  clinicId?: string,
): { historyItems: MedicalRecordInterviewHistoryItem[]; isLoading: boolean; isError: boolean } {
  const { data, isLoading, isError } = useQuery({
    queryKey: queryKeys.medicalRecords.history(petId!, clinicId),
    queryFn: async () => {
      const params: Record<string, string> = { limit: "50", page: "1" };
      if (petId) params.pet_id = petId;
      const { data: res } = await axios.get<MedicalRecordsListResponse>("/v1/medical-records", {
        params,
        ...(clinicId ? { headers: { "X-Clinic-ID": clinicId } } : {}),
      });
      return res.data;
    },
    enabled: !!petId,
    staleTime: QUERY_STALE_TIMES.MEDIUM,
    gcTime: QUERY_GC_TIMES.STANDARD,
  });

  const historyItems: MedicalRecordInterviewHistoryItem[] = (data ?? [])
    // 防衛: 要求した pet/clinic 以外の行が混ざっても投影しない
    .filter((r) => String(r.pet_id ?? "") === String(petId ?? ""))
    .filter((r) => !clinicId || String(r.clinic_id) === clinicId)
    .filter((r) => !excludeRecordId || String(r.id ?? 0) !== excludeRecordId)
    .map(transformToHistoryItem);

  return { historyItems, isLoading, isError };
}
