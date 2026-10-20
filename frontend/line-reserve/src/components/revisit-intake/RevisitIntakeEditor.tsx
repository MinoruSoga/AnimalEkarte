// EMR-474: 再診事前問診の入力・送信・訂正（1 予約 = 1 ペット = 1 form）。
// 契約: backend/docs/api.yaml submitLiffRevisitIntake / getLiffRevisitIntakeOperation、
// technical-spec#revisit-intake-api-selected-v3「submit・結果照会の境界」。
// - 操作 key は送信ごとに新規生成。応答喪失・5xx・claimed・not_found・result_unavailable からは新 key を作らない
// - 結果未確定の間は入力を固定し、同じ key・同じ内容でだけ再送／照会する
// - 成功後の訂正、409 等で結果が確定した後の再送信は新 key
// - 送信後に LINE トークへのメッセージ送信（liff.sendMessages）は行わない
import {
  startTransition,
  useActionState,
  useEffect,
  useRef,
  useState,
  type FormEvent,
} from "react";
import { liffApi } from "../../api/liff-api";
import { PrimaryButton } from "../PrimaryButton";
import { resolveFetchError } from "@/shared-liff/handle-fetch-error";
import type {
  RevisitIntakeFullAnswers,
  RevisitIntakeLIFFRead,
  RevisitIntakeLookupResult,
  RevisitIntakeOwnerCurrentTarget,
} from "../../types/models";
import {
  EMPTY_REVISIT_INTAKE_ANSWERS,
  areRevisitIntakeAnswersEqual,
  type RevisitIntakeDraftStore,
  type RevisitIntakePendingOperation,
} from "../../hooks/use-revisit-intake-drafts";
import {
  buildRevisitIntakeSubmitRequest,
  classifyRevisitIntakeSubmitError,
  firstErrorField,
  revisitIntakeFingerprint,
  validateRevisitIntakeAnswers,
  revisitIntakeRetryAfterSeconds,
  type RevisitIntakeFieldErrors,
} from "../../lib/revisit-intake-submit";
import {
  formatRevisitIntakeDateTime,
  revisitIntakeFieldId,
  type RevisitIntakeFieldName,
} from "../../lib/revisit-intake-options";
import { createRevisitIntakeOperationKey } from "../../lib/revisit-intake-operation-key";
import { RevisitIntakeFields } from "./RevisitIntakeFields";
import { RevisitIntakeAnswerView } from "./RevisitIntakeAnswerView";

export type RevisitIntakeClosedReason =
  | "reception_started"
  | "cancelled"
  | "deleted"
  | "unsupported_lifecycle"
  | null;

type EditorOutcome =
  | { kind: "idle" }
  | {
      kind: "succeeded";
      committedAt: string;
      contentAvailable: boolean;
      submitted: RevisitIntakeFullAnswers;
    }
  | { kind: "unresolved"; message: string | null }
  /** TARGET_CHANGED 後、新しい target snapshot の取得に成功するまで入力・送信を止める */
  | { kind: "target_changed"; atRead: RevisitIntakeLIFFRead }
  /** 401: 再認証が必要。同じ画面からは再試行・新規送信をさせない */
  | { kind: "unauthorized"; message: string; pendingKept: boolean }
  /** 未確定操作の再送が 403/404: ErrorPage へ替えず、結果照会だけ行えるようにする */
  | { kind: "pending_access_denied" }
  | { kind: "invalid"; fieldErrors: RevisitIntakeFieldErrors }
  | { kind: "revision_conflict"; atRead: RevisitIntakeLIFFRead }
  | { kind: "form_version_mismatch"; atRead: RevisitIntakeLIFFRead }
  | { kind: "key_content_mismatch" }
  | { kind: "error"; message: string };

type EditorAction = "submit" | "resend" | "lookup";

