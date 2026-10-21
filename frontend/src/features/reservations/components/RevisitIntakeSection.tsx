import { useEffect, useMemo, useState, type ReactNode } from "react";
import { useQueryClient } from "@tanstack/react-query";
import { CircleDashed, CircleCheck, ClipboardList, Lock, MessageSquareText } from "lucide-react";
import { BADGE, C, ICON, STYLE } from "@/lib/design-tokens";
import { Button } from "@/components/ui/button";
import { EMPTY_VALUE } from "@/constants/empty-value";
import { formatDate } from "@/lib/format/date";
import { formatJSTTime } from "@/lib/jst-date";
import {
  getRetryAfterMs,
  getRevisitIntakeErrorStatus,
  revisitIntakeQueryKey,
  useGetRevisitIntake,
  type RevisitIntakeAnswer,
  type RevisitIntakeHistoryEntry,
  type RevisitIntakeTargetRef,
} from "../api/get-revisit-intake";
import {
  getEnergyAppetiteLabel,
  getExaminationPreferenceLabel,
  getImprovementLabel,
} from "../lib/revisit-intake-options";

interface RevisitIntakeSectionProps {
  /** 表示中の予約 ID。変わると別 query になり前予約の原文は残らない。 */
  appointmentId: string;
  clinicId?: string;
  /** モーダル表示中のみ true。 */
  enabled?: boolean;
}

const SECTION_LABEL = "飼主事前問診（LINE）";
const PROVENANCE_LABEL = "飼主（LINE）";
/** docs/spec/ui-design-guide.md / design-patterns §6.2: 読込は 300ms 超でスケルトン。 */
const SKELETON_DELAY_MS = 300;

/**
 * 読込開始から SKELETON_DELAY_MS 経過後にだけバーを描く（短い読込での点滅防止）。
 * pending の間だけ mount されるため、状態のリセットは unmount で済み effect 内の同期 setState は不要。
 */
function DelayedSkeleton() {
  const [elapsed, setElapsed] = useState(false);
  useEffect(() => {
    const timer = setTimeout(() => setElapsed(true), SKELETON_DELAY_MS);
    return () => clearTimeout(timer);
  }, []);
  if (!elapsed) return null;
  return (
    <>
      <div className={`h-4 w-40 rounded ${C.bgSkeleton} animate-pulse`} />
      <div className={`h-4 w-full rounded ${C.bgSkeleton} animate-pulse`} />
      <div className={`h-4 w-3/4 rounded ${C.bgSkeleton} animate-pulse`} />
    </>
  );
}

/**
 * 直近の失敗に Retry-After があるとき、期限までの手動再試行・続き取得を止める（api.yaml: 尊重する）。
 * 期限は feature のメモリ上だけで持つ。タイマーは「期限到達」を記録するだけで、effect 内で同期 setState しない。
 * 戻り値: 待機中は案内用の秒数、可能なら null。
 */
function useRetryAfterGate(
  error: unknown,
  errorUpdatedAt: number,
  isError: boolean,
): number | null {
  const info = useMemo(() => {
    if (!isError) return null;
    const ms = getRetryAfterMs(error, errorUpdatedAt);
    if (ms === undefined || ms <= 0) return null;
    return { deadline: errorUpdatedAt + ms, seconds: Math.ceil(ms / 1000) };
  }, [error, errorUpdatedAt, isError]);
  const [expiredDeadline, setExpiredDeadline] = useState<number | null>(null);
  useEffect(() => {
    if (!info) return;
    const timer = setTimeout(
      () => setExpiredDeadline(info.deadline),
      Math.max(0, info.deadline - Date.now()),
    );
    return () => clearTimeout(timer);
  }, [info]);
  return info !== null && expiredDeadline !== info.deadline ? info.seconds : null;
}

function formatReceivedAt(iso: string): string {
  const date = formatDate(iso);
  return date === EMPTY_VALUE ? EMPTY_VALUE : `${date} ${formatJSTTime(iso)}`;
}

function textOrEmpty(value: string | null | undefined): string {
  return value === null || value === undefined || value === "" ? EMPTY_VALUE : value;
}

