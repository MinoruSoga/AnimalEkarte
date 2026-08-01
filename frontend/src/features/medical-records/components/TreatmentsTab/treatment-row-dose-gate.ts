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
  requiresConfirm: boolean;
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

/** null / DoseCalcInput 互換を含む正規化（既存 call site の移行用）。 */
function normalizeDoseGateSource(
  source: DoseGateSource | DoseCalcInput | null,
): DoseGateSource {
  if (source === null) return { kind: "missing" };
  if (typeof source === "object" && "kind" in source) {
    return source;
  }
  return { kind: "ready", input: source };
}

export function computeDoseGate(
  source: DoseGateSource | DoseCalcInput | null,
  submittedQty: number,
): DoseGateResult {
  const normalized = normalizeDoseGateSource(source);

  if (normalized.kind === "technical_failure") {
    return {
      isBlocked: true,
      blockReason: DOSE_PARAMS_LOOKUP_FAILED_MESSAGE,
      requiresConfirm: false,
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
      requiresConfirm: false,
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
  if (recommended.eligible && isSignificantDoseDeviation(submittedQty, recommended.quantity)) {
    reasons.push(`推奨値（${recommended.quantity}）から著しく乖離しています`);
  }

  const warning: DoseGateResult["warning"] = submitted.exceedsMax
    ? "exceeds-max"
    : submitted.belowMin
      ? "below-min"
      : "none";

  return {
    isBlocked: submitted.exceedsMax,
    blockReason: submitted.exceedsMax
      ? "実効用量がマスタで設定された絶対上限を超えているため保存できません"
      : "",
    requiresConfirm: reasons.length > 0,
    reason: reasons.join(" / "),
    warning,
    recommendedQuantity: recommended.eligible ? recommended.quantity : null,
    effectiveDoseMg: submitted.effectiveDoseMg,
  };
}
