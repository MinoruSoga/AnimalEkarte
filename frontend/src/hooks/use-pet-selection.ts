import { useState, useMemo, useCallback } from "react";
import type { Pet } from "@/types";

type SelectionMode = "single" | "multiple" | "multiple-same-owner";

export function usePetSelection(initialSelectedPets: Pet[] = [], mode: SelectionMode = "single") {
  const [selectedPets, setSelectedPets] = useState<Pet[]>(initialSelectedPets);

  // js-set-map-lookups: O(1) ルックアップのため Set を事前構築
  const selectedPetIdSet = useMemo(() => new Set(selectedPets.map((p) => p.id)), [selectedPets]);

  const togglePetSelection = useCallback(
    (pet: Pet) => {
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
    },
    [mode],
  );

  const isPetSelected = useCallback((pet: Pet) => selectedPetIdSet.has(pet.id), [selectedPetIdSet]);

  return {
    selectedPets,
    setSelectedPets,
    togglePetSelection,
    isPetSelected,
  };
}