function SectionFrame({ children, busy }: { children: ReactNode; busy?: boolean }) {
  return (
    <section
      aria-labelledby="revisit-intake-heading"
      aria-busy={busy ? true : undefined}
      className="space-y-2"
    >
      <div className="flex items-center gap-1.5">
        <MessageSquareText className={`${ICON.xs} ${C.text40}`} aria-hidden="true" />
        <span id="revisit-intake-heading" className={STYLE.sectionLabel}>
          {SECTION_LABEL}
        </span>
      </div>
      {children}
    </section>
  );
}

function StateBadge({ answered }: { answered: boolean }) {
  const Icon = answered ? CircleCheck : CircleDashed;
  return (
    <span
      className={`inline-flex items-center gap-1 rounded-full border px-2 py-0.5 text-2xs font-semibold ${answered ? BADGE.green : BADGE.gray}`}
    >
      <Icon className={ICON.xxs} aria-hidden="true" />
      {answered ? "回答済み" : "未回答"}
    </span>
  );
}

function ReadOnlyRow({ label, value }: { label: string; value: string }) {
  return (
    <div className="grid grid-cols-[6.5rem_1fr] gap-2 py-1.5">
      <dt className={`text-sm ${C.text60}`}>{label}</dt>
      {/* readonly: ink は落とさない（disabled と区別）。編集・コピー UI は置かない */}
      <dd className={`text-sm ${C.text} whitespace-pre-wrap break-words`}>{value}</dd>
    </div>
  );
}

function AnswerMeta({ answer }: { answer: RevisitIntakeAnswer }) {
  return (
    <p className={`text-xs ${C.text60}`}>
      回答版 {answer.revision} ・ 受信 {formatReceivedAt(answer.received_at)} ・ 出所{" "}
      {PROVENANCE_LABEL}
    </p>
  );
}

function AnswerBody({ answer }: { answer: RevisitIntakeAnswer }) {
  const a = answer.answers;
  return (
    <div className="space-y-1">
      <AnswerMeta answer={answer} />
      <dl className={`divide-y ${C.divideDividerFaint}`}>
        <ReadOnlyRow label="来院目的" value={textOrEmpty(a.visit_purpose)} />
        <ReadOnlyRow label="いつから" value={textOrEmpty(a.onset)} />
        <ReadOnlyRow
          label="元気・食欲"
          value={`元気: ${getEnergyAppetiteLabel(a.energy, EMPTY_VALUE)} ／ 食欲: ${getEnergyAppetiteLabel(a.appetite, EMPTY_VALUE)}`}
        />
        <ReadOnlyRow label="現在の症状" value={textOrEmpty(a.current_symptoms)} />
        <ReadOnlyRow
          label="前回より改善したか"
          value={getImprovementLabel(a.improvement, EMPTY_VALUE)}
        />
        <ReadOnlyRow
          label="検査希望"
          value={getExaminationPreferenceLabel(a.examination_preference, EMPTY_VALUE)}
        />
        <ReadOnlyRow label="一番相談したいこと" value={textOrEmpty(a.main_concern)} />
      </dl>
    </div>
  );
}

function CollapsibleGroup({
  summary,
  children,
  separated,
}: {
  summary: string;
  children: ReactNode;
  separated?: boolean;
}) {
  return (
    <details
      className={`rounded-lg border ${separated ? "border-dashed" : ""} ${C.borderLight} ${separated ? C.bgPage : C.bgSubtle}`}
    >
      <summary
        className={`flex min-h-11 cursor-pointer items-center px-3 text-sm ${C.text80} rounded-lg`}
      >
        {summary}
      </summary>
      <div className="space-y-3 px-3 pb-3">{children}</div>
    </details>
  );
}

function groupPrevious(entries: RevisitIntakeHistoryEntry[]) {
  const groups = new Map<
    string,
    { ref: RevisitIntakeTargetRef; entries: RevisitIntakeHistoryEntry[] }
  >();
  for (const entry of entries) {
    const key = `${entry.target.pet_id}|${entry.target.target_generation}`;
    const group = groups.get(key);
    if (group) group.entries.push(entry);
    else groups.set(key, { ref: entry.target, entries: [entry] });
  }
  return [...groups.values()];
}

function EmptyUnanswered() {
  return (
    <div className={`space-y-1 rounded-lg border ${C.borderLight} ${C.bgPage} p-3`}>
      <p className={`text-sm ${C.text80}`}>この予約の事前問診は届いていません。</p>
      {/* 職員の閲覧権限と飼主の LINE 回答資格は別。資格を示す契約 field が無いため、
          LINE で回答できるとは案内しない（手動予約・未連携の飼主を含む）。 */}
      <p className={`text-xs ${C.text60}`}>問診は任意提出です。</p>
      <p className={`text-xs ${C.text60}`}>
        次の操作: 受付時に聞き取り、補足は「予約内容」またはスタッフ側の記録へ。
      </p>
    </div>
  );
}

