import { beforeEach, describe, expect, it, vi } from "vitest";

vi.mock("@/lib/axios", () => ({
  axios: { get: vi.fn() },
}));

import { axios } from "@/lib/axios";
import { getUnbilledItems } from "./get-unbilled-items";
import type { BackendAccountingItem } from "./types";

const mockedGet = vi.mocked(axios.get);

function buildItem(overrides: Partial<BackendAccountingItem>): BackendAccountingItem {
  return {
    id: 1,
    billing_id: 0,
    category: "treatment" as BackendAccountingItem["category"],
    name: "診察料",
    unit_price: 1000,
    quantity: 1,
    tax_rate: 0.1,
    is_insurance_applicable: false,
    source: "medical_record" as BackendAccountingItem["source"],
    created_at: "2026-07-27T00:00:00Z",
    updated_at: "2026-07-27T00:00:00Z",
    ...overrides,
  };
}

describe("getUnbilledItems", () => {
  beforeEach(() => {
    mockedGet.mockReset();
  });

  it("同じ数値 ID の処置候補とワクチン候補を provenance 種別で区別する", async () => {
    mockedGet.mockResolvedValue({
      data: [
        buildItem({ id: 41, treatment_id: 41 }),
        buildItem({ id: 41, vaccination_id: 41 }),
      ],
    });

    const result = await getUnbilledItems("7");

    expect(result.map((item) => item.id)).toEqual([
      "treatment_41",
      "vaccination_41",
    ]);
  });

  it("具体的な provenance ID がない候補は従来の source と item ID を使う", async () => {
    mockedGet.mockResolvedValue({
      data: [buildItem({ id: 42, source: "manual" as BackendAccountingItem["source"] })],
    });

    const result = await getUnbilledItems("7");

    expect(result[0].id).toBe("manual_42");
  });
});
