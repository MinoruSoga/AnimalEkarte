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
    case "idexx_vetlab":
      return "IDEXX VetLab";
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

export type LabDeviceListenState = "unsupported" | "needs_permission" | "monitoring" | "listening" | "disconnected";

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
    case "monitoring":
      return "自動監視中";
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
      message: "セッションが切れています。機器では再送しないで、再ログインしてください。結果は受信機から自動再試行します",
    };
  }
  if (status === 400) {
    return { label: "失敗（電文不正）", message: "電文を読めませんでした" };
  }
  return {
    label: "失敗（通信エラー）",
    message: "保存を自動再試行しています。機器では再送しないでください",
  };
}

export function requireLabDeviceReceiveResult<T>(results: readonly T[]): T {
  const first = results[0];
  if (first === undefined) {
    throw new Error("empty lab device receive result");
  }
  return first;
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
    case "monitoring":
      return "idle";
    case "disconnected":
      return "idle";
    case "needs_permission":
      return "blocked";
    case "unsupported":
      return "unsupported";
  }
}

// P1: attach レスポンスが実際に persist 済みかどうか判定する。
// status === "persisted" かつ petId あり の場合のみ成功。それ以外は persist 失敗とみなす。
export function isLabDeviceAttachPersisted(card: LabDeviceJobCard): boolean {
  return card.status === "persisted" && card.petId != null;
}

// P2: 原因コードに依存せず、カードが汎用の needs_review 状態かどうか判定する。
// 複数検査種別は分割保存されるため、新規ジョブの needs_review 原因にはならない。
export function labDeviceCardNeedsReview(card: LabDeviceJobCard): boolean {
  return card.status === "needs_review";
}

// F-1: needs_review の原因コードを日本語ラベルに変換する。
// reviewReason が設定されていれば原因固有のメッセージ、未設定なら汎用メッセージを返す。
// null は「needs_review でない」を意味する。
// T001: "lab_device_multiple_exam_types" は複数種別の分割保存に変更したため新規ジョブでは設定されない。
// 旧ジョブとの後方互換のため汎用メッセージへ fallthrough する。
export function labDeviceNeedsReviewReason(card: LabDeviceJobCard): string | null {
  if (card.status !== "needs_review") {
    return null;
  }
  return "確認が必要です（保存できませんでした）";
}

export function isLabDeviceBoardSlotSupported(sourceType: string): boolean {
  return sourceType === "fuji_nx600" || sourceType === "fuji_au10v";
}

export function labDeviceBoardSlotListenState(input: {
  supported: boolean;
  agentConnected: boolean;
  openPorts: number;
  configuredPorts: number;
  hasLiveReceive: boolean;
}): LabDeviceListenState {
  if (!input.supported) {
    return "unsupported";
  }
  if (
    input.agentConnected
    && input.openPorts > 0
    && input.openPorts === input.configuredPorts
  ) {
    return input.hasLiveReceive ? "listening" : "monitoring";
  }
  return "disconnected";
}

export function labDeviceAgentConnectionLabel(input: {
  connected: boolean;
  configuredPorts: number;
  openPorts: number;
}): "切断" | "要確認" | "監視中" {
  if (!input.connected) {
    return "切断";
  }
  if (input.configuredPorts === 0 || input.openPorts < input.configuredPorts) {
    return "要確認";
  }
  return "監視中";
}

export function labDeviceAttachFailureToast(attached: LabDeviceJobCard): string {
  return labDeviceCardNeedsReview(attached)
    ? `保存できませんでした（${labDeviceNeedsReviewReason(attached)}）`
    : "保存できませんでした。未紐付けのままです";
}

export function labDeviceAgentDegradedErrorMessage(
  category: "none" | "discovery_failed" | "port_open_failed" | "queue_write_failed" | "port_close_failed" | "response_write_failed",
): string | null {
  switch (category) {
    case "none":
      return null;
    case "discovery_failed":
      return "USB接続の確認に失敗しました。Macのローカル受信機を再起動してください。";
    case "queue_write_failed":
      return "受信結果を保持できません。追加送信を止めてサポート担当へ連絡してください。";
    case "port_close_failed":
      return "USBポートの終了に失敗しました。Macのローカル受信機を再起動してください。";
    case "response_write_failed":
      return "ローカル受信機との通信に失敗しました。画面を再読み込みしてください。";
    case "port_open_failed":
      return "USBポートを開けません。接続とアクセス権を確認してください。";
  }
}
