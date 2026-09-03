import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { renderHook, waitFor } from "@testing-library/react";
import { createElement, type ReactNode } from "react";
import { beforeEach, describe, expect, it, vi } from "vitest";

import { axios } from "@/lib/axios";
import { queryKeys } from "@/lib/query-keys";
import { transformStaffSelectorItem, useGetStaffs } from "./use-staffs";

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

describe("staff selector query cache (BUG-005 Mode3)", () => {
  beforeEach(() => {
    vi.mocked(axios.get).mockReset();
    vi.mocked(axios.get).mockResolvedValue({ data: rawStaffs });
  });

  it("uses a distinct key from masters.category(staffs)", () => {
    expect(queryKeys.masters.staffSelectorList()).toEqual(["masters", "staffs", "selector-list"]);
    expect(queryKeys.masters.category("staffs")).toEqual(["masters", "staffs"]);
    expect(queryKeys.masters.staffSelectorList()).not.toEqual(queryKeys.masters.category("staffs"));
  });

  it("stores thin selector shape under selector-list and ignores master key contents", async () => {
    const client = new QueryClient({
      defaultOptions: { queries: { retry: false } },
    });

    // Poison master key with an incompatible full-shape-looking object; selector must not read it
    client.setQueryData(queryKeys.masters.category("staffs"), [
      {
        id: "poison",
        name: "Poison Master",
        staffType: "doctor",
        isActive: true,
        email: "poison@example.invalid",
        clinicId: "9",
      },
    ]);

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

    const selectorCached = client.getQueryData(queryKeys.masters.staffSelectorList());
    const masterCached = client.getQueryData(queryKeys.masters.category("staffs"));
    expect(selectorCached).not.toBe(masterCached);
    expect((masterCached as { id: string }[])[0].id).toBe("poison");
  });

  it("master key population order cannot overwrite selector-list after fetch", async () => {
    const client = new QueryClient({
      defaultOptions: { queries: { retry: false } },
    });

    const { result } = renderHook(() => useGetStaffs(), {
      wrapper: createWrapper(client),
    });
    await waitFor(() => expect(result.current.isSuccess).toBe(true));

    client.setQueryData(queryKeys.masters.category("staffs"), [
      {
        id: "m1",
        name: "Master Only",
        staffType: "doctor",
        isActive: true,
        email: "m@example.invalid",
        clinicId: "1",
      },
    ]);

    const selectorCached = client.getQueryData(queryKeys.masters.staffSelectorList()) as {
      id: string;
      email?: string;
    }[];
    expect(selectorCached[0].id).toBe("1");
    expect(selectorCached[0].email).toBeUndefined();
  });
});
