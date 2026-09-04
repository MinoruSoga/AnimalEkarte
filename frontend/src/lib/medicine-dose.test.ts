import { describe, expect, it } from "vitest";

import {
  BodyWeightUnitG,
  BodyWeightUnitKg,
  MedicineCalculationTypeNone,
  MedicineCalculationTypePerWeight,
  MedicineDoseBasisPerAdministration,
  MedicineDoseBasisPerDay,
  MedicineRoundingModeDown,
  MedicineRoundingModeNearest,
  MedicineRoundingModeUp,
  MedicineUnitPerDose,
  MedicineUnitPerML,
  MedicineUnitPerTablet,
} from "@/types/generated/models";

import {
  DOSE_DEVIATION_REL_THRESHOLD,
  calculateDose,
  evaluateSubmittedDose,
  isSignificantDoseDeviation,
  normalizeDoseSpecies,
  normalizeWeightToKg,
  roundQuantity,
  type DoseCalcInput,
} from "./medicine-dose";

// dose_calc_test.go の golden table を仕様の単一の真実として同一ケース表で移植する。
// BE 実装と数値が乖離したらこのテストと dose_calc_test.go の両方を見直すこと。
const FLOAT_TOL = 1e-6;

function expectFloatEq(a: number, b: number) {
  expect(Math.abs(a - b)).toBeLessThanOrEqual(FLOAT_TOL);
}

describe("normalizeDoseSpecies", () => {
  it.each([
    ["犬 漢字", "犬", "dog"],
    ["いぬ ひらがな", "いぬ", "dog"],
    ["dog 英語 大文字混在", "Dog", "dog"],
    ["前後空白 trim", "  cat  ", "cat"],
    ["猫 漢字", "猫", "cat"],
    ["feline 英語", "feline", "cat"],
  ])("%s", (_name, input, want) => {
    expect(normalizeDoseSpecies(input)).toBe(want);
  });

  it.each([
    ["マップ不能種は fail-closed（うさぎ）", "うさぎ"],
    ["マップ不能種は fail-closed（rabbit）", "rabbit"],
    ["空文字は fail-closed", ""],
    ["未知の犬関連語は安全側で fail-closed（狂犬病等の取り違え防止）", "狂犬病"],
  ])("%s", (_name, input) => {
    expect(normalizeDoseSpecies(input)).toBeNull();
  });

  it("犬↔猫の取り違えが起きないことを明示的に保証する", () => {
    const dogInputs = ["犬", "いぬ", "イヌ", "ドッグ", "dog", "canine"];
    for (const input of dogInputs) {
      expect(normalizeDoseSpecies(input)).not.toBe("cat");
    }
    const catInputs = ["猫", "ねこ", "ネコ", "キャット", "cat", "feline"];
    for (const input of catInputs) {
      expect(normalizeDoseSpecies(input)).not.toBe("dog");
    }
  });
});

describe("normalizeWeightToKg", () => {
  it("Kg はそのまま", () => {
    expectFloatEq(normalizeWeightToKg(4.2, BodyWeightUnitKg) ?? NaN, 4.2);
  });
  it("g は kg へ", () => {
    expectFloatEq(normalizeWeightToKg(850, BodyWeightUnitG) ?? NaN, 0.85);
  });
  it("単位未指定（既定 Kg）", () => {
    expectFloatEq(normalizeWeightToKg(3.0, undefined) ?? NaN, 3.0);
  });
  it("0 は fail-closed", () => {
    expect(normalizeWeightToKg(0, BodyWeightUnitKg)).toBeNull();
  });
  it("負値は fail-closed", () => {
    expect(normalizeWeightToKg(-1, BodyWeightUnitKg)).toBeNull();
  });
});

function baseEligibleInput(): DoseCalcInput {
  return {
    calculationType: MedicineCalculationTypePerWeight,
    medicineUnit: MedicineUnitPerTablet,
    strength: 10,
    doseBasis: MedicineDoseBasisPerAdministration,
    dosePerKg: 5,
    weightKg: 4,
  };
}

