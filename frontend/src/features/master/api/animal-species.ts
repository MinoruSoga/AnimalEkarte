import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { axios } from "@/lib/axios";
import { handleApiError } from "@/lib/handle-api-error";
import { QUERY_STALE_TIMES, QUERY_GC_TIMES } from "@/lib/react-query";
import type { AnimalSpecies as ModelAnimalSpecies } from "@/types/generated/models";

// ─────────────────────────────────────────────────
// Transform → domain type (camelCase)
// ─────────────────────────────────────────────────

function transformAnimalSpecies(data: ModelAnimalSpecies) {
  return {
    id: String(data.id ?? 0),
    name: data.name,
    isActive: data.is_active,
    sortOrder: data.sort_order,
    createdAt: data.created_at,
    updatedAt: data.updated_at,
  };
}

export type AnimalSpecies = ReturnType<typeof transformAnimalSpecies>;

// ─────────────────────────────────────────────────
// Request types
// ─────────────────────────────────────────────────

export interface CreateAnimalSpeciesRequest {
  name: string;
  is_active: boolean;
  sort_order: number;
}

export type UpdateAnimalSpeciesRequest = Partial<CreateAnimalSpeciesRequest>;

export interface ReorderAnimalSpeciesRequest {
  ids: number[];
}

// ─────────────────────────────────────────────────
// Query key
// ─────────────────────────────────────────────────

const ANIMAL_SPECIES_KEY = ["masters", "animal-species"] as const;

// ─────────────────────────────────────────────────
// API functions
// ─────────────────────────────────────────────────

export async function listAnimalSpecies(): Promise<AnimalSpecies[]> {
  const { data } = await axios.get<ModelAnimalSpecies[]>(
    "/v1/masters/animal-species",
  );
  return data.map(transformAnimalSpecies);
}

export async function createAnimalSpecies(
  req: CreateAnimalSpeciesRequest,
): Promise<AnimalSpecies> {
  const { data } = await axios.post<ModelAnimalSpecies>(
    "/v1/masters/animal-species",
    req,
  );
  return transformAnimalSpecies(data);
}

export async function updateAnimalSpecies(
  id: string,
  req: UpdateAnimalSpeciesRequest,
): Promise<AnimalSpecies> {
  const { data } = await axios.patch<ModelAnimalSpecies>(
    `/v1/masters/animal-species/${id}`,
    req,
  );
  return transformAnimalSpecies(data);
}

export async function deleteAnimalSpecies(id: string): Promise<void> {
  await axios.delete(`/v1/masters/animal-species/${id}`);
}

export async function reorderAnimalSpecies(
  req: ReorderAnimalSpeciesRequest,
): Promise<void> {
  await axios.patch("/v1/masters/animal-species/reorder", req);
}

// ─────────────────────────────────────────────────
// TanStack Query hooks
// ─────────────────────────────────────────────────

export function useGetAnimalSpecies() {
  return useQuery({
    queryKey: ANIMAL_SPECIES_KEY,
    queryFn: listAnimalSpecies,
    staleTime: QUERY_STALE_TIMES.STATIC,
    gcTime: QUERY_GC_TIMES.LONG,
  });
}

export function useCreateAnimalSpecies() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: createAnimalSpecies,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ANIMAL_SPECIES_KEY });
    },
    onError: (error) => handleApiError(error, "作成"),
  });
}

export function useUpdateAnimalSpecies() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: ({ id, req }: { id: string; req: UpdateAnimalSpeciesRequest }) =>
      updateAnimalSpecies(id, req),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ANIMAL_SPECIES_KEY });
    },
    onError: (error) => handleApiError(error, "更新"),
  });
}

export function useDeleteAnimalSpecies() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: deleteAnimalSpecies,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ANIMAL_SPECIES_KEY });
    },
    onError: (error) => handleApiError(error, "削除"),
  });
}

export function useReorderAnimalSpecies() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: reorderAnimalSpecies,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ANIMAL_SPECIES_KEY });
    },
    onError: (error) => handleApiError(error, "並び替え"),
  });
}
