import { describe, it, expect, vi, beforeEach } from "vitest";

const { mockPost } = vi.hoisted(() => ({
  mockPost: vi.fn(),
}));

vi.mock("@/lib/axios", () => ({
  axios: { post: mockPost },
}));

import { createTreatmentPlanForHospitalization } from "./treatment-plans-write";

describe("createTreatmentPlanForHospitalization", () => {
  beforeEach(() => {
    mockPost.mockReset();
  });

  it("POST /v1/hospitalizations/:id/treatment-plans with wire body", async () => {
    mockPost.mockResolvedValueOnce({
      data: {
        id: "9",
        hospitalization_id: "7",
        treatment_content: "adm rate",
        memo: "memo",
        is_insurance: true,
        unit_price: 990,
        quantity: 1,
        discount_rate: 0,
        discount_amount: 0,
        subtotal: 990,
        sort_order: 0,
        created_at: "2026-07-31T00:00:00+09:00",
        updated_at: "2026-07-31T00:00:00+09:00",
      },
    });

    const result = await createTreatmentPlanForHospitalization("7", {
      treatment_content: "adm rate",
      memo: "memo",
      is_insurance: true,
      unit_price: 990,
      quantity: 1,
      discount_rate: 0,
      discount_amount: 0,
      sort_order: 0,
    });

    expect(mockPost).toHaveBeenCalledWith(
      "/v1/hospitalizations/7/treatment-plans",
      expect.objectContaining({
        treatment_content: "adm rate",
        quantity: 1,
        unit_price: 990,
      }),
    );
    expect(result.id).toBe("9");
  });
});
