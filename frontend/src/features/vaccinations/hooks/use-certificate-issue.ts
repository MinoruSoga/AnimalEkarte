import { useQueryClient } from "@tanstack/react-query";
import { useCallback, useEffect, useLayoutEffect, useMemo, useReducer, useRef, useState } from "react";
import type { CertificateApi } from "../api/certificate-api";
import { newOperationKey } from "../api/certificate-operation-key";
import { boundSucceededRefs, previewMatchesRequest, resultMatchesOperation } from "../api/certificate-binding";
import { CertificateApiError } from "../api/certificate-errors";
import {
  type CertificateContext,
  type CertificateEdition,
  type CertificateKind,
  type CertificateResult,
} from "../api/certificate-schemas";
import {
  decideIssueFailure,
  initialIssueState,
  issueReducer,
  restoredIssueState,
  type IssueNoticeCode,
  type IssuePending,
  type IssueState,
} from "../lib/certificate-issue-reducer";
import { certificateOperationRegistry } from "../lib/certificate-operation-registry";
import {
  draftFromInputs,
  draftToInputs,
  layoutsForKind,
  requestInputsFor,
  selectableKinds,
  type CertificateInputDraft,
  type DraftErrors,
} from "../lib/certificate-inputs";

/**
 * EMR-487: 発行フロー（preview → issue → 結果不明の照会）のオーケストレーション。
 * 状態遷移は lib/certificate-issue-reducer の純関数に集約し、ここは API 呼び出しと副作用だけを持つ。
 */

export interface UseCertificateIssueArgs {
  api: CertificateApi;
  vaccinationId: number;
  context: CertificateContext;
  /** mutation 直前に最新の edit 権限を返す（UI の disabled だけを最終防壁にしない）。 */
  getCanEdit: () => boolean;
  /** 前提変更（version-conflict など）で context を再取得してほしいとき。 */
  onRefetchContext: () => void;
  /** context を取り直し、最新の can_issue を返す（再開の判断に古い値を使わない）。 */
  refetchCanIssue: () => Promise<boolean>;
  /**
   * 新しい operation key を作る直前の関門（EMR-487 Slice A）。再 mount / reload 後の回復 discovery で
   * 未確定の操作が見つかっている・確認できていない間は false。false なら key を作らない。
   */
  ensureRecoveryClear: () => Promise<boolean>;
  /** 自分のフローが同じ key の束縛済みの成功を確認したときに通知する（発見側の保留を外すため）。 */
  onBoundSuccess: (key: string, refs: { editionId: string; artifactId: string }) => void;
}

function initialFor({ context, restored }: { context: CertificateContext; restored: IssuePending | undefined }): IssueState {
  // 画面を離れて戻っても、未確定の発行は同じ key のまま結果不明として復元する（別 key での二重発行を防ぐ）。
  if (restored !== undefined) return restoredIssueState(restored);
  const kind = selectableKinds(context)[0] ?? "normal";
  return initialIssueState({ kind, layout: layoutsForKind(kind)[0] });
}

