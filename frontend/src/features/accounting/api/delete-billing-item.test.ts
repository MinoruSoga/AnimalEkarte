import { beforeEach, describe, expect, it, vi } from "vitest";

import { deleteBillingItem } from "./delete-billing-item";
import { axios } from "@/lib/axios";

vi.mock("@/lib/axios", () => ({
  axios: {
    delete: vi.fn(),
  },
}));

const mockedDelete = vi.mocked(axios.delete);

describe("deleteBillingItem", () => {
  beforeEach(() => {
    mockedDelete.mockReset();
    mockedDelete.mockResolvedValue({ data: undefined });
  });

  it("理由なしは body なしで DELETE する", async () => {
    await deleteBillingItem("12");
    expect(mockedDelete).toHaveBeenCalledWith("/v1/billing-items/12", undefined);
  });

  it("post_close_reason があるとき data body で送る（BUG-021）", async () => {
    await deleteBillingItem("12", { post_close_reason: "締め後削除" });
    expect(mockedDelete).toHaveBeenCalledWith("/v1/billing-items/12", {
      data: { post_close_reason: "締め後削除" },
    });
  });
});
