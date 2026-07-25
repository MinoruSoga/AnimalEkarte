import { calculateDose, evaluateSubmittedDose, isSignificantDoseDeviation } from "@/lib/medicine-dose";
import type { DoseCalcInput } from "@/lib/medicine-dose";

/**
 * #201: TreatmentRow の quantity コミット時に投与量ゲートを判定する。
 *
 * マスタの絶対上限超過だけは保存不能とし、下限未満・推奨値からの乖離は従来どおり
 * 確認理由として返す。input が null（medicine が per_weight でない・species マップ不能・
 * 体重未記録・当該 species の param 未設定）の場合は手動入力・既定の保存挙動を維持する。
 */
export interface DoseGateResult {
  /** マスタの絶対上限超過により保存できないか */
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

export function computeDoseGate(input: DoseCalcInput | null, submittedQty: number): DoseGateResult {
  if (!input) {
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
