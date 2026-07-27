import { useState, useMemo, useCallback } from "react";
import { useLocation, useNavigate } from "react-router";
import { useGetPets } from "@/hooks/use-pet";
import { normalizedIncludes } from "@/lib/normalize-kana";
import type { Pet } from "@/types";
import type { PetSelectionSearchParams } from "@/components/shared/PetSelection/PetSelectionSearchForm";

interface PetSelectionPageConfig {
  /** 選択後の遷移先パス (例: "/examinations/new") */
  selectPath: string;
  /** 戻るボタンの遷移先 (例: "/examinations") */
  backPath: string;
}

const INITIAL_SEARCH_PARAMS: PetSelectionSearchParams = {
  ownerId: "",
  ownerName: "",
  ownerNameKana: "",
  phone: "",
  petName: "",
  petNameKana: "",
  species: "",
  address: "",
};

export function usePetSelectionPage(config: PetSelectionPageConfig) {
  const navigate = useNavigate();
  const location = useLocation();
  const [searchParams, setSearchParams] =
    useState<PetSelectionSearchParams>(INITIAL_SEARCH_PARAMS);

  // error / isLoading を破棄してはならない。破棄すると API 失敗が
  // 「該当0件」と区別できなくなり、利用者に嘘の検索結果を見せる。
  const {
    data: pets = [],
    error,
    isLoading,
  } = useGetPets(undefined, {
    includeDeceased: true,
  });

  const filteredPets = useMemo(() => {
    return pets.filter((pet) => {
      if (searchParams.ownerId && !pet.ownerId.includes(searchParams.ownerId))
        return false;
      if (searchParams.ownerName && !normalizedIncludes(pet.ownerName, searchParams.ownerName))
        return false;
      if (searchParams.ownerNameKana && (!pet.ownerNameKana || !normalizedIncludes(pet.ownerNameKana, searchParams.ownerNameKana)))
        return false;
      if (searchParams.phone && (!pet.phone || !pet.phone.includes(searchParams.phone)))
        return false;
      if (searchParams.address && (!pet.address || !pet.address.includes(searchParams.address)))
        return false;
      if (searchParams.petName && !normalizedIncludes(pet.name, searchParams.petName))
        return false;
      if (searchParams.petNameKana && (!pet.petNameKana || !normalizedIncludes(pet.petNameKana, searchParams.petNameKana)))
        return false;
      if (searchParams.species && !pet.species.includes(searchParams.species))
        return false;
      return true;
    });
  }, [pets, searchParams]);

  // フィルタはリアクティブ（useMemo）のため、ボタン押下時の追加処理は不要
  const handleSearch = useCallback(() => {}, []);

  const handleClear = useCallback(() => {
    setSearchParams(INITIAL_SEARCH_PARAMS);
  }, []);

  const handleSelect = useCallback((pet: Pet) => {
    if (pet.status === "死亡") return;

    const nextParams = new URLSearchParams(location.search);
    nextParams.set("petId", pet.id);
    navigate(`${config.selectPath}?${nextParams.toString()}`, { state: location.state });
  }, [navigate, config.selectPath, location.search, location.state]);

  const handleBack = useCallback(() => {
    navigate(config.backPath);
  }, [navigate, config.backPath]);

  return {
    searchParams,
    setSearchParams,
    filteredPets,
    error,
    isLoading,
    handleSearch,
    handleClear,
    handleSelect,
    handleBack,
  };
}
