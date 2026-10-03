import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";

import { axios } from "@/lib/axios";
import { isPersistedPetId } from "@/lib/pet-id";
import { queryKeys } from "@/lib/query-keys";
import { QUERY_GC_TIMES, QUERY_STALE_TIMES } from "@/lib/react-query";

/**
 * `/v1/pets/{id}/chronic-conditions` の wire 応答行（GET は包みなし bare array、
 * POST/PATCH 応答も同形）。Go ドメインモデル由来の generated/models ではなく
 * このエンドポイント専用 DTO とする（TASK-444-S1）。
 */
export interface PetChronicCondition {
  id: number;
  clinic_id: number;
  pet_id: number;
  condition_code: string;
  condition_name: string;
  diagnosed_at: string;
  notes?: string;
  is_active: boolean;
  created_at: string;
  updated_at: string;
}

/**
 * EMR-248: ペット慢性疾患フラグの CRUD。
 * エンドポイントはフラットな pet ルート `/v1/pets/{id}/chronic-conditions`
 * （clinic スコープは axios 側の X-Clinic-ID ヘッダで処理、パスには含めない）。
 * GET は包みなしの bare array を返す。
 */
export interface CreatePetChronicConditionRequest {
  condition_code: string;
  condition_name: string;
  /** strict YYYY-MM-DD（shared DatePicker 由来の値のみ） */
  diagnosed_at: string;
  notes?: string | null;
  is_active?: boolean;
}

export interface UpdatePetChronicConditionRequest {
  condition_code?: string;
  condition_name?: string;
  diagnosed_at?: string;
  notes?: string | null;
  /** PATCH は nil フィールドを省略するため、無効化は明示的に false を送る */
  is_active?: boolean;
}

export interface CreatePetChronicConditionVariables {
  petId: string;
  request: CreatePetChronicConditionRequest;
}

export interface UpdatePetChronicConditionVariables {
  petId: string;
  conditionId: number;
  request: UpdatePetChronicConditionRequest;
}

export interface DeletePetChronicConditionVariables {
  petId: string;
  conditionId: number;
}

async function getPetChronicConditions(petId: string): Promise<PetChronicCondition[]> {
  const { data } = await axios.get<PetChronicCondition[]>(`/v1/pets/${petId}/chronic-conditions`);
  return data;
}

async function createPetChronicCondition(
  petId: string,
  request: CreatePetChronicConditionRequest,
): Promise<PetChronicCondition> {
  const { data } = await axios.post<PetChronicCondition>(
    `/v1/pets/${petId}/chronic-conditions`,
    request,
  );
  return data;
}

async function updatePetChronicCondition(
  petId: string,
  conditionId: number,
  request: UpdatePetChronicConditionRequest,
): Promise<PetChronicCondition> {
  const { data } = await axios.patch<PetChronicCondition>(
    `/v1/pets/${petId}/chronic-conditions/${conditionId}`,
    request,
  );
  return data;
}

async function deletePetChronicCondition(petId: string, conditionId: number): Promise<void> {
  await axios.delete(`/v1/pets/${petId}/chronic-conditions/${conditionId}`);
}

export function useGetPetChronicConditions(petId: string) {
  // BUG-022 と同じく pending ペットの temp-* など非永続 ID では API を発行しない
  const canFetch = isPersistedPetId(petId);
  return useQuery({
    queryKey: queryKeys.petChronicConditions.list(petId),
    queryFn: () => getPetChronicConditions(petId),
    staleTime: QUERY_STALE_TIMES.STATIC,
    gcTime: QUERY_GC_TIMES.LONG,
    enabled: canFetch,
  });
}

function useInvalidatePetChronicConditions() {
  const queryClient = useQueryClient();
  return (petId: string) =>
    queryClient.invalidateQueries({
      queryKey: queryKeys.petChronicConditions.list(petId),
    });
}

export function useCreatePetChronicCondition() {
  const invalidate = useInvalidatePetChronicConditions();
  return useMutation({
    mutationFn: ({ petId, request }: CreatePetChronicConditionVariables) =>
      createPetChronicCondition(petId, request),
    onSuccess: async (_data, { petId }) => {
      await invalidate(petId);
    },
  });
}

export function useUpdatePetChronicCondition() {
  const invalidate = useInvalidatePetChronicConditions();
  return useMutation({
    mutationFn: ({ petId, conditionId, request }: UpdatePetChronicConditionVariables) =>
      updatePetChronicCondition(petId, conditionId, request),
    onSuccess: async (_data, { petId }) => {
      await invalidate(petId);
    },
  });
}

export function useDeletePetChronicCondition() {
  const invalidate = useInvalidatePetChronicConditions();
  return useMutation({
    mutationFn: ({ petId, conditionId }: DeletePetChronicConditionVariables) =>
      deletePetChronicCondition(petId, conditionId),
    onSuccess: async (_data, { petId }) => {
      await invalidate(petId);
    },
  });
}
