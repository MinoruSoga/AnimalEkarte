import { beforeEach, describe, expect, it, vi } from "vitest";

import { axios } from "@/lib/axios";
import {
  createVaccineMaster,
  getAllVaccinesMaster,
  transformVaccineMaster,
  updateVaccineMaster,
} from "./vaccines-master";

vi.mock("@/lib/axios", () => ({
  axios: {
    get: vi.fn(),
    post: vi.fn(),
    patch: vi.fn(),
  },
}));

const baseVaccine = {
  id: 7,
  clinic_id: 1,
  name: "混合ワクチン",
  price: 5000,
  is_active: true,
  description: "",
  interval: "1year",
  sort_order: 2,
  created_at: "2026-10-20T00:00:00Z",
  updated_at: "2026-10-20T00:00:00Z",
};

describe("vaccine master insurance setting (EMR-455)", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it.each([
    { label: "対象外", wire: true, expected: true },
    { label: "対象", wire: false, expected: false },
    { label: "旧欠落(null)", wire: null, expected: undefined },
    { label: "未返却(undefined)", wire: undefined, expected: undefined },
  ])("keeps $label without coercing to a default", ({ wire, expected }) => {
    const item = transformVaccineMaster({ ...baseVaccine, is_non_insurance: wire });

    expect(item.isNonInsurance).toBe(expected);
    expect(item).toEqual(expect.objectContaining({ id: "7", name: "混合ワクチン", price: 5000 }));
  });

  it("reads list items with their saved insurance setting", async () => {
    vi.mocked(axios.get).mockResolvedValue({
      data: [
        { ...baseVaccine, id: 1, is_non_insurance: false },
        { ...baseVaccine, id: 2, is_non_insurance: null },
      ],
    });

    const items = await getAllVaccinesMaster();

    expect(axios.get).toHaveBeenCalledWith("/v1/masters/vaccines");
    expect(items.map((item) => item.isNonInsurance)).toEqual([false, undefined]);
  });

  it("sends an explicit setting on create and update and reads back the saved value", async () => {
    vi.mocked(axios.post).mockResolvedValue({ data: { ...baseVaccine, is_non_insurance: false } });
    vi.mocked(axios.patch).mockResolvedValue({ data: { ...baseVaccine, is_non_insurance: true } });

    const created = await createVaccineMaster({ name: "混合ワクチン", is_non_insurance: false });
    const updated = await updateVaccineMaster("7", { is_non_insurance: true });

    expect(axios.post).toHaveBeenCalledWith("/v1/masters/vaccines", {
      name: "混合ワクチン",
      is_non_insurance: false,
    });
    expect(axios.patch).toHaveBeenCalledWith("/v1/masters/vaccines/7", { is_non_insurance: true });
    expect(created.isNonInsurance).toBe(false);
    expect(updated.isNonInsurance).toBe(true);
  });

  it("does not add an insurance setting the caller did not send", async () => {
    vi.mocked(axios.patch).mockResolvedValue({ data: { ...baseVaccine, is_non_insurance: null } });

    const updated = await updateVaccineMaster("7", { name: "名称のみ変更" });

    expect(axios.patch).toHaveBeenCalledWith("/v1/masters/vaccines/7", { name: "名称のみ変更" });
    expect(updated.isNonInsurance).toBeUndefined();
  });
});
