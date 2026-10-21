// EMR-474 operation-recovery Slice A: 再読込後に発見した本人の操作の表示と保留。
// - succeeded: 既存 lookup response が持つ範囲（完了と時刻）だけを出す。target / 回答 / 版 / 本文は足さない。
// - unresolved / 結果を読めない: 新しい送信を止め、既存の結果照会（同じ operation key）で再確認できる。
// - closed_without_effect: 参考表示のみ。
// 次操作の解除（fence）・新 key での再開はこの Slice に無い。ここは「止める」と「確かめる」だけ。
import { useEffect, useState } from "react";
import { liffApi } from "../../api/liff-api";
import { resolveFetchError } from "@/shared-liff/handle-fetch-error";
import type { RevisitIntakeRecoveryItem } from "../../api/revisit-intake-recovery-schemas";
import type { RevisitIntakeRecoveryErrorKind } from "../../api/revisit-intake-recovery";
import type { UseRevisitIntakeRecoveryResult } from "../../hooks/use-revisit-intake-recovery";
import { formatRevisitIntakeDateTime } from "../../lib/revisit-intake-options";
import { revisitIntakeRetryAfterSeconds } from "../../lib/revisit-intake-submit";
import { PENDING_UNAUTHORIZED_TEXT } from "./RevisitIntakeDeniedPendingLookup";

const BLOCKED_NOTICE =
  "未確定の操作があるため、新しい回答の送信を一時停止しています。結果を確認できるまでお待ちください。";

// self scope（予約を特定できない発見）は本人の全予約の送信が対象。この予約の結果とは言い切らない。
// Slice A: 回復の確認機能（fence）が無いため、未確定が見つからなくても新しい送信は再開しない。
const FENCE_UNAVAILABLE_NOTICE =
  "新しい回答の送信は、操作の回復確認機能が利用できるようになるまで一時停止しています。これまでの送信結果の確認はできます。";

const BLOCKED_NOTICE_SELF =
  "未確定の送信があるため（予約を特定できていません）、新しい回答の送信を一時停止しています。結果を確認できるまでお待ちください。";

const UNAUTHENTICATED_TEXT =
  "ログインの有効期限が切れました。時間をおいてもう一度確認してください。";

const ERROR_TEXT: Record<RevisitIntakeRecoveryErrorKind, string> = {
  validation: "未確定の操作を確認できませんでした。",
  unauthenticated: UNAUTHENTICATED_TEXT,
  not_authorized: "未確定の操作を確認できませんでした。",
  scope_not_found: "未確定の操作を確認できませんでした。",
  recovery_conflict: "未確定の操作の状態が変わりました。もう一度確認してください。",
  result_unavailable: "未確定の操作の結果を確認できませんでした。時間をおいてもう一度確認してください。",
  unclassified: "未確定の操作を確認できませんでした。",
  contract_violation: "未確定の操作を確認できませんでした。",
  network: "通信できませんでした。電波の良い場所でもう一度確認してください。",
  aborted: "未確定の操作を確認できませんでした。",
};

const MISMATCH_TEXT = "送信結果を確認できませんでした。もう一度お試しください。";

type ItemLookup = { busy: true } | { busy: false; text: string };

function isHold(item: RevisitIntakeRecoveryItem): boolean {
  return item.discovery_state === "unavailable" || item.execution_state === "unresolved";
}

function SucceededCard({ item }: { item: RevisitIntakeRecoveryItem }) {
  const op = item.result?.operation_result;
  if (op === undefined || op.state !== "succeeded") return null;
  const redacted = item.result?.content_available === false;
  return (
    <div className="rounded-lg border border-noah-border bg-white px-4 py-3" role="status">
      <p className="text-sm text-noah-text">回答の送信は完了しています。</p>
      <p className="mt-1 text-xs text-noah-text-sub">
        送信日時 {formatRevisitIntakeDateTime(op.committed_at)}
      </p>
      {redacted ? (
        <p className="mt-1 text-xs text-noah-text-sub">回答内容はこの画面では表示できません。</p>
      ) : null}
    </div>
  );
}

