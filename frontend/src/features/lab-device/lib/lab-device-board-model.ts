import { formatJSTDate } from "@/lib/jst-date";

import type { LabDeviceJobCard, LabDeviceSlot, LabDeviceTodayVisit } from "../api/lab-device";

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

export function labDeviceClockSkewLabel(card: LabDeviceJobCard): string | null {
  return card.clockSkew ? "機器時計がずれています（24時間超）" : null;
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

export function labDeviceCardDayKey(card: LabDeviceJobCard): string {
  const raw = card.measuredAt || card.receivedAt;
  return raw ? formatJSTDate(raw) : "不明";
}

export function groupLabDeviceCardsByDay(
  cards: readonly LabDeviceJobCard[],
): Array<{ day: string; cards: LabDeviceJobCard[] }> {
  const grouped = new Map<string, LabDeviceJobCard[]>();
  for (const card of cards) {
    const day = labDeviceCardDayKey(card);
    const current = grouped.get(day) ?? [];
    grouped.set(day, [...current, card]);
  }
  return [...grouped.entries()]
    .sort((left, right) => right[0].localeCompare(left[0]))
    .map(([day, dayCards]) => ({ day, cards: dayCards }));
}

export function labDeviceReceivedDayLabel(day: string, today: string): string {
  if (day === "不明") {
    return "日時不明";
  }
  return day === today ? `${day}（今日）` : day;
}

export function labDeviceReceivedCards(input: {
  received: readonly LabDeviceJobCard[];
  unlinked: readonly LabDeviceJobCard[];
  saved: readonly LabDeviceJobCard[];
}): LabDeviceJobCard[] {
  if (input.received.length > 0) {
    return [...input.received];
  }
  const seen = new Set<string>();
  const merged: LabDeviceJobCard[] = [];
  for (const card of [...input.unlinked, ...input.saved]) {
    if (seen.has(card.jobId)) {
      continue;
    }
    seen.add(card.jobId);
    merged.push(card);
  }
  return merged;
}

export function labDeviceSelectableTodayVisits(
  visits: readonly LabDeviceTodayVisit[],
): LabDeviceTodayVisit[] {
  return visits.filter((visit) => visit.petId > 0 && !visit.petIsDeceased);
}

export function labDeviceSlotMatchesCard(slot: LabDeviceSlot, card: LabDeviceJobCard): boolean {
  return card.deviceHint === slot.deviceHint || card.sourceType === slot.sourceType;
}

export function labDeviceLatestCardForSlot(
  slot: LabDeviceSlot,
  cards: readonly LabDeviceJobCard[],
): LabDeviceJobCard | undefined {
  return cards.find((card) => labDeviceSlotMatchesCard(slot, card));
}

// 受信 POST の失敗を「機器の電文が悪い」以外も区別して掲示する。
// 送信の瞬間スタッフの目は機器側にあるため、トーストだけでなく機器カードに残す。
export function labDeviceReceiveFailure(status?: number): { label: string; message: string } {
  if (status === 401) {
    return {
      label: "失敗（要ログイン）",
      message: "セッションが切れています。再ログイン後、機器の送信をもう一度押してください",
    };
  }
  if (status === 400) {
    return { label: "失敗（電文不正）", message: "電文を読めませんでした" };
  }
  return {
    label: "失敗（通信エラー）",
    message: "保存できませんでした。機器の送信をもう一度押してください",
  };
}

export function labDeviceLiveReceiveLabel(input: {
  liveLabel?: string;
  latestCard?: LabDeviceJobCard;
}): string {
  if (input.liveLabel) {
    return input.liveLabel;
  }
  if (!input.latestCard) {
    return "未受信";
  }
  return input.latestCard.petName || "未紐付け";
}

export type LabDeviceListenTone = "live" | "idle" | "blocked" | "unsupported";

export function labDeviceListenTone(state: LabDeviceListenState): LabDeviceListenTone {
  switch (state) {
    case "listening":
      return "live";
    case "disconnected":
      return "idle";
    case "needs_permission":
      return "blocked";
    case "unsupported":
      return "unsupported";
  }
}
