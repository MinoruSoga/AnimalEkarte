import { useState } from "react";
import type { Pet } from "@/types";

type SelectionMode = "single" | "multiple" | "multiple-same-owner";

export function usePetSelection(
  initialSelectedPets: Pet[] = [],
  mode: SelectionMode = "single"
) {
  const [selectedPets, setSelectedPets] = useState<Pet[]>(initialSelectedPets);

  const togglePetSelection = (pet: Pet) => {
    setSelectedPets((prev) => {
      const isSelected = prev.some((p) => p.id === pet.id);

      if (isSelected) {
        return prev.filter((p) => p.id !== pet.id);
      } else {
        if (mode === "single") {
          return [pet];
        } else if (mode === "multiple-same-owner") {
          if (prev.length > 0 && prev[0].ownerId !== pet.ownerId) {
            return [pet];
          }
          return [...prev, pet];
        } else {
          return [...prev, pet];
        }
      }
    });
  };

  const isPetSelected = (pet: Pet) => selectedPets.some((p) => p.id === pet.id);

  return {
    selectedPets,
    setSelectedPets,
    togglePetSelection,
    isPetSelected,
  };
}
