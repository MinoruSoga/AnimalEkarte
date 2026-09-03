/**
 * medicine-dose.ts — #201 カルテ薬量（投与量）自動計算の FE 純粋関数群。
 *
 * 仕様は backend/internal/service/dose_calc.go + dose_revalidation.go と完全一致させる
 * （同一ケース表を dose_calc_test.go から移植して検証する。仕様変更時は両方を同時に直すこと）。
 *
 * 安全上の位置づけ: BE は treatment 保存時に医療記録・薬マスタ・vital から独立して同じ計算を
 * 再実行し、監査ログへ記録する（treatment_dose_save.go/dose_revalidation.go）。本モジュールは
 * 保存前の UX（プリフィル・警告・hard gate）専用であり、安全性の最終権威は BE 側にある。
 *
 * 安全設計の核（#201）:
 *   - C1: 安全域(min/max)判定は「丸め後の実効用量(mg)」に対して行う（切上による silent 越え禁止）。
 *   - 両上限: 体重連動上限(weight × max_mg/kg) と 体重非依存上限(absolute_max_dose mg/head) の
 *     小さい方で cap する。
 *   - fail-closed: per_weight 以外・含量不明・体重未記録・per_day で頻度欠落・dose<=0 は
 *     eligible=false（手動入力へフォールバック）。
 *   - species hard gate: normalizeDoseSpecies はマップ不能種を null で返す（犬→猫の silent
 *     フォールバックは絶対に行わない。エイリアスに犬猫を混在させないこと）。
 */
import type {
  BodyWeightUnit,
  MedicineCalculationType,
  MedicineDoseBasis,
  MedicineDoseSpecies,
  MedicineRoundingMode,
  MedicineUnit,
} from "@/types/generated/models";
import {
  BodyWeightUnitG,
  MedicineCalculationTypePerWeight,
  MedicineDoseBasisPerDay,
  MedicineDoseSpeciesCat,
  MedicineDoseSpeciesDog,
  MedicineRoundingModeDown,
  MedicineRoundingModeNearest,
  MedicineRoundingModeUp,
  MedicineUnitPerGram,
  MedicineUnitPerML,
  MedicineUnitPerTablet,
} from "@/types/generated/models";

// 浮動小数の丸め境界での誤検出を防ぐ微小許容値（dose_calc.go roundEpsilon と同値）。
const ROUND_EPSILON = 1e-9;
// 実効用量(mg)の安全域比較に使う許容値（dose_calc.go safetyEpsilon と同値）。
const SAFETY_EPSILON = 1e-6;

/**
 * 「著しい逸脱」とみなす相対乖離のデフォルト（BE dose_revalidation.go
 * doseDeviationRelThreshold と同一値・暫定確定 2026-07-15 PO）。
 */
export const DOSE_DEVIATION_REL_THRESHOLD = 0.2;

// free-text animal_species.name → 正規化種のエイリアス表。
// 安全のため exact-match（trim + 小文字化後）のみ。未知語は fail-closed（null）で manual に落とす。
// 犬・猫を取り違えるエイリアスは絶対に追加しないこと（猫への犬用量は致死的になり得る）。
const DOSE_SPECIES_ALIASES: Record<string, MedicineDoseSpecies> = {
  犬: MedicineDoseSpeciesDog,
  いぬ: MedicineDoseSpeciesDog,
  イヌ: MedicineDoseSpeciesDog,
  ドッグ: MedicineDoseSpeciesDog,
  dog: MedicineDoseSpeciesDog,
  canine: MedicineDoseSpeciesDog,
  猫: MedicineDoseSpeciesCat,
  ねこ: MedicineDoseSpeciesCat,
  ネコ: MedicineDoseSpeciesCat,
  キャット: MedicineDoseSpeciesCat,
  cat: MedicineDoseSpeciesCat,
  feline: MedicineDoseSpeciesCat,
};

/**
 * free-text の種名を計算用の正規化種へ写像する。マップ不能な場合は null（fail-closed）。
 * 犬↔猫の取り違え・silent フォールバックは行わない。
 */
export function normalizeDoseSpecies(
  freeText: string | null | undefined,
): MedicineDoseSpecies | null {
  const key = (freeText ?? "").trim().toLowerCase();
  if (!key) return null;
  return DOSE_SPECIES_ALIASES[key] ?? null;
}

/** 体重を kg に正規化する。weight <= 0 は null（fail-closed）。 */
export function normalizeWeightToKg(weight: number, unit?: BodyWeightUnit | null): number | null {
  if (weight <= 0) return null;
  if (unit === BodyWeightUnitG) return weight / 1000;
  // 既定（Kg・未指定）は kg そのまま。
  return weight;
}

// per_weight 計算で含量分母として解釈可能な単位。per_dose（包/vial）は含量定義が曖昧なため除外。
const ELIGIBLE_MEDICINE_UNITS_FOR_PER_WEIGHT = new Set<MedicineUnit>([
  MedicineUnitPerTablet,
  MedicineUnitPerML,
  MedicineUnitPerGram,
]);

