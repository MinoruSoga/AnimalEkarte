import { beforeEach, describe, expect, it, vi } from "vitest";
import { axios } from "@/lib/axios";
import { getCheckups } from "./get-checkups";

vi.mock("@/lib/axios", () => ({ axios: { get: vi.fn() } }));
const mockedGet = vi.mocked(axios.get);

describe("getCheckups patient history filters", () => {
  beforeEach(() => {
    mockedGet.mockReset();
    mockedGet.mockResolvedValue({ data: { data: [], total: 0, page: 1, limit: 20 } });
  });

  it("sends pet_id so filtering happens before server pagination", async () => {
    await getCheckups({ petId: "42", page: 1, limit: 20 });
    expect(mockedGet).toHaveBeenCalledWith("/v1/checkups", {
      params: { page: 1, limit: 20, pet_id: "42" },
    });
  });
});
