// EMR-477: 飼主事前問診（LINE・読み取り専用）。
// カルテ固定 target の飼主原文を別データ・別区画として表示する。スタッフ編集・複写・クリア・Undo の
// 対象外。書き込み経路を構造で持たないため、onChange / 保存 / 複写系 props は一切受け取らない。
// 業務正本: docs/spec/domains/medical-records/lifecycle.md AC-medical-records-owner-intake-protected。
import { memo, type ReactNode } from "react";

// Internal
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { EMPTY_VALUE } from "@/constants/empty-value";
import { formatDate } from "@/lib/format/date";
import { formatJSTTime } from "@/lib/jst-date";
import { C } from "@/lib/design-tokens";

// Relative
import {
  getRevisitIntakeErrorStatus,
  useGetRecordRevisitIntake,
  type RevisitIntakeAnswer,
  type RevisitIntakeBoundTarget,
  type RevisitIntakeHistoryEntry,
} from "../api/revisit-intake";
import {
  REVISIT_INTAKE_ENERGY_APPETITE_LABELS,
  REVISIT_INTAKE_EXAMINATION_PREFERENCE_LABELS,
  REVISIT_INTAKE_IMPROVEMENT_LABELS,
  REVISIT_INTAKE_UNANSWERED_LABEL,
  resolveRevisitIntakeChoiceLabel,
} from "../lib/revisit-intake-options";

const CARD_TITLE = "飼主事前問診（LINE・読み取り専用）";
const PROVENANCE_LABEL = "飼主（LINE）";

export interface OwnerIntakeReadOnlyCardProps {
  /** 読取に必要な ID のみ。値の変更・保存のコールバックは持たない。 */
  recordId: string | undefined;
  recordClinicId: string | undefined;
  /** カルテの petId。選択中の患者と一致が確認できた時だけ呼び出し側が渡す。 */
  petId: string | undefined;
  isNewRecord?: boolean;
}

function formatReceivedAt(receivedAt: string): string {
  const date = formatDate(receivedAt);
  if (date === EMPTY_VALUE) return EMPTY_VALUE;
  return `${date} ${formatJSTTime(receivedAt)}`;
}

function textOrNull(value: string | null): string | null {
  // 任意自由文は null・空文字ともに表示は未回答。空白のみは原文を書き換えず表示する。
  return value === null || value === "" ? null : value;
}

function AnswerValue({ value }: { value: string | null }) {
  if (value === null) {
    return <span className={C.text60}>{REVISIT_INTAKE_UNANSWERED_LABEL}</span>;
  }
  return <span className={`${C.text} whitespace-pre-wrap break-words`}>{value}</span>;
}

function AnswerFields({ answer }: { answer: RevisitIntakeAnswer }) {
  const a = answer.answers;
  const energy = resolveRevisitIntakeChoiceLabel(REVISIT_INTAKE_ENERGY_APPETITE_LABELS, a.energy);
  const appetite = resolveRevisitIntakeChoiceLabel(
    REVISIT_INTAKE_ENERGY_APPETITE_LABELS,
    a.appetite,
  );
  const energyAppetite =
    energy === null && appetite === null
      ? null
      : `元気: ${energy ?? REVISIT_INTAKE_UNANSWERED_LABEL} / 食欲: ${appetite ?? REVISIT_INTAKE_UNANSWERED_LABEL}`;

  const rows: ReadonlyArray<readonly [string, string | null]> = [
    ["来院目的", textOrNull(a.visit_purpose)],
    ["いつから", textOrNull(a.onset)],
    ["元気・食欲", energyAppetite],
    ["現在の症状", textOrNull(a.current_symptoms)],
    [
      "前回より改善したか",
      resolveRevisitIntakeChoiceLabel(REVISIT_INTAKE_IMPROVEMENT_LABELS, a.improvement),
    ],
    [
      "検査希望",
      resolveRevisitIntakeChoiceLabel(
        REVISIT_INTAKE_EXAMINATION_PREFERENCE_LABELS,
        a.examination_preference,
      ),
    ],
    ["一番相談したいこと", textOrNull(a.main_concern)],
  ];

  return (
    <dl className="grid grid-cols-1 gap-x-4 gap-y-2 text-sm sm:grid-cols-[10rem_1fr]">
      {rows.map(([label, value]) => (
        <div key={label} className="contents">
          <dt className={`font-medium ${C.text60}`}>{label}</dt>
          <dd className="m-0">
            <AnswerValue value={value} />
          </dd>
        </div>
      ))}
    </dl>
  );
}

function ProvenanceLine({ answer }: { answer: RevisitIntakeAnswer }) {
  return (
    <p className={`text-xs ${C.text60}`}>
      出所: {PROVENANCE_LABEL} / 回答版: 第{answer.revision}版 / 受信:{" "}
      {formatReceivedAt(answer.received_at)} / 項目定義 v{answer.form_definition_version}
    </p>
  );
}

function CardShell({
  badge,
  children,
  busy = false,
}: {
  badge?: string;
  children: ReactNode;
  busy?: boolean;
}) {
  return (
    <section
      aria-label={CARD_TITLE}
      aria-busy={busy}
      data-testid="owner-intake-card"
      className={`${C.bgWhite} border ${C.borderMedium} rounded-lg p-4 flex flex-col gap-3`}
    >
      <div className="flex flex-wrap items-center gap-2">
        <h2 className={`text-sm font-semibold ${C.text}`}>{CARD_TITLE}</h2>
        {badge ? <Badge variant="outline">{badge}</Badge> : null}
      </div>
      {children}
    </section>
  );
}

