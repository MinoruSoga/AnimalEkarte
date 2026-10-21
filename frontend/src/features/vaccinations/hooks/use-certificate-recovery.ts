import { useCallback, useEffect, useLayoutEffect, useMemo, useReducer, useRef } from "react";
import type { CertificateApi } from "../api/certificate-api";
import { boundSucceededRefs, resultMatchesOperation } from "../api/certificate-binding";
import { CertificateApiError, type CertificateErrorKind } from "../api/certificate-errors";
import { CERTIFICATE_RECOVERY_FENCE_AVAILABLE } from "../lib/certificate-recovery-gate";
import {
  discoverCertificateOperations,
  type CertificateRecoveryDiscovery,
  type CertificateRecoveryTarget,
} from "../lib/certificate-recovery-discovery";
import {
  deriveRecoveryView,
  type RecoveryLookupNotice,
  type RecoveryRefs,
  type RecoveryView,
} from "../lib/certificate-recovery-view";

/**
 * EMR-487 Slice A: 再 mount / reload 後に、職員本人の未確定・完了済みの issue / reprint を発見して保留を決める。
 *
 * - メモリのみ。operation key・世代・本文を localStorage / sessionStorage / IndexedDB / URL に置かない。
 * - 既存の in-memory pending-key 登録簿には書かない・消さない。発見した key から IssuePending / ReprintPending
 *   （本文を持たない）を作らない。発見した操作の再 lookup は同じ key の lookupOperation だけで、再送はしない。
 * - discovery が空・complete でも新 key を作る根拠にしない。ここは「既知の阻害要因があるか」だけを返す。
 */

interface LookupState {
  readonly checking: boolean;
  readonly notice: RecoveryLookupNotice | null;
}

export interface RecoveryState {
  readonly status: "idle" | "loading" | "ready" | "error";
  readonly run: number;
  readonly discovery: CertificateRecoveryDiscovery | null;
  readonly errorKind: CertificateErrorKind | null;
  readonly resolved: ReadonlyMap<string, RecoveryRefs>;
  readonly lookups: Readonly<Record<string, LookupState>>;
}

type RecoveryAction =
  | { type: "restart" }
  | { type: "settled"; run: number; discovery: CertificateRecoveryDiscovery }
  | { type: "failed"; run: number; errorKind: CertificateErrorKind }
  | { type: "lookupStart"; run: number; key: string }
  | { type: "lookupDone"; run: number; key: string; refs: RecoveryRefs | null; notice: RecoveryLookupNotice | null };

export function initialRecoveryState(enabled: boolean): RecoveryState {
  return {
    status: enabled ? "loading" : "idle",
    run: 0,
    discovery: null,
    errorKind: null,
    resolved: new Map(),
    lookups: {},
  };
}

export function recoveryReducer(state: RecoveryState, action: RecoveryAction): RecoveryState {
  switch (action.type) {
    case "restart":
      return { ...initialRecoveryState(true), run: state.run + 1 };
    case "settled":
      if (action.run !== state.run) return state;
      return { ...state, status: "ready", discovery: action.discovery, errorKind: null };
    case "failed":
      if (action.run !== state.run) return state;
      return { ...state, status: "error", discovery: null, errorKind: action.errorKind };
    case "lookupStart":
      if (action.run !== state.run) return state;
      return { ...state, lookups: { ...state.lookups, [action.key]: { checking: true, notice: null } } };
    case "lookupDone": {
      if (action.run !== state.run) return state;
      const resolved = new Map(state.resolved);
      if (action.refs !== null) resolved.set(action.key, action.refs);
      return {
        ...state,
        resolved,
        lookups: { ...state.lookups, [action.key]: { checking: false, notice: action.notice } },
      };
    }
  }
}

type Outcome =
  | { kind: "ready"; discovery: CertificateRecoveryDiscovery }
  | { kind: "error"; errorKind: CertificateErrorKind };

export interface UseCertificateRecoveryArgs {
  api: CertificateApi;
  target: CertificateRecoveryTarget;
  /** false なら discovery を走らせず、開始は常に不可（fail-closed）。 */
  enabled: boolean;
}

