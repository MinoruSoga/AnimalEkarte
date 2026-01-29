import { useState, useMemo } from "react";
import { MOCK_PETS } from "@/config/mock-data";
import type { Pet } from "@/types";

type SelectionMode = "single" | "multiple" | "multiple-same-owner";

interface UsePetSelectionOptions {
  returnAllOnEmpty?: boolean;
  maxResults?: number;
}

export function usePetSelection(
  initialSelectedPets: Pet[] = [],
  mode: SelectionMode = "single",
  options: UsePetSelectionOptions = { returnAllOnEmpty: true, maxResults: 100 }
) {
  const [searchQuery, setSearchQuery] = useState("");
  const [selectedPets, setSelectedPets] = useState<Pet[]>(initialSelectedPets);

  const filteredPets = useMemo(() => {
    // If no query and we shouldn't return all, return empty
    if (!searchQuery && !options.returnAllOnEmpty) {
      return [];
    }

    // If no query and we CAN return all, return all (up to max)
    if (!searchQuery) {
      return MOCK_PETS.slice(0, options.maxResults || 100);
    }

    const lowerQuery = searchQuery.toLowerCase();
    const results = MOCK_PETS.filter(
      (pet) =>
        pet.ownerName.toLowerCase().includes(lowerQuery) ||
        pet.name.toLowerCase().includes(lowerQuery) ||
        pet.species.toLowerCase().includes(lowerQuery) ||
        // Add ID search
        pet.ownerId.toLowerCase().includes(lowerQuery) ||
        // Add ID search
        pet.id.toLowerCase().includes(lowerQuery)
    );

    return results.slice(0, options.maxResults || 100);
  }, [searchQuery, options.returnAllOnEmpty, options.maxResults]);

  const togglePetSelection = (pet: Pet) => {
    setSelectedPets((prev) => {
      const isSelected = prev.some((p) => p.id === pet.id);

      if (isSelected) {
        // Deselect
        return prev.filter((p) => p.id !== pet.id);
      } else {
        // Select
        if (mode === "single") {
          return [pet];
        } else if (mode === "multiple-same-owner") {
          // If selecting a pet from a different owner, reset selection to just this new pet
          // Otherwise add to selection
          if (prev.length > 0 && prev[0].ownerId !== pet.ownerId) {
            return [pet];
          }
          return [...prev, pet];
        } else {
          // multiple (unrestricted)
          return [...prev, pet];
        }
      }
    });
  };

  const isPetSelected = (pet: Pet) => selectedPets.some((p) => p.id === pet.id);

  return {
    searchQuery,
    setSearchQuery,
    selectedPets,
    setSelectedPets,
    filteredPets,
    togglePetSelection,
    isPetSelected,
  };
}
