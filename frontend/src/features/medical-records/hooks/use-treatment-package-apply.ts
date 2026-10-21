/**
 * EMR-524: package preview/apply/lookup/再表示の react-query 接続（契約mock段階）。
 * runtime 未提供ではクエリを発行しない（TREATMENT_PACKAGE_RUNTIME_AVAILABLE=false）。
 * 成功後は application IDs を確認してから plan 一覧を再取得し、
 * 再取得失敗を apply 失敗にしない。
 */
import { useCallback, useMemo, useRef, useState } from "react";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { QUERY_STALE_TIMES, QUERY_GC_TIMES } from "@/lib/react-query";
import {
  applyTreatmentPackage,
  listTreatmentPackageCandidates,
  getMedicalRecordOperationResult,
  listMedicalRecordPlanItems,
  listTreatmentPackageApplications,
  previewTreatmentPackage,
  type MedicalRecordPlanOperationResult,
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
export function useTreatmentPackageApply({
  scope,
  editContext,
  runtimeAvailable = TREATMENT_PACKAGE_RUNTIME_AVAILABLE,
}: UseTreatmentPackageApplyArgs) {
  const qc = useQueryClient();
  const scopeRef = useRef(scope);
  scopeRef.current = scope;
  const [state, setState] = useState<ApplyState>({ kind: "idle" });
  const heldKeyRef = useRef<string | null>(null);

  const lookupKey = useCallback(
    async (operationKey: string): Promise<MedicalRecordPlanOperationResult | null> => {
      const current = scopeRef.current;
      try {
        const result = await getMedicalRecordOperationResult(
          current.recordId,
          operationKey,
          current.clinicId,
        );
        return result;
      } catch {
        return null;
      }
    },
    [],
  );

  const apply = useMutation({
    mutationFn: async (command: { body: TreatmentPackageApply; operationKey: string }) => {
      const current = scopeRef.current;
      return applyTreatmentPackage(current.recordId, command.body, current.clinicId);
    },
  });

  const submit = useCallback(
    async (preview: Parameters<typeof buildApplyCommand>[0]["preview"]) => {
      if (!runtimeAvailable || editContext === null) return;
      // 不明保留の key がある限り新しい明示送信はしない（同 key 照会を先に解決）。
      if (heldKeyRef.current !== null) return;
      const command = buildApplyCommand({ preview, editContext, scope });
      heldKeyRef.current = command.operationKey;
      setState({ kind: "submitting", operationKey: command.operationKey });
      const heldScope = scopeKey(scope);
      try {
        const result = await apply.mutateAsync({ body: command.body, operationKey: command.operationKey });
        if (scopeKey(scopeRef.current) !== heldScope) return; // 切替後の遅延応答は混ぜない
        const resolution = resolveOperationResult(result);
        if (resolution.kind === "success") {
          heldKeyRef.current = null;
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
        heldKeyRef.current = null;
        setState({ kind: "failed", status: resolution.status });
      } catch {
        // 通信断/timeout/parse 失敗: 元 key 照会へ。新 key の盲目再送はしない。
        if (scopeKey(scopeRef.current) !== heldScope) return;
        setState({ kind: "unknown", operationKey: command.operationKey });
      }
    },
    [apply, editContext, qc, runtimeAvailable, scope],
  );

  /** 結果不明を元 key で照会。成功確定まで新 key は発行しない。 */
  const resolveUnknown = useCallback(async () => {
    const held = heldKeyRef.current;
    if (held === null) return;
    const heldScope = scopeKey(scopeRef.current);
    const result = await lookupKey(held);
    if (scopeKey(scopeRef.current) !== heldScope) return;
    if (result === null) return; // 取得失敗は unknown 維持
    const resolution = resolveOperationResult(result);
    if (resolution.kind === "success") {
      heldKeyRef.current = null;
      setState({
        kind: "succeeded",
        applicationId: resolution.result.result_refs.application_id,
        planItemsVersion: resolution.result.committed_versions.plan_items,
      });
      qc.invalidateQueries({ queryKey: packageApplyKeys.applications(scope.recordId) });
      qc.invalidateQueries({ queryKey: packageApplyKeys.planItems(scope.recordId) });
      return;
    }
    if (resolution.kind === "retry-lookup") return; // まだ不明
    heldKeyRef.current = null;
    setState({ kind: "failed", status: resolution.status });
  }, [lookupKey, qc, scope.recordId]);

  /** scope 切替・明示 reset で保留状態を破棄（遅延成功を別対象へ混ぜない）。 */
  const reset = useCallback(() => {
    heldKeyRef.current = null;
    setState({ kind: "idle" });
  }, []);

  return useMemo(
    () => ({ state, submit, resolveUnknown, reset, heldOperationKey: heldKeyRef.current }),
    [state, submit, resolveUnknown, reset],
  );
}