export function useCertificateRecovery({ api, target, enabled }: UseCertificateRecoveryArgs) {
  const { intent, vaccinationId } = target;
  const editionId = target.intent === "reprint" ? target.editionId : null;
  const [state, dispatch] = useReducer(recoveryReducer, enabled, initialRecoveryState);
  // 非同期ハンドラが最新の state を読むための ref（render 中には書かない）。
  const stateRef = useRef(state);
  useLayoutEffect(() => {
    stateRef.current = state;
  }, [state]);
  const inflightRef = useRef<Promise<Outcome> | null>(null);
  const outcomeRef = useRef<Outcome | null>(null);
  const resolvedRef = useRef<ReadonlyMap<string, RecoveryRefs>>(new Map());
  const lookupBusyRef = useRef(new Set<string>());

  useEffect(() => {
    if (!enabled) return;
    const controller = new AbortController();
    const run = state.run;
    const discoveryTarget: CertificateRecoveryTarget =
      intent === "issue" ? { intent, vaccinationId } : { intent, vaccinationId, editionId: editionId ?? 0 };
    const promise: Promise<Outcome> = discoverCertificateOperations(api, discoveryTarget, controller.signal).then(
      (discovery): Outcome => ({ kind: "ready", discovery }),
      (error: unknown): Outcome => ({
        kind: "error",
        errorKind: error instanceof CertificateApiError ? error.kind : "network",
      }),
    );
    inflightRef.current = promise;
    void promise.then((outcome) => {
      if (controller.signal.aborted) return;
      if (inflightRef.current === promise) inflightRef.current = null;
      outcomeRef.current = outcome;
      if (outcome.kind === "ready") dispatch({ type: "settled", run, discovery: outcome.discovery });
      else dispatch({ type: "failed", run, errorKind: outcome.errorKind });
    });
    return () => controller.abort();
  }, [api, enabled, intent, vaccinationId, editionId, state.run]);

  const view: RecoveryView | null = useMemo(
    () => (state.discovery === null ? null : deriveRecoveryView(state.discovery, state.resolved)),
    [state.discovery, state.resolved],
  );

  /** discovery を取り直す。確認できていない間は新しい操作を始めさせない。 */
  const refresh = useCallback(() => {
    outcomeRef.current = null;
    resolvedRef.current = new Map();
    lookupBusyRef.current.clear();
    dispatch({ type: "restart" });
  }, []);

  /**
   * 新しい operation key を作る直前の関門。進行中の discovery があれば待つ。
   * 確認済みで保留なしのときだけ true。未取得・失敗・保留中は false（key を作らない）。
   * true は「既知の阻害要因がない」だけで、新 key が安全である証明ではない。
   */
  const ensureClear = useCallback(async (): Promise<boolean> => {
    // fence が無い間は、発見の結果にかかわらず新しい操作を始めない（Slice A は表示と再照会のみ）。
    if (!CERTIFICATE_RECOVERY_FENCE_AVAILABLE) return false;
    const pending = inflightRef.current;
    const outcome = pending !== null ? await pending : outcomeRef.current;
    if (outcome === null || outcome.kind !== "ready") return false;
    return !deriveRecoveryView(outcome.discovery, resolvedRef.current).hold;
  }, []);

  /** 発見した操作の結果を、同じ operation key の既存 lookup で再照会する（再送はしない）。 */
  const lookupItem = useCallback(
    async (key: string) => {
      const discovery = stateRef.current.discovery;
      if (discovery === null || lookupBusyRef.current.has(key)) return;
      const run = stateRef.current.run;
      lookupBusyRef.current.add(key);
      dispatch({ type: "lookupStart", run, key });
      let refs: RecoveryRefs | null = null;
      let notice: RecoveryLookupNotice | null = "unavailable";
      try {
        const result = await api.lookupOperation(key, discovery.target.intent);
        if (!resultMatchesOperation(result, key)) {
          notice = "mismatch";
        } else if (result.status === "not-succeeded") {
          notice = "not-succeeded";
        } else {
          refs = boundSucceededRefs(result, {
            operationKey: key,
            vaccinationId: discovery.target.vaccinationId,
            editionId: discovery.target.intent === "reprint" ? String(discovery.target.editionId) : undefined,
          });
          notice = refs === null ? "mismatch" : null;
        }
      } catch {
        // 通信失敗・権限・未 claim 応答を含め、成功を確認できない限り保留のまま。
        notice = "unavailable";
      } finally {
        lookupBusyRef.current.delete(key);
      }
      if (refs !== null && run === stateRef.current.run) {
        resolvedRef.current = new Map(resolvedRef.current).set(key, refs);
      }
      dispatch({ type: "lookupDone", run, key, refs, notice });
    },
    [api],
  );

  /**
   * この画面の既存フローが、同じ key の成功を要求へ束縛して確認できたとき。保留を外せるのは
   * 「その key の束縛済みの成功」だけ（not-succeeded・scope-not-found・通信失敗では呼ばれない）。
   */
  const adoptSuccess = useCallback((key: string, refs: RecoveryRefs) => {
    resolvedRef.current = new Map(resolvedRef.current).set(key, refs);
    dispatch({ type: "lookupDone", run: stateRef.current.run, key, refs, notice: null });
  }, []);

  const canStart = CERTIFICATE_RECOVERY_FENCE_AVAILABLE && (state.status === "loading" || (state.status === "ready" && view !== null && !view.hold));

  return { state, view, canStart, ensureClear, refresh, lookupItem, adoptSuccess };
}

export type CertificateRecovery = ReturnType<typeof useCertificateRecovery>;
