import { describe, it, expect, vi, beforeEach } from "vitest";

vi.mock("@/lib/axios", () => ({
  axios: { get: vi.fn() },
}));

import { axios } from "@/lib/axios";
import { ownersLoader } from "./loaders";

const mockedGet = vi.mocked(axios.get);

describe("ownersLoader — empty-owner-pet-row deceasedAt", () => {
  beforeEach(() => {
    mockedGet.mockReset();
  });

  it("ペットなし飼主の空ペット行は deceasedAt === undefined", async () => {
    mockedGet.mockResolvedValue({
      data: {
        data: [
          {
            id: 1,
            clinic_id: 1,
            owner_name: "ヤマダ タロウ",
            owner_name_kana: "ヤマダ タロウ",
            address1: "東京都",
            address2: "",
            phone: "090-0000-0000",
            pets: [],
          },
        ],
        total: 1,
        page: 1,
        limit: 100,
      },
    });

    const result = await ownersLoader({
      request: new Request("http://localhost/owners"),
    });

    expect(result.pets).toHaveLength(1);
    expect(result.pets[0].deceasedAt).toBeUndefined();
  });
});
