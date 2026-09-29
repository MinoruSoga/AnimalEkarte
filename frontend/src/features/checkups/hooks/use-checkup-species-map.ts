import { useCallback, useMemo } from "react";
import { useQueries } from "@tanstack/react-query";
import { axios } from "@/lib/axios";
import { isPersistedPetId } from "@/lib/pet-id";
import { queryKeys } from "@/lib/query-keys";
import { QUERY_GC_TIMES, QUERY_STALE_TIMES } from "@/lib/react-query";
import { transformBackendPetToFrontend } from "@/lib/transforms/pet";
import type { Pet } from "@/types";
import type { PetResponse } from "@/types/generated/pet-responses";
import type { CheckupRecord } from "../api/transforms";

/**
 * EMR-223: 表示中ページの行から「petId → 動物種名」マップを組み立てる。
 *
 * GET /v1/checkups の行は動物種を持たない（BackendCheckupGlobal に species 系フィールド無し）ため、
 * 行の petId をキーにペット詳細を useQueries で引き、Pet.species（animal_species.name）を使う。
 * queryKey は useGetPet（src/hooks/use-pet.ts）と同一の queryKeys.pets.detail、queryFn も同じ
 * GET /v1/pets/:id + transformBackendPetToFrontend 形なのでキャッシュ・無効化を共有する。
 * BE findPetByID は AnimalSpecies を Preload 済みで、種名はレスポンス埋め込みで届く。
 *
 * enabled=false の間は一切 fetch しない（種フィルタ未適用時のリクエストを発生させない）。
 * フィルタ適用中に未解決の petId はマップに載らず、filterCheckupsBySpecies がその行を除外する。
 */
export function useCheckupSpeciesMap(
  checkups: CheckupRecord[],
  enabled: boolean,
): Map<string, string> {
  // BUG-022 系ガード: ローカル pending の temp-* ID は /v1/pets/:id に送らない
  const petIds = useMemo(
    () => [
      ...new Set(checkups.map((c) => c.petId).filter((id): id is string => isPersistedPetId(id))),
    ],
    [checkups],
  );

  // combine は results / 関数参照が変わった時だけ再評価され、それ以外の render では
  // 前回の Map をそのまま返す（下流の useMemo が無駄に破棄されない）。
  const combineSpeciesMap = useCallback(
    (results: readonly { data?: Pet }[]) => {
      const map = new Map<string, string>();
      results.forEach((result, index) => {
        const petId = petIds[index];
        const name = result.data?.species;
        if (petId && name) map.set(petId, name);
      });
      return map;
    },
    [petIds],
  );

  return useQueries({
    queries: petIds.map((petId) => ({
      queryKey: queryKeys.pets.detail(petId),
      queryFn: async (): Promise<Pet> => {
        const { data } = await axios.get<PetResponse>(`/v1/pets/${petId}`);
        return transformBackendPetToFrontend(data);
      },
      enabled,
      staleTime: QUERY_STALE_TIMES.STATIC,
      gcTime: QUERY_GC_TIMES.LONG,
    })),
    combine: combineSpeciesMap,
  });
}
