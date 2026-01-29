import { useState, useMemo } from "react";
import { useNavigate } from "react-router";
import { MOCK_PETS } from "@/config/mock-data";
import type { Pet } from "@/types";
import type { PetSelectionSearchParams } from "@/components/shared/PetSelection";

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
  const [searchParams, setSearchParams] =
    useState<PetSelectionSearchParams>(INITIAL_SEARCH_PARAMS);

  const filteredPets = useMemo(() => {
    return MOCK_PETS.filter((pet) => {
      if (searchParams.ownerId && !pet.ownerId.includes(searchParams.ownerId))
        return false;
      if (
        searchParams.ownerName &&
        !pet.ownerName.includes(searchParams.ownerName)
      )
        return false;
      if (searchParams.phone && (!pet.phone || !pet.phone.includes(searchParams.phone)))
        return false;
      if (searchParams.petName && !pet.name.includes(searchParams.petName))
        return false;
      if (searchParams.species && !pet.species.includes(searchParams.species))
        return false;
      return true;
    });
  }, [searchParams]);

  const handleSearch = () => {
    // Reactive filter — no-op until API integration
  };

  const handleSelect = (pet: Pet) => {
    navigate(`${config.selectPath}?petId=${pet.id}`);
  };

  const handleBack = () => {
    navigate(config.backPath);
  };

  return {
    searchParams,
    setSearchParams,
    filteredPets,
    handleSearch,
    handleSelect,
    handleBack,
  };
}
