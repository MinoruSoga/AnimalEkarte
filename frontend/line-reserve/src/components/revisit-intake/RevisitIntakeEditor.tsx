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
  | { kind: "unauthorized"; message: string }
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
  onRefetch: () => void;
  onDirtyChange: (dirty: boolean) => void;
  onTargetChanged: () => void;
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
  authBlocked,
  onResend,
  onLookup,
}: {
  pending: RevisitIntakePendingOperation;
  message: string | null;
  busy: boolean;
  resendBlocked: boolean;
  authBlocked: boolean;
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
      <p className="text-xs text-noah-warning-text-sub">
        結果が確定するまで、入力内容は変更できません。
      </p>
      <div className="flex flex-col gap-2">
        {canResend ? (
          <PrimaryButton onClick={onResend} disabled={busy || resendBlocked || authBlocked}>
            同じ内容で再送信
          </PrimaryButton>
        ) : null}
        <button
          type="button"
          onClick={onLookup}
          disabled={busy || authBlocked}
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
  onRefetch,
}: {
  message: string;
  refetching: boolean;
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
        disabled={refetching}
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
  onRefetch,
  onDirtyChange,
  onTargetChanged,
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
  // 429 Retry-After の期限（epoch ms）。期限までは「同じ内容で再送信」を無効にする。
  const [resendAvailableAt, setResendAvailableAt] = useState<number | null>(null);
  const inFlightRef = useRef(false);
  const alertRef = useRef<HTMLDivElement>(null);

  function persist(
    nextValues: RevisitIntakeFullAnswers,
    nextPending: RevisitIntakePendingOperation | null,
  ) {
    if (nextPending === null && areRevisitIntakeAnswersEqual(nextValues, baseline)) {
      draftStore.clear(targetKey);
    } else {
      draftStore.set(targetKey, { values: nextValues, pending: nextPending });
    }
  }

  function updatePending(next: RevisitIntakePendingOperation | null) {
    setPending(next);
    persist(values, next);
  }

  function applyResult(
    result: ResolvedLookupResult,
    op: RevisitIntakePendingOperation,
  ): EditorOutcome {
    if (result.operation_result.state === "succeeded") {
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
    updatePending({ ...op, status: result.operation_result.state });
    return { kind: "unresolved", message: null };
  }

  async function lookup(op: RevisitIntakePendingOperation): Promise<EditorOutcome> {
    try {
      const result = await liffApi.getRevisitIntakeOperation(clinicId, op.key, idToken);
      if (isLookupNotFound(result)) {
        updatePending({ ...op, status: "not_found" });
        return { kind: "unresolved", message: null };
      }
      return applyResult(result, op);
    } catch (err: unknown) {
      updatePending(op);
      const info = resolveFetchError(err, "送信結果の確認");
      if (info.status === 401) return { kind: "unauthorized", message: info.message };
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
          if (fresh) updatePending(null);
          return { kind: "unauthorized", message: resolveFetchError(err, "問診の送信").message };
        case "access_denied":
          onAccessDenied();
          return { kind: "idle" };
        case "rate_limited":
          updatePending({ ...op, status: "unknown" });
          setResendAvailableAt(
            kind.retryAfterSeconds !== null && kind.retryAfterSeconds > 0
              ? Date.now() + kind.retryAfterSeconds * 1000
              : null,
          );
          return { kind: "unresolved", message: rateLimitedMessage(kind.retryAfterSeconds) };
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
        const fieldErrors = validateRevisitIntakeAnswers(values);
        if (firstErrorField(fieldErrors) !== null) {
          return { kind: "invalid", fieldErrors };
        }
        const body = buildRevisitIntakeSubmitRequest(current, read.form_definition_version, values);
        const op: RevisitIntakePendingOperation = {
          key: createRevisitIntakeOperationKey(),
          fingerprint: revisitIntakeFingerprint(body),
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
        if (resendAvailableAt !== null && Date.now() < resendAvailableAt) return previous;
        return await send(pending, false);
      }
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
          onRefetch={onRefetch}
        />
      ) : null}

      {targetChangeLocked ? (
        <RefetchBanner
          message="予約の対象が変更されました。最新の内容を読み込むまで送信できません。"
          refetching={refetching}
          onRefetch={onRefetch}
        />
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
          resendBlocked={resendAvailableAt !== null}
          authBlocked={authBlocked}
          onResend={() => trigger("resend")}
          onLookup={() => trigger("lookup")}
        />
      ) : null}

      {answered && current.content_available && current.latest_answer ? (
        <details
          open={outcome.kind === "revision_conflict" && !needsRefetch}
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
          disabled={isPending || needsRefetch || targetChangeLocked || authBlocked}
        >
          {isPending ? "送信中..." : answered ? "訂正して再送信" : "回答を送信する"}
        </PrimaryButton>
      ) : null}
    </form>
  );
}
