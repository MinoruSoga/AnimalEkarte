import { useQuery } from "@tanstack/react-query";
import { axios } from "@/lib/axios";
import { QUERY_STALE_TIMES, QUERY_GC_TIMES } from "@/lib/react-query";
import { queryKeys } from "@/lib/query-keys";
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
  reservationType: "/v1/masters/reservation-types",
  procedure: "/v1/masters/procedures",
  hospitalization: "/v1/masters/hospitalization-plans",
  insurance: "/v1/masters/insurances",
  cage: "/v1/masters/cages",
  staff: "/v1/masters/staffs",
  trimmingCourse: "/v1/masters/trimming-courses",
  trimmingOption: "/v1/masters/trimming-options",
  diagnosisType: "/v1/masters/diagnosis-types",
  diagnosisName: "/v1/masters/diagnosis-names",
  checkup: "/v1/masters/checkup-types",
  occupations: "/v1/masters/occupations",
  inquiryTemplate: "/v1/masters/inquiry-templates",
};

export function transformGenericMasterItem(data: GenericMasterBackendItem): MasterItem {
  return {
    id: String(data.id),
    name: data.name,
    price: data.price != null ? Number(data.price) : 0,
    status: data.is_active ? "active" : "inactive",
    description: data.description ?? "",
    // Consultation-specific fields
    timeCondition: data.time_condition != null ? String(data.time_condition) : undefined,
    duration: data.duration != null ? Number(data.duration) : null,
  };
}

const getMasterItemsByEndpoint = async (endpoint: string): Promise<MasterItem[]> => {
  const { data } = await axios.get<GenericMasterBackendItem[]>(endpoint);
  return data.map(transformGenericMasterItem);
};

export const useGetMasterItemsByCategory = (category: string) => {
  const endpoint = MASTER_CATEGORY_ENDPOINT[category];
  return useQuery({
    queryKey: queryKeys.masters.category(category),
    queryFn: () => getMasterItemsByEndpoint(endpoint),
    enabled: !!endpoint,
    staleTime: QUERY_STALE_TIMES.STATIC,
    gcTime: QUERY_GC_TIMES.LONG,
  });
};

