import { describe, expect, it } from "vitest";

import { clinicalGeneralRoutes } from "./clinical-general-routes";

describe("BUG-20260906-003 aggregation route", () => {
  it("/aggregation は element と lazy Component を同一 route に置かない", () => {
    const route = clinicalGeneralRoutes.find((entry) => entry.path === "/aggregation");
    expect(route).toBeDefined();
    expect(route?.lazy).toBeUndefined();
    expect(route?.element).toBeTruthy();
    expect(route?.children?.[0]?.index).toBe(true);
    expect(route?.children?.[0]?.lazy).toEqual(expect.any(Function));
  });
});
