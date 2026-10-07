import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { axios } from "@/lib/axios";
import { handleApiError } from "@/lib/handle-api-error";
import { QUERY_STALE_TIMES, QUERY_GC_TIMES } from "@/lib/react-query";
import { queryKeys } from "@/lib/query-keys";
import type { Cage as ModelCage } from "@/types/generated/models";

// Strict union types (models.ts uses string, but we keep these for form safety)
export type CageType = "icu" | "dog" | "cat" | "general";
export type CageSize = "small" | "medium" | "large";

// Request types (derived from models.ts, with strict unions for cage_type/cage_size)
type CageRequestBase = Omit<
  ModelCage,
  "id" | "clinic_id" | "created_at" | "updated_at" | "cage_type" | "cage_size"
> & {
  cage_type: CageType;
  cage_size: CageSize;
};

export type CreateCageRequest = Pick<CageRequestBase, "name" | "cage_type" | "cage_size"> &
  Partial<Omit<CageRequestBase, "name" | "cage_type" | "cage_size">>;

export type UpdateCageRequest = Partial<CageRequestBase>;

// EMR-306: merged_into_cage_id は UNIT-U1 の codegen まで generated ModelCage に
// 存在しないため、wire 応答のみローカル intersection で拡張する（本ファイルは
// generated-models-import-allowlist 登録済み）。codegen で実フィールドが生成
// された後も intersection は実フィールドと同型で競合せず、mapping 行はそのまま
// 実フィールドを参照するため動作は変わらない。
type CageWire = ModelCage & { merged_into_cage_id?: number | null };

function transformCage(data: CageWire) {
  return {
    id: String(data.id ?? 0),
    name: data.name,
    cageType: data.cage_type as CageType,
    cageSize: data.cage_size as CageSize,
    price: data.price ?? 0,
    isActive: data.is_active,
    description: data.description,
    sortOrder: data.sort_order,
    mergedIntoCageId: data.merged_into_cage_id ?? null,
    createdAt: data.created_at,
    updatedAt: data.updated_at,
  };
}

export type Cage = ReturnType<typeof transformCage>;

const getAllCages = async (): Promise<Cage[]> => {
  const { data } = await axios.get<CageWire[]>("/v1/masters/cages");
  return data.map(transformCage);
};

export const useGetAllCages = () => {
  return useQuery({
    queryKey: queryKeys.masters.category("cages"),
    queryFn: getAllCages,
    staleTime: QUERY_STALE_TIMES.STATIC,
    gcTime: QUERY_GC_TIMES.LONG,
  });
};

const createCage = async (req: CreateCageRequest): Promise<Cage> => {
  const { data } = await axios.post<CageWire>("/v1/masters/cages", req);
  return transformCage(data);
};

export const useCreateCage = () => {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: createCage,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.masters.category("cages") });
    },
    onError: (error) => handleApiError(error, "作成"),
  });
};

const updateCage = async (id: string, req: UpdateCageRequest): Promise<Cage> => {
  const { data } = await axios.patch<CageWire>(`/v1/masters/cages/${id}`, req);
  return transformCage(data);
};

export const useUpdateCage = () => {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: ({ id, req }: { id: string; req: UpdateCageRequest }) => updateCage(id, req),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.masters.category("cages") });
    },
    onError: (error) => handleApiError(error, "更新"),
  });
};

const deleteCage = async (id: string): Promise<void> => {
  await axios.delete(`/v1/masters/cages/${id}`);
};

export type ReorderCagesRequest = { ids: number[] };

const reorderCages = async (req: ReorderCagesRequest): Promise<void> => {
  await axios.patch("/v1/masters/cages/reorder", req);
};

export const useDeleteCage = () => {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: deleteCage,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.masters.category("cages") });
    },
    onError: (error) => handleApiError(error, "削除"),
  });
};

export const useReorderCages = () => {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (req: ReorderCagesRequest) => reorderCages(req),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.masters.category("cages") });
    },
    onError: (error) => handleApiError(error, "並び替え"),
  });
};
