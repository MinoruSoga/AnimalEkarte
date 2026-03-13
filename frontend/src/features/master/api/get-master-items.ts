import { useQuery } from "@tanstack/react-query";
import { axios } from "@/lib/axios";
import type { MasterItem } from "@/types";

/** Common fields present in ALL backend master entities */
export interface GenericMasterBackendItem {
  id: string;
  code?: string;
  name: string;
  price?: number | null;
  is_active: boolean;
  description?: string;
  sort_order?: number;
  created_at: string;
  updated_at: string;
  // allow extra entity-specific fields
  [key: string]: unknown;
}

/** Maps frontend category key → backend endpoint path */
export const MASTER_CATEGORY_ENDPOINT: Record<string, string> = {
  examination: "/v1/masters/examination-types",
  vaccine: "/v1/masters/vaccines",
  medicine: "/v1/masters/medicines",
  consultation: "/v1/masters/consultations",
  serviceType: "/v1/masters/service-types",
  procedure: "/v1/masters/procedures",
  hospitalization: "/v1/masters/hospitalization-plans",
  insurance: "/v1/masters/insurances",
  cage: "/v1/masters/cages",
  trimmingCourse: "/v1/masters/trimming-courses",
  trimmingOption: "/v1/masters/trimming-options",
  diagnosisCategory: "/v1/masters/diagnosis-categories",
  diagnosisName: "/v1/masters/diagnosis-names",
  checkup: "/v1/masters/checkup-types",
};

export function transformGenericMasterItem(data: GenericMasterBackendItem): MasterItem {
  return {
    id: data.id,
    name: data.name,
    price: data.price != null ? Number(data.price) : 0,
    status: data.is_active ? "active" : "inactive",
    description: data.description ?? "",
  };
}

export const getMasterItemsByEndpoint = async (endpoint: string): Promise<MasterItem[]> => {
  const { data } = await axios.get<GenericMasterBackendItem[]>(endpoint);
  return data.map(transformGenericMasterItem);
};

export const useGetMasterItemsByCategory = (category: string) => {
  const endpoint = MASTER_CATEGORY_ENDPOINT[category];
  return useQuery({
    queryKey: ["masterItems", category],
    queryFn: () => getMasterItemsByEndpoint(endpoint),
    enabled: !!endpoint,
  });
};

// Keep this for backward compat but return empty (no global fetch endpoint anymore)
export const getMasterItems = async (): Promise<MasterItem[]> => [];
export const useGetMasterItems = () =>
  useQuery({ queryKey: ["masterItems"], queryFn: getMasterItems });
