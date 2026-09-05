import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { renderHook, waitFor } from "@testing-library/react";
import { createElement, type ReactNode } from "react";
import { beforeEach, describe, expect, it, vi } from "vitest";

import { axios } from "@/lib/axios";
import { queryKeys } from "@/lib/query-keys";
import { useGetStaffs as useGetSelectorStaffs } from "@/hooks/use-staffs";
import { transformStaff, useGetStaffs } from "./staffs";

vi.mock("@/lib/axios", () => ({
  axios: {
    get: vi.fn(),
    post: vi.fn(),
    patch: vi.fn(),
    delete: vi.fn(),
  },
}));
vi.mock("@/lib/handle-api-error", () => ({
  handleApiError: vi.fn(),
}));

function createWrapper(client: QueryClient) {
  return function Wrapper({ children }: { children: ReactNode }) {
    return createElement(QueryClientProvider, { client }, children);
  };
}

describe("master staff transform (fail-closed staff_type)", () => {
  it("never maps missing/null/unknown staff_type to doctor", () => {
    const cases = [
      { staff_type: "doctor" as const, want: "doctor" },
      { staff_type: "resource" as const, want: "resource" },
      { staff_type: "nurse" as const, want: "nurse" },
      { staff_type: "trimmer" as const, want: "trimmer" },
      { staff_type: undefined, want: "" },
      { staff_type: null, want: "" },
      { staff_type: "bogus", want: "bogus" },
    ];

    for (const tt of cases) {
      const row = transformStaff({
        id: 1,
        name: "X",
        is_active: true,
        staff_type: tt.staff_type as never,
      });
      expect(row.staffType).toBe(tt.want);
    }
  });

  it("keeps full master shape fields that selector list does not carry", () => {
    const row = transformStaff({
      id: 9,
      name: "Clinic Doctor",
      is_active: true,
      staff_type: "doctor",
      clinic_id: 3,
      email: "doc@example.invalid",
      reservation_display_name: "Dr X",
    } as never);
    expect(row).toMatchObject({
      id: "9",
      name: "Clinic Doctor",
      staffType: "doctor",
      email: "doc@example.invalid",
      clinicId: "3",
      reservationDisplayName: "Dr X",
    });
  });

  it("uses category(staffs) as the raw cache key (selector shares it via select)", () => {
    expect(queryKeys.masters.category("staffs")).toEqual(["masters", "staffs"]);
    expect(queryKeys.masters.staffSelectorList()).toEqual(["masters", "staffs", "selector-list"]);
  });

  it("master full shape is produced by select, not stored as the raw cache value", () => {
    const client = new QueryClient();
    const rawRows = [
      {
        id: 1,
        name: "Doc",
        is_active: true,
        staff_type: "doctor",
        email: "a@example.invalid",
      },
    ];
    client.setQueryData(queryKeys.masters.category("staffs"), rawRows);

    const cached = client.getQueryData(queryKeys.masters.category("staffs")) as {
      email?: string;
      is_active?: boolean;
    }[];
    expect(cached[0].email).toBe("a@example.invalid");
    expect(cached[0].is_active).toBe(true);
  });

  it("useGetStaffs fetches raw ModelStaff into category(staffs) and selects full shape", async () => {
    vi.mocked(axios.get).mockResolvedValue({
      data: [
        {
          id: 1,
          name: "Doc",
          is_active: true,
          staff_type: "doctor",
          email: "doc@example.invalid",
          clinic_id: 2,
        },
      ],
    });
    const client = new QueryClient({
      defaultOptions: { queries: { retry: false } },
    });

    const { result } = renderHook(() => useGetStaffs(), {
      wrapper: createWrapper(client),
    });
    await waitFor(() => expect(result.current.isSuccess).toBe(true));

    expect(result.current.data?.[0]).toMatchObject({
      id: "1",
      name: "Doc",
      staffType: "doctor",
      email: "doc@example.invalid",
    });
    expect(client.getQueryData(queryKeys.masters.category("staffs"))).toEqual([
      {
        id: 1,
        name: "Doc",
        is_active: true,
        staff_type: "doctor",
        email: "doc@example.invalid",
        clinic_id: 2,
      },
    ]);
  });

  it("selector and master CRUD share one HTTP fetch", async () => {
    vi.mocked(axios.get).mockResolvedValue({
      data: [
        { id: 1, name: "Doc", is_active: true, staff_type: "doctor", email: "d@example.invalid" },
      ],
    });
    const client = new QueryClient({
      defaultOptions: { queries: { retry: false } },
    });
    const wrapper = createWrapper(client);

    const selector = renderHook(() => useGetSelectorStaffs(), { wrapper });
    const master = renderHook(() => useGetStaffs(), { wrapper });

    await waitFor(() => {
      expect(selector.result.current.isSuccess).toBe(true);
      expect(master.result.current.isSuccess).toBe(true);
    });

    expect(axios.get).toHaveBeenCalledTimes(1);
    expect(selector.result.current.data?.[0]).toMatchObject({ id: "1", staffType: "doctor" });
    expect(Object.prototype.hasOwnProperty.call(selector.result.current.data?.[0], "email")).toBe(
      false,
    );
    expect(master.result.current.data?.[0]).toMatchObject({
      id: "1",
      email: "d@example.invalid",
    });
  });
});

beforeEach(() => {
  vi.mocked(axios.get).mockReset();
});
