import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { renderHook, waitFor } from "@testing-library/react";
import { createElement, type ReactNode } from "react";
import { beforeEach, describe, expect, it, vi } from "vitest";

import { axios } from "@/lib/axios";
import { queryKeys } from "@/lib/query-keys";
import { STAFFS_RAW_QUERY_KEY, transformStaffSelectorItem, useGetStaffs } from "./use-staffs";

vi.mock("@/lib/axios", () => ({
  axios: {
    get: vi.fn(),
  },
}));

const rawStaffs = [
  {
    id: 1,
    name: "Active Doctor",
    is_active: true,
    staff_type: "doctor",
    occupation: { name: "獣医師" },
  },
  {
    id: 2,
    name: "Resource Row",
    is_active: true,
    staff_type: "resource",
  },
  {
    id: 3,
    name: "Missing Type",
    is_active: true,
  },
  {
    id: 4,
    name: "Null Type",
    is_active: true,
    staff_type: null,
  },
  {
    id: 5,
    name: "Unknown Type",
    is_active: true,
    staff_type: "unknown-role",
  },
  {
    id: 6,
    name: "Inactive Doctor",
    is_active: false,
    staff_type: "doctor",
  },
] as const;

function createWrapper(client: QueryClient) {
  return function Wrapper({ children }: { children: ReactNode }) {
    return createElement(QueryClientProvider, { client }, children);
  };
}

describe("use-staffs selector transform (fail-closed staff_type)", () => {
  it("does not fail-open missing, null, resource, or unknown staff_type to doctor", () => {
    const items = rawStaffs.map((row) => transformStaffSelectorItem(row as never));
    expect(items.map((s) => s.staffType)).toEqual([
      "doctor",
      "resource",
      "",
      "",
      "unknown-role",
      "doctor",
    ]);
    const doctorCandidates = items.filter((s) => s.staffType === "doctor" && s.isActive);
    expect(doctorCandidates).toEqual([expect.objectContaining({ id: "1", name: "Active Doctor" })]);
  });
});

describe("staff selector shares raw masters.staffs cache (STG P0-3)", () => {
  beforeEach(() => {
    vi.mocked(axios.get).mockReset();
    vi.mocked(axios.get).mockResolvedValue({ data: rawStaffs });
  });

  it("keeps the unused selector-list key distinct for prefix-invalidate docs", () => {
    expect(queryKeys.masters.staffSelectorList()).toEqual(["masters", "staffs", "selector-list"]);
    expect(STAFFS_RAW_QUERY_KEY).toEqual(["masters", "staffs"]);
  });

  it("stores raw ModelStaff under category(staffs) and selects the thin shape", async () => {
    const client = new QueryClient({
      defaultOptions: { queries: { retry: false } },
    });

    const { result } = renderHook(() => useGetStaffs(), {
      wrapper: createWrapper(client),
    });
    await waitFor(() => expect(result.current.isSuccess).toBe(true));

    expect(result.current.data?.map((s) => s.id)).toEqual(["1", "2", "3", "4", "5", "6"]);
    expect(result.current.data?.[0]).toMatchObject({
      id: "1",
      staffType: "doctor",
      isActive: true,
    });
    expect(Object.prototype.hasOwnProperty.call(result.current.data?.[0], "email")).toBe(false);

    const rawCached = client.getQueryData(STAFFS_RAW_QUERY_KEY) as {
      id: number;
      is_active?: boolean;
    }[];
    expect(rawCached[0].id).toBe(1);
    expect(rawCached[0].is_active).toBe(true);
  });
});
