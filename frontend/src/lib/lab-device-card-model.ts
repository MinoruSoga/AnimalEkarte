import type { LabDeviceJobCard } from "@/hooks/use-lab-device-unlinked";

/**
 * FE-RC-015 followup3: LabDeviceUnlinkedBanner / Board で共有するカード表示純関数。
 */

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

export function labDeviceClockSkewLabel(card: LabDeviceJobCard): string | null {
  return card.clockSkew ? "機器時計がずれています（24時間超）" : null;
}

// P1: attach レスポンスが実際に persist 済みかどうか判定する。
// status === "persisted" かつ petId あり の場合のみ成功。それ以外は persist 失敗とみなす。
export function isLabDeviceAttachPersisted(card: LabDeviceJobCard): boolean {
  return card.status === "persisted" && card.petId != null;
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
