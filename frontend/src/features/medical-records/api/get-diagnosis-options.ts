import { useQuery } from "@tanstack/react-query";
import { axios } from "@/lib/axios";
import { QUERY_STALE_TIMES, QUERY_GC_TIMES } from "@/lib/react-query";
import { queryKeys } from "@/lib/query-keys";
import type { DiagnosisType, DiagnosisName } from "@/types/generated/models";

interface PaginatedResponse<T> {
  data: T[];
  total: number;
  page: number;
  limit: number;
}

function transformDiagnosisTypeOption(item: DiagnosisType) {
  return {
    id: Number(item.id ?? 0),
    name: item.name,
  };
}
type DiagnosisTypeOption = ReturnType<typeof transformDiagnosisTypeOption>;

function transformDiagnosisNameOption(item: DiagnosisName) {
  return {
    id: Number(item.id ?? 0),
    name: item.name,
  };
}
export type DiagnosisNameOption = ReturnType<typeof transformDiagnosisNameOption>;

const getDiagnosisTypes = async (): Promise<DiagnosisTypeOption[]> => {
  const { data } = await axios.get<DiagnosisType[] | PaginatedResponse<DiagnosisType>>(
    "/v1/masters/diagnosis-types",
    { params: { limit: 100 } },
  );
  const items = Array.isArray(data) ? data : (data.data ?? []);
  return items.map(transformDiagnosisTypeOption);
};

const getDiagnosisNames = async (typeId?: number | null): Promise<DiagnosisNameOption[]> => {
  const params: Record<string, unknown> = { limit: 100 };
  if (typeId) params.type_id = typeId;
  const { data } = await axios.get<DiagnosisName[] | PaginatedResponse<DiagnosisName>>(
    "/v1/masters/diagnosis-names",
    { params },
  );
  const items = Array.isArray(data) ? data : (data.data ?? []);
  return items.map(transformDiagnosisNameOption);
};

export const useGetDiagnosisTypes = () =>
  useQuery({
    queryKey: queryKeys.masters.category("diagnosis-types"),
    queryFn: getDiagnosisTypes,
    staleTime: QUERY_STALE_TIMES.STATIC,
    gcTime: QUERY_GC_TIMES.LONG,
  });

export const useGetDiagnosisNames = (typeId?: number | null) =>
  useQuery({
    queryKey: queryKeys.masters.diagnosisNames(typeId),
    queryFn: () => getDiagnosisNames(typeId),
    staleTime: QUERY_STALE_TIMES.STATIC,
    gcTime: QUERY_GC_TIMES.LONG,
  });
