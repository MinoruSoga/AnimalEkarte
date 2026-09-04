import { describe, expect, it } from "vitest";

import { isPersistedPetId } from "./pet-id";

describe("isPersistedPetId (BUG-022)", () => {
  it.each(["1", "7", "42", "pet-synth-1", "9007199254740991"])("accepts non-temp id %j", (id) => {
    expect(isPersistedPetId(id)).toBe(true);
  });

  it.each(["", "temp-1", "temp-1710000000000", null, undefined])(
    "rejects empty or temp id %j",
    (id) => {
      expect(isPersistedPetId(id)).toBe(false);
    },
  );
});