describe("calculateDose eligibility (fail-closed)", () => {
  const cases: [string, (input: DoseCalcInput) => DoseCalcInput][] = [
    [
      "calculation_type=none",
      (input) => ({ ...input, calculationType: MedicineCalculationTypeNone }),
    ],
    ["medicine_unit nil", (input) => ({ ...input, medicineUnit: null })],
    [
      "medicine_unit per_dose は不適格",
      (input) => ({ ...input, medicineUnit: MedicineUnitPerDose }),
    ],
    ["strength nil", (input) => ({ ...input, strength: null })],
    ["strength <= 0", (input) => ({ ...input, strength: 0 })],
    ["dose_per_kg <= 0", (input) => ({ ...input, dosePerKg: 0 })],
    ["weight <= 0（未記録）", (input) => ({ ...input, weightKg: 0 })],
    [
      "per_day で frequency 欠落",
      (input) => ({ ...input, doseBasis: MedicineDoseBasisPerDay, frequencyPerDay: null }),
    ],
    [
      "per_day で frequency <= 0",
      (input) => ({ ...input, doseBasis: MedicineDoseBasisPerDay, frequencyPerDay: 0 }),
    ],
  ];

  it.each(cases)("%s", (_name, mutate) => {
    const got = calculateDose(mutate(baseEligibleInput()));
    expect(got.eligible).toBe(false);
    expect(got.ineligibleReason).not.toBe("");
  });
});

describe("calculateDose calculation", () => {
  const cases: {
    name: string;
    input: DoseCalcInput;
    wantQty: number;
    wantEffMg: number;
    wantBelowMin?: boolean;
    wantExceedsMax?: boolean;
  }[] = [
    {
      name: "基本: 錠剤 per_administration 丸めなし",
      input: {
        calculationType: MedicineCalculationTypePerWeight,
        medicineUnit: MedicineUnitPerTablet,
        strength: 10,
        doseBasis: MedicineDoseBasisPerAdministration,
        dosePerKg: 5,
        weightKg: 4,
      },
      wantQty: 2,
      wantEffMg: 20,
    },
    {
      name: "液剤: per_ml で小数数量（silent 丸めなし）",
      input: {
        calculationType: MedicineCalculationTypePerWeight,
        medicineUnit: MedicineUnitPerML,
        strength: 50,
        doseBasis: MedicineDoseBasisPerAdministration,
        dosePerKg: 10,
        weightKg: 2.5,
      },
      wantQty: 0.5,
      wantEffMg: 25,
    },
    {
      name: "丸め上げ（上限未設定なので逸脱なし）",
      input: {
        calculationType: MedicineCalculationTypePerWeight,
        medicineUnit: MedicineUnitPerTablet,
        strength: 20,
        doseBasis: MedicineDoseBasisPerAdministration,
        dosePerKg: 5,
        weightKg: 3,
        roundingStep: 1,
        roundingMode: MedicineRoundingModeUp,
      },
      wantQty: 1,
      wantEffMg: 20,
    },
    {
      name: "C1 回帰: 切上で実効用量が上限を越える → exceedsMax フラグ（silent 越え禁止）",
      input: {
        calculationType: MedicineCalculationTypePerWeight,
        medicineUnit: MedicineUnitPerTablet,
        strength: 10,
        doseBasis: MedicineDoseBasisPerAdministration,
        dosePerKg: 5,
        weightKg: 1.7,
        maxMgPerKg: 5,
        roundingStep: 1,
        roundingMode: MedicineRoundingModeUp,
      },
      // rawMg=8.5, cap=8.5, ceil(0.85)=1錠, effMg=10 > 8.5 → exceedsMax
      wantQty: 1,
      wantEffMg: 10,
      wantExceedsMax: true,
    },
    {
      name: "両上限: 大型犬で absolute_max_dose が binding（小さい方で cap）",
      input: {
        calculationType: MedicineCalculationTypePerWeight,
        medicineUnit: MedicineUnitPerTablet,
        strength: 100,
        doseBasis: MedicineDoseBasisPerAdministration,
        dosePerKg: 10,
        weightKg: 40,
        maxMgPerKg: 10,
        absoluteMaxDose: 300,
      },
      // rawMg=400, weightCap=400, absCap=300 → cap=300 → 3錠, effMg=300（境界一致でフラグなし）
      wantQty: 3,
      wantEffMg: 300,
    },
    {
      name: "min クランプ: dose_per_kg が下限未満 → 下限へ引上げ",
      input: {
        calculationType: MedicineCalculationTypePerWeight,
        medicineUnit: MedicineUnitPerTablet,
        strength: 10,
        doseBasis: MedicineDoseBasisPerAdministration,
        dosePerKg: 2,
        minMgPerKg: 5,
        weightKg: 4,
      },
      wantQty: 2,
      wantEffMg: 20,
    },
    {
      name: "max クランプ: dose_per_kg が上限超過 → 上限へ引下げ",
      input: {
        calculationType: MedicineCalculationTypePerWeight,
        medicineUnit: MedicineUnitPerTablet,
        strength: 10,
        doseBasis: MedicineDoseBasisPerAdministration,
        dosePerKg: 20,
        maxMgPerKg: 5,
        weightKg: 2,
      },
      wantQty: 1,
      wantEffMg: 10,
    },
    {
      name: "belowMin: 丸め下げで実効用量が下限割れ → フラグ",
      input: {
        calculationType: MedicineCalculationTypePerWeight,
        medicineUnit: MedicineUnitPerTablet,
        strength: 10,
        doseBasis: MedicineDoseBasisPerAdministration,
        dosePerKg: 5,
        minMgPerKg: 4,
        weightKg: 1.5,
        roundingStep: 1,
        roundingMode: MedicineRoundingModeDown,
      },
      // rawMg=7.5, floor(0.75)=0錠, effMg=0 < lower(6) → belowMin
      wantQty: 0,
      wantEffMg: 0,
      wantBelowMin: true,
    },
    {
      name: "per_day 基準: frequency で 1回量へ按分",
      input: {
        calculationType: MedicineCalculationTypePerWeight,
        medicineUnit: MedicineUnitPerTablet,
        strength: 25,
        doseBasis: MedicineDoseBasisPerDay,
        frequencyPerDay: 2,
        dosePerKg: 10,
        weightKg: 5,
      },
      // basisFactor=0.5, rawMg=10*5*0.5=25, 25/25=1錠
      wantQty: 1,
      wantEffMg: 25,
    },
    {
      name: "per_day 基準: absolute_max_dose を1日総量とみなし按分して cap",
      input: {
        calculationType: MedicineCalculationTypePerWeight,
        medicineUnit: MedicineUnitPerML,
        strength: 50,
        doseBasis: MedicineDoseBasisPerDay,
        frequencyPerDay: 2,
        dosePerKg: 40,
        weightKg: 10,
        absoluteMaxDose: 300,
      },
      // basisFactor=0.5, rawMg=40*10*0.5=200, absCap=300*0.5=150 → cap=150 → 3mL, effMg=150
      wantQty: 3,
      wantEffMg: 150,
    },
  ];

  it.each(cases.map((c) => [c.name, c] as const))("%s", (_name, tt) => {
    const got = calculateDose(tt.input);
    expect(got.eligible).toBe(true);
    expectFloatEq(got.quantity, tt.wantQty);
    expectFloatEq(got.effectiveDoseMg, tt.wantEffMg);
    expect(got.belowMin).toBe(tt.wantBelowMin ?? false);
    expect(got.exceedsMax).toBe(tt.wantExceedsMax ?? false);
  });
});