/** per_weight 計算の入力。weightKg は normalizeWeightToKg 済み（kg・>0 期待）。 */
export interface DoseCalcInput {
  calculationType: MedicineCalculationType;
  /** strength の分母。per_tablet/per_ml/per_gram のみ適格（per_dose/未設定は不適格） */
  medicineUnit?: MedicineUnit | null;
  /** mg/単位 */
  strength?: number | null;
  doseBasis: MedicineDoseBasis;
  /** per_day 基準時に必須（> 0） */
  frequencyPerDay?: number | null;
  /** mg/kg */
  dosePerKg: number;
  /** 安全域下限（任意） */
  minMgPerKg?: number | null;
  /** 体重連動上限（任意） */
  maxMgPerKg?: number | null;
  /** 体重非依存 mg/head 上限（任意） */
  absoluteMaxDose?: number | null;
  /** 丸め単位（未指定=丸めなし） */
  roundingStep?: number | null;
  roundingMode?: MedicineRoundingMode | null;
  weightKg: number;
}

/** 計算結果と安全フラグ。eligible=false の場合は計算値を使わず手動入力へ。 */
export interface DoseCalcResult {
  eligible: boolean;
  ineligibleReason: string;

  /** clamp(dose_per_kg, min, max) */
  effectiveRate: number;
  /** 丸め・cap 前の 1回投与あたり mg */
  rawMg: number;
  /** 両上限適用後・丸め前 */
  cappedMg: number;
  /** 丸め後のプリフィル数量（単位数） */
  quantity: number;
  /** quantity × strength（丸め後の実効用量。C1 判定の基準） */
  effectiveDoseMg: number;

  /** 実効用量が安全域下限を下回る */
  belowMin: boolean;
  /** 実効用量が両上限の小さい方を上回る（切上による silent 越えを必ず検出） */
  exceedsMax: boolean;
}

function ineligible(reason: string): DoseCalcResult {
  return {
    eligible: false,
    ineligibleReason: reason,
    effectiveRate: 0,
    rawMg: 0,
    cappedMg: 0,
    quantity: 0,
    effectiveDoseMg: 0,
    belowMin: false,
    exceedsMax: false,
  };
}

function doseEligibility(input: DoseCalcInput): string | null {
  if (input.calculationType !== MedicineCalculationTypePerWeight) {
    return "calculation_type is not per_weight";
  }
  if (!input.medicineUnit) {
    return "medicine_unit is required for per_weight";
  }
  if (!ELIGIBLE_MEDICINE_UNITS_FOR_PER_WEIGHT.has(input.medicineUnit)) {
    return "medicine_unit is not eligible for per_weight";
  }
  if (input.strength == null || input.strength <= 0) {
    return "strength is missing or non-positive";
  }
  if (input.dosePerKg <= 0) {
    return "dose_per_kg is non-positive";
  }
  if (input.weightKg <= 0) {
    return "weight is not recorded";
  }
  if (
    input.doseBasis === MedicineDoseBasisPerDay &&
    (!input.frequencyPerDay || input.frequencyPerDay <= 0)
  ) {
    return "frequency_per_day is required for per_day basis";
  }
  return null;
}

function doseBasisFactor(input: DoseCalcInput): number {
  if (
    input.doseBasis === MedicineDoseBasisPerDay &&
    input.frequencyPerDay &&
    input.frequencyPerDay > 0
  ) {
    return 1 / input.frequencyPerDay;
  }
  return 1;
}

function clampRate(rate: number, lo?: number | null, hi?: number | null): number {
  let r = rate;
  if (lo != null && r < lo) r = lo;
  if (hi != null && r > hi) r = hi;
  return r;
}

/** value を step 刻みで mode 方向に丸める。step/mode が未指定なら丸めない。 */
export function roundQuantity(
  value: number,
  step?: number | null,
  mode?: MedicineRoundingMode | null,
): number {
  if (step == null || mode == null || step <= 0) return value;
  const n = value / step;
  switch (mode) {
    case MedicineRoundingModeUp:
      return Math.ceil(n - ROUND_EPSILON) * step;
    case MedicineRoundingModeDown:
      return Math.floor(n + ROUND_EPSILON) * step;
    case MedicineRoundingModeNearest:
      return Math.round(n) * step;
    default:
      return value;
  }
}

interface UpperCap {
  upperCapMg: number;
  hasUpperCap: boolean;
}

/** 1回投与あたりの上限（両上限の小さい方）。設定済みの上限がなければ hasUpperCap=false。 */
function perAdministrationUpperCap(input: DoseCalcInput, basisFactor: number): UpperCap {
  let upper = Number.POSITIVE_INFINITY;
  let has = false;
  if (input.maxMgPerKg != null) {
    const weightCap = input.maxMgPerKg * input.weightKg * basisFactor;
    if (weightCap < upper) upper = weightCap;
    has = true;
  }
  if (input.absoluteMaxDose != null) {
    let absCap = input.absoluteMaxDose;
    if (input.doseBasis === MedicineDoseBasisPerDay) {
      // 1日総量とみなし 1回あたりへ按分（per_day で frequency は適格性で > 0 を保証済み）。
      absCap *= basisFactor;
    }
    if (absCap < upper) upper = absCap;
    has = true;
  }
  return { upperCapMg: has ? upper : 0, hasUpperCap: has };
}

