import { describe, it, expect, vi, beforeEach } from "vitest";
import { renderHook, waitFor } from "@testing-library/react";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import type { ReactNode } from "react";
import type { TreatmentPlanResponse } from "@/types/generated/hospitalization-responses";

const { mockGet } = vi.hoisted(() => ({
  mockGet: vi.fn(),
}));

vi.mock("@/lib/axios", () => ({
  axios: { get: mockGet },
}));

import { useGetTreatmentPlans } from "./get-treatment-plans";

function createWrapper() {
  const client = new QueryClient({
    defaultOptions: { queries: { retry: false } },
  });
  return function Wrapper({ children }: { children: ReactNode }) {
    return <QueryClientProvider client={client}>{children}</QueryClientProvider>;
  };
}

const wireFixture: TreatmentPlanResponse[] = [
  {
    id: "42",
    hospitalization_id: "7",
    treatment_content: "入院料",
    memo: "",
    is_insurance: true,
    unit_price: 1000,
    quantity: 1,
    discount_rate: 0,
    discount_amount: 0,
    subtotal: 1000,
    sort_order: 0,
    created_at: "2026-07-23T00:00:00+09:00",
    updated_at: "2026-07-23T00:00:00+09:00",
  },
];

describe("useGetTreatmentPlans", () => {
  beforeEach(() => {
    mockGet.mockReset();
  });

  it("GET /v1/hospitalizations/:id/treatment-plans を叩き wire 配列を返す", async () => {
    mockGet.mockResolvedValueOnce({ data: wireFixture });

    const { result } = renderHook(() => useGetTreatmentPlans("7"), {
      wrapper: createWrapper(),
    });

    await waitFor(() => expect(result.current.isSuccess).toBe(true));

    expect(mockGet).toHaveBeenCalledWith("/v1/hospitalizations/7/treatment-plans");
    expect(result.current.data).toEqual(wireFixture);
    // wire id は string（models.TreatmentPlan の number ではない）
    expect(typeof result.current.data?.[0]?.id).toBe("string");
  });

  it("hospitalizationId 未指定なら query を起動しない", () => {
    const { result } = renderHook(() => useGetTreatmentPlans(undefined), {
      wrapper: createWrapper(),
    });

    expect(result.current.fetchStatus).toBe("idle");
    expect(mockGet).not.toHaveBeenCalled();
  });
});
