// EMR-474: LIFF 再診事前問診（予約単位。1 予約 = 1 ペット = 1 form）。
// 業務正本: docs/spec/domains/reservations/pages/reservation-management.md#line-revisit-intake-20261017
// 技術契約: docs/design/domains/reservations/technical-spec.md#revisit-intake-api-selected-v3
// 予約 ID は App の state で受け取り URL へ載せない。入力 draft は target 別 memory だけに保持する。
import { useCallback, useState } from "react";
import { BackButton } from "../components/BackButton";
import { Spinner } from "@/shared-liff/Spinner";
import { ErrorPage } from "@/shared-liff/ErrorPage";
import {
  isRevisitIntakeAccessDenied,
  useRevisitIntakeRead,
} from "../hooks/use-revisit-intake-read";
import {
  revisitIntakeTargetKey,
  type RevisitIntakeDraftStore,
} from "../hooks/use-revisit-intake-drafts";
import {
  EDIT_BLOCK_FALLBACK_MESSAGE,
  EDIT_BLOCK_MESSAGES,
  REVISIT_INTAKE_SUPPORTED_FORM_VERSION,
  formatRevisitIntakeDateTime,
} from "../lib/revisit-intake-options";
import {
  RevisitIntakeEditor,
  type RevisitIntakeClosedReason,
} from "../components/revisit-intake/RevisitIntakeEditor";
import {
  PENDING_UNAUTHORIZED_TEXT,
  RevisitIntakeDeniedPendingLookup,
} from "../components/revisit-intake/RevisitIntakeDeniedPendingLookup";
import { RevisitIntakeHistory } from "../components/revisit-intake/RevisitIntakeHistory";
import { RevisitIntakeAnswerView } from "../components/revisit-intake/RevisitIntakeAnswerView";
import { RevisitIntakeRecoveryPanel } from "../components/revisit-intake/RevisitIntakeRecoveryPanel";
import { useRevisitIntakeRecovery } from "../hooks/use-revisit-intake-recovery";
import type { RevisitIntakeHistoryEntry, RevisitIntakeOwnerCurrentTarget } from "../types/models";

/** 予約一覧・完了画面から渡す表示用の予約概要（API DTO は変えない）。 */
export interface RevisitIntakeReservationSummary {
  /**
   * 表示用 snapshot の対象ペット ID（分かる導線だけ渡す）。petName は petId が現在 target の pet と
   * 一致するときだけ表示し、不明・不一致なら出さない（pet 名は推定しない）。日時・コースは予約に紐づくため表示する。
   */
  petId?: string;
  petName?: string;
  dateLabel?: string;
  courseName?: string;
}

interface RevisitIntakePageProps {
  clinicId: string;
  idToken: string;
  appointmentId: string;
  summary?: RevisitIntakeReservationSummary;
  draftStore: RevisitIntakeDraftStore;
  onBack: () => void;
}

const ACCESS_DENIED_MESSAGE =
  "この予約の問診を表示できません。予約一覧から開き直すか、医院へお問い合わせください。";

const CLOSED_NOTICES: Record<Exclude<RevisitIntakeClosedReason, null>, string> = {
  reception_started:
    "受付が始まったため、問診は送信されませんでした。受付後の変更はスタッフへお伝えください。",
  cancelled: "予約がキャンセルされているため、問診は送信されませんでした。",
  deleted: "予約が取り消されているため、問診は送信されませんでした。",
  unsupported_lifecycle:
    "現在の予約状態では問診を送信できませんでした。変更があれば医院へお伝えください。",
};

function AnswerTimestamps({
  current,
  history,
}: {
  current: RevisitIntakeOwnerCurrentTarget;
  history: RevisitIntakeHistoryEntry[];
}) {
  const latest = current.latest_answer;
  if (!latest) return null;
  if (latest.revision === "1") {
    return (
      <p className="text-xs text-noah-text-sub">
        回答日時 {formatRevisitIntakeDateTime(latest.received_at)}
      </p>
    );
  }
  const first = history.find(
    (h) =>
      h.target_relation === "current" &&
      h.target.target_id === current.ref.target_id &&
      h.answer.revision === "1",
  );
  return (
    <div className="space-y-0.5 text-xs text-noah-text-sub">
      {first ? <p>回答日時 {formatRevisitIntakeDateTime(first.answer.received_at)}</p> : null}
      <p>最終訂正日時 {formatRevisitIntakeDateTime(latest.received_at)}</p>
    </div>
  );
}