/**
 * #201 の計算モデルに従い 1回投与あたりの数量と実効用量を算出する。
 *
 * 計算順序:
 *   effective_rate = clamp(dose_per_kg, min, max)
 *   basis_factor   = 1 (per_administration) | 1/frequency_per_day (per_day)
 *   raw_mg         = effective_rate × weight_kg × basis_factor
 *   capped_mg      = min(raw_mg, weight×max_mg/kg×basis_factor, absolute_cap)  ← 設定済みの上限のみ
 *   quantity       = round(capped_mg ÷ strength, step, mode)
 *   effective_mg   = quantity × strength
 *   安全域判定      = effective_mg（C1: 丸め後）を上限/下限と比較
 */
export function calculateDose(input: DoseCalcInput): DoseCalcResult {
  const reason = doseEligibility(input);
  if (reason) return ineligible(reason);

  const strength = input.strength as number;
  const basisFactor = doseBasisFactor(input);
  const effectiveRate = clampRate(input.dosePerKg, input.minMgPerKg, input.maxMgPerKg);
  const rawMg = effectiveRate * input.weightKg * basisFactor;

  let cappedMg = rawMg;
  const { upperCapMg, hasUpperCap } = perAdministrationUpperCap(input, basisFactor);
  if (hasUpperCap && upperCapMg < cappedMg) cappedMg = upperCapMg;

  const quantity = roundQuantity(cappedMg / strength, input.roundingStep, input.roundingMode);
  const effectiveDoseMg = quantity * strength;

  let belowMin = false;
  let exceedsMax = false;
  // C1: 安全域判定は丸め後の実効用量(mg)で行う。
  if (hasUpperCap && effectiveDoseMg > upperCapMg + SAFETY_EPSILON) {
    exceedsMax = true;
  }
  if (input.minMgPerKg != null) {
    const lowerMg = input.minMgPerKg * input.weightKg * basisFactor;
    if (effectiveDoseMg < lowerMg - SAFETY_EPSILON) {
      belowMin = true;
    }
  }

  return {
    eligible: true,
    ineligibleReason: "",
    effectiveRate,
    rawMg,
    cappedMg,
    quantity,
    effectiveDoseMg,
    belowMin,
    exceedsMax,
  };
}

/**
 * 「著しい逸脱」（獣医の手動上書き検出）。BE EvaluateSavedDose の
 * DeviatesFromComputed と同一ロジック（computed に対する相対乖離）。
 * hard gate 表示の判定に使う。computedQty<=0（eligible=false 相当）は逸脱判定しない。
 */
export function isSignificantDoseDeviation(submittedQty: number, computedQty: number): boolean {
  if (computedQty <= 0) return false;
  const rel = Math.abs((submittedQty - computedQty) / computedQty);
  return rel > DOSE_DEVIATION_REL_THRESHOLD;
}

/** submittedQty（獣医が実際に入力した／保存しようとしている数量）を安全域に対して評価した結果。 */
export interface SubmittedDoseEvaluation {
  /** submittedQty × strength（実際に投与される実効用量） */
  effectiveDoseMg: number;
  upperCapMg: number | null;
  hasUpperCap: boolean;
  /** 実効用量が両上限の小さい方を上回る（丸め境界越え・手動入力いずれも検出） */
  exceedsMax: boolean;
  /** 実効用量が安全域下限を下回る */
  belowMin: boolean;
}

/**
 * 手動入力・丸め境界越えを含む「実際に保存されようとしている quantity」を安全域に対して評価する。
 * BE dose_revalidation.go の EvaluateSavedDose と同一ロジック（保存値の実効 mg を判定基準にする）。
 * calculateDose は「推奨プリフィル値」を計算するだけなので、hard gate は必ず本関数で
 * submittedQty（画面上の現在値）を評価すること — calculateDose の結果だけを見ると
 * 手動上書き後の危険な値を見逃す。
 */
export function evaluateSubmittedDose(
  input: DoseCalcInput,
  submittedQty: number,
): SubmittedDoseEvaluation {
  const strength = input.strength ?? 0;
  const basisFactor = doseBasisFactor(input);
  const effectiveDoseMg = submittedQty * strength;

  const { upperCapMg, hasUpperCap } = perAdministrationUpperCap(input, basisFactor);
  let exceedsMax = false;
  if (hasUpperCap && effectiveDoseMg > upperCapMg + SAFETY_EPSILON) {
    exceedsMax = true;
  }
  let belowMin = false;
  if (input.minMgPerKg != null) {
    const lowerMg = input.minMgPerKg * input.weightKg * basisFactor;
    if (effectiveDoseMg < lowerMg - SAFETY_EPSILON) {
      belowMin = true;
    }
  }

  return {
    effectiveDoseMg,
    upperCapMg: hasUpperCap ? upperCapMg : null,
    hasUpperCap,
    exceedsMax,
    belowMin,
  };
}
