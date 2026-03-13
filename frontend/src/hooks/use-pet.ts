import { useState } from "react";
import { useGetPets, useGetPet } from "@/features/pets/api";
import type { Pet } from "@/types";

export { useGetPet };

/**
 * Hook for searching/listing pets, optionally filtered by ownerId.
 * Wraps useGetPets with loading and error state.
 */
export function usePetSearch(ownerId?: string) {
  const { data: pets, isLoading, error, isPending } = useGetPets(ownerId);

  return {
    pets: pets ?? [],
    isLoading,
    error,
    isPending,
  };
}

/**
 * Hook for managing a single selected pet in UI state.
 */
export function usePetSelection() {
  const [selectedPet, setSelectedPet] = useState<Pet | null>(null);

  const selectPet = (pet: Pet) => {
    setSelectedPet(pet);
  };

  const clearSelection = () => {
    setSelectedPet(null);
  };

  return {
    selectedPet,
    selectPet,
    clearSelection,
  };
}

/**
 * Hook for fetching a single pet by ID with React Query cache.
 * Wraps useGetPet for ergonomic usage.
 */
export function usePetInfo(petId: string) {
  const { data: pet, isLoading, error } = useGetPet(petId);

  return {
    pet,
    isLoading,
    error,
  };
}
