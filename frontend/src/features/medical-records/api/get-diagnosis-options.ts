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
  const { data } = await axios.get<DiagnosisCategory[] | PaginatedResponse<DiagnosisCategory>>(
    "/v1/masters/diagnosis-categories",
    { params: { limit: 100 } },
  );
  const items = Array.isArray(data) ? data : (data.data ?? []);
  return items.map((item) => ({
    id: Number(item.id ?? 0),
    name: item.name,
  }));
};

export const getDiagnosisNames = async (categoryId?: number | null): Promise<DiagnosisNameOption[]> => {
  const params: Record<string, unknown> = { limit: 100 };
  if (categoryId) params.category_id = categoryId;
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

export const useGetDiagnosisCategories = () =>
  useQuery({
    queryKey: ["masters", "diagnosis-categories"],
    queryFn: getDiagnosisCategories,
    staleTime: QUERY_STALE_TIMES.STATIC,
    gcTime: QUERY_GC_TIMES.LONG,
  });

export const useGetDiagnosisNames = (categoryId?: number | null) =>
  useQuery({
    queryKey: ["masters", "diagnosis-names", categoryId ?? null],
    queryFn: () => getDiagnosisNames(categoryId),
    staleTime: QUERY_STALE_TIMES.STATIC,
    gcTime: QUERY_GC_TIMES.LONG,
  });
