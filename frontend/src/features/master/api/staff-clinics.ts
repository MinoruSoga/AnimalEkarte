import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { axios } from "@/lib/axios";
import { handleApiError } from "@/lib/handle-api-error";
import { QUERY_STALE_TIMES, QUERY_GC_TIMES } from "@/lib/react-query";

import { STAFFS_QUERY_KEY } from "./staffs";

// ─────────────────────────────────────────────────
// Clinics list (for staff assignment UI)
// ─────────────────────────────────────────────────

export interface ClinicSummary {
  id: string;
  name: string;
}

const getClinicsListKey = (scope?: "all") =>
  ["clinics-list", scope ?? "assigned"] as const;

export function useGetClinicsList(scope?: "all") {
  return useQuery({
    queryKey: getClinicsListKey(scope),
    queryFn: async (): Promise<ClinicSummary[]> => {
      const params = scope ? { scope } : undefined;
      const { data } = await axios.get<Array<{ id: number; name: string }>>(
        "/v1/clinics",
        { params },
      );
      return data.map((c) => ({ id: String(c.id), name: c.name }));
    },
    staleTime: QUERY_STALE_TIMES.STATIC,
    gcTime: QUERY_GC_TIMES.LONG,
  });
}

// ─────────────────────────────────────────────────
// Staff Clinic Assignments API
// ─────────────────────────────────────────────────

const STAFF_CLINICS_KEY = (staffId: string) =>
  [...STAFFS_QUERY_KEY, staffId, "clinics"] as const;

export function useGetStaffClinics(staffId: string | null) {
  return useQuery({
    queryKey: STAFF_CLINICS_KEY(staffId ?? ""),
    queryFn: async (): Promise<string[]> => {
      const { data } = await axios.get<{ clinic_ids: number[] }>(
        `/v1/masters/staffs/${staffId}/clinics`,
      );
      return (data.clinic_ids ?? []).map(String);
    },
    enabled: staffId !== null,
    staleTime: QUERY_STALE_TIMES.STATIC,
    gcTime: QUERY_GC_TIMES.LONG,
  });
}

export function useUpdateStaffClinics() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: async ({
      staffId,
      clinicIds,
    }: {
      staffId: string;
      clinicIds: string[];
    }) => {
      await axios.put(`/v1/masters/staffs/${staffId}/clinics`, {
        clinic_ids: clinicIds.map((id) => parseInt(id, 10)),
      });
    },
    onSuccess: (_data, variables) => {
      queryClient.invalidateQueries({
        queryKey: STAFF_CLINICS_KEY(variables.staffId),
      });
    },
    onError: (error) => handleApiError(error, "設定"),
  });
}
