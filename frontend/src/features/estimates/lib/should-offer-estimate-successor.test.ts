import { describe, expect, it } from "vitest";

import {
  isValidSuccessorReason,
  shouldOfferEstimateSuccessor,
} from "./should-offer-estimate-successor";

describe("shouldOfferEstimateSuccessor / isValidSuccessorReason", () => {
  it("estimate successor offer rules match locked create policy", () => {
    expect(shouldOfferEstimateSuccessor({ canCreate: true, status: "approved" })).toBe(true);
    expect(shouldOfferEstimateSuccessor({ canCreate: true, status: "rejected" })).toBe(true);

    expect(shouldOfferEstimateSuccessor({ canCreate: true, status: "draft" })).toBe(false);
    expect(shouldOfferEstimateSuccessor({ canCreate: true, status: "sent" })).toBe(false);

    for (const status of ["draft", "sent", "approved", "rejected"] as const) {
      expect(shouldOfferEstimateSuccessor({ canCreate: false, status })).toBe(false);
    }

    expect(isValidSuccessorReason("x")).toBe(true);
    expect(isValidSuccessorReason("  reason  ")).toBe(true);
    expect(isValidSuccessorReason("a".repeat(500))).toBe(true);
    expect(isValidSuccessorReason("あ".repeat(500))).toBe(true);
    expect(isValidSuccessorReason("😀".repeat(500))).toBe(true);

    expect(isValidSuccessorReason("")).toBe(false);
    expect(isValidSuccessorReason("   ")).toBe(false);
    expect(isValidSuccessorReason("a".repeat(501))).toBe(false);
    expect(isValidSuccessorReason("あ".repeat(501))).toBe(false);
    expect(isValidSuccessorReason("😀".repeat(501))).toBe(false);
    expect(isValidSuccessorReason(null)).toBe(false);
    expect(isValidSuccessorReason(undefined)).toBe(false);
    expect(isValidSuccessorReason(1)).toBe(false);
    expect(isValidSuccessorReason({})).toBe(false);
  });
});
