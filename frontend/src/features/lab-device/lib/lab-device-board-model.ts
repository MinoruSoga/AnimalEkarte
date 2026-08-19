import type { LabDeviceJobCard, LabDeviceSlot } from "../api/lab-device";

export function labDeviceSourceLabel(sourceType: string): string {
  switch (sourceType) {
    case "fuji_nx600":
      return "NX600";
    case "fuji_au10v":
      return "AU10V";
    case "arkray_pu4010":
      return "PU-4010";
    default:
      return sourceType;
  }
}

export function labDeviceCardTitle(card: LabDeviceJobCard): string {
  return card.deviceHint || labDeviceSourceLabel(card.sourceType);
}

export function labDeviceHasUnmapped(card: LabDeviceJobCard): boolean {
  return card.unmappedItemCount > 0 || card.items.some((item) => item.needsReview);
}

export function labDeviceUnmappedMasterHref(sourceType: string): string {
  const params = new URLSearchParams({ source: sourceType, from: "board" });
  return `/settings/lab-device-item-masters?${params.toString()}`;
}

export function labDevicePortStorageKey(slotKey: string): string {
  return `lab-device-port:${slotKey}`;
}

export function isWebSerialSupported(): boolean {
  return typeof navigator !== "undefined" && "serial" in navigator;
}

export function findSlotByHint(slots: LabDeviceSlot[], hint: string): LabDeviceSlot | undefined {
  const normalized = hint.trim().toLowerCase();
  return slots.find((slot) => slot.deviceHint.toLowerCase() === normalized || slot.key === normalized);
}

export type LabDeviceListenState = "unsupported" | "needs_permission" | "listening" | "disconnected";

export function labDeviceListenState(input: {
  serialSupported: boolean;
  hasStoredPort: boolean;
  connected: boolean;
}): LabDeviceListenState {
  if (!input.serialSupported) {
    return "unsupported";
  }
  if (!input.hasStoredPort) {
    return "needs_permission";
  }
  return input.connected ? "listening" : "disconnected";
}

export function labDeviceBoardLinkLabel(states: readonly LabDeviceListenState[]): "受信中" | "切断" {
  return states.includes("listening") ? "受信中" : "切断";
}

export function labDeviceSlotListenLabel(state: LabDeviceListenState): string {
  switch (state) {
    case "listening":
      return "受信中";
    case "needs_permission":
      return "未許可";
    case "unsupported":
      return "非対応";
    case "disconnected":
      return "切断";
  }
}