function UnansweredEmptyState() {
  // 空状態は「何が無いか／条件／次の操作」の 3 点（ui-design-guide §2）。
  return (
    <div className={`text-sm ${C.text} flex flex-col gap-1`}>
      <p>この予約の飼主事前問診の回答はありません。</p>
      <p className={C.text60}>飼主が LINE から送信すると、ここに原文が表示されます。</p>
      <p className={C.text60}>
        診察で確認した内容やスタッフの補足は、問診タブに記録してください。
      </p>
    </div>
  );
}

function HistoryList({
  entries,
  hasNextPage,
  isFetchingNextPage,
  onLoadMore,
}: {
  entries: RevisitIntakeHistoryEntry[];
  hasNextPage: boolean;
  isFetchingNextPage: boolean;
  onLoadMore: () => void;
}) {
  if (entries.length === 0 && !hasNextPage) return null;
  return (
    <details className={`border-t ${C.borderLight} pt-3`}>
      <summary className={`cursor-pointer text-sm font-medium ${C.text} min-h-11 flex items-center`}>
        過去の回答版（{entries.length}
        {hasNextPage ? "件以上" : "件"}）
      </summary>
      <div className="flex flex-col gap-4 mt-2">
        {entries.map((entry) => (
          <div
            key={entry.answer.answer_id}
            className={`flex flex-col gap-2 border ${C.borderLight} rounded-md p-3`}
            data-testid="owner-intake-history-entry"
          >
            <AnswerFields answer={entry.answer} />
            <ProvenanceLine answer={entry.answer} />
          </div>
        ))}
        {hasNextPage ? (
          <div>
            <Button
              type="button"
              variant="outline"
              onClick={onLoadMore}
              disabled={isFetchingNextPage}
            >
              {isFetchingNextPage ? "読み込み中…" : "さらに表示"}
            </Button>
          </div>
        ) : null}
      </div>
    </details>
  );
}

function BoundBody({
  target,
  history,
  hasNextPage,
  isFetchingNextPage,
  onLoadMore,
}: {
  target: RevisitIntakeBoundTarget;
  history: RevisitIntakeHistoryEntry[];
  hasNextPage: boolean;
  isFetchingNextPage: boolean;
  onLoadMore: () => void;
}) {
  const latest = target.latest_answer;
  const canShowBody = target.answer_state === "answered" && target.content_available && !!latest;

  // 固定 target 内の旧 revision だけを出す。他 target の行・現在版との重複は出さない。
  const pastEntries = history.filter(
    (entry) =>
      entry.target.target_id === target.ref.target_id &&
      entry.answer.answer_id !== latest?.answer_id,
  );

  if (target.answer_state === "unanswered") {
    return (
      <CardShell badge="未回答">
        <UnansweredEmptyState />
      </CardShell>
    );
  }

  if (!canShowBody) {
    return (
      <CardShell badge="回答済み">
        <p className={`text-sm ${C.text}`}>回答済み（本文は閲覧できません）</p>
      </CardShell>
    );
  }

  return (
    <CardShell badge={`回答済み 第${latest.revision}版`}>
      <AnswerFields answer={latest} />
      <ProvenanceLine answer={latest} />
      <HistoryList
        entries={pastEntries}
        hasNextPage={hasNextPage}
        isFetchingNextPage={isFetchingNextPage}
        onLoadMore={onLoadMore}
      />
    </CardShell>
  );
}

export const OwnerIntakeReadOnlyCard = memo(function OwnerIntakeReadOnlyCard({
  recordId,
  recordClinicId,
  petId,
  isNewRecord = false,
}: OwnerIntakeReadOnlyCardProps) {
  const query = useGetRecordRevisitIntake({
    recordId,
    clinicId: recordClinicId,
    petId,
    enabled: !isNewRecord,
  });

  // 新規カルテ・ID 未確定（患者切替の遷移途中を含む）では何も取得・表示しない。
  if (isNewRecord || !recordId || !recordClinicId || !petId) return null;

  if (query.isPending) {
    return (
      <CardShell busy>
        <p role="status" className={`text-sm ${C.text60}`}>
          飼主事前問診を読み込み中…
        </p>
      </CardShell>
    );
  }

  if (query.isError) {
    const status = getRevisitIntakeErrorStatus(query.error);
    if (status === 404) return null;
    if (status === 403) {
      return (
        <p className={`text-xs ${C.text60}`} data-testid="owner-intake-forbidden-note">
          飼主事前問診を表示する権限がありません。
        </p>
      );
    }
    return (
      <CardShell>
        <div role="alert" className="flex flex-wrap items-center gap-3 text-sm">
          <span className={C.text}>飼主事前問診を取得できませんでした。</span>
          <Button
            type="button"
            variant="outline"
            onClick={() => void query.refetch()}
            disabled={query.isFetching}
          >
            再試行
          </Button>
        </div>
      </CardShell>
    );
  }

  const first = query.data.pages[0];

  if (first.binding === "unbound") {
    return (
      <CardShell badge="対応不明">
        <p className={`text-sm ${C.text}`}>対応する予約回答を特定できません</p>
        <p className={`text-xs ${C.text60}`}>
          このカルテは予約の問診回答と結び付いていないため、未回答とは判断できません。
        </p>
      </CardShell>
    );
  }

  if (first.binding === "mismatch") {
    // 本文は一切出さない。
    return (
      <CardShell badge="出所不明">
        <p className={`text-sm ${C.text}`}>
          予約回答の出所を確認できないため、飼主の回答は表示しません。
        </p>
      </CardShell>
    );
  }

  const history = query.data.pages.flatMap((page) =>
    page.binding === "bound" ? page.history : [],
  );

  return (
    <BoundBody
      target={first.target}
      history={history}
      hasNextPage={query.hasNextPage}
      isFetchingNextPage={query.isFetchingNextPage}
      onLoadMore={() => void query.fetchNextPage()}
    />
  );
});
