import { useGetPets, useGetPet } from "@/features/pets";

export { useGetPet };

/**
 * Shared hook for searching/listing pets, optionally filtered by ownerId.
 * Wraps useGetPets for cross-feature use.
 */
export function useSearchPets(ownerId?: string) {
  const { data: pets, isLoading, error, isPending } = useGetPets(ownerId);

  return {
    pets: pets ?? [],
    isLoading,
    error,
    isPending,
  };
}
