import { describe, expect, it } from "vitest";

import {
  MedicineCalculationTypePerWeight,
  MedicineDoseBasisPerAdministration,
  MedicineUnitPerTablet,
} from "@/types/generated/models";
import type { DoseCalcInput } from "@/lib/medicine-dose";
import { DOSE_PARAMS_LOOKUP_FAILED_MESSAGE } from "../../api/medicine-dose-lookup";

import { computeDoseGate, resolveDoseGateSource } from "./treatment-row-dose-gate";

describe("computeDoseGate", () => {
  it("missing（手動入力対象外）は常に requiresDeviationReason=false", () => {
    const got = computeDoseGate({ kind: "missing" }, 5);
    expect(got.requiresDeviationReason).toBe(false);
    expect(got.isBlocked).toBe(false);
    expect(got.warning).toBe("none");
  });

  it("technical_failure は isBlocked=true で固定文言のみ（upstream body なし）", () => {
    const got = computeDoseGate({ kind: "technical_failure" }, 1);
    expect(got.isBlocked).toBe(true);
    expect(got.blockReason).toBe(DOSE_PARAMS_LOOKUP_FAILED_MESSAGE);
    expect(got.blockReason).not.toMatch(/SQLSTATE|timeout|stack|http/i);
    expect(got.requiresDeviationReason).toBe(false);
    expect(got.warning).toBe("none");
  });

  it("missing は technical_failure と型で区別され保存をブロックしない", () => {
    const missing = computeDoseGate({ kind: "missing" }, 5);
    const failed = computeDoseGate({ kind: "technical_failure" }, 5);
    expect(missing.isBlocked).toBe(false);
    expect(failed.isBlocked).toBe(true);
    expect(failed.blockReason).not.toBe(missing.blockReason);
    expect(failed.blockReason.length).toBeGreaterThan(0);
  });

  it("resolveDoseGateSource: authority.failed を doseCalcInput=null と同一視しない", () => {
    expect(resolveDoseGateSource(null, { status: "failed" })).toEqual({
      kind: "technical_failure",
    });
    expect(resolveDoseGateSource(null, { status: "success", params: [] })).toEqual({
      kind: "missing",
    });
    expect(resolveDoseGateSource(null, { status: "idle" })).toEqual({ kind: "missing" });
  });

  it("resolveDoseGateSource: pending は technical failure ではなく missing 扱い（保存継続）", () => {
    expect(resolveDoseGateSource(null, { status: "pending" })).toEqual({ kind: "missing" });
    // 取得中に評価可能な input が既にあるなら ready を維持する。
    const input: DoseCalcInput = {
      calculationType: MedicineCalculationTypePerWeight,
      medicineUnit: MedicineUnitPerTablet,
      strength: 10,
      dosePerKg: 5,
      weightKg: 4,
      maxMgPerKg: 10,
    };
    expect(resolveDoseGateSource(input, { status: "pending" })).toEqual({
      kind: "ready",
      input,
    });
  });

  it("安全域内・推奨値と一致する submitted は gate 不要", () => {
    const input: DoseCalcInput = {
      calculationType: MedicineCalculationTypePerWeight,
      medicineUnit: MedicineUnitPerTablet,
      strength: 10,
      doseBasis: MedicineDoseBasisPerAdministration,
      dosePerKg: 5,
      weightKg: 4,
      maxMgPerKg: 10,
    };
    // 推奨値 = 2錠(20mg)
    const got = computeDoseGate({ kind: "ready", input }, 2);
    expect(got.requiresDeviationReason).toBe(false);
    expect(got.isBlocked).toBe(false);
    expect(got.recommendedQuantity).toBe(2);
  });

  it("上限超過だけが物理ブロックになる（丸め境界越え含む）", () => {
    const input: DoseCalcInput = {
      calculationType: MedicineCalculationTypePerWeight,
      medicineUnit: MedicineUnitPerTablet,
      strength: 10,
      doseBasis: MedicineDoseBasisPerAdministration,
      dosePerKg: 5,
      weightKg: 1.7,
      maxMgPerKg: 5,
    };
    const got = computeDoseGate({ kind: "ready", input }, 2); // 20mg > upperCap(8.5mg)
    expect(got.requiresDeviationReason).toBe(false);
    expect(got.isBlocked).toBe(true);
    expect(got.warning).toBe("exceeds-max");
    expect(got.reason).toContain("上限");
  });

  it("下限割れは警告対象だが物理ブロックしない（理由必須）", () => {
    const input: DoseCalcInput = {
      calculationType: MedicineCalculationTypePerWeight,
      medicineUnit: MedicineUnitPerTablet,
      strength: 10,
      doseBasis: MedicineDoseBasisPerAdministration,
      dosePerKg: 5,
      weightKg: 1.5,
      minMgPerKg: 4,
      maxMgPerKg: 10,
    };
    const got = computeDoseGate({ kind: "ready", input }, 0);
    expect(got.requiresDeviationReason).toBe(true);
    expect(got.isBlocked).toBe(false);
    expect(got.warning).toBe("below-min");
  });

  it("安全域内の推奨値からの著しい乖離は warning=none でも理由必須", () => {
    const input: DoseCalcInput = {
      calculationType: MedicineCalculationTypePerWeight,
      medicineUnit: MedicineUnitPerTablet,
      strength: 10,
      doseBasis: MedicineDoseBasisPerAdministration,
      dosePerKg: 5,
      weightKg: 4,
      maxMgPerKg: 100, // 上限に十分な余裕を持たせ exceedsMax を発生させない
    };
    // 推奨=2錠。5錠は乖離率150%で閾値(20%)超過だが upperCap(400mg)には収まる。
    const got = computeDoseGate({ kind: "ready", input }, 5);
    expect(got.requiresDeviationReason).toBe(true);
    expect(got.isBlocked).toBe(false);
    expect(got.warning).toBe("none");
    expect(got.reason).toContain("推奨値");
  });
});
