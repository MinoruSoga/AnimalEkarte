import { describe, expect, it } from "vitest";

import { resolveTrimmingActiveFlag } from "./trimming";

describe("resolveTrimmingActiveFlag", () => {
  it("prefers is_active from the course wire", () => {
    expect(resolveTrimmingActiveFlag({ is_active: true, status: "inactive" })).toBe(true);
    expect(resolveTrimmingActiveFlag({ is_active: false, status: "active" })).toBe(false);
  });

  it("falls back to MasterItem.status when isActive is missing", () => {
    expect(resolveTrimmingActiveFlag({ status: "active" })).toBe(true);
    expect(resolveTrimmingActiveFlag({ status: "inactive" })).toBe(false);
  });
});
