import { useMemo } from "react";
import { useQuery } from "@tanstack/react-query";
import { axios } from "@/lib/axios";
import { QUERY_STALE_TIMES, QUERY_GC_TIMES } from "@/lib/react-query";
import { queryKeys } from "@/lib/query-keys";
import type { MasterItem } from "@/types";
import { normalizeKana } from "@/lib/normalize-kana";

// ─────────────────────────────────────────────────
// Types
// ─────────────────────────────────────────────────

/** Common fields present in ALL backend master entities */
interface GenericMasterBackendItem {
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

// ─────────────────────────────────────────────────
// Constants
// ─────────────────────────────────────────────────

/** Maps frontend category key → backend endpoint path */
const MASTER_CATEGORY_ENDPOINT: Record<string, string> = {
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

// ─────────────────────────────────────────────────
// Transform
// ─────────────────────────────────────────────────

function transformGenericMasterItem(data: GenericMasterBackendItem): MasterItem {
  return {
    id: String(data.id),
    name: data.name,
    price: data.price != null ? Number(data.price) : 0,
    status: data.is_active ? "active" : "inactive",
    description: data.description ?? "",
    // Consultation-specific fields
    timeCondition: data.time_condition != null ? String(data.time_condition) : undefined,
    duration: data.duration != null ? Number(data.duration) : null,
    // TrimmingCourse-specific field (#73)
    courseTypeId: data.course_type_id != null ? String(data.course_type_id) : undefined,
  };
}

// ─────────────────────────────────────────────────
// API function
// ─────────────────────────────────────────────────

const getMasterItemsByEndpoint = async (endpoint: string): Promise<MasterItem[]> => {
  const { data } = await axios.get<GenericMasterBackendItem[]>(endpoint);
  return data.map(transformGenericMasterItem);
};

// ─────────────────────────────────────────────────
// Hooks
// ─────────────────────────────────────────────────

/**
 * Read-only hook to fetch master items by category key.
 * The source of truth for generic master item fetching across all features.
 */
function useGetMasterItemsByCategory(category: string) {
  const endpoint = MASTER_CATEGORY_ENDPOINT[category];
  return useQuery({
    queryKey: queryKeys.masters.category(category),
    queryFn: () => getMasterItemsByEndpoint(endpoint),
    enabled: !!endpoint,
    staleTime: QUERY_STALE_TIMES.STATIC,
    gcTime: QUERY_GC_TIMES.LONG,
  });
}

/** Read-only master data hook for cross-feature consumption. Returns filtered items by category. */
export function useGetMasterItems(category?: string, searchTerm?: string) {
  const resolvedCategory = category && category !== "all" ? category : "";
  const { data: categoryItems = [], isLoading } = useGetMasterItemsByCategory(resolvedCategory);

  const filteredItems = useMemo(() => {
    if (!searchTerm) return categoryItems;
    const normalizedTerm = normalizeKana(searchTerm).toLowerCase();
    return categoryItems.filter(
      (i) =>
        normalizeKana(i.name).toLowerCase().includes(normalizedTerm) ||
        (i.description
          ? normalizeKana(i.description).toLowerCase().includes(normalizedTerm)
          : false),
    );
  }, [categoryItems, searchTerm]);

  return { data: filteredItems, isLoading };
}
