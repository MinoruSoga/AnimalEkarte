import { describe, expect, it, vi, beforeEach } from "vitest";

import {
  completeAccounting,
  createAccountingCompletionIdempotencyKey,
} from "./complete-accounting";

const postMock = vi.fn();

vi.mock("@/lib/axios", () => ({
  axios: {
    post: (...args: unknown[]) => postMock(...args),
  },
}));

describe("completeAccounting", () => {
  beforeEach(() => {
    postMock.mockReset();
    postMock.mockResolvedValue({
      data: {
        id: 99,
        clinic_id: 1,
        status: "completed",
        scheduled_date: "2026-08-01",
        subtotal: 1000,
        tax_total: 100,
        total_amount: 1100,
        has_insurance: false,
        memo: "",
        created_at: "2026-08-01T00:00:00Z",
        updated_at: "2026-08-01T00:00:00Z",
        total_refunded_amount: 0,
        items: [],
        payments: [],
      },
    });
  });

  it("POST /v1/accountings/complete を1回だけ呼び Idempotency-Key を付与する", async () => {
    const key = "11111111-1111-4111-8111-111111111111";
    await completeAccounting(
      {
        pet_id: 1,
        owner_id: 2,
        scheduled_date: "2026-08-01T00:00:00+09:00",
        items: [
          {
            category: "other",
            name: "診察料",
            unit_price: 1000,
            quantity: 1,
            discount_rate: 10,
            discount_amount: 100,
            tax_type: "excluded",
            tax_rate: 0.1,
            is_insurance_applicable: false,
            source: "manual",
          },
        ],
        payment_splits: [
          {
            method: "cash",
            amount: 1100,
            received_amount: 2000,
            change_amount: 900,
          },
        ],
      },
      key,
    );

    expect(postMock).toHaveBeenCalledTimes(1);
    expect(postMock).toHaveBeenCalledWith(
      "/v1/accountings/complete",
      expect.objectContaining({
        pet_id: 1,
        owner_id: 2,
        items: expect.arrayContaining([
          expect.objectContaining({
            name: "診察料",
            unit_price: 1000,
            discount_rate: 10,
            discount_amount: 100,
          }),
        ]),
      }),
      {
        headers: { "Idempotency-Key": key },
      },
    );
  });

  it("同一 Idempotency-Key で2回 POST しても key は呼び出し側が再利用できる", async () => {
    const key = "22222222-2222-4222-8222-222222222222";
    const body = {
      pet_id: 1,
      owner_id: 2,
      scheduled_date: "2026-08-01T00:00:00+09:00",
      items: [
        {
          category: "examination",
          name: "診察",
          unit_price: 1000,
          quantity: 1,
          tax_type: "excluded",
          tax_rate: 0.1,
          is_insurance_applicable: false,
          source: "manual",
        },
      ],
    };
    await completeAccounting(body, key);
    await completeAccounting(body, key);
    expect(postMock).toHaveBeenCalledTimes(2);
    expect(postMock.mock.calls[0]?.[2]).toEqual({ headers: { "Idempotency-Key": key } });
    expect(postMock.mock.calls[1]?.[2]).toEqual({ headers: { "Idempotency-Key": key } });
  });

  it("createAccountingCompletionIdempotencyKey は UUID 形式を返す", () => {
    const key = createAccountingCompletionIdempotencyKey();
    expect(key).toMatch(
      /^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$/i,
    );
  });
});