describe("roundQuantity", () => {
  // mode 列は未知値ケース("invalid_mode")を含めて検証するため意図的に string | null で緩める。
  const cases: [string, number, number | null, string | null, number][] = [
    ["step nil は丸めない", 1.234, null, MedicineRoundingModeUp, 1.234],
    ["mode nil は丸めない", 1.234, 1, null, 1.234],
    ["step も mode も nil", 1.234, null, null, 1.234],
    ["step<=0 は丸めない", 1.234, 0, MedicineRoundingModeUp, 1.234],
    ["step 負値は丸めない", 1.234, -1, MedicineRoundingModeUp, 1.234],
    ["Up: 切上", 1.1, 1, MedicineRoundingModeUp, 2],
    ["Down: 切下", 1.9, 1, MedicineRoundingModeDown, 1],
    ["Nearest: 四捨五入（切上側）", 1.5, 1, MedicineRoundingModeNearest, 2],
    ["Nearest: 四捨五入（切下側）", 1.4, 1, MedicineRoundingModeNearest, 1],
    ["Nearest: step=0.5刻み", 1.3, 0.5, MedicineRoundingModeNearest, 1.5],
    ["未知の mode は丸めない（fail-safe）", 1.234, 1, "invalid_mode", 1.234],
  ];

  it.each(cases)("%s", (_name, value, step, mode, want) => {
    const got = roundQuantity(value, step, mode as Parameters<typeof roundQuantity>[2]);
    expectFloatEq(got, want);
  });
});

