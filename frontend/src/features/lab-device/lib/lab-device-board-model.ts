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
