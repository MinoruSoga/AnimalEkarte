import { useQuery } from "@tanstack/react-query";
import { axios } from "@/lib/axios";
import { QUERY_STALE_TIMES, QUERY_GC_TIMES } from "@/lib/react-query";
import type { DiagnosisCategory, DiagnosisName } from "@/types/generated/models";

export interface DiagnosisCategoryOption {
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

export const getDiagnosisCategories = async (): Promise<DiagnosisCategoryOption[]> => {
  const { data } = await axios.get<PaginatedResponse<DiagnosisCategory>>(
    "/v1/masters/diagnosis-categories",
    { params: { limit: 100 } },
  );
  return (data.data ?? []).map((item) => ({
    id: Number(item.id ?? 0),
    name: item.name,
  }));
};

export const getDiagnosisNames = async (): Promise<DiagnosisNameOption[]> => {
  const { data } = await axios.get<PaginatedResponse<DiagnosisName>>(
    "/v1/masters/diagnosis-names",
    { params: { limit: 100 } },
  );
  return (data.data ?? []).map((item) => ({
    id: Number(item.id ?? 0),
    name: item.name,
  }));
};

export const useGetDiagnosisCategories = () =>
  useQuery({
    queryKey: ["masters", "diagnosis-categories"],
    queryFn: getDiagnosisCategories,
    staleTime: QUERY_STALE_TIMES.STATIC,
    gcTime: QUERY_GC_TIMES.LONG,
  });

export const useGetDiagnosisNames = () =>
  useQuery({
    queryKey: ["masters", "diagnosis-names"],
    queryFn: getDiagnosisNames,
    staleTime: QUERY_STALE_TIMES.STATIC,
    gcTime: QUERY_GC_TIMES.LONG,
  });
