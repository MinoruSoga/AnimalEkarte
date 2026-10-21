/**
 * EMR-524: package apply の C06 operation 状態機械（純粋ロジック・契約mock段階）。
 * 契約正本: docs/work/emr/emr-518-package-api-contract-20261020.md と
 * backend/docs/api.yaml `TreatmentPackageApply` / `MedicalRecordPlanOperationResultResponse`
 * （採用 SHA f1ded6ec367459d91baec08414ce780df39e9a3f）。
 *
 * - operation_key は明示 apply の時だけ生成する。連打・応答断は同 key 照会。
 * - in-progress / result-unavailable / not-succeeded を成功へ変換しない。
 * - 内容変更は元 unknown 解決後に別 key の別確認。2回目の明示 apply は別 application。
 * - 医院/患者/record 切替後の遅延成功を別対象へ混ぜない（scopeKey で破棄）。
 * - 財務 unknown は warning、owner/classification 未解決は blocking。total null を 0 にしない。
 */
import type {
  MedicalRecordPlanOperationResult,
  OperationSuccess,
  TreatmentPackagePreview,
} from "../api/treatment-package-applications";

/** runtime apply/preview/application/plan-items/lookup の BE は未登録。
 *  既定で入口を無効にする明示的ゲート。環境変数では自動有効化しない。 */
export const TREATMENT_PACKAGE_RUNTIME_AVAILABLE = false as const;

/** apply に必要な編集 context（EMR-494 公開 wire に合わせる adapter 経由で供給）。
 *  mock では contract fixture 相当値を注入できる。null = 供給不可 → apply を送らない。 */
export interface TreatmentPackageEditContext {
  editSessionRef: string;
  editEpoch: string;
  /** plan_items 依存版（owner_expected_version）。未初期化は null ではなく供給不可。 */
  ownerExpectedVersion: string;
}

/** 操作が属する対象。切替を検出するための同一性キー。 */
export interface OperationScope {
  recordId: string;
  clinicId?: string;
  petId?: string;
}

export function scopeKey(scope: OperationScope): string {
  return `${scope.clinicId ?? ""}/${scope.petId ?? ""}/${scope.recordId}`;
}

export type ApplyPhase =
  | { kind: "idle" }
  | { kind: "awaiting-preview"; packageId: string; packageVersion: string }
  | {
      kind: "previewed";
      preview: TreatmentPackagePreview;
      blocking: readonly BlockingIssue[];
      warnings: readonly WarningIssue[];
    }
  | { kind: "submitting"; operationKey: string; scopeKey: string }
  | { kind: "unknown"; operationKey: string; scopeKey: string; reason: string }
  | { kind: "succeeded"; operationKey: string; scopeKey: string; result: OperationSuccess }
  | { kind: "failed"; operationKey: string; scopeKey: string; status: string };

export interface BlockingIssue {
  componentKey: string;
  code: string;
  message: string;
}
export interface WarningIssue {
  componentKey: string;
  code: string;
  message: string;
  requiredAt: "selection" | "execution" | "billing";
}

/** preview を行別 blocking/warning へ分類する。applicable=false かつ
 *  blocking 0件（CAS scope未ready等）は server 側で既に applicable:false。 */
export function classifyPreview(preview: TreatmentPackagePreview): {
  blocking: BlockingIssue[];
  warnings: WarningIssue[];
  applicable: boolean;
} {
  const blocking: BlockingIssue[] = [];
  const warnings: WarningIssue[] = [];
  for (const item of preview.items) {
    for (const issue of item.issues) {
      if (issue.severity === "blocking") {
        blocking.push({ componentKey: issue.component_key, code: issue.code, message: issue.message });
      } else {
        warnings.push({
          componentKey: issue.component_key,
          code: issue.code,
          message: issue.message,
          requiredAt: issue.required_at,
        });
      }
    }
  }
  return { blocking, warnings, applicable: preview.applicable && blocking.length === 0 };
}

/** 送信可否の最終境界。view 権・編集 holder/epoch・preview applicable が揃わないと送らない。 */
export function canSubmitApply(args: {
  runtimeAvailable: boolean;
  editContext: TreatmentPackageEditContext | null;
  preview: TreatmentPackagePreview | null;
  canCreate: boolean;
}): { ok: true } | { ok: false; reason: string } {
  if (!args.runtimeAvailable) return { ok: false, reason: "runtime 未提供" };
  if (!args.canCreate) return { ok: false, reason: "作成権限がありません" };
  if (args.editContext === null) return { ok: false, reason: "編集コンテキスト未提供" };
  if (args.preview === null || !classifyPreview(args.preview).applicable) {
    return { ok: false, reason: "適用不可のプレビューです" };
  }
  return { ok: true };
}

export interface ApplyCommand {
  operationKey: string;
  body: {
    operation_key: string;
    edit_session_ref: string;
    edit_epoch: string;
    owner_expected_version: string;
    package_id: string;
    package_version: string;
    preview_token: string;
  };
  scopeKey: string;
}

/** 明示 apply 時のみ operation key を1個生成して body を組み立てる。 */
export function buildApplyCommand(args: {
  preview: TreatmentPackagePreview;
  editContext: TreatmentPackageEditContext;
  scope: OperationScope;
  generateKey?: () => string;
}): ApplyCommand {
  const key = (args.generateKey ?? (() => crypto.randomUUID()))();
  return {
    operationKey: key,
    scopeKey: scopeKey(args.scope),
    body: {
      operation_key: key,
      edit_session_ref: args.editContext.editSessionRef,
      edit_epoch: args.editContext.editEpoch,
      owner_expected_version: args.editContext.ownerExpectedVersion,
      package_id: args.preview.package_id,
      package_version: args.preview.package_version,
      preview_token: args.preview.preview_token,
    },
  };
}

/** scope 一致のみ受理。切替後の遅延応答は破棄する。 */
export function resultInScope(
  currentScopeKey: string,
  heldScopeKey: string,
): boolean {
  return currentScopeKey === heldScopeKey;
}

export type Resolution =
  | { kind: "success"; result: OperationSuccess }
  | { kind: "retry-lookup" }
  | { kind: "failed"; status: string };

/**
 * operation 応答を解釈する。succeeded/replayed のみ成功。
 * in-progress/result-unavailable は同 key 照会継続。他の非成功は失敗確定。
 */
export function resolveOperationResult(result: MedicalRecordPlanOperationResult): Resolution {
  if (result.status === "succeeded" || result.status === "replayed") {
    return { kind: "success", result };
  }
  if (result.status === "operation-in-progress" || result.status === "result-unavailable") {
    return { kind: "retry-lookup" };
  }
  return { kind: "failed", status: result.status };
}

/** 通信断/timeout/parse 失敗は「結果不明」→ 元 key 照会へ。盲目的な新 key 作成はしない。 */
export function resolutionForTransportFailure(): Resolution {
  return { kind: "retry-lookup" };
}
