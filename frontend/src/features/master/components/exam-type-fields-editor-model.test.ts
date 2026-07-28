import { describe, expect, it } from "vitest";

import {
  buildReferenceRangeRequest,
  validateReferenceRangeDrafts,
  type ReferenceRangeDraft,
} from "./exam-type-fields-editor-model";

const baseDraft = (overrides: Partial<ReferenceRangeDraft> = {}): ReferenceRangeDraft => ({
  animalSpeciesId: "2",
  mode: "numeric",
  min: "",
  max: "",
  ...overrides,
});

describe("exam type reference range validation", () => {
  it("rejects numeric and qualitative bounds coexisting", () => {
    expect(validateReferenceRangeDrafts([
      baseDraft({ min: "1", qualitativeMin: "(-)" }),
    ])).toBe("数値範囲と定性範囲は同時に指定できません");
  });

  it("rejects unsupported and reversed qualitative bounds", () => {
    expect(validateReferenceRangeDrafts([
      baseDraft({ mode: "qualitative", min: "unknown", max: "(+)" }),
    ])).toBe("定性値は (-)、(±)、(+)、(++)、(+++) から選択してください");
    expect(validateReferenceRangeDrafts([
      baseDraft({ mode: "qualitative", min: "(++)", max: "(+)" }),
    ])).toBe("定性範囲の下限は上限以下にしてください");
  });

  it("rejects non-finite and reversed numeric bounds", () => {
    expect(validateReferenceRangeDrafts([
      baseDraft({ min: "Infinity" }),
    ])).toBe("数値範囲には有限の数値を入力してください");
    expect(validateReferenceRangeDrafts([
      baseDraft({ min: "10", max: "5" }),
    ])).toBe("数値範囲の下限は上限以下にしてください");
  });

  it("builds numeric and qualitative replacement payloads", () => {
    expect(buildReferenceRangeRequest([
      baseDraft({ min: "5", max: "10" }),
      baseDraft({
        animalSpeciesId: "3",
        mode: "qualitative",
        min: "（ - ）",
        max: "(+)",
      }),
    ])).toEqual([
      { animal_species_id: 2, ref_min: 5, ref_max: 10 },
      { animal_species_id: 3, qualitative_min: "(-)", qualitative_max: "(+)" },
    ]);
  });
});