export function RevisitIntakeSection({
  appointmentId,
  clinicId,
  enabled = true,
}: RevisitIntakeSectionProps) {
  const query = useGetRevisitIntake({ appointmentId, clinicId, enabled });
  const queryClient = useQueryClient();
  const pages = query.data?.pages;
  const retryWaitSeconds = useRetryAfterGate(query.error, query.errorUpdatedAt, query.isError);
  // 401/403/404 は cached pages の有無に関係なく最優先（再取得・追加ページでセッション失効・権限喪失した場合も本文を出さない）。
  const errorStatus = query.isError ? getRevisitIntakeErrorStatus(query.error) : undefined;
  const denied = errorStatus === 401 || errorStatus === 403 || errorStatus === 404;

  // ページ間で対象（ref）または回答状態が変わったら後続ページを混ぜない。
  const targetChanged = useMemo(() => {
    if (!pages || pages.length < 2) return false;
    const key = (t: (typeof pages)[number]["current_target"]) =>
      t
        ? [
            t.ref.appointment_id,
            t.ref.pet_id,
            t.ref.target_generation,
            t.ref.target_id,
            t.head_revision,
            t.answer_state,
            t.latest_answer?.answer_id ?? null,
            t.content_available,
          ].join("|")
        : "";
    const first = key(pages[0].current_target);
    return pages.some((p) => key(p.current_target) !== first);
  }, [pages]);

  const { currentTarget, availability, currentHistory, previousHistory } = useMemo(() => {
    const first = pages?.[0];
    const history = (pages ?? []).flatMap((p) => p.history);
    const latestId = first?.current_target?.latest_answer?.answer_id;
    return {
      availability: first?.availability,
      currentTarget: first?.current_target ?? null,
      // 最新回答は本体に出すので、訂正履歴は旧 revision のみ。
      currentHistory: history.filter(
        (h) => h.target_relation === "current" && h.answer.answer_id !== latestId,
      ),
      previousHistory: history.filter((h) => h.target_relation === "previous"),
    };
  }, [pages]);

  if (query.isPending) {
    return (
      <SectionFrame busy>
        <div role="status" className="space-y-2">
          <span className="sr-only">問診を読み込み中</span>
          <DelayedSkeleton />
        </div>
      </SectionFrame>
    );
  }

  // 再取得（reconnect 等）が失敗したのに取得済み本文を「最新」として出し続けない。
  // 追加ページ取得の失敗だけは、取得済みページを残して下部の注記で扱う。
  const refetchFailed = query.isError && !!pages && !query.isFetchNextPageError;

  if (denied || (query.isError && !pages) || refetchFailed) {
    const status = errorStatus;
    if (status === 404) {
      // 対象予約の問診を読めない: セクションは出さず 1 行のみ（存在可否の詳細は出さない）。
      return <p className={`text-xs ${C.text60}`}>問診情報はありません。</p>;
    }
    if (status === 401) {
      // セッション失効: 本文・件数・再試行は出さず、再ログインを促す 1 行のみ。
      return (
        <SectionFrame>
          <p role="alert" className={`flex items-center gap-1.5 text-sm ${C.text60}`}>
            <Lock className={ICON.sm} aria-hidden="true" />
            セッションが切れました。再ログインしてください。
          </p>
        </SectionFrame>
      );
    }
    if (status === 403) {
      return (
        <SectionFrame>
          <p className={`flex items-center gap-1.5 text-sm ${C.text60}`}>
            <Lock className={ICON.sm} aria-hidden="true" />
            閲覧権限がありません。
          </p>
        </SectionFrame>
      );
    }
    return (
      <SectionFrame>
        <div className="flex items-center justify-between gap-2" role="alert">
          <p className={`text-sm ${C.textWarning}`}>問診を取得できませんでした。</p>
          <Button
            type="button"
            variant="outline"
            size="sm"
            className="min-h-11"
            disabled={query.isFetching || retryWaitSeconds !== null}
            onClick={() => void query.refetch()}
          >
            再試行
          </Button>
          {retryWaitSeconds !== null ? (
            <p className={`text-xs ${C.text60}`}>約{retryWaitSeconds}秒後に再試行できます</p>
          ) : null}
        </div>
      </SectionFrame>
    );
  }

  if (targetChanged) {
    return (
      <SectionFrame>
        <div className="flex items-center justify-between gap-2" role="alert">
          <p className={`text-sm ${C.textWarning}`}>
            対象または回答状態が変わりました。再読み込みしてください。
          </p>
          <Button
            type="button"
            variant="outline"
            size="sm"
            className="min-h-11"
            onClick={() =>
              void queryClient.resetQueries({
                queryKey: revisitIntakeQueryKey(appointmentId, clinicId),
              })
            }
          >
            再読み込み
          </Button>
        </div>
      </SectionFrame>
    );
  }

  const answered = currentTarget?.answer_state === "answered";
  const latest = currentTarget?.latest_answer ?? null;
  const previousGroups = groupPrevious(previousHistory);

  return (
    <SectionFrame>
      {availability === "target_unresolved" || currentTarget === null ? (
        <p className={`flex items-center gap-1.5 text-sm ${C.text60}`}>
          <ClipboardList className={ICON.sm} aria-hidden="true" />
          対象を特定できません。
        </p>
      ) : (
        <div className="space-y-2">
          <div className="flex items-center gap-2">
            <StateBadge answered={answered} />
          </div>
          {!answered ? <EmptyUnanswered /> : null}
          {answered && latest && currentTarget.content_available ? (
            <div className={`rounded-lg border ${C.borderLight} ${C.bgSubtle} p-3`}>
              <AnswerBody answer={latest} />
            </div>
          ) : null}
          {answered && (!latest || !currentTarget.content_available) ? (
            <p className={`flex items-center gap-1.5 text-sm ${C.text60}`}>
              <Lock className={ICON.sm} aria-hidden="true" />
              回答済み（本文は閲覧できません）
            </p>
          ) : null}
        </div>
      )}

      {currentHistory.length > 0 ? (
        <CollapsibleGroup
          summary={`訂正履歴${query.hasNextPage ? "" : `（${currentHistory.length}件）`}`}
        >
          {currentHistory.map((h) => (
            <div
              key={h.answer.answer_id}
              className={`border-t pt-2 first:border-t-0 first:pt-0 ${C.borderDivider}`}
            >
              <AnswerBody answer={h.answer} />
            </div>
          ))}
        </CollapsibleGroup>
      ) : null}

      {previousGroups.length > 0 ? (
        <CollapsibleGroup
          separated
          summary={`以前の対象ペットの回答${query.hasNextPage ? "" : `（${previousHistory.length}件）`}`}
        >
          <p className={`text-xs ${C.text60}`}>現在の予約対象ではありません。参照のみです。</p>
          {previousGroups.map((g) => (
            <div key={`${g.ref.pet_id}|${g.ref.target_generation}`} className="space-y-2">
              <p className={`text-xs font-semibold ${C.text70}`}>
                ペットID {g.ref.pet_id} ・ 世代 {g.ref.target_generation}
              </p>
              {g.entries.map((h) => (
                <div
                  key={h.answer.answer_id}
                  className={`border-t pt-2 first:border-t-0 first:pt-0 ${C.borderDivider}`}
                >
                  <AnswerBody answer={h.answer} />
                </div>
              ))}
            </div>
          ))}
        </CollapsibleGroup>
      ) : null}

      {query.hasNextPage ? (
        <div className="space-y-1">
          {query.isFetchNextPageError ? (
            <p role="alert" className={`text-xs ${C.textWarning}`}>
              続きを取得できませんでした。
            </p>
          ) : null}
          <Button
            type="button"
            variant="outline"
            size="sm"
            className="min-h-11"
            disabled={query.isFetchingNextPage || retryWaitSeconds !== null}
            aria-busy={query.isFetchingNextPage ? true : undefined}
            onClick={() => void query.fetchNextPage()}
          >
            {query.isFetchingNextPage ? "読み込み中…" : "さらに表示"}
          </Button>
          {retryWaitSeconds !== null ? (
            <p className={`text-xs ${C.text60}`}>約{retryWaitSeconds}秒後に再試行できます</p>
          ) : null}
        </div>
      ) : null}
    </SectionFrame>
  );
}
