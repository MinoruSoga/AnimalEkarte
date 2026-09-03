import { describe, expect, it } from "vitest";
import { renderHook, waitFor } from "@testing-library/react";
import { http, HttpResponse } from "msw";
import { PALETTE } from "@/lib/design-tokens";
import { server } from "@/testing/mocks/node";
import { createTestWrapper } from "@/testing/TestUtils";
import { useReservationTypeColorMap } from "./use-reservation-type-color-map";

const ACTIVE_TYPE = {
  id: 1,
  name: "一般診療",
  color: "#111111",
  is_active: true,
  duration_minutes: 30,
  sort_order: 1,
  is_internal: false,
  category: "general",
};

const INACTIVE_TYPE = {
  id: 2,
  name: "旧コース",
  color: "#ff6600",
  is_active: false,
  duration_minutes: 60,
  sort_order: 2,
  is_internal: false,
  category: "general",
};

describe("useReservationTypeColorMap (BUG-016)", () => {
  it("includes inactive categories in colorMap with isInactive and original color", async () => {
    server.use(
      http.get("/api/v1/masters/reservation-types", () =>
        HttpResponse.json([ACTIVE_TYPE, INACTIVE_TYPE]),
      ),
      http.get("/api/v1/masters/reservation-type-groups", () => HttpResponse.json([])),
    );

    const { result } = renderHook(() => useReservationTypeColorMap(), {
      wrapper: createTestWrapper(),
    });

    await waitFor(() => {
      expect(result.current.colorMap.has(INACTIVE_TYPE.name)).toBe(true);
    });

    const inactiveColor = result.current.colorMap.get(INACTIVE_TYPE.name);
    expect(inactiveColor).toBeDefined();
    expect(inactiveColor?.isInactive).toBe(true);
    expect(inactiveColor?.hex.toLowerCase()).toBe("#ff6600");
    expect(inactiveColor?.hex).not.toBe(PALETTE.grayMedium);
    expect(inactiveColor?.style.backgroundColor).not.toBe(PALETTE.mutedBg);

    const activeColor = result.current.colorMap.get(ACTIVE_TYPE.name);
    expect(activeColor).toBeDefined();
    expect(activeColor?.isInactive).not.toBe(true);
    expect(activeColor?.hex.toLowerCase()).toBe("#111111");
  });

  it("keeps inactive categories out of activeGroupEntries / activeEntries", async () => {
    server.use(
      http.get("/api/v1/masters/reservation-types", () =>
        HttpResponse.json([ACTIVE_TYPE, INACTIVE_TYPE]),
      ),
      http.get("/api/v1/masters/reservation-type-groups", () => HttpResponse.json([])),
    );

    const { result } = renderHook(() => useReservationTypeColorMap(), {
      wrapper: createTestWrapper(),
    });

    await waitFor(() => {
      expect(result.current.colorMap.size).toBeGreaterThan(0);
    });

    const legendNames = result.current.activeGroupEntries.map((entry) => entry.name);
    expect(legendNames).toContain(ACTIVE_TYPE.name);
    expect(legendNames).not.toContain(INACTIVE_TYPE.name);
    expect(result.current.activeEntries.map((entry) => entry.name)).not.toContain(
      INACTIVE_TYPE.name,
    );
  });
});
