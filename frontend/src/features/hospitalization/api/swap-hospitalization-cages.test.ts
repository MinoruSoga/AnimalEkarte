import { beforeEach, describe, expect, it, vi } from "vitest";

const { mockPost } = vi.hoisted(() => ({ mockPost: vi.fn() }));

vi.mock("@/lib/axios", () => ({ axios: { post: mockPost } }));

import { swapHospitalizationCages } from "./swap-hospitalization-cages";

describe("swapHospitalizationCages", () => {
  beforeEach(() => {
    mockPost.mockReset();
    mockPost.mockResolvedValue({ data: { hospitalizations: [] } });
  });

  it.each([
    { ids: ["7", "3"], cages: ["17", "13"] },
    { ids: ["3", "7"], cages: ["13", "17"] },
  ])("preserves the cage snapshot correspondence for $ids", async ({ ids, cages }) => {
    await swapHospitalizationCages(ids[0], ids[1], [cages[0], cages[1]]);

    expect(mockPost).toHaveBeenCalledWith("/v1/hospitalizations/cage-swap", {
      hospitalization_ids: ids.map(Number),
      expected_cage_ids: cages.map(Number),
    });
  });
});
