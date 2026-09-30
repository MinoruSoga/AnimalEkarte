import { describe, expect, it } from "vitest";

import type { TrimmingUI } from "@/lib/transforms/trimming";
import type { InterviewHistoryItem } from "../types";
import {
  TRIMMING_TIMELINE_ID_PREFIX,
  mergePetTimelineItems,
  toTrimmingTimelineItem,
} from "./pet-timeline";

const makeTrimming = (overrides: Partial<TrimmingUI> = {}): TrimmingUI => ({
  id: "7",
  reservationTypeId: "",
  hasDetail: true,
  date: "2026-01-03",
  petId: "5",
  ownerId: "9",
  petNumber: "",
  petName: "モモ",
  ownerName: "飼主",
  species: "犬",
  breed: "",
  weight: "",
  styleRequest: "シャンプーカット",
  staff: "鈴木",
  status: "完了",
  staffId: "",
  courseId: "",
  courseName: "",
  optionIds: [],
  bw: "",
  bwUnit: "Kg",
  bt: "",
  usedShampoo: "",
  usedRibbon: "",
  remarks: "嫌がりなし",
  styleImage: undefined,
  completedImage: undefined,
  ...overrides,
});

const makeMedicalItem = (overrides: Partial<InterviewHistoryItem> = {}): InterviewHistoryItem => ({
  id: "11",
  date: "2026/01/02",
  author: "田中",
  type: "確定済",
  title: "消化器症状",
  content: "嘔吐2回",
  sortDate: "2026-01-02T00:00:00Z",
  ...overrides,
});

describe("toTrimmingTimelineItem（NO32 統合タイムライン）", () => {
  it("トリミング記録をタイムライン行へ変換し /trimming/:id へリンクする", () => {
    const item = toTrimmingTimelineItem(makeTrimming());

    expect(item.id).toBe(`${TRIMMING_TIMELINE_ID_PREFIX}7`);
    expect(item.href).toBe("/trimming/7");
    expect(item.type).toBe("トリミング（完了）");
    expect(item.title).toBe("シャンプーカット");
    expect(item.content).toBe("嫌がりなし");
    expect(item.author).toBe("鈴木");
    expect(item.date).toBe("2026/01/03");
    expect(item.sortDate).toBe("2026-01-03");
    // トリミング記録は問診複写の対象外
    expect(item.copySource).toBeUndefined();
  });

  it("styleRequest が空ならコース名、両方空なら「トリミング」をタイトルにする", () => {
    expect(
      toTrimmingTimelineItem(makeTrimming({ styleRequest: "", courseName: "カットコース" })).title,
    ).toBe("カットコース");
    expect(toTrimmingTimelineItem(makeTrimming({ styleRequest: "", courseName: "" })).title).toBe(
      "トリミング",
    );
  });

  it("remarks/staff が空ならフォールバック表示にする", () => {
    const item = toTrimmingTimelineItem(makeTrimming({ remarks: "", staff: "" }));

    expect(item.content).toBe("（記録なし）");
    expect(item.author).toBe("-");
  });

  it("status がバッジに反映される（予約/キャンセルも履歴として見える）", () => {
    expect(toTrimmingTimelineItem(makeTrimming({ status: "予約" })).type).toBe(
      "トリミング（予約）",
    );
    expect(toTrimmingTimelineItem(makeTrimming({ status: "キャンセル" })).type).toBe(
      "トリミング（キャンセル）",
    );
  });
});

describe("mergePetTimelineItems（NO32 統合タイムライン）", () => {
  it("カルテとトリミングを sortDate 降順で1本にマージする", () => {
    const merged = mergePetTimelineItems(
      [
        makeMedicalItem({ id: "1", sortDate: "2026-01-01T00:00:00Z" }),
        makeMedicalItem({ id: "3", sortDate: "2026-01-05T00:00:00Z" }),
      ],
      [
        makeTrimming({ id: "2", date: "2026-01-04" }),
        makeTrimming({ id: "4", date: "2026-01-06" }),
      ],
    );

    expect(merged.map((i) => i.id)).toEqual([
      `${TRIMMING_TIMELINE_ID_PREFIX}4`,
      "3",
      `${TRIMMING_TIMELINE_ID_PREFIX}2`,
      "1",
    ]);
  });

  it("同日はカルテ行を先に並べる（stable sort）", () => {
    const merged = mergePetTimelineItems(
      [makeMedicalItem({ id: "1", sortDate: "2026-01-03T10:00:00+09:00" })],
      [makeTrimming({ id: "9", date: "2026-01-03" })],
    );

    expect(merged.map((i) => i.id)).toEqual(["1", `${TRIMMING_TIMELINE_ID_PREFIX}9`]);
  });

  it("トリミングが無いペットはカルテ履歴だけを返す", () => {
    const merged = mergePetTimelineItems([makeMedicalItem()], []);

    expect(merged).toHaveLength(1);
    expect(merged[0].id).toBe("11");
  });

  it("カルテが無いペットはトリミング履歴だけを返す", () => {
    const merged = mergePetTimelineItems([], [makeTrimming({ id: "9" })]);

    expect(merged).toHaveLength(1);
    expect(merged[0].id).toBe(`${TRIMMING_TIMELINE_ID_PREFIX}9`);
  });

  it("両方空なら空配列を返す", () => {
    expect(mergePetTimelineItems([], [])).toEqual([]);
  });

  it("日付未設定（sortDate 空）の行は末尾に回る", () => {
    const merged = mergePetTimelineItems(
      [makeMedicalItem({ id: "1", sortDate: "2026-01-02" })],
      [makeTrimming({ id: "9", date: "" })],
    );

    expect(merged.map((i) => i.id)).toEqual(["1", `${TRIMMING_TIMELINE_ID_PREFIX}9`]);
    expect(merged[1].date).toBe("-");
  });

  it("カルテ行の copySource と既定遷移先（href 未指定）はそのまま維持する", () => {
    const merged = mergePetTimelineItems(
      [
        makeMedicalItem({
          copySource: { recordId: "11", chiefComplaint: "元気がない" },
        }),
      ],
      [],
    );

    expect(merged[0].copySource).toEqual({ recordId: "11", chiefComplaint: "元気がない" });
    expect(merged[0].href).toBeUndefined();
  });
});
