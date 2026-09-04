import { afterEach, beforeEach, describe, expect, it } from "vitest";
import { renderHook, waitFor } from "@testing-library/react";
import { http, HttpResponse } from "msw";
import { useQueryClient } from "@tanstack/react-query";
import { CURRENT_CLINIC_STORAGE_KEY } from "@/lib/current-clinic";
import { queryKeys } from "@/lib/query-keys";
import { server } from "@/testing/mocks/node";
import { createTestWrapper } from "@/testing/TestUtils";
import {
  useGetReservationAvailableTimes,
  useGetReservationStaffs,
  useGetReservationTypesGrouped,
} from "./use-reservation-types";

const CLINIC_ID = "17";

beforeEach(() => {
  localStorage.setItem(CURRENT_CLINIC_STORAGE_KEY, CLINIC_ID);
});

afterEach(() => {
  localStorage.removeItem(CURRENT_CLINIC_STORAGE_KEY);
});

describe("useGetReservationStaffs", () => {
  it("projects the positive capability contract without propagating legacy exclusion fields", async () => {
    server.use(
      http.get(`/api/v1/clinics/${CLINIC_ID}/reservation-staffs`, () =>
        HttpResponse.json([
          {
            id: 8,
            name: "予約担当",
            is_active: true,
            capable_courses: [{ id: 31, name: "一般診療" }],
            excluded_courses: [{ id: 32, name: "互換面" }],
            staff_type: "doctor",
          },
        ]),
      ),
    );

    const { result } = renderHook(() => useGetReservationStaffs(), {
      wrapper: createTestWrapper(),
    });

    await waitFor(() => expect(result.current.isSuccess).toBe(true));

    expect(result.current.data).toEqual([
      {
        id: 8,
        name: "予約担当",
        is_active: true,
        capable_courses: [{ id: 31, name: "一般診療" }],
      },
    ]);
    expect(result.current.data?.[0]).not.toHaveProperty("excluded_courses");
  });

  it("normalizes a missing capability array to empty and ignores legacy exclusions", async () => {
    server.use(
      http.get(`/api/v1/clinics/${CLINIC_ID}/reservation-staffs`, () =>
        HttpResponse.json([
          {
            id: 9,
            name: "未設定担当",
            is_active: true,
            excluded_courses: [],
          },
        ]),
      ),
    );

    const { result } = renderHook(() => useGetReservationStaffs(), {
      wrapper: createTestWrapper(),
    });

    await waitFor(() => expect(result.current.isSuccess).toBe(true));

    expect(result.current.data).toEqual([
      {
        id: 9,
        name: "未設定担当",
        is_active: true,
        capable_courses: [],
      },
    ]);
    expect(result.current.data?.[0]).not.toHaveProperty("excluded_courses");
  });

  it("normalizes a null capability array to empty without propagating legacy exclusions", async () => {
    server.use(
      http.get(`/api/v1/clinics/${CLINIC_ID}/reservation-staffs`, () =>
        HttpResponse.json([
          {
            id: 10,
            name: "null 設定担当",
            is_active: true,
            capable_courses: null,
            excluded_courses: [{ id: 41, name: "互換面" }],
          },
        ]),
      ),
    );

    const { result } = renderHook(() => useGetReservationStaffs(), {
      wrapper: createTestWrapper(),
    });

    await waitFor(() => expect(result.current.isSuccess).toBe(true));

    expect(result.current.data).toEqual([
      {
        id: 10,
        name: "null 設定担当",
        is_active: true,
        capable_courses: [],
      },
    ]);
    expect(result.current.data?.[0]).not.toHaveProperty("excluded_courses");
  });
});

const RESERVATION_TYPES_PAYLOAD = [
  {
    id: 1,
    name: "一般診療",
    color: "#111111",
    is_active: true,
    duration_minutes: 30,
    sort_order: 1,
    is_internal: false,
    category: "general",
    group_id: 10,
    group: { id: 10, name: "診療", color: "#111111" },
  },
  {
    id: 2,
    name: "旧コース",
    color: "#222222",
    is_active: false,
    duration_minutes: 60,
    sort_order: 2,
    is_internal: false,
    category: "general",
    group_id: 10,
    group: { id: 10, name: "診療", color: "#111111" },
  },
  {
    id: 3,
    name: "別の無効コース",
    color: "#333333",
    is_active: false,
    duration_minutes: 45,
    sort_order: 3,
    is_internal: false,
    category: "general",
    group_id: 10,
    group: { id: 10, name: "診療", color: "#111111" },
  },
];

describe("useGetReservationTypesGrouped (BUG-015)", () => {
  it("default: keeps only active reservation types", async () => {
    server.use(
      http.get("/api/v1/masters/reservation-types", () =>
        HttpResponse.json(RESERVATION_TYPES_PAYLOAD),
      ),
    );

    const { result } = renderHook(() => useGetReservationTypesGrouped(), {
      wrapper: createTestWrapper(),
    });

    await waitFor(() => expect(result.current.isSuccess).toBe(true));

    const names = result.current.data?.flatMap((g) => g.types.map((t) => t.name)) ?? [];
    expect(names).toEqual(["一般診療"]);
  });

  it("selectedTypeId: keeps the selected inactive type and excludes other inactive types", async () => {
    server.use(
      http.get("/api/v1/masters/reservation-types", () =>
        HttpResponse.json(RESERVATION_TYPES_PAYLOAD),
      ),
    );

    const { result } = renderHook(() => useGetReservationTypesGrouped(2), {
      wrapper: createTestWrapper(),
    });

    await waitFor(() => expect(result.current.isSuccess).toBe(true));

    const types = result.current.data?.flatMap((g) => g.types) ?? [];
    expect(types.map((t) => t.name).sort()).toEqual(["一般診療", "旧コース"].sort());
    expect(types.find((t) => t.id === 2)?.is_active).toBe(false);
    expect(types.some((t) => t.id === 3)).toBe(false);
  });
});

describe("useGetReservationAvailableTimes (BUG-015)", () => {
  it("sets meta.silentError so QueryCache can skip the global fetch toast", async () => {
    server.use(
      http.get("/api/v1/reservations/available-times", () =>
        HttpResponse.json([{ start_time: "0945", end_time: "1045" }]),
      ),
    );

    const { result } = renderHook(
      () => {
        const query = useGetReservationAvailableTimes("2", "2026-06-01", null);
        const queryClient = useQueryClient();
        return { query, queryClient };
      },
      { wrapper: createTestWrapper() },
    );

    await waitFor(() => expect(result.current.query.isSuccess).toBe(true));

    const cached = result.current.queryClient.getQueryCache().find({
      queryKey: queryKeys.reservations.availableTimes("2", "2026-06-01"),
    });
    expect(cached?.meta?.silentError).toBe(true);
  });
});
