import { describe, expect, it, afterEach } from "vitest";
import { http, HttpResponse } from "msw";
import { server } from "@/testing/mocks/node";
import { listReservationTypes } from "./reservation-types";

afterEach(() => {
  server.resetHandlers();
});

function makeRaw(
  id: number,
  name: string,
  isActive = true,
  overrides: Record<string, unknown> = {},
) {
  return {
    id,
    clinic_id: 1,
    name,
    is_active: isActive,
    color: "#3B82F6",
    description: "",
    sort_order: id,
    created_at: "2026-05-29T00:00:00Z",
    updated_at: "2026-05-29T00:00:00Z",
    reservation_display_name: "",
    duration_minutes: 15,
    short_name: "",
    show_short_name: false,
    reservation_visible: true,
    reservation_comment: "",
    reservation_image_url: "",
    reservation_day_option: "none",
    is_internal: false,
    category: "general",
    ...overrides,
  };
}

describe("listReservationTypes", () => {
  it("子なし root: isLeaf=true, depth=0, childIds=[]", async () => {
    server.use(
      http.get("/api/v1/masters/reservation-types", () =>
        HttpResponse.json([makeRaw(1, "一般診療")]),
      ),
    );

    const result = await listReservationTypes();

    expect(result).toHaveLength(1);
    expect(result[0]).toMatchObject({
      id: "1",
      name: "一般診療",
      isLeaf: true,
      depth: 0,
      childIds: [],
      parentId: undefined,
      parentName: undefined,
    });
  });

  it("子ノード: isLeaf=true, depth=1, parentId あり", async () => {
    server.use(
      http.get("/api/v1/masters/reservation-types", () =>
        HttpResponse.json([
          makeRaw(1, "LINEコース", true, {
            children: [makeRaw(2, "初診コース")],
          }),
        ]),
      ),
    );

    const result = await listReservationTypes();
    const child = result.find((t) => t.id === "2");

    expect(child).toBeDefined();
    expect(child).toMatchObject({
      id: "2",
      name: "初診コース",
      isLeaf: true,
      depth: 1,
      parentId: "1",
      parentName: "LINEコース",
      childIds: [],
    });
  });

  it("子あり root: isLeaf=false, depth=0, childIds=[...]", async () => {
    server.use(
      http.get("/api/v1/masters/reservation-types", () =>
        HttpResponse.json([
          makeRaw(1, "LINEコース", true, {
            children: [makeRaw(2, "初診コース"), makeRaw(3, "再診コース")],
          }),
        ]),
      ),
    );

    const result = await listReservationTypes();
    const root = result.find((t) => t.id === "1");

    expect(root).toBeDefined();
    expect(root).toMatchObject({
      isLeaf: false,
      depth: 0,
      childIds: ["2", "3"],
    });
  });

  it("2件の root + 各2件の children → 6件 flat", async () => {
    server.use(
      http.get("/api/v1/masters/reservation-types", () =>
        HttpResponse.json([
          makeRaw(1, "コースA", true, {
            children: [makeRaw(2, "コースA-1"), makeRaw(3, "コースA-2")],
          }),
          makeRaw(4, "コースB", true, {
            children: [makeRaw(5, "コースB-1"), makeRaw(6, "コースB-2")],
          }),
        ]),
      ),
    );

    const result = await listReservationTypes();

    expect(result).toHaveLength(6);
    expect(result.map((t) => t.id)).toEqual(["1", "2", "3", "4", "5", "6"]);
  });

  it("BE の順序が維持される（root → その子 → 次 root の順）", async () => {
    server.use(
      http.get("/api/v1/masters/reservation-types", () =>
        HttpResponse.json([
          makeRaw(10, "ルートA", true, {
            children: [makeRaw(20, "葉A-1"), makeRaw(30, "葉A-2")],
          }),
          makeRaw(40, "ルートB"),
        ]),
      ),
    );

    const result = await listReservationTypes();

    expect(result.map((t) => t.id)).toEqual(["10", "20", "30", "40"]);
  });
});
