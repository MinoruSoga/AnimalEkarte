import { useQuery } from "@tanstack/react-query";
import { axios } from "@/lib/axios";
import { QUERY_STALE_TIMES, QUERY_GC_TIMES } from "@/lib/react-query";
import type { ServiceType as ModelServiceType } from "@/types/generated/models";

// ─────────────────────────────────────────────────
// Frontend display type
// ─────────────────────────────────────────────────

export interface ServiceType {
  id: string;
  clinicId: string;
  name: string;
  /** HEX color code (e.g. "#3B82F6") stored in DB */
  color: string;
  isActive: boolean;
  description: string;
  sortOrder: number;
  createdAt: string;
  updatedAt: string;
}

// ─────────────────────────────────────────────────
// Transform
// ─────────────────────────────────────────────────

function transformServiceType(data: ModelServiceType): ServiceType {
  return {
    id: String(data.id ?? 0),
    clinicId: String(data.clinic_id ?? 0),
    name: data.name,
    color: data.color,
    isActive: data.is_active,
    description: data.description,
    sortOrder: data.sort_order,
    createdAt: data.created_at,
    updatedAt: data.updated_at,
  };
}

// ─────────────────────────────────────────────────
// Query keys
// ─────────────────────────────────────────────────

export const SERVICE_TYPES_QUERY_KEY = ["masters", "service-types"] as const;

// ─────────────────────────────────────────────────
// API functions
// ─────────────────────────────────────────────────

export async function listServiceTypes(): Promise<ServiceType[]> {
  const { data } = await axios.get<ModelServiceType[]>("/v1/masters/service-types");
  return data.map(transformServiceType);
}

// ─────────────────────────────────────────────────
// TanStack Query hooks
// ─────────────────────────────────────────────────

export function useListServiceTypes() {
  return useQuery({
    queryKey: SERVICE_TYPES_QUERY_KEY,
    queryFn: listServiceTypes,
    staleTime: QUERY_STALE_TIMES.STATIC,
    gcTime: QUERY_GC_TIMES.LONG,
  });
}
