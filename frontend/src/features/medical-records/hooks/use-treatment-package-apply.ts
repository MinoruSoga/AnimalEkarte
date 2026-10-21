/**
 * EMR-524: package preview/apply/lookup/再表示の react-query 接続（契約mock段階）。
 * runtime 未提供ではクエリを発行しない（TREATMENT_PACKAGE_RUNTIME_AVAILABLE=false）。
 * 成功後は application IDs を確認してから plan 一覧を再取得し、
 * 再取得失敗を apply 失敗にしない。
 */
import { useCallback, useEffect, useMemo, useRef, useState } from "react";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { QUERY_STALE_TIMES, QUERY_GC_TIMES } from "@/lib/react-query";
import {
  applyTreatmentPackage,
  listTreatmentPackageCandidates,
  getMedicalRecordOperationResult,
  listMedicalRecordPlanItems,
  listTreatmentPackageApplications,
  previewTreatmentPackage,
  type TreatmentPackageApply,
} from "../api/treatment-package-applications";
import {
  TREATMENT_PACKAGE_RUNTIME_AVAILABLE,
  buildApplyCommand,
  resolveOperationResult,
  scopeKey,
  type OperationScope,
  type TreatmentPackageEditContext,
} from "../lib/treatment-package-operation-state";

export const packageApplyKeys = {
  applications: (recordId: string) => ["treatment-package-applications", recordId] as const,
  planItems: (recordId: string) => ["medical-record-plan-items", recordId] as const,
};

export function useTreatmentPackageCandidates(clinicId?: string) {
  return useQuery({
    queryKey: ["treatment-package-candidates", clinicId ?? ""] as const,
    queryFn: () => listTreatmentPackageCandidates(clinicId),
    enabled: TREATMENT_PACKAGE_RUNTIME_AVAILABLE,
    staleTime: QUERY_STALE_TIMES.STATIC,
    gcTime: QUERY_GC_TIMES.LONG,
  });
}

export function useMedicalRecordPlanItems(recordId: string | null, clinicId?: string) {
  return useQuery({
    queryKey: [...packageApplyKeys.planItems(recordId ?? ""), clinicId ?? ""] as const,
    queryFn: () => listMedicalRecordPlanItems(recordId as string, clinicId),
    enabled: recordId !== null && TREATMENT_PACKAGE_RUNTIME_AVAILABLE,
    staleTime: QUERY_STALE_TIMES.REALTIME,
    gcTime: QUERY_GC_TIMES.STANDARD,
  });
}

export function useTreatmentPackageApplications(recordId: string | null, clinicId?: string) {
  return useQuery({
    queryKey: [...packageApplyKeys.applications(recordId ?? ""), clinicId ?? ""] as const,
    queryFn: () => listTreatmentPackageApplications(recordId as string, clinicId),
    enabled: recordId !== null && TREATMENT_PACKAGE_RUNTIME_AVAILABLE,
    staleTime: QUERY_STALE_TIMES.REALTIME,
    gcTime: QUERY_GC_TIMES.STANDARD,
  });
}

export function useTreatmentPackagePreview(recordId: string | null, clinicId?: string) {
  return useMutation({
    mutationFn: ({ packageId, packageVersion }: { packageId: string; packageVersion: string }) =>
      previewTreatmentPackage(
        recordId as string,
        { package_id: packageId, package_version: packageVersion },
        clinicId,
      ),
  });
}

export type ApplyState =
  | { kind: "idle" }
  | { kind: "submitting"; operationKey: string }
  | { kind: "unknown"; operationKey: string }
  | { kind: "succeeded"; applicationId: string; planItemsVersion: string }
  | { kind: "failed"; status: string };

export interface UseTreatmentPackageApplyArgs {
  scope: OperationScope;
  editContext: TreatmentPackageEditContext | null;
  runtimeAvailable?: boolean;
}

/**
 * 明示 apply → C06 応答解釈 → 不明は同 key 照会 → 成功後 application IDs 確認 →
 * plan 一覧再取得。scope 切替後の遅延応答は破棄する。
 */
interface PendingOperation {
  operationKey: string;
  recordId: string;
  clinicId?: string;
}

