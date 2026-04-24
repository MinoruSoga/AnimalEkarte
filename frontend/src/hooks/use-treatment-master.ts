import { useQuery } from "@tanstack/react-query";
import { axios } from "@/lib/axios";
import { QUERY_STALE_TIMES, QUERY_GC_TIMES } from "@/lib/react-query";
import {
  transformVaccine,
  transformCheckupType,
  transformConsultation,
  transformProcedure,
} from "@/lib/transforms/treatment";
import type {
  VaccineItem,
  CheckupTypeItem,
  ConsultationItem,
  ProcedureItem,
} from "@/lib/transforms/treatment";
import type {
  Vaccine,
  CheckupType,
  Consultation,
  Procedure,
} from "@/types/generated/models";

export type { VaccineItem, CheckupTypeItem, ConsultationItem, ProcedureItem };

// ─────────────────────────────────────────────────
// Read-only hooks for treatment master data
// Used by cross-feature consumers that need reference data without CRUD.
// Query keys match the master feature's hooks to share the React Query cache.
// ─────────────────────────────────────────────────

export function useGetAllVaccinesMaster() {
  return useQuery({
    queryKey: ["masters", "vaccines"] as const,
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
    queryKey: ["masters", "checkup-types"] as const,
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
    queryKey: ["masters", "consultations"] as const,
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
    queryKey: ["masters", "procedures"] as const,
    queryFn: async (): Promise<ProcedureItem[]> => {
      const { data } = await axios.get<Procedure[]>("/v1/masters/procedures");
      return data.map(transformProcedure);
    },
    staleTime: QUERY_STALE_TIMES.STATIC,
    gcTime: QUERY_GC_TIMES.LONG,
  });
}