interface RevisitIntakeEditorProps {
  clinicId: string;
  idToken: string;
  appointmentId: string;
  read: RevisitIntakeLIFFRead;
  current: RevisitIntakeOwnerCurrentTarget;
  targetKey: string;
  draftStore: RevisitIntakeDraftStore;
  refetching: boolean;
  /** 直近の GET が失敗している（read は旧 snapshot のまま）。最新取得までは送信しない */
  readError: boolean;
  /** 読込 route の Retry-After 期限中 */
  retryBlocked: boolean;
  onRefetch: () => void;
  onDirtyChange: (dirty: boolean) => void;
  /** 送信処理中（inFlight）と結果未確定（unresolved = pending あり）をページの離脱確認へ通知する */
  onPendingChange: (inFlight: boolean, unresolved: boolean) => void;
  onTargetChanged: () => void;
  /** 送信は成功したが、その後に予約の対象が変わっていた（新しい target を読み直す） */
  onSucceededThenTargetChanged: () => void;
  onIntakeClosed: (reason: RevisitIntakeClosedReason) => void;
  onAccessDenied: () => void;
}

type ResolvedLookupResult = Exclude<
  RevisitIntakeLookupResult,
  { operation_result: { state: "not_found" } }
>;

function isLookupNotFound(
  result: RevisitIntakeLookupResult,
): result is Exclude<RevisitIntakeLookupResult, ResolvedLookupResult> {
  return result.operation_result.state === "not_found";
}

const LOOKUP_FAILED_MESSAGE =
  "送信結果を確認できませんでした。通信状況をご確認のうえ、もう一度お試しください。";

const PENDING_UNAUTHORIZED_MESSAGE =
  "ログインの有効期限が切れました。送信結果が未確定のため、この画面を閉じずに、時間をおいてもう一度「送信結果を確認」してください。";

const MISMATCHED_RESULT_MESSAGE =
  "送信結果を確認できませんでした。二重に保存されないよう、同じ内容で再送信するか結果を確認してください。";

function lookupRetryMessage(seconds: number): string {
  return `ただいま混み合っています。${seconds}秒ほど待ってから、送信結果を確認してください。`;
}

function serviceUnavailableMessage(seconds: number): string {
  return `ただいま混み合っています。${seconds}秒ほど待ってから、同じ内容で再送信するか送信結果を確認してください。`;
}

function internalBackoffMessage(seconds: number): string {
  return `ただいま混み合っています。${seconds}秒ほど待ってから、送信結果を確認してください。結果を確認するまで再送信はできません。`;
}

function rateLimitedMessage(seconds: number | null): string {
  return seconds !== null && seconds > 0
    ? `アクセスが集中しています。${seconds}秒ほど待ってから、同じ内容で再送信してください。`
    : "アクセスが集中しています。しばらく待ってから、同じ内容で再送信してください。";
}

function PendingPanel({
  pending,
  message,
  busy,
  resendBlocked,
  lookupBlocked,
  onResend,
  onLookup,
}: {
  pending: RevisitIntakePendingOperation;
  message: string | null;
  busy: boolean;
  resendBlocked: boolean;
  lookupBlocked: boolean;
  onResend: () => void;
  onLookup: () => void;
}) {
  const text: Record<RevisitIntakePendingOperation["status"], string> = {
    unknown:
      "送信結果をまだ確認できていません。二重に保存されないよう、同じ内容で再送信するか結果を確認してください。",
    claimed_uncommitted:
      "送信は受け付けられましたが、保存の完了を確認できていません。同じ内容でもう一度送信できます。",
    not_found:
      "送信の記録が見つかりませんでした（届いていない可能性があります）。同じ内容でもう一度送信できます。",
    result_unavailable:
      "結果を確認できません。二重に保存されないよう、新しい送信は行えません。時間をおいて結果を再確認するか、医院へお問い合わせください。",
  };
  const canResend = pending.status !== "result_unavailable";
  return (
    <div
      role="alert"
      className="space-y-3 rounded-lg border border-noah-warning-border bg-noah-warning-box-bg px-4 py-3"
    >
      <p className="text-sm text-noah-warning-text">{text[pending.status]}</p>
      {message !== null ? <p className="text-sm text-noah-warning-text-sub">{message}</p> : null}
      {pending.lookupRequired === true && canResend ? (
        <p className="text-sm text-noah-warning-text-sub">送信結果を確認してから再送信できます。</p>
      ) : null}
      <p className="text-xs text-noah-warning-text-sub">
        結果が確定するまで、入力内容は変更できません。
      </p>
      <div className="flex flex-col gap-2">
        {canResend ? (
          <PrimaryButton onClick={onResend} disabled={busy || resendBlocked}>
            同じ内容で再送信
          </PrimaryButton>
        ) : null}
        <button
          type="button"
          onClick={onLookup}
          disabled={busy || lookupBlocked}
          className="min-h-11 w-full rounded-lg border border-noah-border-input bg-white px-4 text-sm font-semibold text-noah-teal-dark disabled:opacity-50"
        >
          送信結果を確認
        </button>
      </div>
    </div>
  );
}

