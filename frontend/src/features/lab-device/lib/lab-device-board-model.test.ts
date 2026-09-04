import { describe, expect, it } from "vitest";

import {
  parseLabDeviceSlots,
  type LabDeviceJobCard,
  type LabDeviceSlot,
  type LabDeviceTodayVisit,
} from "../api/lab-device";
import {
  groupLabDeviceCardsByDay,
  isLabDeviceAttachPersisted,
  labDeviceBoardLinkLabel,
  labDeviceCardNeedsReview,
  labDeviceCardTitle,
  labDeviceClockSkewLabel,
  labDeviceHasUnmapped,
  labDeviceLatestCardForSlot,
  labDeviceListenState,
  labDeviceListenTone,
  labDeviceLiveReceiveLabel,
  labDeviceNeedsReviewReason,
  labDeviceReceiveFailure,
  requireLabDeviceReceiveResult,
  labDeviceReceivedCards,
  labDeviceReceivedDayLabel,
  labDeviceSelectableTodayVisits,
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
  clockSkew: false,
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

  it("warns when device clock skew exceeds 24 hours", () => {
    expect(labDeviceClockSkewLabel(card())).toBeNull();
    expect(labDeviceClockSkewLabel(card({ clockSkew: true }))).toBe(
      "機器時計がずれています（24時間超）",
    );
  });

  it("flags unmapped or needs-review cards", () => {
    expect(labDeviceHasUnmapped(card())).toBe(false);
    expect(labDeviceHasUnmapped(card({ unmappedItemCount: 1 }))).toBe(true);
    expect(
      labDeviceHasUnmapped(
        card({
          items: [
            {
              deviceItemCode: "ZZZ",
              valueRaw: "1",
              unit: "",
              flag: "",
              needsReview: true,
              sortOrder: 0,
            },
          ],
        }),
      ),
    ).toBe(true);
  });

  it("treats authorized open ports as listening and the rest as disconnected", () => {
    expect(
      labDeviceListenState({
        serialSupported: false,
        hasStoredPort: false,
        connected: false,
      }),
    ).toBe("unsupported");
    expect(
      labDeviceListenState({
        serialSupported: true,
        hasStoredPort: false,
        connected: false,
      }),
    ).toBe("needs_permission");
    expect(
      labDeviceListenState({
        serialSupported: true,
        hasStoredPort: true,
        connected: false,
      }),
    ).toBe("disconnected");
    expect(
      labDeviceListenState({
        serialSupported: true,
        hasStoredPort: true,
        connected: true,
      }),
    ).toBe("listening");
    expect(labDeviceBoardLinkLabel(["needs_permission", "disconnected"])).toBe("切断");
    expect(labDeviceBoardLinkLabel(["needs_permission", "listening"])).toBe("受信中");
    expect(labDeviceSlotListenLabel("needs_permission")).toBe("未許可");
    expect(labDeviceSlotListenLabel("listening")).toBe("受信中");
    expect(labDeviceSlotListenLabel("monitoring")).toBe("自動監視中");
  });

  it("groups received cards by JST day, newest first", () => {
    const grouped = groupLabDeviceCardsByDay([
      card({ jobId: "old", receivedAt: "2026-08-17T23:30:00+09:00" }),
      card({ jobId: "today-1", receivedAt: "2026-08-19T08:00:00+09:00" }),
      card({ jobId: "today-2", measuredAt: "2026-08-19T11:00:00+09:00" }),
    ]);
    expect(grouped.map((row) => row.day)).toEqual(["2026-08-19", "2026-08-17"]);
    expect(grouped[0]?.cards.map((item) => item.jobId)).toEqual(["today-1", "today-2"]);
    expect(labDeviceReceivedDayLabel("2026-08-19", "2026-08-19")).toBe("2026-08-19（今日）");
    expect(labDeviceReceivedDayLabel("2026-08-17", "2026-08-19")).toBe("2026-08-17");
  });

  it("prefers the board received list and drops deceased or pet-less visits", () => {
    const received = [card({ jobId: "r1" })];
    expect(
      labDeviceReceivedCards({
        received,
        unlinked: [card({ jobId: "u1" })],
        saved: [card({ jobId: "s1" })],
      }),
    ).toEqual(received);
    expect(
      labDeviceReceivedCards({
        received: [],
        unlinked: [card({ jobId: "u1" })],
        saved: [card({ jobId: "s1" }), card({ jobId: "u1" })],
      }).map((item) => item.jobId),
    ).toEqual(["u1", "s1"]);

    const visits: LabDeviceTodayVisit[] = [
      {
        recordId: 1,
        petId: 11,
        petName: "タロウ",
        ownerName: "山田",
        species: "犬",
        doctorName: "佐藤",
        visitType: "再診",
      },
      {
        recordId: 2,
        petId: 12,
        petName: "亡",
        ownerName: "鈴木",
        species: "猫",
        doctorName: "佐藤",
        visitType: "初診",
        petIsDeceased: true,
      },
      {
        recordId: 3,
        petId: 0,
        petName: "",
        ownerName: "田中",
        species: "",
        doctorName: "",
        visitType: "",
      },
    ];
    expect(labDeviceSelectableTodayVisits(visits).map((visit) => visit.petId)).toEqual([11]);
  });

  it("picks the newest received card and live label for each device slot", () => {
    const slot: LabDeviceSlot = {
      key: "au10v",
      sourceType: "fuji_au10v",
      deviceHint: "AU10V",
      baud: 9600,
    };
    const latest = labDeviceLatestCardForSlot(slot, [
      card({ jobId: "nx", sourceType: "fuji_nx600", deviceHint: "NX600" }),
      card({ jobId: "au", sourceType: "fuji_au10v", deviceHint: "AU10V", petName: "タロウ" }),
    ]);
    expect(latest?.jobId).toBe("au");
    expect(labDeviceLiveReceiveLabel({ liveLabel: "受信", latestCard: latest })).toBe("受信");
    expect(labDeviceLiveReceiveLabel({ latestCard: latest })).toBe("タロウ");
    expect(labDeviceLiveReceiveLabel({})).toBe("未受信");
    expect(labDeviceListenTone("listening")).toBe("live");
    expect(labDeviceListenTone("disconnected")).toBe("idle");
    expect(labDeviceListenTone("needs_permission")).toBe("blocked");
    expect(labDeviceListenTone("unsupported")).toBe("unsupported");
  });

  it("スロット JSON の baud と parity を読み、不正な parity は落とす", () => {
    const slots = parseLabDeviceSlots(
      '[{"key":"pu4010","source_type":"arkray_pu4010","device_hint":"PU-4010","baud":2400,"parity":"even"},' +
        '{"key":"nx600","source_type":"fuji_nx600","device_hint":"NX600","baud":9600},' +
        '{"key":"x","source_type":"fuji_au10v","device_hint":"AU10V","baud":9600,"parity":"bogus"}]',
    );
    expect(slots[0]).toMatchObject({ key: "pu4010", baud: 2400, parity: "even" });
    expect(slots[1]!.parity).toBeUndefined();
    expect(slots[2]!.parity).toBeUndefined();
  });

  // FE-RC-037: サーバ由来 JSON が配列でない場合、無検証キャストせず空配列にフォールバックする。
  it.each([
    ["オブジェクト", '{"key":"nx600"}'],
    ["数値", "42"],
    ["非JSON文字列", "not json"],
    [
      "配列内の非オブジェクト要素は無視する",
      '[1, "x", {"key":"nx600","source_type":"fuji_nx600","device_hint":"NX600"}]',
    ],
  ])("不正な形状(%s)は空配列または有効要素のみへ落とす", (_name, json) => {
    expect(() => parseLabDeviceSlots(json)).not.toThrow();
  });

  it("配列内の非オブジェクト要素を無視しつつ有効な要素は変換する", () => {
    const slots = parseLabDeviceSlots(
      '[1, "x", {"key":"nx600","source_type":"fuji_nx600","device_hint":"NX600"}]',
    );
    expect(slots).toHaveLength(1);
    expect(slots[0]).toMatchObject({ key: "nx600", baud: 9600 });
  });

  it("受信失敗を要因別のラベルと案内に分ける", () => {
    expect(labDeviceReceiveFailure(401).label).toBe("失敗（要ログイン）");
    expect(labDeviceReceiveFailure(401).message).toContain("再ログイン");
    expect(labDeviceReceiveFailure(401).message).toContain("再送しない");
    expect(labDeviceReceiveFailure(400)).toEqual({
      label: "失敗（電文不正）",
      message: "電文を読めませんでした",
    });
    expect(labDeviceReceiveFailure(500).label).toBe("失敗（通信エラー）");
    expect(labDeviceReceiveFailure(500).message).toContain("自動再試行");
    expect(labDeviceReceiveFailure(500).message).toContain("再送しない");
    expect(labDeviceReceiveFailure(undefined).label).toBe("失敗（通信エラー）");
  });

  it("空の受信成功レスポンスをACK可能な結果として扱わない", () => {
    expect(() => requireLabDeviceReceiveResult([])).toThrow("empty lab device receive result");
    expect(requireLabDeviceReceiveResult(["job-1"])).toBe("job-1");
  });

  // P1: attach 後 persist 失敗判定
  it("isLabDeviceAttachPersisted: status=persisted かつ petId あり のみ成功と判定する", () => {
    expect(isLabDeviceAttachPersisted(card({ status: "persisted", petId: 1 }))).toBe(true);
    expect(isLabDeviceAttachPersisted(card({ status: "persisted", petId: undefined }))).toBe(false);
    expect(isLabDeviceAttachPersisted(card({ status: "linked", petId: 1 }))).toBe(false);
    expect(isLabDeviceAttachPersisted(card({ status: "needs_review", petId: 1 }))).toBe(false);
    expect(isLabDeviceAttachPersisted(card({ status: "received", petId: undefined }))).toBe(false);
  });

  // P2: needs_review カード判定
  it("labDeviceCardNeedsReview: status=needs_review のみ true", () => {
    expect(labDeviceCardNeedsReview(card({ status: "needs_review" }))).toBe(true);
    expect(labDeviceCardNeedsReview(card({ status: "persisted" }))).toBe(false);
    expect(labDeviceCardNeedsReview(card({ status: "received" }))).toBe(false);
    expect(labDeviceCardNeedsReview(card({ status: "linked" }))).toBe(false);
  });

  // F-1 (T001 更新): needs_review 原因コードを日本語ラベルに変換する。
  // "lab_device_multiple_exam_types" は複数種別の分割保存に変更したため新規ジョブでは設定されない。
  // 旧ジョブ互換で汎用メッセージへ fallthrough する。
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
