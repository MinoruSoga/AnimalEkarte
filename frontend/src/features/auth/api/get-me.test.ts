import { describe, expect, it } from "vitest";

import { QUERY_STALE_TIMES } from "@/lib/react-query";

import { ME_QUERY_CACHE } from "./get-me";

describe("useGetMe cache policy (STG P0-2)", () => {
  it("does not poll, refetch on focus, or use a 10s session window", () => {
    expect(ME_QUERY_CACHE.staleTime).toBe(QUERY_STALE_TIMES.SESSION);
    expect(ME_QUERY_CACHE.staleTime).toBe(5 * 60 * 1000);
    expect(ME_QUERY_CACHE.refetchOnWindowFocus).toBe(false);
    expect(ME_QUERY_CACHE.refetchInterval).toBe(false);
  });
});
