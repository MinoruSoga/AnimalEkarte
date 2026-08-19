import { describe, expect, it } from "vitest";

import type { LabDeviceJobCard } from "../api/lab-device";
import {
  labDeviceBoardLinkLabel,
  labDeviceCardTitle,
  labDeviceHasUnmapped,
  labDeviceListenState,
  labDeviceSlotListenLabel,
  labDeviceSourceLabel,
  labDeviceUnmappedMasterHref,
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

  it("opens the matching device side panel from an unmapped chip", () => {
    expect(labDeviceUnmappedMasterHref("fuji_nx600")).toBe(
      "/settings/lab-device-item-masters?source=fuji_nx600&from=board",
    );
  });

  it("flags unmapped or needs-review cards", () => {
    expect(labDeviceHasUnmapped(card())).toBe(false);
    expect(labDeviceHasUnmapped(card({ unmappedItemCount: 1 }))).toBe(true);
    expect(labDeviceHasUnmapped(card({
      items: [{ deviceItemCode: "ZZZ", valueRaw: "1", unit: "", flag: "", needsReview: true, sortOrder: 0 }],
    }))).toBe(true);
  });

  it("treats authorized open ports as listening and the rest as disconnected", () => {
    expect(labDeviceListenState({
      serialSupported: false,
      hasStoredPort: false,
      connected: false,
    })).toBe("unsupported");
    expect(labDeviceListenState({
      serialSupported: true,
      hasStoredPort: false,
      connected: false,
    })).toBe("needs_permission");
    expect(labDeviceListenState({
      serialSupported: true,
      hasStoredPort: true,
      connected: false,
    })).toBe("disconnected");
    expect(labDeviceListenState({
      serialSupported: true,
      hasStoredPort: true,
      connected: true,
    })).toBe("listening");
    expect(labDeviceBoardLinkLabel(["needs_permission", "disconnected"])).toBe("切断");
    expect(labDeviceBoardLinkLabel(["needs_permission", "listening"])).toBe("受信中");
    expect(labDeviceSlotListenLabel("needs_permission")).toBe("未許可");
    expect(labDeviceSlotListenLabel("listening")).toBe("受信中");
  });
});
