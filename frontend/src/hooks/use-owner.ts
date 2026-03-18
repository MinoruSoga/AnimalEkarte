import { useGetOwner } from "@/features/owners/api/get-owner";

/**
 * Hook for fetching a single owner by ID with React Query cache.
 * Wraps useGetOwner for ergonomic usage across features.
 */
export function useOwnerInfo(ownerId: string) {
  const { data: owner, isLoading, error } = useGetOwner(ownerId);

  return {
    owner,
    isLoading,
    error,
  };
}