function RefetchBanner({
  message,
  refetching,
  blocked,
  onRefetch,
}: {
  message: string;
  refetching: boolean;
  /** route 共通の Retry-After 期限中（refetch は無視されるため押せないようにする） */
  blocked: boolean;
  onRefetch: () => void;
}) {
  return (
    <div
      role="alert"
      className="space-y-2 rounded-lg border border-noah-danger-border bg-noah-danger-bg px-4 py-3"
    >
      <p className="text-sm text-noah-danger-text">{message}</p>
      <button
        type="button"
        onClick={onRefetch}
        disabled={refetching || blocked}
        aria-busy={refetching}
        className="min-h-11 w-full rounded-lg border border-noah-border-input bg-white px-4 text-sm font-semibold text-noah-teal-dark disabled:opacity-50"
      >
        {refetching ? "読み込み中..." : "最新の状態を読み込む"}
      </button>
    </div>
  );
}

export function RevisitIntakeEditor({
  clinicId,
  idToken,
  appointmentId,
  read,
  current,
  targetKey,
  draftStore,
  refetching,
  readError,
  retryBlocked,
  onRefetch,
  onDirtyChange,
  onPendingChange,
  onTargetChanged,
  onSucceededThenTargetChanged,
  onIntakeClosed,
  onAccessDenied,
}: RevisitIntakeEditorProps) {
  // 本文が可視の最新回答だけを prefill する。不可視 head・旧自分回答・previous 履歴は流し込まない。
  const baseline =
    current.content_available && current.latest_answer
      ? current.latest_answer.answers
      : EMPTY_REVISIT_INTAKE_ANSWERS;
  const [values, setValues] = useState<RevisitIntakeFullAnswers>(
    () => draftStore.get(targetKey)?.values ?? baseline,
  );
  const [pending, setPending] = useState<RevisitIntakePendingOperation | null>(
    () => draftStore.get(targetKey)?.pending ?? null,
  );
  // 429 / 503 の Retry-After 期限（epoch ms）。期限までは「同じ内容で再送信」を無効にする。
  // draftStore の pending から復元する（期限切れは下の effect で即解除）。
  const [resendAvailableAt, setResendAvailableAt] = useState<number | null>(
    () => draftStore.get(targetKey)?.pending?.resendAvailableAt ?? null,
  );
  // 同じ target のまま head（版・本文可否・回答 ID）が変わったら、local draft も未確定操作も無い場合に限り
  // 入力を新しい baseline へ同期する（render 中の state 調整。draftStore に残る draft は保持する）。
  const baselineToken = [
    current.head_revision ?? "",
    String(current.content_available),
    current.latest_answer?.answer_id ?? "",
  ].join("|");
  // draft の元になった baseline。保存済み draft はそのときの token を復元する。
  const [draftBaseToken, setDraftBaseToken] = useState(
    () => draftStore.get(targetKey)?.baselineToken ?? baselineToken,
  );
  const [syncedBaselineToken, setSyncedBaselineToken] = useState(baselineToken);
  if (syncedBaselineToken !== baselineToken) {
    setSyncedBaselineToken(baselineToken);
    if (pending === null && draftStore.get(targetKey) === undefined) {
      setValues(baseline);
      setDraftBaseToken(baselineToken);
    }
  }
  const staleDraft =
    pending === null &&
    draftBaseToken !== baselineToken &&
    !areRevisitIntakeAnswersEqual(values, baseline);

  function acknowledgeLatest() {
    setDraftBaseToken(baselineToken);
    persist(values, pending, baselineToken);
  }
  // 409 REVISION_CONFLICT で未成功が確定した操作の key。同内容（expected_revision を除く fingerprint が同じ）の
  // 再送信は同 key で再検証し、内容が変わったときだけ新 key にする（technical-spec#revisit-intake-api-selected-v3）。
  // 自動再送はしない。回答が成功した後の訂正は常に新 key（成功時は pending 経由で消える）。
  const [conflictedOp, setConflictedOp] = useState<{ key: string; fingerprint: string } | null>(
    null,
  );
  const inFlightRef = useRef(false);
  const alertRef = useRef<HTMLDivElement>(null);

  function persist(
    nextValues: RevisitIntakeFullAnswers,
    nextPending: RevisitIntakePendingOperation | null,
    nextBaseToken: string = draftBaseToken,
  ) {
    if (nextPending === null && areRevisitIntakeAnswersEqual(nextValues, baseline)) {
      draftStore.clear(targetKey);
    } else {
      draftStore.set(targetKey, {
        values: nextValues,
        pending: nextPending,
        baselineToken: nextBaseToken,
      });
    }
  }

  function updatePending(next: RevisitIntakePendingOperation | null) {
    setPending(next);
    persist(values, next);
  }

  /**
   * 応答が要求した操作・target のものかを確かめる（fail-closed）。operation_id が key と違う、
   * または成功応答の current_state / result_refs が要求した appointment・pet・世代と違うなら適用しない。
   */
  function resultMatchesOperation(
    result: ResolvedLookupResult,
    op: RevisitIntakePendingOperation,
  ): boolean {
    if (result.operation_result.operation_id !== op.key) return false;
    if (result.operation_result.state !== "succeeded") return true;
    // operation_result は初回の不変結果（api.yaml RevisitIntakeSuccessResponse）。送信した target との対応は
    // result_refs と committed_versions で確かめ、現在 snapshot の current_state では判定しない。
    const refs = result.operation_result.result_refs;
    if (refs && (refs.appointment_id !== appointmentId || refs.pet_id !== op.body.pet_id)) {
      return false;
    }
    const versions = result.operation_result.committed_versions;
    if (versions && versions.target_generation !== op.body.target_generation) return false;
    return true;
  }

  /** 成功時の現在 snapshot が送信した target と別か（成功後に対象が変わった）。 */
  function currentStateMovedAway(result: ResolvedLookupResult, op: RevisitIntakePendingOperation) {
    const state = result.current_state;
    return (
      state !== null &&
      (state.ref.appointment_id !== appointmentId ||
        state.ref.pet_id !== op.body.pet_id ||
        state.ref.target_generation !== op.body.target_generation)
    );
  }

  function applyResult(
    result: ResolvedLookupResult,
    op: RevisitIntakePendingOperation,
  ): EditorOutcome {
    if (!resultMatchesOperation(result, op)) {
      updatePending({ ...op, status: "unknown" });
      return { kind: "unresolved", message: MISMATCHED_RESULT_MESSAGE };
    }
    if (result.operation_result.state === "succeeded" && currentStateMovedAway(result, op)) {
      // 送信は成功済み。新しい target の state を送信した target のものとして適用せず、
      // 未確定操作と draft を消して既存の再取得経路で新しい target を読み直す。
      setPending(null);
      draftStore.clear(targetKey);
      onSucceededThenTargetChanged();
      return { kind: "idle" };
    }
    if (result.operation_result.state === "succeeded") {
      const refs = result.operation_result.result_refs;
      const versions = result.operation_result.committed_versions;
      // 成功した回答を次の訂正の元にする（再取得後の head と同じ token）。
      if (refs && versions) setDraftBaseToken(`${versions.answer_revision}|true|${refs.answer_id}`);
      setPending(null);
      setValues(op.body.answers);
      draftStore.clear(targetKey);
      onRefetch();
      return {
        kind: "succeeded",
        committedAt: result.operation_result.committed_at,
        contentAvailable: result.content_available,
        submitted: op.body.answers,
      };
    }
    updatePending({ ...op, status: result.operation_result.state, lookupRequired: false });
    return { kind: "unresolved", message: null };
  }

  async function lookup(op: RevisitIntakePendingOperation): Promise<EditorOutcome> {
    try {
      const result = await liffApi.getRevisitIntakeOperation(clinicId, op.key, idToken);
      if (isLookupNotFound(result)) {
        if (result.operation_result.operation_id !== op.key) {
          updatePending({ ...op, status: "unknown" });
          return { kind: "unresolved", message: MISMATCHED_RESULT_MESSAGE };
        }
        updatePending({ ...op, status: "not_found", lookupRequired: false });
        return { kind: "unresolved", message: null };
      }
      return applyResult(result, op);
    } catch (err: unknown) {
      const info = resolveFetchError(err, "送信結果の確認");
      if (info.status === 401) {
        updatePending(op);
        return { kind: "unauthorized", message: PENDING_UNAUTHORIZED_MESSAGE, pendingKept: true };
      }
      // 照会 route の 429 / Retry-After 付き 503 も期限まで再送・照会を止める。
      const retryAfter = revisitIntakeRetryAfterSeconds(err);
      if (retryAfter !== null) {
        const availableAt = Date.now() + retryAfter * 1000;
        // 照会自体が 500 + Retry-After なら、期限後も成功する照会まで再送を許さない。
        updatePending(
          info.status === 500
            ? { ...op, resendAvailableAt: availableAt, lookupRequired: true }
            : { ...op, resendAvailableAt: availableAt },
        );
        setResendAvailableAt(availableAt);
        return { kind: "unresolved", message: lookupRetryMessage(retryAfter) };
      }
      updatePending(op);
      return { kind: "unresolved", message: LOOKUP_FAILED_MESSAGE };
    }
  }

  async function send(op: RevisitIntakePendingOperation, fresh: boolean): Promise<EditorOutcome> {
    try {
      const result = await liffApi.submitRevisitIntake(
        clinicId,
        appointmentId,
        op.body,
        op.key,
        idToken,
      );
      return applyResult(result, op);
    } catch (err: unknown) {
      const kind = classifyRevisitIntakeSubmitError(err);
      switch (kind.type) {
        case "invalid":
          updatePending(null);
          return { kind: "invalid", fieldErrors: kind.fieldErrors };
        case "revision_conflict":
          setConflictedOp({ key: op.key, fingerprint: op.fingerprint });
          updatePending(null);
          return { kind: "revision_conflict", atRead: read };
        case "form_version_mismatch":
          updatePending(null);
          return { kind: "form_version_mismatch", atRead: read };
        case "target_changed":
          updatePending(null);
          onTargetChanged();
          return { kind: "target_changed", atRead: read };
        case "intake_closed":
          updatePending(null);
          onIntakeClosed(kind.reason);
          return { kind: "idle" };
        case "key_content_mismatch":
          updatePending(null);
          return { kind: "key_content_mismatch" };
        case "payload_too_large":
          updatePending(null);
          return {
            kind: "error",
            message: "送信内容が大きすぎます。入力内容を短くしてから送信してください。",
          };
        case "unsupported_media_type":
          updatePending(null);
          return {
            kind: "error",
            message: "送信形式に問題があります。LINEアプリから開き直してお試しください。",
          };
        case "unauthorized":
          // 認証前に拒否されるため、新規送信なら結果未確定の操作は残さない。
          if (fresh) {
            updatePending(null);
            return {
              kind: "unauthorized",
              message: resolveFetchError(err, "問診の送信").message,
              pendingKept: false,
            };
          }
          // 未確定操作の key は memory にしか無い。LINE を開き直すと key を失い二重送信になりうるため、
          // 閉じないよう案内して pending を保持する（LIFF 再認証の共通処理が決まるまでは案内のみ）。
          return { kind: "unauthorized", message: PENDING_UNAUTHORIZED_MESSAGE, pendingKept: true };
        case "access_denied":
          // 新規送信の 403/404 は処理前の拒否。未確定の操作として残さない（再送では保持する）。
          if (fresh) {
            updatePending(null);
            onAccessDenied();
            return { kind: "idle" };
          }
          // 照会 route は現予約の閲覧権限と分離されている。結果確認の導線を残す。
          return { kind: "pending_access_denied" };
        case "rate_limited":
        case "service_unavailable": {
          const availableAt =
            kind.retryAfterSeconds !== null && kind.retryAfterSeconds > 0
              ? Date.now() + kind.retryAfterSeconds * 1000
              : null;
          updatePending({ ...op, status: "unknown", resendAvailableAt: availableAt });
          setResendAvailableAt(availableAt);
          return {
            kind: "unresolved",
            message:
              kind.type === "rate_limited"
                ? rateLimitedMessage(kind.retryAfterSeconds)
                : serviceUnavailableMessage(kind.retryAfterSeconds),
          };
        }
        case "unknown_with_backoff": {
          // 500 + Retry-After: 結果不明のまま期限まで待ち、期限後も照会が済むまで再送を許さない。
          const availableAt = Date.now() + kind.retryAfterSeconds * 1000;
          updatePending({
            ...op,
            status: "unknown",
            resendAvailableAt: availableAt,
            lookupRequired: true,
          });
          setResendAvailableAt(availableAt);
          return { kind: "unresolved", message: internalBackoffMessage(kind.retryAfterSeconds) };
        }
        case "unknown":
          // 応答喪失・5xx: 失敗確定ではない。同じ key で結果を照会する。
          return lookup({
            ...op,
            status: op.status === "result_unavailable" ? op.status : "unknown",
          });
      }
    }
  }

  async function runAction(previous: EditorOutcome, action: EditorAction): Promise<EditorOutcome> {
    try {
      if (action === "submit") {
        if (pending !== null) return previous;
        // 旧 snapshot の head_revision を expected_revision に使わない（成功後の再取得中・失敗時）。
        if (refetching || readError) return previous;
        // 古い回答を元にした draft は、最新回答の確認が明示されるまで送らない（黙って上書きしない）。
        if (staleDraft) return previous;
        const fieldErrors = validateRevisitIntakeAnswers(values);
        if (firstErrorField(fieldErrors) !== null) {
          return { kind: "invalid", fieldErrors };
        }
        const body = buildRevisitIntakeSubmitRequest(current, read.form_definition_version, values);
        const fingerprint = revisitIntakeFingerprint(body);
        const reuseKey = conflictedOp !== null && conflictedOp.fingerprint === fingerprint;
        setConflictedOp(null);
        const op: RevisitIntakePendingOperation = {
          key: reuseKey ? conflictedOp.key : createRevisitIntakeOperationKey(),
          fingerprint,
          body,
          status: "unknown",
        };
        // 応答が失われても同じ key で照会・再送できるよう、送信前に登録する。
        updatePending(op);
        return await send(op, true);
      }
      if (pending === null) return previous;
      if (action === "resend") {
        if (pending.status === "result_unavailable") return previous;
        if (pending.lookupRequired === true) return previous;
        if (resendAvailableAt !== null && Date.now() < resendAvailableAt) return previous;
        if (
          (previous.kind === "unauthorized" && previous.pendingKept) ||
          previous.kind === "pending_access_denied"
        ) {
          return previous;
        }
        return await send(pending, false);
      }
      if (resendAvailableAt !== null && Date.now() < resendAvailableAt) return previous;
      return await lookup(pending);
    } finally {
      inFlightRef.current = false;
    }
  }

  const [outcome, dispatch, isPending] = useActionState<EditorOutcome, EditorAction>(runAction, {
    kind: "idle",
  });

  function trigger(action: EditorAction) {
    if (inFlightRef.current) return; // 二重送信抑止（再描画前の連打も止める）
    inFlightRef.current = true;
    startTransition(() => dispatch(action));
  }

  function handleSubmit(event: FormEvent<HTMLFormElement>) {
    // <form action> の自動 reset で制御入力が巻き戻らないよう onSubmit から action を起動する。
    event.preventDefault();
    trigger("submit");
  }

  function handleChange<K extends RevisitIntakeFieldName>(
    field: K,
    value: RevisitIntakeFullAnswers[K],
  ) {
    const next = { ...values, [field]: value };
    setValues(next);
    persist(next, pending);
  }

  const dirty =
    !areRevisitIntakeAnswersEqual(values, baseline) &&
    !(outcome.kind === "succeeded" && areRevisitIntakeAnswersEqual(values, outcome.submitted));

  useEffect(() => {
    onDirtyChange(dirty);
  }, [dirty, onDirtyChange]);
  useEffect(() => () => onDirtyChange(false), [onDirtyChange]);
  useEffect(() => {
    onPendingChange(isPending, pending !== null);
  }, [isPending, pending, onPendingChange]);
  useEffect(() => () => onPendingChange(false, false), [onPendingChange]);

  useEffect(() => {
    if (outcome.kind !== "invalid") return;
    const field = firstErrorField(outcome.fieldErrors);
    if (field !== null) {
      document.getElementById(revisitIntakeFieldId(field))?.focus();
    } else {
      alertRef.current?.focus();
    }
  }, [outcome]);

  useEffect(() => {
    if (resendAvailableAt === null) return;
    const timer = setTimeout(
      () => setResendAvailableAt(null),
      Math.max(0, resendAvailableAt - Date.now()),
    );
    return () => clearTimeout(timer);
  }, [resendAvailableAt]);

  const fieldErrors = outcome.kind === "invalid" ? outcome.fieldErrors : {};
  const needsRefetch =
    (outcome.kind === "revision_conflict" || outcome.kind === "form_version_mismatch") &&
    outcome.atRead === read;
  // 旧 target の snapshot のままでは送らない（再取得中・再取得失敗時も含む）。
  const targetChangeLocked = outcome.kind === "target_changed" && outcome.atRead === read;
  const authBlocked = outcome.kind === "unauthorized";
  // 未確定操作で 401 / 403・404 を受けた後は再送させず、結果照会だけ許す。
  const pendingResendLocked =
    (outcome.kind === "unauthorized" && outcome.pendingKept) ||
    outcome.kind === "pending_access_denied";
  const inputsLocked = isPending || pending !== null || targetChangeLocked;
  const answered = current.answer_state === "answered";

  return (
    <form onSubmit={handleSubmit} noValidate aria-busy={isPending} className="space-y-5">
      <div className="rounded-lg bg-white/60 px-4 py-3 text-xs leading-relaxed text-noah-text-sub">
        <p>
          回答は任意です。回答しなくても予約・受付には影響しません。送信する場合のみ「来院目的」の入力が必要です。
        </p>
        <p className="mt-1">
          入力内容はこの画面を開いている間だけ一時的に保持されます。LINEを閉じたり再読み込みしたりすると、送信前の入力は失われます。
        </p>
      </div>

      {outcome.kind === "succeeded" ? (
        <div
          role="status"
          className="rounded-lg border border-noah-info-border bg-noah-info-box-bg px-4 py-3"
        >
          <p className="text-sm font-medium text-noah-info-text">
            {outcome.contentAvailable
              ? `回答を送信しました（${formatRevisitIntakeDateTime(outcome.committedAt)}）。`
              : "回答の送信は完了しています。"}
          </p>
          <p className="mt-1 text-xs text-noah-info-text-sub">
            受付前であれば、この画面から訂正して再送信できます。
          </p>
        </div>
      ) : null}

      {outcome.kind === "revision_conflict" ? (
        needsRefetch ? (
          <RefetchBanner
            message="ほかの操作で回答が更新されています。入力内容はそのまま残っています。最新の状態を読み込み、保存済みの回答を確認してから再送信してください。"
            refetching={refetching}
            blocked={retryBlocked}
            onRefetch={onRefetch}
          />
        ) : (
          <p role="status" className="text-sm text-noah-text">
            最新の状態を読み込みました。保存済みの最新回答を確認のうえ、必要であれば再送信してください。
          </p>
        )
      ) : null}

      {outcome.kind === "form_version_mismatch" && needsRefetch ? (
        <RefetchBanner
          message="問診の項目が更新されています。最新の状態を読み込んでから送信してください。"
          refetching={refetching}
          blocked={retryBlocked}
          onRefetch={onRefetch}
        />
      ) : null}

      {readError && !targetChangeLocked && !needsRefetch ? (
        <RefetchBanner
          message="最新の状態を読み込めませんでした。読み込みに成功するまで送信できません。"
          refetching={refetching}
          blocked={retryBlocked}
          onRefetch={onRefetch}
        />
      ) : null}

      {targetChangeLocked ? (
        <RefetchBanner
          message="予約の対象が変更されました。最新の内容を読み込むまで送信できません。"
          refetching={refetching}
          blocked={retryBlocked}
          onRefetch={onRefetch}
        />
      ) : null}

      {outcome.kind === "pending_access_denied" ? (
        <div
          role="alert"
          className="rounded-lg border border-noah-danger-border bg-noah-danger-bg px-4 py-3"
        >
          <p className="text-sm text-noah-danger-text">
            現在この予約を閲覧できません。送信結果の確認だけ行えます。
          </p>
        </div>
      ) : null}

      {outcome.kind === "unauthorized" ? (
        <div
          role="alert"
          className="rounded-lg border border-noah-danger-border bg-noah-danger-bg px-4 py-3"
        >
          <p className="text-sm text-noah-danger-text">{outcome.message}</p>
        </div>
      ) : null}

      {outcome.kind === "key_content_mismatch" ? (
        <div
          role="alert"
          className="rounded-lg border border-noah-danger-border bg-noah-danger-bg px-4 py-3"
        >
          <p className="text-sm text-noah-danger-text">
            送信内容を確認できませんでした。内容をご確認のうえ、もう一度送信してください。
          </p>
        </div>
      ) : null}

      {outcome.kind === "error" ? (
        <div
          role="alert"
          className="rounded-lg border border-noah-danger-border bg-noah-danger-bg px-4 py-3"
        >
          <p className="text-sm text-noah-danger-text">{outcome.message}</p>
        </div>
      ) : null}

      {outcome.kind === "invalid" && firstErrorField(outcome.fieldErrors) === null ? (
        <div
          ref={alertRef}
          tabIndex={-1}
          role="alert"
          className="rounded-lg border border-noah-danger-border bg-noah-danger-bg px-4 py-3"
        >
          <p className="text-sm text-noah-danger-text">入力内容を確認してください。</p>
        </div>
      ) : null}

      {pending !== null && !isPending ? (
        <PendingPanel
          pending={pending}
          message={outcome.kind === "unresolved" ? outcome.message : null}
          busy={isPending}
          resendBlocked={
            resendAvailableAt !== null || pending.lookupRequired === true || pendingResendLocked
          }
          lookupBlocked={resendAvailableAt !== null}
          onResend={() => trigger("resend")}
          onLookup={() => trigger("lookup")}
        />
      ) : null}

      {staleDraft ? (
        <div
          role="alert"
          className="space-y-2 rounded-lg border border-noah-warning-border bg-noah-warning-box-bg px-4 py-3"
        >
          <p className="text-sm text-noah-warning-text">
            {current.content_available
              ? "この入力は古い回答を元にしています。最新の回答を確認してから送信してください。"
              : "この入力は古い回答を元にしています。最新の回答はこの画面では表示できません。送信すると、この入力が最新の回答として保存されます。"}
          </p>
          <button
            type="button"
            onClick={acknowledgeLatest}
            className="min-h-11 w-full rounded-lg border border-noah-border-input bg-white px-4 text-sm font-semibold text-noah-teal-dark"
          >
            {current.content_available ? "最新の回答を確認した" : "この入力で続ける"}
          </button>
        </div>
      ) : null}

      {answered && current.content_available && current.latest_answer ? (
        <details
          open={staleDraft || (outcome.kind === "revision_conflict" && !needsRefetch)}
          className="rounded-lg border border-noah-border bg-white px-4 py-2"
        >
          <summary className="flex min-h-11 cursor-pointer items-center text-sm font-medium text-noah-teal-dark">
            保存済みの最新回答を見る
          </summary>
          <RevisitIntakeAnswerView answers={current.latest_answer.answers} />
        </details>
      ) : null}

      <RevisitIntakeFields
        values={values}
        errors={fieldErrors}
        disabled={inputsLocked}
        onChange={handleChange}
      />

      {pending === null || isPending ? (
        <PrimaryButton
          type="submit"
          disabled={
            isPending ||
            needsRefetch ||
            targetChangeLocked ||
            authBlocked ||
            refetching ||
            readError ||
            staleDraft
          }
        >
          {isPending ? "送信中..." : answered ? "訂正して再送信" : "回答を送信する"}
        </PrimaryButton>
      ) : null}
    </form>
  );
}
