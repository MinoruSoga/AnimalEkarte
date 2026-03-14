import { useQuery } from "@tanstack/react-query";
import { axios } from "@/lib/axios";

// ─────────────────────────────────────────────────
// Backend types (snake_case) - service_type_response.go 準拠
// ─────────────────────────────────────────────────

export interface BackendServiceType {
  id: number;
  clinic_id: number;
  name: string;
  /** HEX color code stored in DB (e.g. "#3B82F6") */
  color: string;
  is_active: boolean;
  description: string;
  sort_order: number;
  created_at: string;
  updated_at: string;
}

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

function transformServiceType(data: BackendServiceType): ServiceType {
  return {
    id: String(data.id),
    clinicId: String(data.clinic_id),
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
  const { data } = await axios.get<BackendServiceType[]>("/v1/masters/service-types");
  return data.map(transformServiceType);
}

// ─────────────────────────────────────────────────
// TanStack Query hooks
// ─────────────────────────────────────────────────

export function useListServiceTypes() {
  return useQuery({
    queryKey: SERVICE_TYPES_QUERY_KEY,
    queryFn: listServiceTypes,
  });
}
