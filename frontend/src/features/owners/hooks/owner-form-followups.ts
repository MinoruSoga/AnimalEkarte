export async function runOwnerCreateFollowups<T>(
  invalidateOwners: () => Promise<unknown>,
  createPets: ReadonlyArray<() => Promise<T>>,
): Promise<PromiseSettledResult<T>[]> {
  const [, petResults] = await Promise.all([
    invalidateOwners(),
    Promise.allSettled(createPets.map((createPet) => createPet())),
  ]);
  return petResults;
}
