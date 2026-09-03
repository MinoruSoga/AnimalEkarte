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
} from "@/lib/transforms/treatment";
import type { Medicine } from "@/lib/transforms/medicine";
import type {
  Vaccine,
  CheckupType,
  Consultation,
  Procedure,
  Medicine as MedicineModel,
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
      const { data } = await axios.get<Vaccine[]>("/v1/masters/vaccines");
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

export function useGetAllConsultations() {
  return useQuery({
    queryKey: queryKeys.masters.category("consultations"),
    queryFn: async (): Promise<ConsultationItem[]> => {
      const { data } = await axios.get<Consultation[]>("/v1/masters/consultations");
      return data.map(transformConsultation);
    },
    staleTime: QUERY_STALE_TIMES.STATIC,
    gcTime: QUERY_GC_TIMES.LONG,
  });
}

export function useGetAllProcedures() {
  return useQuery({
    queryKey: queryKeys.masters.category("procedures"),
    queryFn: async (): Promise<ProcedureItem[]> => {
      const { data } = await axios.get<Procedure[]>("/v1/masters/procedures");
      return data.map(transformProcedure);
    },
    staleTime: QUERY_STALE_TIMES.STATIC,
    gcTime: QUERY_GC_TIMES.LONG,
  });
}

/**
 * #201: TreatmentSearchDialog（薬剤カテゴリ）・カルテ投与量プレビューの両方から参照する薬剤マスタ。
 * queryKey は features/master/api/medicines.ts の useGetAllMedicines と同一
 * (queryKeys.masters.category("medicines")) にしてキャッシュを共有する。
 */
export function useGetAllMedicinesMaster() {
  return useQuery({
    queryKey: queryKeys.masters.category("medicines"),
    queryFn: async (): Promise<Medicine[]> => {
      const { data } = await axios.get<MedicineModel[]>("/v1/masters/medicines");
      return data.map(transformBackendMedicineToFrontend);
    },
    staleTime: QUERY_STALE_TIMES.STATIC,
    gcTime: QUERY_GC_TIMES.LONG,
  });
}

export interface HospitalizationPlanRefItem {
  id: string;
  name: string;
}

/**
 * FE-RC-015: hospitalization/CarePlanRefSelect が features/master を直接 import
 * していた feature-to-feature import を解消するための読み取り専用参照。
 * queryKey は features/master/api/hospitalization-plans.ts の
 * useGetAllHospitalizationPlans と同一にしてキャッシュを共有する。
 */
export function useGetAllHospitalizationPlansMaster() {
  return useQuery({
    queryKey: queryKeys.masters.category("hospitalization-plans"),
    queryFn: async (): Promise<HospitalizationPlanRefItem[]> => {
      const { data } = await axios.get<HospitalizationPlanModel[]>("/v1/masters/hospitalization-plans");
      return data.map((item) => ({ id: String(item.id), name: item.name }));
    },
    staleTime: QUERY_STALE_TIMES.STATIC,
    gcTime: QUERY_GC_TIMES.LONG,
  });
}