export function RevisitIntakeRecoveryPanel({
  recovery,
  clinicId,
  idToken,
  fenceNoticeRelevant = false,
}: {
  recovery: UseRevisitIntakeRecoveryResult;
  clinicId: string;
  idToken: string;
  /** 新規送信の入口（editable な target）があるときだけ、fence 待ちの停止告知を出す。 */
  fenceNoticeRelevant?: boolean;
}) {
  const [lookups, setLookups] = useState<Record<string, ItemLookup>>({});
  const { refresh } = recovery;
  // 429 / Retry-After 付き 500・503 の期限（memory のみ）。期限までその item の照会 button を止める。
  const [retryUntil, setRetryUntil] = useState<Record<string, number>>({});
  const [now, setNow] = useState(() => Date.now());
  useEffect(() => {
    const future = Object.values(retryUntil).filter((t) => t > now);
    if (future.length === 0) return;
    const timer = setTimeout(() => setNow(Date.now()), Math.max(0, Math.min(...future) - Date.now()));
    return () => clearTimeout(timer);
  }, [retryUntil, now]);

  async function relookup(key: string) {
    setLookups((prev) => ({ ...prev, [key]: { busy: true } }));
    let text: string;
    let refreshAfter = true;
    try {
      // 既存の結果照会（同じ operation key）。新しい key は作らない。
      const result = await liffApi.getRevisitIntakeOperation(clinicId, key, idToken);
      const op = result.operation_result;
      if (op.operation_id !== key) {
        text = MISMATCH_TEXT;
        refreshAfter = false;
      } else if (op.state === "succeeded") {
        text = "送信は完了していました。最新の状態を確認しています。";
      } else if (op.state === "not_found") {
        text = "送信の記録は見つかりませんでした。結果が確定するまで新しい回答は送信できません。";
      } else if (op.state === "claimed_uncommitted") {
        text = "保存はまだ確認できていません。時間をおいてもう一度確認してください。";
      } else {
        text = "結果を確認できません。時間をおいてもう一度確認してください。";
      }
    } catch (err: unknown) {
      const info = resolveFetchError(err, "送信結果の確認");
      text = info.status === 401 ? PENDING_UNAUTHORIZED_TEXT : info.message;
      refreshAfter = false;
      const retryAfter = revisitIntakeRetryAfterSeconds(err);
      if (retryAfter !== null) {
        const until = Date.now() + retryAfter * 1000;
        text = `${retryAfter}秒後に再確認できます`;
        setRetryUntil((prev) => ({ ...prev, [key]: until }));
        setNow(Date.now());
      }
    }
    setLookups((prev) => ({ ...prev, [key]: { busy: false, text } }));
    // 表示と保留の更新は発見をやり直して得た結果だけに従う（照会の結果から保留を外さない）。
    if (refreshAfter) refresh();
  }

  if (recovery.status === "error") {
    return (
      <div
        role="alert"
        className="space-y-2 rounded-lg border border-noah-warning-border bg-noah-warning-box-bg px-4 py-3"
      >
        <p className="text-sm text-noah-warning-text">
          {recovery.errorKind !== null ? ERROR_TEXT[recovery.errorKind] : ERROR_TEXT.unclassified}
        </p>
        <p className="text-sm text-noah-warning-text">
          確認できるまで、新しい回答は送信できません。
        </p>
        <button
          type="button"
          onClick={refresh}
          className="min-h-11 w-full rounded-lg border border-noah-border-input bg-white px-4 text-sm font-semibold text-noah-teal-dark"
        >
          もう一度確認する
        </button>
      </div>
    );
  }

  const { discovery, summary } = recovery;
  if (discovery === null || summary === null) return null;

  const isSelf = discovery.scope === "self";
  const holdItems = discovery.items.filter(isHold);
  // self scope の成功・終了は別の予約のものを含み得るため、この予約の状態として表示しない。
  const succeeded = isSelf
    ? []
    : discovery.items.filter(
        (i) => i.discovery_state === "result" && i.execution_state === "succeeded",
      );
  const closed = isSelf
    ? []
    : discovery.items.filter((i) => i.execution_state === "closed_without_effect");
  const latestSucceeded = succeeded[0];
  const showBlocked = summary.hasOperationHold || (summary.newSubmissionBlocked && fenceNoticeRelevant);
  if (!showBlocked && latestSucceeded === undefined && closed.length === 0) {
    return null;
  }

  return (
    <section aria-label="未確定の操作の確認" className="space-y-3">
      {showBlocked ? (
        <div
          role="alert"
          className="space-y-2 rounded-lg border border-noah-warning-border bg-noah-warning-box-bg px-4 py-3"
        >
          <p className="text-sm text-noah-warning-text">
            {!summary.hasOperationHold
              ? FENCE_UNAVAILABLE_NOTICE
              : isSelf
                ? BLOCKED_NOTICE_SELF
                : BLOCKED_NOTICE}
          </p>
          {summary.reasons.includes("scan_incomplete") ? (
            <p className="text-sm text-noah-warning-text">
              操作の履歴をすべて確認できていません。もう一度確認してください。
            </p>
          ) : null}
          {summary.reasons.includes("active_operation") && holdItems.length === 0 ? (
            <p className="text-sm text-noah-warning-text">
              直前の操作の確認が済むまで、この画面からは新しい回答を送信できません。
            </p>
          ) : null}
          {holdItems.length === 0 && summary.hasOperationHold ? (
            <button
              type="button"
              onClick={refresh}
              disabled={recovery.status === "checking"}
              className="min-h-11 w-full rounded-lg border border-noah-border-input bg-white px-4 text-sm font-semibold text-noah-teal-dark disabled:opacity-50"
            >
              もう一度確認する
            </button>
          ) : null}
        </div>
      ) : null}

      {holdItems.map((item) => {
        const state = lookups[item.operation_key];
        const busy = state?.busy === true;
        const waiting = (retryUntil[item.operation_key] ?? 0) > now;
        return (
          <div
            key={item.operation_key}
            className="space-y-2 rounded-lg border border-noah-warning-border bg-white px-4 py-3"
          >
            <p className="text-sm text-noah-text">
              {item.discovery_state === "unavailable"
                ? "送信結果を読み取れていません。"
                : "送信結果が未確定です。"}
            </p>
            <p className="text-xs text-noah-text-sub">
              送信日時 {formatRevisitIntakeDateTime(item.claimed_at)}
            </p>
            {state !== undefined && !state.busy ? (
              <p role="status" className="text-sm text-noah-text">
                {state.text}
              </p>
            ) : null}
            <button
              type="button"
              onClick={() => void relookup(item.operation_key)}
              disabled={busy || waiting}
              className="min-h-11 w-full rounded-lg border border-noah-border-input bg-white px-4 text-sm font-semibold text-noah-teal-dark disabled:opacity-50"
            >
              {busy ? "確認中..." : "送信結果を確認"}
            </button>
          </div>
        );
      })}

      {latestSucceeded !== undefined ? <SucceededCard item={latestSucceeded} /> : null}

      {closed.length > 0 ? (
        <p className="text-xs text-noah-text-sub">
          反映されずに終了した送信が{closed.length}件あります（参考）。
        </p>
      ) : null}
    </section>
  );
}
