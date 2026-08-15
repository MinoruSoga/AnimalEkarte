import { describe, expect, it } from "vitest";

import { parseInternalPath } from "./internal-navigation";

describe("parseInternalPath", () => {
  it.each([
    ["/", "/"],
    ["/owners/300588?tab=summary", "/owners/300588?tab=summary"],
    ["/reset-password#token=fragment-token", "/reset-password#token=fragment-token"],
  ])("accepts same-origin path %s", (candidate, expected) => {
    expect(parseInternalPath(candidate)).toBe(expected);
  });

  it.each([
    null,
    undefined,
    "",
    "owners",
    "https://evil.example",
    "//evil.example",
    "/\\evil.example",
    "/%5Cevil.example",
    "/owners\n/300588",
    " /owners",
  ])("rejects an unsafe navigation candidate %s", (candidate) => {
    expect(parseInternalPath(candidate)).toBeNull();
  });
});
