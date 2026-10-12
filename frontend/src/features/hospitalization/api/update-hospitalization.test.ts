import { beforeEach, describe, expect, it, vi } from "vitest";

const { mockPatch } = vi.hoisted(() => ({ mockPatch: vi.fn() }));

vi.mock("@/lib/axios", () => ({ axios: { patch: mockPatch } }));

import { updateHospitalization } from "./update-hospitalization";

describe("updateHospitalization", () => {
  beforeEach(() => {
    mockPatch.mockReset();
    mockPatch.mockResolvedValue({ data: { id: 7, status: "admitted" } });
  });

  it("serializes cage_id and expected_cage_id as numbers", async () => {
    await updateHospitalization("7", {
      cage_id: "12",
      expected_cage_id: "10",
    });

    const payload = mockPatch.mock.calls[0][1] as Record<string, unknown>;
    expect(payload.cage_id).toBe(12);
    expect(payload.expected_cage_id).toBe(10);
    expect(typeof payload.cage_id).toBe("number");
    expect(typeof payload.expected_cage_id).toBe("number");
  });

  it("forwards expected_cage_is_null unchanged as a boolean", async () => {
    await updateHospitalization("7", {
      cage_id: "12",
      expected_cage_is_null: true,
    });

    const payload = mockPatch.mock.calls[0][1] as Record<string, unknown>;
    expect(payload.cage_id).toBe(12);
    expect(payload.expected_cage_is_null).toBe(true);
    expect(typeof payload.expected_cage_is_null).toBe("boolean");
    expect(payload).not.toHaveProperty("expected_cage_id");
  });

  it("sends clear_cage_id with the displayed expected snapshot", async () => {
    await updateHospitalization("7", {
      clear_cage_id: true,
      expected_cage_id: "10",
    });

    expect(mockPatch).toHaveBeenCalledWith("/v1/hospitalizations/7", {
      clear_cage_id: true,
      expected_cage_id: 10,
    });
  });

  it("omits empty-string cage and expected values", async () => {
    await updateHospitalization("7", { cage_id: "", expected_cage_id: "" });

    const payload = mockPatch.mock.calls[0][1] as Record<string, unknown>;
    expect(payload).not.toHaveProperty("cage_id");
    expect(payload).not.toHaveProperty("expected_cage_id");
  });
});
