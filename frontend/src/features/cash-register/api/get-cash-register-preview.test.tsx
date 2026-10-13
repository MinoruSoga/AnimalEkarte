import { describe, it, expect, vi, beforeEach } from "vitest";
import { renderHook, waitFor } from "@testing-library/react";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import type { ReactNode } from "react";

vi.mock("@/lib/axios", () => ({
  axios: { get: vi.fn() },
}));

import { axios } from "@/lib/axios";
import { useGetCashRegisterPreview } from "./get-cash-register-preview";

const mockedGet = vi.mocked(axios.get);

function wrapper({ children }: { children: ReactNode }) {
  const qc = new QueryClient({ defaultOptions: { queries: { retry: false } } });
  return <QueryClientProvider client={qc}>{children}</QueryClientProvider>;
}

// EMR-354: backend が明細行に返す category_amounts（split 支払額の部門別 GROSS 按分）を
// transformCloseBillingDetail が categoryAmounts として transport すること。
// RED 時点では transform がキーを捨てるため undefined になり失敗する。
// 欠落時は undefined のまま（旧ペイロード後方互換 → 表示側で記録なし）。

const DETAIL_BASE = {
  paid_at: "2026-05-01T01:00:00Z",
  owner_name: "田中太郎",
  pet_name: "ポチ",
  is_hospitalization: false,
  category: "examination",
  refund_amount: 0,
};

const PAYLOAD = {
  date: "2026-05-01",
  period: "am",
  period_start: "2026-05-01T09:00:00+09:00",
  period_end: "2026-05-01T13:00:00+09:00",
  is_already_closed: false,
  is_holiday: false,
  aggregate: {
    categories: {},
    payment_methods: [],
    theoretical_cash: 0,
    tax_breakdown: {
      standard: { taxable_amount: 0, tax_amount: 0 },
      reduced: { taxable_amount: 0, tax_amount: 0 },
    },
    unclassified_other_count: 0,
  },
  billing_details: [
    {
      ...DETAIL_BASE,
      billing_id: 7,
      payment_method_id: 2,
      payment_method_name: "現金",
      billing_amount: 1000,
      net_amount: 1000,
      category_amounts: { examination: 667, food: 333 },
    },
    {
      ...DETAIL_BASE,
      billing_id: 7,
      payment_method_id: 1,
      payment_method_name: "クレジットカード",
      billing_amount: 3000,
      net_amount: 3000,
      category_amounts: { examination: 2000, food: 1000 },
    },
    {
      ...DETAIL_BASE,
      billing_id: 8,
      payment_method_id: 2,
      payment_method_name: "現金",
      billing_amount: 500,
      net_amount: 500,
      // category_amounts 欠落（旧バージョンのペイロード想定）
    },
  ],
};

describe("useGetCashRegisterPreview — EMR-354 category_amounts transport", () => {
  beforeEach(() => {
    mockedGet.mockReset();
  });

  it("billing_details[].category_amounts を categoryAmounts として保持する", async () => {
    mockedGet.mockResolvedValue({ data: PAYLOAD });

    const { result } = renderHook(() => useGetCashRegisterPreview("2026-05-01", "am", true), {
      wrapper,
    });

    await waitFor(() => expect(result.current.isSuccess).toBe(true));

    const details = result.current.data?.billingDetails;
    expect(details).toHaveLength(3);
    expect(details?.[0]?.categoryAmounts).toEqual({
      examination: 667,
      food: 333,
    });
    expect(details?.[1]?.categoryAmounts).toEqual({
      examination: 2000,
      food: 1000,
    });
  });

  it("category_amounts 欠落行は undefined のまま（記録なし表示の区別子）", async () => {
    mockedGet.mockResolvedValue({ data: PAYLOAD });

    const { result } = renderHook(() => useGetCashRegisterPreview("2026-05-01", "am", true), {
      wrapper,
    });

    await waitFor(() => expect(result.current.isSuccess).toBe(true));

    const row = result.current.data?.billingDetails?.[2];
    expect(row?.billingId).toBe(8);
    expect(row?.categoryAmounts).toBeUndefined();
  });
});
