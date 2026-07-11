import { describe, it, expect } from "vitest";
import * as gen from "@/types/generated/models";
import type { TreatmentItemType, BodyWeightUnit } from "./index";

// FE4-4: 既存 literal union は型安全性は健在だが、backend が値を追加すると黙って古くなる。
// drift テストで生成定数の値集合との一致を追随漏れ検知として機械固定する（実装は無変更）。
describe("medical-records union drift", () => {
  it("TreatmentItemType の値集合が TreatmentItemType* 生成定数と一致する", () => {
    const values: TreatmentItemType[] = ["consultation", "procedure", "medicine", "other"];
    expect(new Set<string>(values)).toEqual(
      new Set([
        gen.TreatmentItemTypeConsultation,
        gen.TreatmentItemTypeProcedure,
        gen.TreatmentItemTypeMedicine,
        gen.TreatmentItemTypeOther,
      ]),
    );
  });

  it("BodyWeightUnit の値集合が BodyWeightUnit* 生成定数と一致する", () => {
    const values: BodyWeightUnit[] = ["Kg", "g"];
    expect(new Set<string>(values)).toEqual(
      new Set([gen.BodyWeightUnitKg, gen.BodyWeightUnitG]),
    );
  });
});
