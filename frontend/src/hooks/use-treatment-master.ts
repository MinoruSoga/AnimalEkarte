import { useQuery } from "@tanstack/react-query";
import { axios } from "@/lib/axios";
import { QUERY_STALE_TIMES, QUERY_GC_TIMES } from "@/lib/react-query";
import { queryKeys } from "@/lib/query-keys";
import {
  transformVaccine,
  transformCheckupType,
  transformConsultation,
  transformProcedure,
} from "@/lib/transforms/treatment";
import { transformBackendMedicineToFrontend } from "@/lib/transforms/medicine";
import type {
  VaccineItem,
  CheckupTypeItem,
  ConsultationItem,
  ProcedureItem,
  ConsultationWire,
  ProcedureWire,
  VaccineWire,
} from "@/lib/transforms/treatment";
import type { Medicine } from "@/lib/transforms/medicine";
import type { LegacyMedicineResponse } from "@/types/medicine";
import type {
  CheckupType,
  HospitalizationPlan as HospitalizationPlanModel,
} from "@/types/generated/models";

export type { VaccineItem, CheckupTypeItem, ConsultationItem, ProcedureItem };

// ─────────────────────────────────────────────────
// Read-only hooks for treatment master data
// Used by cross-feature consumers that need reference data without CRUD.
// Query keys match the master feature's hooks to share the React Query cache.
// ─────────────────────────────────────────────────

export function useGetAllVaccinesMaster() {
  return useQuery({
    queryKey: queryKeys.masters.category("vaccines"),
    queryFn: async (): Promise<VaccineItem[]> => {
      const { data } = await axios.get<VaccineWire[]>("/v1/masters/vaccines");
      return data.map(transformVaccine);
    },
    staleTime: QUERY_STALE_TIMES.STATIC,
    gcTime: QUERY_GC_TIMES.LONG,
  });
}

export function useGetAllCheckupTypes() {
  return useQuery({
    queryKey: queryKeys.masters.category("checkup-types"),
    queryFn: async (): Promise<CheckupTypeItem[]> => {
      const { data } = await axios.get<CheckupType[]>("/v1/masters/checkup-types");
      return data.map(transformCheckupType);
    },
    staleTime: QUERY_STALE_TIMES.STATIC,
    gcTime: QUERY_GC_TIMES.LONG,
  });
}

/**
 * 診察マスタ一覧。truthy な `clinicId` を明示するとクリニック別キーと
 * `X-Clinic-ID` ヘッダ付き GET になる。省略・空文字なら従来のグローバルキーと
 * 暗黙の共通ヘッダを維持し、他クリニックデータへのフォールバックはしない。
 */
export function useGetAllConsultations({ clinicId }: { clinicId?: string } = {}) {
  return useQuery({
    queryKey: clinicId
      ? queryKeys.masters.categoryForClinic("consultations", clinicId)
      : queryKeys.masters.category("consultations"),
    queryFn: async (): Promise<ConsultationItem[]> => {
      const { data } = clinicId
        ? await axios.get<ConsultationWire[]>("/v1/masters/consultations", {
            headers: { "X-Clinic-ID": clinicId },
          })
        : await axios.get<ConsultationWire[]>("/v1/masters/consultations");
      return data.map(transformConsultation);
    },
    staleTime: QUERY_STALE_TIMES.STATIC,
    gcTime: QUERY_GC_TIMES.LONG,
  });
}

/**
 * 処置マスタ一覧。truthy な `clinicId` でクリニック別キー＋`X-Clinic-ID` 付き GET、
 * 省略・空文字なら従来のグローバルキーと暗黙ヘッダ。`enabled: false` の間は
 * 自動取得を止める（手動 refetch は可能）。他クリニックデータへのフォールバックはしない。
 */
export function useGetAllProcedures({
  enabled = true,
  clinicId,
}: { enabled?: boolean; clinicId?: string } = {}) {
  return useQuery({
    enabled,
    queryKey: clinicId
      ? queryKeys.masters.categoryForClinic("procedures", clinicId)
      : queryKeys.masters.category("procedures"),
    queryFn: async (): Promise<ProcedureItem[]> => {
      const { data } = clinicId
        ? await axios.get<ProcedureWire[]>("/v1/masters/procedures", {
            headers: { "X-Clinic-ID": clinicId },
          })
        : await axios.get<ProcedureWire[]>("/v1/masters/procedures");
      return data.map(transformProcedure);
    },
    staleTime: QUERY_STALE_TIMES.STATIC,
    gcTime: QUERY_GC_TIMES.LONG,
  });
}

/**
 * #201: TreatmentSearchDialog（薬剤カテゴリ）・カルテ投与量プレビューの両方から参照する薬剤マスタ。
 * clinicId 指定時はクリニック別の queryKey とヘッダを使い、省略時は
 * features/master/api/medicines.ts と同じグローバルキーのキャッシュを共有する。
 */
export const getAllMedicinesMaster = async (clinicId?: string): Promise<Medicine[]> => {
  const { data } = clinicId
    ? await axios.get<LegacyMedicineResponse[]>("/v1/masters/medicines", {
        headers: { "X-Clinic-ID": clinicId },
      })
    : await axios.get<LegacyMedicineResponse[]>("/v1/masters/medicines");
  return data.map(transformBackendMedicineToFrontend);
};

export function useGetAllMedicinesMaster({
  enabled = true,
  clinicId,
}: { enabled?: boolean; clinicId?: string } = {}) {
  return useQuery({
    enabled,
    queryKey: clinicId
      ? queryKeys.masters.categoryForClinic("medicines", clinicId)
      : queryKeys.masters.category("medicines"),
    queryFn: () => getAllMedicinesMaster(clinicId),
    staleTime: QUERY_STALE_TIMES.STATIC,
    gcTime: QUERY_GC_TIMES.LONG,
  });
}

export interface HospitalizationPlanRefItem {
  id: string;
  name: string;
  /** マスタ単価。0 は有限値として保持し欠損/空欄に畳まない(API 省略時は 0)。 */
  price: number;
}

/**
 * FE-RC-015: hospitalization/CarePlanRefSelect が features/master を直接 import
 * していた feature-to-feature import を解消するための読み取り専用参照。
 * queryKey は features/master/api/hospitalization-plans.ts の
 * useGetAllHospitalizationPlans と同一にしてキャッシュを共有する。
 */
export function useGetAllHospitalizationPlansMaster({
  enabled = true,
}: { enabled?: boolean } = {}) {
  return useQuery({
    enabled,
    queryKey: queryKeys.masters.category("hospitalization-plans"),
    queryFn: async (): Promise<HospitalizationPlanRefItem[]> => {
      const { data } = await axios.get<HospitalizationPlanModel[]>(
        "/v1/masters/hospitalization-plans",
      );
      return data.map((item) => ({
        id: String(item.id),
        name: item.name,
        price: item.price ?? 0,
      }));
    },
    staleTime: QUERY_STALE_TIMES.STATIC,
    gcTime: QUERY_GC_TIMES.LONG,
  });
}
