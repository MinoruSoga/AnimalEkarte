import { afterEach, beforeEach, describe, expect, it } from "vitest";
import { renderHook, waitFor } from "@testing-library/react";
import { http, HttpResponse } from "msw";
import { CURRENT_CLINIC_STORAGE_KEY } from "@/lib/current-clinic";
import { server } from "@/testing/mocks/node";
import { createTestWrapper } from "@/testing/utils";
import { useGetReservationStaffs } from "./use-reservation-types";

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
