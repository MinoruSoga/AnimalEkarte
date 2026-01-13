import { useState, useMemo } from "react";
import { useNavigate } from "react-router-dom";
import { Pet } from "../../../types";
import { MOCK_PETS } from "../../../lib/constants";

export const useHospitalizationPetSelection = () => {
  const navigate = useNavigate();
  const [searchParams, setSearchParams] = useState({
    ownerId: "",
    ownerName: "",
    ownerNameKana: "",
    phone: "",
    petName: "",
    petNameKana: "",
    species: "",
    address: "",
  });

  const filteredPets = useMemo(() => {
    return MOCK_PETS.filter((pet) => {
      if (searchParams.ownerId && !pet.ownerId.includes(searchParams.ownerId)) return false;
      if (searchParams.ownerName && !pet.ownerName.includes(searchParams.ownerName)) return false;
      if (searchParams.phone && (!pet.phone || !pet.phone.includes(searchParams.phone))) return false;
      if (searchParams.petName && !pet.name.includes(searchParams.petName)) return false;
      if (searchParams.species && !pet.species.includes(searchParams.species)) return false;
      return true;
    });
  }, [searchParams]);

  const handleSearch = () => {
    console.log("検索実行:", searchParams);
  };

  const handleSelect = (pet: Pet) => {
    navigate(`/hospitalization/new?petId=${pet.id}`);
  };

  const handleBack = () => {
    navigate("/hospitalization");
  };

  return {
    searchParams,
    setSearchParams,
    filteredPets,
    handleSearch,
    handleSelect,
    handleBack
  };
};
