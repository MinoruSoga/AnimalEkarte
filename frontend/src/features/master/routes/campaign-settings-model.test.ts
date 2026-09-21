import { describe, expect, it } from "vitest";

import type { CampaignFormData } from "../lib/campaign-side-panel-model";
import { buildCampaignCreateRequest, buildCampaignUpdateRequest } from "./campaign-settings-model";

function makeFormData(overrides: Partial<CampaignFormData> = {}): CampaignFormData {
  return {
    name: "春キャンペーン",
    startDate: "2026-03-01",
    endDate: "2026-03-31",
    discountType: "amount",
    discountValue: 1200,
    isActive: true,
    targetCategories: ["goods"],
    targetItemIds: ["11"],
    ...overrides,
  };
}

describe("buildCampaignCreateRequest", () => {
  it("persists empty-like discount_value 0 as discount not unit price", () => {
    const request = buildCampaignCreateRequest(makeFormData({ discountValue: 0 }));
    expect(request.discount_value).toBe(0);
    expect(request.discount_type).toBe("amount");
    expect(request).not.toHaveProperty("unit_price");
    expect(request).not.toHaveProperty("price");
  });

  it("persists amount 1200 and rate 10 discount payloads", () => {
    const amount = buildCampaignCreateRequest(
      makeFormData({ discountType: "amount", discountValue: 1200 }),
    );
    expect(amount.discount_type).toBe("amount");
    expect(amount.discount_value).toBe(1200);

    const rate = buildCampaignCreateRequest(
      makeFormData({ discountType: "rate", discountValue: 10 }),
    );
    expect(rate.discount_type).toBe("rate");
    expect(rate.discount_value).toBe(10);
  });
});

describe("buildCampaignUpdateRequest", () => {
  it("persists edited discount_value without inventing a unit price field", () => {
    const request = buildCampaignUpdateRequest(
      makeFormData({ discountType: "amount", discountValue: 2300 }),
    );
    expect(request.discount_value).toBe(2300);
    expect(request).not.toHaveProperty("price");
  });
});
