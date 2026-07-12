import { calculateDose, evaluateSubmittedDose, isSignificantDoseDeviation } from "@/lib/medicine-dose";
import type { DoseCalcInput } from "@/lib/medicine-dose";

/**
 * #201: TreatmentRow の quantity コミット時に確認ダイアログ（hard gate）を出すべきかを判定する。
 * PRODUCT_PHILOSOPHY の「確認ダイアログ禁止」より SPECIFICATION 2.1 の臨床安全を優先する適用例
 * ——警告に格下げしないこと（致死性用量の素通りを意味する）。
 *
 * input が null（medicine が per_weight でない・species マップ不能・体重未記録・
 * 当該 species の param 未設定）の場合は常に requiresConfirm=false（手動入力・既定挙動）。
 */
export interface DoseGateResult {
  requiresConfirm: boolean;
  /** ダイアログ本文に表示する理由（複数該当時は " / " 区切り） */
  reason: string;
  warning: "none" | "below-min" | "exceeds-max";
  /** ユーザーへの参考表示用（推奨値との比較） */
  recommendedQuantity: number | null;
  effectiveDoseMg: number;
}

export function computeDoseGate(input: DoseCalcInput | null, submittedQty: number): DoseGateResult {
  if (!input) {
    return { requiresConfirm: false, reason: "", warning: "none", recommendedQuantity: null, effectiveDoseMg: 0 };
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
    requiresConfirm: reasons.length > 0,
    reason: reasons.join(" / "),
    warning,
    recommendedQuantity: recommended.eligible ? recommended.quantity : null,
    effectiveDoseMg: submitted.effectiveDoseMg,
  };
}