describe("calculateDose C1 no silent breach", () => {
  // strength=10 mg/錠, max=5 mg/kg, 切上, 体重を変えて ceil が上限を越える領域を走査。
  it.each([1.1, 1.3, 1.7, 1.9, 2.1, 2.3, 2.7, 2.9])("weight=%s", (w) => {
    const input: DoseCalcInput = {
      calculationType: MedicineCalculationTypePerWeight,
      medicineUnit: MedicineUnitPerTablet,
      strength: 10,
      doseBasis: MedicineDoseBasisPerAdministration,
      dosePerKg: 5,
      weightKg: w,
      maxMgPerKg: 5,
      roundingStep: 1,
      roundingMode: MedicineRoundingModeUp,
    };
    const got = calculateDose(input);
    const upper = 5 * w; // weight × max_mg/kg（basisFactor=1）
    if (got.effectiveDoseMg > upper + 1e-6) {
      expect(got.exceedsMax).toBe(true);
    } else {
      expect(got.exceedsMax).toBe(false);
    }
  });
});

describe("evaluateSubmittedDose", () => {
  it("丸め境界越え: 手動入力で上限を越えたら exceedsMax=true（silent 越え禁止）", () => {
    const input: DoseCalcInput = {
      calculationType: MedicineCalculationTypePerWeight,
      medicineUnit: MedicineUnitPerTablet,
      strength: 10,
      doseBasis: MedicineDoseBasisPerAdministration,
      dosePerKg: 5,
      weightKg: 1.7,
      maxMgPerKg: 5,
    };
    // upperCap = 5*1.7 = 8.5mg. submitted=2錠→20mg > 8.5mg
    const got = evaluateSubmittedDose(input, 2);
    expect(got.exceedsMax).toBe(true);
    expect(got.effectiveDoseMg).toBe(20);
    expect(got.upperCapMg).not.toBeNull();
  });

  it("submitted が安全域内なら exceedsMax=false", () => {
    const input: DoseCalcInput = {
      calculationType: MedicineCalculationTypePerWeight,
      medicineUnit: MedicineUnitPerTablet,
      strength: 10,
      doseBasis: MedicineDoseBasisPerAdministration,
      dosePerKg: 5,
      weightKg: 4,
      maxMgPerKg: 10,
    };
    const got = evaluateSubmittedDose(input, 2);
    expect(got.exceedsMax).toBe(false);
    expect(got.belowMin).toBe(false);
  });

  it("submitted が下限を割ると belowMin=true", () => {
    const input: DoseCalcInput = {
      calculationType: MedicineCalculationTypePerWeight,
      medicineUnit: MedicineUnitPerTablet,
      strength: 10,
      doseBasis: MedicineDoseBasisPerAdministration,
      dosePerKg: 5,
      weightKg: 1.5,
      minMgPerKg: 4,
    };
    // lowerMg = 4*1.5 = 6mg. submitted=0 (未入力相当) → 0mg < 6mg
    const got = evaluateSubmittedDose(input, 0);
    expect(got.belowMin).toBe(true);
  });

  it("上限未設定なら hasUpperCap=false・upperCapMg=null", () => {
    const input: DoseCalcInput = {
      calculationType: MedicineCalculationTypePerWeight,
      medicineUnit: MedicineUnitPerTablet,
      strength: 10,
      doseBasis: MedicineDoseBasisPerAdministration,
      dosePerKg: 5,
      weightKg: 4,
    };
    const got = evaluateSubmittedDose(input, 100);
    expect(got.hasUpperCap).toBe(false);
    expect(got.upperCapMg).toBeNull();
    expect(got.exceedsMax).toBe(false);
  });
});

describe("isSignificantDoseDeviation", () => {
  it("閾値(0.20)以内は逸脱ではない", () => {
    expect(isSignificantDoseDeviation(2.2, 2)).toBe(false);
  });
  it("閾値(0.20)超過は逸脱", () => {
    expect(isSignificantDoseDeviation(2.5, 2)).toBe(true);
  });
  it("computed<=0 は逸脱判定しない", () => {
    expect(isSignificantDoseDeviation(5, 0)).toBe(false);
  });
  it("BE dose_revalidation.go と同一閾値を使用する", () => {
    expect(DOSE_DEVIATION_REL_THRESHOLD).toBe(0.2);
  });
});
