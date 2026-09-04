import { describe, expect, it } from "vitest";

import type { LabDeviceJobCard } from "@/hooks/use-lab-device-unlinked";
import {
  isLabDeviceAttachPersisted,
  labDeviceCardTitle,
  labDeviceClockSkewLabel,
  labDeviceNeedsReviewReason,
  labDeviceSourceLabel,
} from "@/lib/lab-device-card-model";

const card = (patch: Partial<LabDeviceJobCard> = {}): LabDeviceJobCard => ({
  jobId: "j1",
  sourceType: "fuji_au10v",
  deviceHint: "AU10V",
  status: "received",
  specimenIdRaw: "TEST1",
  itemCount: 1,
  unmappedItemCount: 0,
  clockSkew: false,
  items: [],
  ...patch,
});

describe("lab-device-card-model", () => {
  it("labDeviceSourceLabel: 既知のソース種別を日本語ラベルに変換する", () => {
    expect(labDeviceSourceLabel("fuji_nx600")).toBe("NX600");
    expect(labDeviceSourceLabel("fuji_au10v")).toBe("AU10V");
    expect(labDeviceSourceLabel("arkray_pu4010")).toBe("PU-4010");
    expect(labDeviceSourceLabel("idexx_vetlab")).toBe("IDEXX VetLab");
    expect(labDeviceSourceLabel("unknown_device")).toBe("unknown_device");
  });

  it("labDeviceCardTitle: deviceHint があればそれを返し、なければ sourceLabel にフォールバックする", () => {
    expect(labDeviceCardTitle(card({ deviceHint: "NX600" }))).toBe("NX600");
    expect(labDeviceCardTitle(card({ deviceHint: "" }))).toBe("AU10V");
    expect(labDeviceCardTitle(card({ deviceHint: "", sourceType: "fuji_nx600" }))).toBe("NX600");
  });

  it("labDeviceClockSkewLabel: clockSkew=true のときのみ警告文字列、false で null", () => {
    expect(labDeviceClockSkewLabel(card())).toBeNull();
    expect(labDeviceClockSkewLabel(card({ clockSkew: true }))).toBe(
      "機器時計がずれています（24時間超）",
    );
  });

  it("isLabDeviceAttachPersisted: status=persisted かつ petId あり のみ成功と判定する", () => {
    expect(isLabDeviceAttachPersisted(card({ status: "persisted", petId: 1 }))).toBe(true);
    expect(isLabDeviceAttachPersisted(card({ status: "persisted", petId: undefined }))).toBe(false);
    expect(isLabDeviceAttachPersisted(card({ status: "linked", petId: 1 }))).toBe(false);
    expect(isLabDeviceAttachPersisted(card({ status: "needs_review", petId: 1 }))).toBe(false);
    expect(isLabDeviceAttachPersisted(card({ status: "received", petId: undefined }))).toBe(false);
  });

  it("labDeviceNeedsReviewReason: needs_review は汎用メッセージ、非 needs_review で null", () => {
    expect(labDeviceNeedsReviewReason(card({ status: "needs_review" }))).toBe(
      "確認が必要です（保存できませんでした）",
    );
    expect(
      labDeviceNeedsReviewReason(
        card({ status: "needs_review", reviewReason: "lab_device_multiple_exam_types" }),
      ),
    ).toBe("確認が必要です（保存できませんでした）");
    expect(
      labDeviceNeedsReviewReason(card({ status: "needs_review", reviewReason: "other_code" })),
    ).toBe("確認が必要です（保存できませんでした）");
    expect(labDeviceNeedsReviewReason(card({ status: "persisted" }))).toBeNull();
    expect(labDeviceNeedsReviewReason(card({ status: "received" }))).toBeNull();
  });
});