export function useTreatmentPackageApply({
  scope,
  editContext,
  runtimeAvailable = TREATMENT_PACKAGE_RUNTIME_AVAILABLE,
}: UseTreatmentPackageApplyArgs) {
  const qc = useQueryClient();
  const scopeRef = useRef(scope);
  scopeRef.current = scope;
  const currentScopeKey = scopeKey(scope);
  const [state, setState] = useState<ApplyState>({ kind: "idle" });
  // scopeKey → 未解決の pending operation。key と発行時 scope を不可分に保持し、
  // 照会は発行時 scope へ固定する（切替後の対象へ混ぜない）。
  const pendingRef = useRef<Map<string, PendingOperation>>(new Map());

  // scope 切替: 旧 scope の保留表示を新対象へ混ぜない。元 scope に戻れば同 key 照会を再開できる。
  useEffect(() => {
    const held = pendingRef.current.get(currentScopeKey);
    setState(
      held === undefined ? { kind: "idle" } : { kind: "unknown", operationKey: held.operationKey },
    );
  }, [currentScopeKey]);

  const apply = useMutation({
    mutationFn: async (command: { body: TreatmentPackageApply; operationKey: string }) => {
      const current = scopeRef.current;
      return applyTreatmentPackage(current.recordId, command.body, current.clinicId);
    },
  });

  const submit = useCallback(
    async (preview: Parameters<typeof buildApplyCommand>[0]["preview"]) => {
      if (!runtimeAvailable || editContext === null) return;
      const sk = scopeKey(scope);
      // 同 scope に未解決 key がある限り新しい明示送信はしない（同 key 照会を先に解決）。
      if (pendingRef.current.has(sk)) return;
      const command = buildApplyCommand({ preview, editContext, scope });
      pendingRef.current.set(sk, {
        operationKey: command.operationKey,
        recordId: scope.recordId,
        clinicId: scope.clinicId,
      });
      setState({ kind: "submitting", operationKey: command.operationKey });
      try {
        const outcome = await apply.mutateAsync({
          body: command.body,
          operationKey: command.operationKey,
        });
        if (scopeKey(scopeRef.current) !== sk) return; // 切替後の遅延応答は混ぜない
        if (outcome.kind === "unavailable") {
          setState({ kind: "unknown", operationKey: command.operationKey });
          return;
        }
        if (outcome.kind === "error-envelope") {
          // 前段 error envelope: 5xx は一時障害として照会継続、4xx は確定失敗。
          if (outcome.httpStatus >= 500) {
            setState({ kind: "unknown", operationKey: command.operationKey });
            return;
          }
          pendingRef.current.delete(sk);
          setState({ kind: "failed", status: outcome.code ?? `http-${outcome.httpStatus}` });
          return;
        }
        const resolution = resolveOperationResult(outcome.result);
        if (resolution.kind === "success") {
          pendingRef.current.delete(sk);
          const applicationId = resolution.result.result_refs.application_id;
          // 成功確認 → plan 一覧再取得。再取得失敗は apply 失敗にしない。
          setState({
            kind: "succeeded",
            applicationId,
            planItemsVersion: resolution.result.committed_versions.plan_items,
          });
          qc.invalidateQueries({ queryKey: packageApplyKeys.applications(scope.recordId) });
          qc.invalidateQueries({ queryKey: packageApplyKeys.planItems(scope.recordId) });
          return;
        }
        if (resolution.kind === "retry-lookup") {
          setState({ kind: "unknown", operationKey: command.operationKey });
          return;
        }
        pendingRef.current.delete(sk);
        setState({ kind: "failed", status: resolution.status });
      } catch {
        // 通信断/timeout/parse 失敗: 元 key 照会へ。新 key の盲目再送はしない。
        if (scopeKey(scopeRef.current) !== sk) return;
        setState({ kind: "unknown", operationKey: command.operationKey });
      }
    },
    [apply, editContext, qc, runtimeAvailable, scope],
  );

  /** 結果不明を元 key で照会。照会先は発行時 scope に固定し、成功確定まで新 key は発行しない。 */
  const resolveUnknown = useCallback(async () => {
    const sk = scopeKey(scopeRef.current);
    const held = pendingRef.current.get(sk);
    if (held === undefined) return;
    const outcome = await getMedicalRecordOperationResult(
      held.recordId,
      held.operationKey,
      held.clinicId,
    );
    if (scopeKey(scopeRef.current) !== sk) return;
    if (outcome.kind === "unavailable") return; // 取得失敗は unknown 維持
    if (outcome.kind === "error-envelope") {
      if (outcome.httpStatus >= 500) return;
      pendingRef.current.delete(sk);
      setState({ kind: "failed", status: outcome.code ?? `http-${outcome.httpStatus}` });
      return;
    }
    const resolution = resolveOperationResult(outcome.result);
    if (resolution.kind === "success") {
      pendingRef.current.delete(sk);
      setState({
        kind: "succeeded",
        applicationId: resolution.result.result_refs.application_id,
        planItemsVersion: resolution.result.committed_versions.plan_items,
      });
      qc.invalidateQueries({ queryKey: packageApplyKeys.applications(held.recordId) });
      qc.invalidateQueries({ queryKey: packageApplyKeys.planItems(held.recordId) });
      return;
    }
    if (resolution.kind === "retry-lookup") return; // まだ不明
    pendingRef.current.delete(sk);
    setState({ kind: "failed", status: resolution.status });
  }, [qc]);

  /** 明示 reset で現在 scope の保留状態を破棄（遅延成功を別対象へ混ぜない）。 */
  const reset = useCallback(() => {
    pendingRef.current.delete(scopeKey(scopeRef.current));
    setState({ kind: "idle" });
  }, []);

  const heldOperationKey = pendingRef.current.get(currentScopeKey)?.operationKey ?? null;
  return useMemo(
    () => ({ state, submit, resolveUnknown, reset, heldOperationKey }),
    [state, submit, resolveUnknown, reset, heldOperationKey],
  );
}
