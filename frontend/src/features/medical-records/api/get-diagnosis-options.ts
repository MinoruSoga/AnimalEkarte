import { useQuery } from "@tanstack/react-query";
import { axios } from "@/lib/axios";
import { QUERY_STALE_TIMES, QUERY_GC_TIMES } from "@/lib/react-query";
import type { DiagnosisType, DiagnosisName } from "@/types/generated/models";

export interface DiagnosisTypeOption {
  id: number;
  name: string;
}

export interface DiagnosisNameOption {
  id: number;
  name: string;
}

interface PaginatedResponse<T> {
  data: T[];
  total: number;
  page: number;
  limit: number;
}

export const getDiagnosisTypes = async (): Promise<DiagnosisTypeOption[]> => {
  const { data } = await axios.get<DiagnosisType[] | PaginatedResponse<DiagnosisType>>(
    "/v1/masters/diagnosis-types",
    { params: { limit: 100 } },
  );
  const items = Array.isArray(data) ? data : (data.data ?? []);
  return items.map((item) => ({
    id: Number(item.id ?? 0),
    name: item.name,
  }));
};

export const getDiagnosisNames = async (typeId?: number | null): Promise<DiagnosisNameOption[]> => {
  const params: Record<string, unknown> = { limit: 100 };
  if (typeId) params.type_id = typeId;
  const { data } = await axios.get<DiagnosisName[] | PaginatedResponse<DiagnosisName>>(
    "/v1/masters/diagnosis-names",
    { params },
  );
  const items = Array.isArray(data) ? data : (data.data ?? []);
  return items.map((item) => ({
    id: Number(item.id ?? 0),
    name: item.name,
  }));
};

export const useGetDiagnosisTypes = () =>
  useQuery({
    queryKey: ["masters", "diagnosis-types"],
    queryFn: getDiagnosisTypes,
    staleTime: QUERY_STALE_TIMES.STATIC,
    gcTime: QUERY_GC_TIMES.LONG,
  });

export const useGetDiagnosisNames = (typeId?: number | null) =>
  useQuery({
    queryKey: ["masters", "diagnosis-names", typeId ?? null],
    queryFn: () => getDiagnosisNames(typeId),
    staleTime: QUERY_STALE_TIMES.STATIC,
    gcTime: QUERY_GC_TIMES.LONG,
  });