function StatusBadge({ answered }: { answered: boolean }) {
  return (
    <span
      className={`shrink-0 whitespace-nowrap rounded-full px-2 py-1 text-xs font-medium ${
        answered
          ? "bg-noah-success-bg text-noah-success-text"
          : "bg-noah-neutral-bg text-noah-neutral-text"
      }`}
    >
      {answered ? "回答済み" : "未回答"}
    </span>
  );
}

export function RevisitIntakePage({
  clinicId,
  idToken,
  appointmentId,
  summary,
  draftStore,
  onBack,
}: RevisitIntakePageProps) {
  const intake = useRevisitIntakeRead(clinicId, appointmentId, idToken);
  const [accessDenied, setAccessDenied] = useState(false);
  // 操作回復 Slice A: 再読込後の未確定操作を発見する。発見が済むまで・保留理由がある間は新規送信を止める。
  // 0 件・complete・経過時間から送信の許可を導かない（recovery.newSubmissionBlocked が false でも許可ではない）。
  const recoveryTarget = intake.read?.current_target?.ref ?? null;
  const recovery = useRevisitIntakeRecovery({
    clinicId,
    idToken,
    target:
      recoveryTarget === null
        ? null
        : {
            appointmentId: recoveryTarget.appointment_id,
            petId: recoveryTarget.pet_id,
            targetGeneration: recoveryTarget.target_generation,
          },
    selfScope:
      accessDenied ||
      intake.error !== null ||
      (intake.read !== null && intake.read.current_target === null),
  });
  const [notice, setNotice] = useState<string | null>(null);
  const [dirty, setDirty] = useState(false);
  const [leavePrompt, setLeavePrompt] = useState<"dirty" | "unresolved" | "in_flight" | null>(null);
  const [submitInFlight, setSubmitInFlight] = useState(false);
  const [submitUnresolved, setSubmitUnresolved] = useState(false);
  const [, setPendingEpoch] = useState(0);
  const { refetch } = intake;
  // store は reactive でないため、照会で pending が消えたら再描画と再取得を起こす。
  const handlePendingResolved = useCallback(() => {
    setPendingEpoch((n) => n + 1);
    refetch();
  }, [refetch]);

  const handleDirtyChange = useCallback((next: boolean) => setDirty(next), []);
  const handlePendingChange = useCallback((inFlight: boolean, unresolved: boolean) => {
    setSubmitInFlight(inFlight);
    setSubmitUnresolved(unresolved);
  }, []);
  const handleTargetChanged = useCallback(() => {
    setNotice(
      "予約の対象が変更されたため、問診は送信されませんでした。最新の内容を読み込みます。以前の入力は引き継がれません。",
    );
    refetch();
  }, [refetch]);
  const handleSucceededThenTargetChanged = useCallback(() => {
    setNotice(
      "回答の送信は完了しています。その後、予約の対象が変更されたため最新の内容を読み込みます。",
    );
    refetch();
  }, [refetch]);
  const handleIntakeClosed = useCallback(
    (reason: RevisitIntakeClosedReason) => {
      setNotice(reason === null ? EDIT_BLOCK_FALLBACK_MESSAGE : CLOSED_NOTICES[reason]);
      refetch();
    },
    [refetch],
  );
  const handleAccessDenied = useCallback(() => setAccessDenied(true), []);

  // renderCurrent が RevisitIntakeEditor を載せる条件と同一（ここを変えるときは両方を揃える）。
  const isEditorMounted = () => {
    const { read: r } = intake;
    const cur = r?.current_target ?? null;
    if (r === null || cur === null) return false;
    if (recovery.newSubmissionBlocked) return false;
    if (r.form_definition_version !== REVISIT_INTAKE_SUPPORTED_FORM_VERSION) return false;
    if (r.availability === "target_unresolved" || !cur.editable) return false;
    const pending = draftStore.findPending(clinicId, appointmentId);
    return (
      pending === undefined || pending.targetKey === revisitIntakeTargetKey(clinicId, cur.ref)
    );
  };

  // アプリ内の離脱だけを確認する（LINE 終了・OS による破棄は保証しない）。
  // 送信中は戻らせない。結果未確定・未送信の入力は、それぞれ正確な文言で確認する。
  const handleBack = () => {
    if (submitInFlight) {
      setLeavePrompt("in_flight");
      return;
    }
    // editor が外れている間は submitUnresolved が false に戻る。memory store の未確定送信（送信済み）を先に見る。
    if (
      submitUnresolved ||
      (!isEditorMounted() && draftStore.findPending(clinicId, appointmentId) !== undefined)
    ) {
      setLeavePrompt("unresolved");
      return;
    }
    // editor が外れている間（再取得中・失敗中・読み取り専用・未対応形式・target 不明・別 target の未確定送信）は
    // dirty が false に戻る。editor が draft を代表していないときは memory store の未送信 draft を見る。
    if (dirty || (!isEditorMounted() && draftStore.hasDraft(clinicId, appointmentId))) {
      setLeavePrompt("dirty");
      return;
    }
    onBack();
  };
  const leave = () => {
    setLeavePrompt(null);
    onBack();
  };

  if (accessDenied || isRevisitIntakeAccessDenied(intake.error)) {
    // 照会 route は閲覧権限と分離されている。未確定の送信が残っていれば key を失わないよう結果照会だけ出す。
    const deniedPending = draftStore.findPending(clinicId, appointmentId);
    // self scope の発見が失敗または保留のときだけ殻を出す。成功済みの送信は別予約のものを含み得るため、
    // それだけで「この予約に回答済み」と読める表示にはしない（従来どおり ErrorPage）。
    const deniedRecoveryVisible =
      recovery.status === "error" ||
      (recovery.summary !== null && recovery.summary.hasOperationHold);
    if (deniedPending === undefined && !deniedRecoveryVisible) {
      return <ErrorPage message={ACCESS_DENIED_MESSAGE} onAction={onBack} actionLabel="戻る" />;
    }
    return (
      <div className="min-h-screen bg-noah-teal-light flex flex-col">
        <div className="max-w-md mx-auto w-full flex flex-col flex-1">
          <div className="px-4 pt-4">
            {/* 未確定の送信（memory のみの key）が残る間は、確認なしに離れない。 */}
            <BackButton
              onClick={() => {
                // 照会成功で store が空になった直後は、古い描画のまま確認を挟まない。
                if (draftStore.findPending(clinicId, appointmentId) === undefined) onBack();
                else setLeavePrompt("unresolved");
              }}
            />
            {leavePrompt === "unresolved" ? (
              <div
                role="alert"
                className="mb-3 space-y-2 rounded-lg border border-noah-warning-border bg-noah-warning-box-bg px-4 py-3"
              >
                <p className="text-sm text-noah-warning-text">
                  送信結果が未確定です。この画面を離れると、予約一覧にこの予約が表示されない場合は結果を確認できなくなります。
                </p>
                <div className="flex gap-2">
                  <button
                    type="button"
                    onClick={leave}
                    className="min-h-11 flex-1 rounded-lg bg-noah-teal px-3 text-sm font-semibold text-white"
                  >
                    このまま戻る
                  </button>
                  <button
                    type="button"
                    onClick={() => setLeavePrompt(null)}
                    className="min-h-11 flex-1 rounded-lg border border-noah-border bg-white px-3 text-sm text-noah-text"
                  >
                    この画面に留まる
                  </button>
                </div>
              </div>
            ) : null}
            <h1 className="text-lg font-bold text-noah-teal-dark">再診の事前問診</h1>
          </div>
          <div className="flex-1 space-y-4 px-4 py-4">
            <p role="status" className="text-sm text-noah-text">
              {ACCESS_DENIED_MESSAGE}
            </p>
            <RevisitIntakeRecoveryPanel recovery={recovery} clinicId={clinicId} idToken={idToken} />
            {deniedPending !== undefined ? (
              <RevisitIntakeDeniedPendingLookup
                clinicId={clinicId}
                idToken={idToken}
                appointmentId={appointmentId}
                targetKey={deniedPending.targetKey}
                pending={deniedPending.pending}
                draftStore={draftStore}
              />
            ) : null}
          </div>
        </div>
      </div>
    );
  }

  const { read } = intake;
  const current = read?.current_target ?? null;
  // pet 名は petId が与えられ現在 target の pet と一致するときだけ表示する（検証できない名前は出さない）。
  const summaryPetName =
    summary?.petName &&
    summary.petId !== undefined &&
    current !== null &&
    summary.petId === current.ref.pet_id
      ? summary.petName
      : undefined;
  const summaryParts = [summaryPetName, summary?.dateLabel, summary?.courseName].filter(Boolean);

  // 予約単位の未確定送信。現 target・編集可否・形式対応に関わらず、結果照会の入口は常に残す
  // （照会 route は閲覧・編集権限と独立。key を失うと結果を確かめられず新 key の送信が起こり得る）。
  const appointmentPending = draftStore.findPending(clinicId, appointmentId);

  function renderPendingLookup(
    reason: "read_only" | "other_target" | "read_failed",
    found: NonNullable<typeof appointmentPending>,
  ) {
    return (
      <RevisitIntakeDeniedPendingLookup
        key={found.targetKey}
        clinicId={clinicId}
        idToken={idToken}
        appointmentId={appointmentId}
        targetKey={found.targetKey}
        pending={found.pending}
        draftStore={draftStore}
        reason={reason}
        onSucceeded={handlePendingResolved}
      />
    );
  }

  function renderCurrent() {
    if (!read) return null;
    // 編集できる状態では、未確定操作の発見が済むまで status・editor を出さない（発見前の送信を防ぐ）。
    if (
      current !== null &&
      current.editable &&
      recovery.status === "checking" &&
      recovery.discovery === null
    ) {
      return (
        <div className="py-12 text-center" role="status">
          <Spinner size="sm" borderColorClassName="border-noah-teal" />
          <p className="text-sm text-noah-text-sub">未確定の操作を確認中...</p>
        </div>
      );
    }
    if (read.form_definition_version !== REVISIT_INTAKE_SUPPORTED_FORM_VERSION) {
      return (
        <div className="space-y-3">
          <p role="status" className="rounded-lg bg-white px-4 py-3 text-sm text-noah-text">
            この問診の形式にはまだ対応していません。LINEアプリを最新にしてから開き直してください。
          </p>
          {appointmentPending !== undefined
            ? renderPendingLookup("read_only", appointmentPending)
            : null}
        </div>
      );
    }
    if (read.availability === "target_unresolved" || current === null) {
      return (
        <div className="space-y-3">
          <p role="status" className="rounded-lg bg-white px-4 py-3 text-sm text-noah-text">
            {EDIT_BLOCK_MESSAGES.target_unresolved}
          </p>
          {appointmentPending !== undefined
            ? renderPendingLookup("read_only", appointmentPending)
            : null}
        </div>
      );
    }
    const answered = current.answer_state === "answered";
    const targetKey = revisitIntakeTargetKey(clinicId, current.ref);
    // 編集不可でも、未確定の送信が残っていれば key を失わないよう結果照会だけ出す（照会 route は編集可否と独立）。
    const readOnlyPending = current.editable ? undefined : appointmentPending;
    // 別 target（旧世代・別ペット）の未確定送信が残っている間は editor を載せない。新 key の送信を起こさず、
    // 旧 operation の結果が succeeded と確定するまで照会だけを出す（not_found 等からは新 key を許さない）。
    const otherTargetPending =
      current.editable &&
      appointmentPending !== undefined &&
      appointmentPending.targetKey !== targetKey
        ? appointmentPending
        : undefined;
    return (
      <div className="space-y-4">
        <div className="rounded-xl border border-noah-border bg-white p-4">
          <div className="flex items-start justify-between gap-2">
            <p className="text-sm font-semibold text-noah-text">問診の状態</p>
            <StatusBadge answered={answered} />
          </div>
          {answered && current.content_available ? (
            <div className="mt-2">
              <AnswerTimestamps current={current} history={intake.history} />
            </div>
          ) : null}
          {answered && !current.content_available ? (
            <p className="mt-2 text-sm text-noah-text-sub">
              回答済みです。回答内容はこの画面では表示できません。
              {current.editable ? "新しく送信すると、最新の回答として保存されます。" : ""}
            </p>
          ) : null}
        </div>

        {otherTargetPending !== undefined ? (
          renderPendingLookup("other_target", otherTargetPending)
        ) : current.editable && recovery.newSubmissionBlocked ? (
          // Editor を外す間も、同 target の memory pending は結果照会だけ出す（再送・新規送信は出さない）。
          appointmentPending !== undefined && appointmentPending.targetKey === targetKey ? (
            renderPendingLookup("read_only", appointmentPending)
          ) : null
        ) : current.editable ? (
          <RevisitIntakeEditor
            key={targetKey}
            clinicId={clinicId}
            idToken={idToken}
            appointmentId={appointmentId}
            read={read}
            current={current}
            targetKey={targetKey}
            draftStore={draftStore}
            refetching={intake.refetching}
            readError={intake.error !== null}
            retryBlocked={intake.retryBlocked}
            onRefetch={refetch}
            onDirtyChange={handleDirtyChange}
            onPendingChange={handlePendingChange}
            onTargetChanged={handleTargetChanged}
            onSucceededThenTargetChanged={handleSucceededThenTargetChanged}
            onIntakeClosed={handleIntakeClosed}
            onAccessDenied={handleAccessDenied}
          />
        ) : (
          <div className="space-y-3">
            <p
              role="status"
              className="rounded-lg border border-noah-warning-border bg-noah-warning-box-bg px-4 py-3 text-sm text-noah-warning-text"
            >
              {current.edit_block_reason !== null
                ? EDIT_BLOCK_MESSAGES[current.edit_block_reason]
                : EDIT_BLOCK_FALLBACK_MESSAGE}
            </p>
            {readOnlyPending !== undefined
              ? renderPendingLookup("read_only", readOnlyPending)
              : null}
            {current.content_available && current.latest_answer ? (
              <div className="rounded-xl border border-noah-border bg-white px-4 py-2">
                <p className="pt-2 text-sm font-semibold text-noah-text">最新の回答</p>
                <RevisitIntakeAnswerView answers={current.latest_answer.answers} />
              </div>
            ) : null}
          </div>
        )}
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-noah-teal-light flex flex-col">
      <div className="max-w-md mx-auto w-full flex flex-col flex-1">
        <div className="px-4 pt-4">
          <BackButton onClick={handleBack} />
          {leavePrompt === "in_flight" && submitInFlight ? (
            <p
              role="alert"
              className="mb-3 rounded-lg border border-noah-warning-border bg-noah-warning-box-bg px-4 py-3 text-sm text-noah-warning-text"
            >
              送信処理中です。結果が確定するまでお待ちください。
            </p>
          ) : null}
          {leavePrompt === "unresolved" || leavePrompt === "dirty" ? (
            <div
              role="alert"
              className="mb-3 space-y-2 rounded-lg border border-noah-warning-border bg-noah-warning-box-bg px-4 py-3"
            >
              <p className="text-sm text-noah-warning-text">
                {leavePrompt === "unresolved"
                  ? "送信結果が未確定です。戻っても再開時に結果を確認できます（入力内容と送信状態は、LINEを閉じるまでこの端末に一時的に残ります）。"
                  : "回答はまだ送信されていません。このまま戻りますか？"}
              </p>
              <div className="flex gap-2">
                <button
                  type="button"
                  onClick={leave}
                  className="min-h-11 flex-1 rounded-lg bg-noah-teal px-3 text-sm font-semibold text-white"
                >
                  {leavePrompt === "unresolved" ? "このまま戻る" : "送信せずに戻る"}
                </button>
                <button
                  type="button"
                  onClick={() => setLeavePrompt(null)}
                  className="min-h-11 flex-1 rounded-lg border border-noah-border bg-white px-3 text-sm text-noah-text"
                >
                  {leavePrompt === "unresolved" ? "この画面に留まる" : "入力を続ける"}
                </button>
              </div>
            </div>
          ) : null}
          <h1 className="text-lg font-bold text-noah-teal-dark">再診の事前問診</h1>
          {summaryParts.length > 0 ? (
            <p className="mt-1 text-sm text-noah-text-sub">{summaryParts.join(" / ")}</p>
          ) : null}
        </div>

        <div className="flex-1 space-y-4 px-4 py-4">
          {intake.loading ? (
            <div className="py-12 text-center" role="status">
              <Spinner size="sm" borderColorClassName="border-noah-teal" />
              <p className="text-sm text-noah-text-sub">読み込み中...</p>
            </div>
          ) : null}

          {intake.error !== null ? (
            <div className="py-4 text-center text-noah-danger">
              <p role="alert">
                {/* 読込 401 で LINE の再起動を案内すると memory-only の key を失う。未確定送信があるときは閉じさせない。 */}
                {intake.error.status === 401 && read === null && appointmentPending !== undefined
                  ? PENDING_UNAUTHORIZED_TEXT
                  : intake.error.message}
              </p>
              {intake.error.canRetry ? (
                <>
                  {intake.retryBlocked && intake.error.retryAfterSeconds !== null ? (
                    <p className="mt-1 text-sm text-noah-text-sub">
                      {`混み合っています。${intake.error.retryAfterSeconds}秒ほど待ってから再試行してください。`}
                    </p>
                  ) : null}
                  <button
                    type="button"
                    onClick={refetch}
                    disabled={intake.retryBlocked}
                    className="mt-2 min-h-11 px-3 text-sm font-medium text-noah-teal-dark underline disabled:opacity-50"
                  >
                    再試行
                  </button>
                </>
              ) : null}
            </div>
          ) : null}

          {intake.error !== null && read === null && appointmentPending !== undefined
            ? renderPendingLookup("read_failed", appointmentPending)
            : null}

          {notice !== null ? (
            <div
              role="alert"
              className="flex items-start justify-between gap-3 rounded-lg border border-noah-danger-border bg-noah-danger-bg px-4 py-3"
            >
              <p className="flex-1 text-sm text-noah-danger-text">{notice}</p>
              <button
                type="button"
                onClick={() => setNotice(null)}
                aria-label="閉じる"
                className="min-h-11 min-w-11 font-bold leading-none text-noah-danger-text"
              >
                ×
              </button>
            </div>
          ) : null}

          {intake.refetching ? (
            <p role="status" className="text-xs text-noah-text-sub">
              最新の状態を読み込み中...
            </p>
          ) : null}

          <RevisitIntakeRecoveryPanel
            recovery={recovery}
            clinicId={clinicId}
            idToken={idToken}
            fenceNoticeRelevant={current !== null && current.editable}
          />

          {renderCurrent()}

          {read && read.form_definition_version === REVISIT_INTAKE_SUPPORTED_FORM_VERSION ? (
            <RevisitIntakeHistory
              history={intake.history}
              hasMore={intake.nextCursor !== null}
              loadingMore={intake.loadingMore}
              loadMoreError={intake.loadMoreError}
              loadMoreRetryAfterSeconds={intake.loadMoreRetryAfterSeconds}
              onLoadMore={intake.loadMore}
            />
          ) : null}
        </div>
      </div>
    </div>
  );
}
