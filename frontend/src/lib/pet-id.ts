/**
 * Local pending pets on /owners/new use client-only IDs `temp-${Date.now()}`
 * (BUG-022). Those must never be sent in /v1/pets/:id paths.
 * Server pet IDs are opaque strings (decimal ints in current BE); do not
 * over-constrain shape beyond rejecting the known client temp prefix.
 */
export function isPersistedPetId(petId: string | null | undefined): boolean {
  if (petId == null || petId === "") {
    return false;
  }
  return !petId.startsWith("temp-");
}
