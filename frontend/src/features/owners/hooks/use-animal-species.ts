import { useMemo } from "react";
import { useQuery } from "@tanstack/react-query";
import { axios } from "@/lib/axios";
import { QUERY_STALE_TIMES, QUERY_GC_TIMES } from "@/lib/react-query";
import type { AnimalSpecies as BackendAnimalSpecies } from "@/types/generated/models";

interface AnimalSpeciesOption extends BackendAnimalSpecies {
  label: string;
  isInactive?: boolean;
}

/**
 * 動物種マスタを取得
 * ペット編集モード時には削除済み種類も取得し、グレーアウト表示用のラベルを付与
 *
 * BUG-321: 削除済み動物種の表示状態を確定化
 */
export const useAnimalSpecies = (opts?: { includeInactive?: boolean }) => {
  const { data: allSpecies, isLoading, isError, error } = useQuery({
    queryKey: ["masters", "animal-species", opts?.includeInactive ? "all" : "active"],
    queryFn: async () => {
      const { data } = await axios.get<BackendAnimalSpecies[]>(
        "/v1/masters/animal-species"
      );
      return data;
    },
    staleTime: QUERY_STALE_TIMES.STATIC,
    gcTime: QUERY_GC_TIMES.LONG,
  });

  // 編集モード時に削除済み種類を判別し、ラベル付与
  const speciesOptions = useMemo<AnimalSpeciesOption[]>(() => {
    if (!allSpecies) return [];

    return allSpecies.map((species) => {
      const isInactive = species.is_active === false;
      return {
        ...species,
        isInactive,
        label: isInactive ? `${species.name} (利用不可)` : species.name,
      };
    });
  }, [allSpecies]);

  // 新規ペット登録時は有効な種類のみ
  const activeSpecies = useMemo<AnimalSpeciesOption[]>(() => {
    return speciesOptions.filter((s) => !s.isInactive);
  }, [speciesOptions]);

  return {
    allSpecies: speciesOptions,
    activeSpecies,
    isLoading,
    isError,
    error,
  };
};
