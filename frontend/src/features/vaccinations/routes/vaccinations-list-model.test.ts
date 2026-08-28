import { describe, expect, it } from "vitest";

import { vaccinationCreateHref, vaccinationListDetailHref } from "./vaccinations-list-model";

describe("vaccinationListDetailHref", () => {
  it("カルテに紐づく接種は予防接種タブ付きカルテへ行く", () => {
    expect(
      vaccinationListDetailHref({ id: "vac-1", medicalRecordId: "mr-1" }),
    ).toBe("/medical-records/mr-1?tab=%E4%BA%88%E9%98%B2%E6%8E%A5%E7%A8%AE&vaccinationId=vac-1");
  });

  it("未紐付けの接種は接種詳細画面へ行く", () => {
    expect(vaccinationListDetailHref({ id: "vac-1" })).toBe("/vaccinations/vac-1");
  });
});

describe("vaccinationCreateHref", () => {
  it("新規接種は独立フォーム /vaccinations/new?petId= へ行く（BUG-501）", () => {
    expect(vaccinationCreateHref("pet-1")).toBe("/vaccinations/new?petId=pet-1");
  });
});