export function useCertificateIssue({
  api,
  vaccinationId,
  context,
  getCanEdit,
  onRefetchContext,
  refetchCanIssue,
  ensureRecoveryClear,
  onBoundSuccess,
}: UseCertificateIssueArgs) {
  const queryClient = useQueryClient();
  const registry = useMemo(() => certificateOperationRegistry(queryClient), [queryClient]);
  const [state, dispatch] = useReducer(issueReducer, { context, restored: registry.issuePending.get(vaccinationId) }, initialFor);
  const [fieldErrors, setFieldErrors] = useReducer(
    (_: DraftErrors, next: DraftErrors) => next,
    {} as DraftErrors,
  );
  // 編集のたびに進める。遅れて返った preview が編集後の画面に載らないようにする。
  const epochRef = useRef(0);
  // 結果の照会は同時に 1 本だけ（重なった照会が先に確定した結果を後から上書きしない）。
  const lookupBusyRef = useRef(false);
  // key を作る前の関門（非同期）の間に二重に進まないための印。
  const mintingRef = useRef(false);
  const [checking, setChecking] = useState(false);
  // 非同期ハンドラが最新の state を読むための ref（render 中には書かない）。
  const stateRef = useRef(state);
  useLayoutEffect(() => {
    stateRef.current = state;
  }, [state]);

  // 未確定の key が無くなった（発行済み・前提変更・破棄・新規開始）ら、登録簿からも外す。
  // 未確定のまま unmount しても外さない（戻ったときに復元するため）。
  useEffect(() => {
    if (state.pending === null || state.phase === "issued") registry.issuePending.delete(vaccinationId);
  }, [registry, vaccinationId, state.pending, state.phase]);

  const edit = useCallback(
    (patch: { kind?: CertificateKind; layout?: IssueState["layout"]; draft?: Partial<CertificateInputDraft> }) => {
      epochRef.current += 1;
      setFieldErrors({});
      const nextKind = patch.kind;
      dispatch({
        type: "edit",
        patch: nextKind && !patch.layout ? { ...patch, layout: layoutsForKind(nextKind)[0] } : patch,
      });
    },
    [],
  );

  /** 最新発行内容を入力欄へコピーする（職員の明示操作のみ。自動更新・自動発行ではない）。 */
  const copyFrom = useCallback((edition: CertificateEdition) => {
    epochRef.current += 1;
    setFieldErrors({});
    dispatch({ type: "replaceDraft", draft: draftFromInputs(edition.snapshot.inputs) });
  }, []);

  const fail = useCallback(
    (notice: IssueNoticeCode, kind: "stale" | "blocked" | "previewFail" | "unknown", refetch = false) => {
      if (refetch) onRefetchContext();
      if (kind === "stale") dispatch({ type: "stale", notice });
      else if (kind === "blocked") dispatch({ type: "blocked", notice });
      else if (kind === "unknown") dispatch({ type: "issueUnknown", notice });
      else dispatch({ type: "previewFail", notice });
    },
    [onRefetchContext],
  );

  const requestPreview = useCallback(async () => {
    const current = stateRef.current;
    if (!getCanEdit()) {
      fail("forbidden", "blocked");
      return;
    }
    const converted = draftToInputs(current.draft);
    if (converted.inputs === null) {
      setFieldErrors(converted.errors);
      return;
    }
    setFieldErrors({});
    const request = {
      kind: current.kind,
      layout: current.layout,
      inputs: requestInputsFor(current.kind, converted.inputs),
      preconditions: {
        source_version: context.source_version,
        latest_edition_id: context.latest[current.kind]?.id ?? null,
      },
    };
    const epoch = epochRef.current;
    dispatch({ type: "previewStart" });
    try {
      const preview = await api.previewForIssue(vaccinationId, request);
      if (epoch !== epochRef.current) return;
      // 形が正しいだけの別接種・別様式・別 CAS の preview を、確認対象として載せない。
      const bound = previewMatchesRequest(preview, {
        vaccinationId,
        kind: request.kind,
        sourceVersion: request.preconditions.source_version,
        latestEditionId: request.preconditions.latest_edition_id,
        inputs: request.inputs,
      });
      if (!bound) return fail("preview-failed", "previewFail");
      dispatch({ type: "previewOk", preview, request });
    } catch (error) {
      if (epoch !== epochRef.current) return;
      if (error instanceof CertificateApiError) {
        if (error.kind === "forbidden") return fail("forbidden", "blocked");
        if (error.kind === "unauthenticated") return fail("unauthenticated", "blocked");
        if (error.kind === "version-conflict") return fail("version-conflict", "previewFail", true);
        if (error.kind === "dependency-changed") return fail("dependency-changed", "previewFail", true);
        if (error.kind === "validation") return fail("validation", "previewFail");
        if (error.kind === "not-found") return fail("not-found", "previewFail", true);
      }
      fail("preview-failed", "previewFail");
    }
  }, [api, vaccinationId, context, getCanEdit, fail]);

  const applyResult = useCallback(
    (result: CertificateResult, pending: IssuePending): void => {
      // 待っている操作以外の結果は、成功でも「成功なし」でも採用しない（結果不明のまま照会に戻す）。
      if (!resultMatchesOperation(result, pending.key)) {
        dispatch({ type: "issueUnknown", notice: "contract-violation" });
        return;
      }
      if (result.status === "not-succeeded") {
        dispatch({ type: "lookupNotSucceeded", notice: "not-succeeded" });
        return;
      }
      const refs = boundSucceededRefs(result, { operationKey: pending.key, vaccinationId });
      if (refs === null) {
        dispatch({ type: "issueUnknown", notice: "contract-violation" });
        return;
      }
      dispatch({ type: "issueOk", refs: { ...refs, replayed: result.status === "replayed" } });
      onBoundSuccess(pending.key, refs);
      onRefetchContext();
    },
    [onRefetchContext, onBoundSuccess, vaccinationId],
  );

  const lookup = useCallback(
    async (pending: IssuePending) => {
      if (lookupBusyRef.current) return;
      lookupBusyRef.current = true;
      setChecking(true);
      try {
        applyResult(await api.lookupOperation(pending.key, "issue"), pending);
      } catch (error) {
        if (!(error instanceof CertificateApiError)) return fail("unavailable", "unknown");
        if (error.kind === "scope-not-found") {
          dispatch({ type: "lookupNotSucceeded", notice: "not-claimed" });
        } else if (error.kind === "forbidden") fail("forbidden", "blocked");
        else if (error.kind === "unauthenticated") fail("unauthenticated", "blocked");
        else if (error.kind === "operation-in-progress") fail("busy", "unknown");
        else fail("unavailable", "unknown");
      } finally {
        lookupBusyRef.current = false;
        setChecking(false);
      }
    },
    [api, fail, applyResult],
  );

  const execute = useCallback(
    async (pending: IssuePending) => {
      try {
        const result = await api.issue(vaccinationId, pending.request);
        if (result.status === "not-succeeded") {
          // POST が not-succeeded を返すのは契約外。成功とも失敗とも推定せず照会に回す。
          dispatch({ type: "issueUnknown", notice: "unavailable" });
          await lookup(pending);
          return;
        }
        applyResult(result, pending);
      } catch (error) {
        const decision = decideIssueFailure(error);
        if (decision.type === "stale") fail(decision.notice, "stale", decision.refetchContext);
        else if (decision.type === "blocked") fail(decision.notice, "blocked");
        else {
          dispatch({ type: "issueUnknown", notice: decision.notice });
          await lookup(pending); // 結果不明は同じ key で一度だけ自動照会（以降は職員操作）
        }
      }
    },
    [api, vaccinationId, lookup, applyResult, fail],
  );

  const issue = useCallback(async () => {
    if (mintingRef.current) return;
    mintingRef.current = true;
    let pending: IssuePending | null = null;
    try {
      // 回復 discovery の関門。保留・未確認なら key を作らない（空/complete を許可とみなさない）。
      if (!(await ensureRecoveryClear())) return;
      const current = stateRef.current;
      if (current.phase !== "previewed" || current.preview === null || current.previewRequest === null) return;
      if (!getCanEdit()) return fail("forbidden", "blocked");
      // key は確認済み preview 1 つに束縛して、発行の直前に 1 度だけ作る。
      const key = newOperationKey();
      pending = {
        key,
        request: {
          ...current.previewRequest,
          operation_key: key,
          preview_fingerprint: current.preview.preview_fingerprint,
          preview_display_issue_date: current.preview.snapshot.display_issue_date,
        },
      };
      // POST を送る前に登録簿へ置く。以後に画面を離れても、この key を失わずに照会できる。
      registry.issuePending.set(vaccinationId, pending);
      dispatch({ type: "issueStart", pending });
    } finally {
      mintingRef.current = false;
    }
    if (pending !== null) await execute(pending);
  }, [execute, getCanEdit, fail, registry, vaccinationId, ensureRecoveryClear]);

  /** 結果不明の手動照会（同 key）。 */
  const recheck = useCallback(async () => {
    const pending = stateRef.current.pending;
    if (pending === null) return;
    await lookup(pending);
  }, [lookup]);

  /**
   * 権限なし・未認証で止まった後の再開。権限が戻っていることを確認できたときだけ解除する。
   * 未確定の key があれば捨てず、同じ key の照会から再開する。
   */
  const resumeAfterBlocked = useCallback(async () => {
    const current = stateRef.current;
    if (current.phase !== "blocked") return;
    const canIssue = await refetchCanIssue();
    if (!getCanEdit() || !canIssue) return;
    dispatch({ type: "unblock" });
    if (current.pending !== null) await lookup(current.pending);
  }, [getCanEdit, lookup, refetchCanIssue]);

  /** 照会で「成功なし」確定後の再試行。同 key・同 request のまま。 */
  const retrySameKey = useCallback(async () => {
    const pending = stateRef.current.pending;
    if (pending === null) return;
    if (!getCanEdit()) return fail("forbidden", "blocked");
    dispatch({ type: "retryStart" });
    await execute(pending);
  }, [execute, getCanEdit, fail]);

  /** 「成功なし」確定後に、内容を直して再 preview へ進む（別 key になる）。 */
  const discardAndEdit = useCallback(() => {
    epochRef.current += 1;
    dispatch({ type: "acknowledgeResolved" });
  }, []);

  /** 発行済みの画面から、新しい発行を始める。 */
  const startNew = useCallback(() => {
    epochRef.current += 1;
    setFieldErrors({});
    dispatch({ type: "startNew" });
  }, []);

  return { state, fieldErrors, checking, edit, copyFrom, startNew, requestPreview, issue, recheck, retrySameKey, discardAndEdit, resumeAfterBlocked };
}
