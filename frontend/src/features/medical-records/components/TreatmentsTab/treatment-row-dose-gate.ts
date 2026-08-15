import { calculateDose, evaluateSubmittedDose, isSignificantDoseDeviation } from "@/lib/medicine-dose";
import type { DoseCalcInput } from "@/lib/medicine-dose";

import {
  DOSE_PARAMS_LOOKUP_FAILED_MESSAGE,
  type DoseParamsAuthority,
} from "../../api/medicine-dose-lookup";

/**
 * #201: TreatmentRow の quantity コミット時に投与量ゲートを判定する。
 *
 * マスタの絶対上限超過だけは保存不能とし、下限未満・推奨値からの乖離は従来どおり
 * 確認理由として返す。
 *
 * TASK-025: DoseGateSource で technical failure と missing data を型で区別する。
 * - ready: DoseCalcInput が揃い臨床評価可能
 * - missing: 体重未記録・species 解決不能・param 未設定等 → 手動入力・保存継続（TASK-033 まで）
 * - technical_failure: dose-params 取得失敗 → 保存停止 + 固定文言
 */
export type DoseGateSource =
  | { kind: "ready"; input: DoseCalcInput }
  | { kind: "missing" }
  | { kind: "technical_failure" };

export interface DoseGateResult {
  /** マスタの絶対上限超過、または technical failure により保存できないか */
  isBlocked: boolean;
  /** 保存不能時にインライン表示する理由 */
  blockReason: string;
  /**
   * TASK-377: 上限内の下限割れ / 推奨値からの著しい乖離で free-text 逸脱理由が必須か。
   * 上限超過（isBlocked）では false（理由で解除しない）。旧 requiresConfirm を reason-required 契約へ同期。
   */
  requiresDeviationReason: boolean;
  /** 非ブロックの注意事項を含む理由（複数該当時は " / " 区切り） */
  reason: string;
  warning: "none" | "below-min" | "exceeds-max";
  /** ユーザーへの参考表示用（推奨値との比較） */
  recommendedQuantity: number | null;
  effectiveDoseMg: number;
}

/**
 * DoseCalcInput と DoseParamsAuthority からゲート入力を組み立てる。
 * authority.failed を doseCalcInput=null（missing）と同一視しない。
 */
export function resolveDoseGateSource(
  doseCalcInput: DoseCalcInput | null,
  authority: DoseParamsAuthority,
): DoseGateSource {
  if (authority.status === "failed") return { kind: "technical_failure" };
  if (doseCalcInput) return { kind: "ready", input: doseCalcInput };
  return { kind: "missing" };
}

/**
 * TASK-025: 引数は DoseGateSource のみ。`DoseCalcInput | null` を受け付けてはならない。
 * null を許すと「技術障害を missing と同一視して保存を通す」経路が型の上で復活する。
 */
export function computeDoseGate(
  source: DoseGateSource,
  submittedQty: number,
): DoseGateResult {
  const normalized = source;

  if (normalized.kind === "technical_failure") {
    return {
      isBlocked: true,
      blockReason: DOSE_PARAMS_LOOKUP_FAILED_MESSAGE,
      requiresDeviationReason: false,
      reason: "",
      warning: "none",
      recommendedQuantity: null,
      effectiveDoseMg: 0,
    };
  }

  if (normalized.kind === "missing") {
    return {
      isBlocked: false,
      blockReason: "",
      requiresDeviationReason: false,
      reason: "",
      warning: "none",
      recommendedQuantity: null,
      effectiveDoseMg: 0,
    };
  }

  const input = normalized.input;
  const recommended = calculateDose(input);
  const submitted = evaluateSubmittedDose(input, submittedQty);

  const reasons: string[] = [];
  if (submitted.exceedsMax) reasons.push("実効用量が安全域の上限を超えています");
  if (submitted.belowMin) reasons.push("実効用量が安全域の下限を下回っています");
  const significantDeviation =
    recommended.eligible && isSignificantDoseDeviation(submittedQty, recommended.quantity);
  if (significantDeviation) {
    reasons.push(`推奨値（${recommended.quantity}）から著しく乖離しています`);
  }

  const warning: DoseGateResult["warning"] = submitted.exceedsMax
    ? "exceeds-max"
    : submitted.belowMin
      ? "below-min"
      : "none";

  // 上限超過は hard block。理由必須は上限内の下限割れ / 著しい乖離のみ。
  const requiresDeviationReason =
    !submitted.exceedsMax && (submitted.belowMin || significantDeviation);

  return {
    isBlocked: submitted.exceedsMax,
    blockReason: submitted.exceedsMax
      ? "実効用量がマスタで設定された絶対上限を超えているため保存できません"
      : "",
    requiresDeviationReason,
    reason: reasons.join(" / "),
    warning,
    recommendedQuantity: recommended.eligible ? recommended.quantity : null,
    effectiveDoseMg: submitted.effectiveDoseMg,
  };
}
