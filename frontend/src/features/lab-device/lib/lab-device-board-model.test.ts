import { describe, expect, it } from "vitest";

import type { LabDeviceJobCard } from "../api/lab-device";
import {
  labDeviceCardTitle,
  labDeviceHasUnmapped,
  labDeviceSourceLabel,
} from "./lab-device-board-model";

const card = (patch: Partial<LabDeviceJobCard> = {}): LabDeviceJobCard => ({
  jobId: "j1",
  sourceType: "fuji_au10v",
  deviceHint: "AU10V",
  status: "received",
  specimenIdRaw: "TEST1",
  itemCount: 1,
  unmappedItemCount: 0,
  items: [],
  ...patch,
});

describe("lab-device-board-model", () => {
  it("labels the three Joto devices", () => {
    expect(labDeviceSourceLabel("fuji_nx600")).toBe("NX600");
    expect(labDeviceSourceLabel("fuji_au10v")).toBe("AU10V");
    expect(labDeviceSourceLabel("arkray_pu4010")).toBe("PU-4010");
    expect(labDeviceCardTitle(card({ deviceHint: "" }))).toBe("AU10V");
  });

  it("flags unmapped or needs-review cards", () => {
    expect(labDeviceHasUnmapped(card())).toBe(false);
    expect(labDeviceHasUnmapped(card({ unmappedItemCount: 1 }))).toBe(true);
    expect(labDeviceHasUnmapped(card({
      items: [{ deviceItemCode: "ZZZ", valueRaw: "1", unit: "", flag: "", needsReview: true, sortOrder: 0 }],
    }))).toBe(true);
  });
});
