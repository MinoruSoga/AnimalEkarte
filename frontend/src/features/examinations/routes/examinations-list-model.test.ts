import { describe, expect, it } from "vitest";

import { examinationCreateHref, examinationListDetailHref } from "./examinations-list-model";

describe("examinationListDetailHref", () => {
  it("カルテに紐づく検査は検査タブ付きカルテへ行く", () => {
    expect(
      examinationListDetailHref({ id: "exam-1", medicalRecordId: "mr-1" }),
    ).toBe("/medical-records/mr-1?tab=%E6%A4%9C%E6%9F%BB&examId=exam-1");
  });

  it("未紐付けの検査は検査詳細画面へ行く", () => {
    expect(examinationListDetailHref({ id: "exam-1" })).toBe("/examinations/exam-1");
  });
});

describe("examinationCreateHref", () => {
  it("新規検査は当日カルテの検査タブへ行く", () => {
    expect(examinationCreateHref("pet-1")).toBe(
      "/medical-records/new?petId=pet-1&tab=%E6%A4%9C%E6%9F%BB",
    );
  });
});
