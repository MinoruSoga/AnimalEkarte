import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { renderHook, waitFor } from "@testing-library/react";
import { createElement, type ReactNode } from "react";
import { beforeEach, describe, expect, it, vi } from "vitest";

import { axios } from "@/lib/axios";
import { queryKeys } from "@/lib/query-keys";
import { useStaffValidation } from "./use-staff-validation";

vi.mock("@/lib/axios", () => ({
  axios: {
    get: vi.fn(),
  },
}));

function createWrapper(client: QueryClient) {
  return function Wrapper({ children }: { children: ReactNode }) {
    return createElement(QueryClientProvider, { client }, children);
  };
}

describe("useStaffValidation (STG P0-3)", () => {
  beforeEach(() => {
    vi.mocked(axios.get).mockReset();
    vi.mocked(axios.get).mockResolvedValue({
      data: [
        { id: 1, name: "Active Doctor", is_active: true, staff_type: "doctor" },
        { id: 2, name: "Inactive Doctor", is_active: false, staff_type: "doctor" },
      ],
    });
  });

  it("shares the raw staffs cache and does not fetch masters.staff", async () => {
    const client = new QueryClient({
      defaultOptions: { queries: { retry: false } },
    });
    const { result } = renderHook(() => useStaffValidation(), {
      wrapper: createWrapper(client),
    });

    await waitFor(() => expect(result.current.isValidStaff("Active Doctor")).toBe(true));

    expect(result.current.isValidStaff("Inactive Doctor")).toBe(false);
    expect(vi.mocked(axios.get)).toHaveBeenCalledTimes(1);
    expect(vi.mocked(axios.get).mock.calls[0]?.[0]).toBe("/v1/masters/staffs");
    expect(client.getQueryData(queryKeys.masters.category("staffs"))).toBeDefined();
    expect(client.getQueryData(queryKeys.masters.category("staff"))).toBeUndefined();
  });
});
