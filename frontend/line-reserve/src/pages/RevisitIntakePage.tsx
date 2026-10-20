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
import { RevisitIntakeHistory } from "../components/revisit-intake/RevisitIntakeHistory";
import { RevisitIntakeAnswerView } from "../components/revisit-intake/RevisitIntakeAnswerView";
import type { RevisitIntakeHistoryEntry, RevisitIntakeOwnerCurrentTarget } from "../types/models";

/** 予約一覧・完了画面から渡す表示用の予約概要（API DTO は変えない）。 */
export interface RevisitIntakeReservationSummary {
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
  const [notice, setNotice] = useState<string | null>(null);
  const [dirty, setDirty] = useState(false);
  const [confirmingLeave, setConfirmingLeave] = useState(false);
  const { refetch } = intake;

  const handleDirtyChange = useCallback((next: boolean) => setDirty(next), []);
  const handleTargetChanged = useCallback(() => {
    setNotice(
      "予約の対象が変更されたため、問診は送信されませんでした。最新の内容を読み込みます。以前の入力は引き継がれません。",
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

  // アプリ内の離脱だけを確認する（LINE 終了・OS による破棄は保証しない）。
  const handleBack = () => {
    if (dirty) {
      setConfirmingLeave(true);
      return;
    }
    onBack();
  };

  if (accessDenied || isRevisitIntakeAccessDenied(intake.error)) {
    return <ErrorPage message={ACCESS_DENIED_MESSAGE} onAction={onBack} actionLabel="戻る" />;
  }

  const { read } = intake;
  const current = read?.current_target ?? null;

  function renderCurrent() {
    if (!read) return null;
    if (read.form_definition_version !== REVISIT_INTAKE_SUPPORTED_FORM_VERSION) {
      return (
        <p role="status" className="rounded-lg bg-white px-4 py-3 text-sm text-noah-text">
          この問診の形式にはまだ対応していません。LINEアプリを最新にしてから開き直してください。
        </p>
      );
    }
    if (read.availability === "target_unresolved" || current === null) {
      return (
        <p role="status" className="rounded-lg bg-white px-4 py-3 text-sm text-noah-text">
          {EDIT_BLOCK_MESSAGES.target_unresolved}
        </p>
      );
    }
    const answered = current.answer_state === "answered";
    const targetKey = revisitIntakeTargetKey(clinicId, current.ref);
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

        {current.editable ? (
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
            onRefetch={refetch}
            onDirtyChange={handleDirtyChange}
            onTargetChanged={handleTargetChanged}
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
          {confirmingLeave ? (
            <div
              role="alert"
              className="mb-3 space-y-2 rounded-lg border border-noah-warning-border bg-noah-warning-box-bg px-4 py-3"
            >
              <p className="text-sm text-noah-warning-text">
                回答はまだ送信されていません。このまま戻りますか？
              </p>
              <div className="flex gap-2">
                <button
                  type="button"
                  onClick={() => {
                    setConfirmingLeave(false);
                    onBack();
                  }}
                  className="min-h-11 flex-1 rounded-lg bg-noah-teal px-3 text-sm font-semibold text-white"
                >
                  送信せずに戻る
                </button>
                <button
                  type="button"
                  onClick={() => setConfirmingLeave(false)}
                  className="min-h-11 flex-1 rounded-lg border border-noah-border bg-white px-3 text-sm text-noah-text"
                >
                  入力を続ける
                </button>
              </div>
            </div>
          ) : null}
          <h1 className="text-lg font-bold text-noah-teal-dark">再診の事前問診</h1>
          {summary && (summary.petName || summary.dateLabel || summary.courseName) ? (
            <p className="mt-1 text-sm text-noah-text-sub">
              {[summary.petName, summary.dateLabel, summary.courseName].filter(Boolean).join(" / ")}
            </p>
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
              <p role="alert">{intake.error.message}</p>
              {intake.error.canRetry ? (
                <button
                  type="button"
                  onClick={refetch}
                  className="mt-2 min-h-11 px-3 text-sm font-medium text-noah-teal-dark underline"
                >
                  再試行
                </button>
              ) : null}
            </div>
          ) : null}

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

          {renderCurrent()}

          {read ? (
            <RevisitIntakeHistory
              history={intake.history}
              hasMore={intake.nextCursor !== null}
              loadingMore={intake.loadingMore}
              loadMoreError={intake.loadMoreError}
              onLoadMore={intake.loadMore}
            />
          ) : null}
        </div>
      </div>
    </div>
  );
}
