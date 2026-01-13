import { useMemo } from "react";
import { Pet } from "../../../types";
import { MOCK_PETS } from "../../../lib/constants";

export function usePets(searchTerm: string) {
  const filteredPets = useMemo(() => {
    let data = MOCK_PETS;
    
    if (searchTerm) {
        const lowerTerm = searchTerm.toLowerCase();
        data = data.filter((pet) => {
          return (
            pet.ownerName.toLowerCase().includes(lowerTerm) ||
            pet.ownerId.includes(lowerTerm) ||
            pet.name.toLowerCase().includes(lowerTerm) ||
            (pet.species && pet.species.toLowerCase().includes(lowerTerm))
          );
        });
    }
    return data;
  }, [searchTerm]);

  return { data: filteredPets };
}
