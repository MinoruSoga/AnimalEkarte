import { useGetOwner as _useGetOwner } from "@/features/owners";

/**
 * Shared hook for fetching a single owner by ID.
 * Wraps features/owners useGetOwner for cross-feature use (avoids feature→feature import).
 */
export function useGetOwner(ownerId: string) {
  const { data: owner, isLoading, error } = _useGetOwner(ownerId);

  return {
    owner,
    isLoading,
    error,
  };
}
