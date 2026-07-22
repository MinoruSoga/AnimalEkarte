import { describe, expect, it, vi } from "vitest";

import { runOwnerCreateFollowups } from "./owner-form-followups";

function deferred<T>() {
  let resolve!: (value: T) => void;
  const promise = new Promise<T>((res) => {
    resolve = res;
  });
  return { promise, resolve };
}

describe("runOwnerCreateFollowups", () => {
  it("owner一覧のinvalidateを待たずに全pet作成を開始する", async () => {
    const invalidation = deferred<void>();
    const firstPet = deferred<string>();
    const secondPet = deferred<string>();
    const invalidateOwners = vi.fn(() => invalidation.promise);
    const createFirstPet = vi.fn(() => firstPet.promise);
    const createSecondPet = vi.fn(() => secondPet.promise);

    const pending = runOwnerCreateFollowups(invalidateOwners, [createFirstPet, createSecondPet]);

    expect(invalidateOwners).toHaveBeenCalledOnce();
    expect(createFirstPet).toHaveBeenCalledOnce();
    expect(createSecondPet).toHaveBeenCalledOnce();

    firstPet.resolve("first");
    secondPet.resolve("second");
    invalidation.resolve();

    await expect(pending).resolves.toEqual([
      { status: "fulfilled", value: "first" },
      { status: "fulfilled", value: "second" },
    ]);
  });
});
