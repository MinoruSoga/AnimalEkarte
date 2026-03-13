import { useGetPets, useGetPet } from "@/features/pets/api";

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
