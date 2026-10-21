import type { CertificateErrorKind } from "../api/certificate-errors";
import type { RecoveryLookupNotice } from "../hooks/use-certificate-recovery";

/** EMR-487 Slice A: 回復 discovery の画面文言。機械判定は state / kind で行い、BE の本文は表示に使わない。 */

export type RecoveryOperationLabel = "発行" | "再印刷";

export const RECOVERY_INTENT_LABEL = { issue: "発行", reprint: "再印刷" } as const;

export const RECOVERY_LOADING = "未確定の操作がないか確認しています…";

export function recoveryErrorText(kind: CertificateErrorKind | null, label: RecoveryOperationLabel): string {
  const base = `未確定の${label}操作を確認できないため、確認できるまで新しい${label}は開始できません。`;
  if (kind === "forbidden") return `${base}この記録を確認する権限がない可能性があります。`;
  if (kind === "unauthenticated") return `${base}再ログイン後にやり直してください。`;
  return base;
}

export function recoveryHoldText(label: RecoveryOperationLabel): { title: string; detail: string } {
  return {
    title: `結果が確定していない${label}操作があります`,
    detail: `遅れて完了する可能性があります。結果を再照会してください。確認できるまで新しい${label}は開始できません（一覧が空でも、取得が完了しても、次の操作が安全とは限りません）。`,
  };
}

export const RECOVERY_TRUNCATED_TEXT =
  "確認できていない古い操作が残っている可能性があります。確認できるまで新しい操作は開始できません。";
export const RECOVERY_ACTIVE_UNVERIFIED_TEXT =
  "進行中の操作を一覧から確認できません。再確認してください。確認できるまで新しい操作は開始できません。";

export const RECOVERY_HELD_REASON = {
  unresolved: "結果未確認",
  unavailable: "結果を取得できません",
} as const;

export const RECOVERY_LOOKUP_NOTICE: Record<RecoveryLookupNotice, string> = {
  "not-succeeded":
    "現時点では成功を確認できません。遅れて完了する可能性があるため、保留のままです。",
  unavailable: "結果を確認できませんでした。保留のままです。時間をおいて再照会してください。",
  mismatch: "応答が対象の操作と一致しないため採用しません。保留のままです。",
};

export const RECOVERY_CLOSED_TEXT = "効果なしで終結済みの操作です（参考表示）。";
export const RECOVERY_OWN_PENDING_TEXT = "この画面で進行中の操作です。上の通知から結果を照会してください。";
